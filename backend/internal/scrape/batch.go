package scrape

import (
	"cmp"
	"context"
	"fmt"
	"maps"
	"slices"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/theodoretliu/crossword-leaderboard/internal/store"
)

// Batcher keeps an in-memory copy of what the database already holds, records
// only what scrapes reveal to be new (a first solve time for a user and date, a
// new user, or a renamed one), and writes those changes in periodic Flushes.
//
// This exists because Neon bills for every minute the database is awake and
// suspends it after 5 idle minutes: re-scraping NYT every few seconds is free,
// but writing (or even reading) the database that often would keep it awake
// permanently. Nearly every scrape changes nothing, so it never touches the
// database at all.
type Batcher struct {
	save func(ctx context.Context, users []Result, solves []solve) error

	mu sync.Mutex
	// names and stored mirror the database.
	names  map[int64]string
	stored map[solveKey]struct{}
	// pendingUsers and pendingSolves are changes not yet written.
	pendingUsers  map[int64]string
	pendingSolves map[solveKey]int32
}

type solveKey struct {
	NYTUserID int64
	Day       int64 // days since the Unix epoch, UTC
}

func keyFor(nytUserID int64, date time.Time) solveKey {
	return solveKey{NYTUserID: nytUserID, Day: StartOfDay(date).Unix() / 86400}
}

func (k solveKey) date() time.Time { return time.Unix(k.Day*86400, 0).UTC() }

// NewBatcher loads the stored users and solves (one burst of database work at
// startup) and returns a Batcher that writes to pool.
func NewBatcher(ctx context.Context, pool *pgxpool.Pool) (*Batcher, error) {
	q := store.New(pool)
	users, err := q.ListKnownUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading users: %w", err)
	}
	solves, err := q.ListStoredSolves(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading solves: %w", err)
	}

	b := newBatcher(func(ctx context.Context, users []Result, solves []solve) error {
		return save(ctx, pool, users, solves)
	})
	for _, u := range users {
		b.names[u.NytUserID] = u.Name
	}
	for _, s := range solves {
		b.stored[keyFor(s.NytUserID, s.Date)] = struct{}{}
	}
	return b, nil
}

func newBatcher(save func(context.Context, []Result, []solve) error) *Batcher {
	return &Batcher{
		save:          save,
		names:         map[int64]string{},
		stored:        map[solveKey]struct{}{},
		pendingUsers:  map[int64]string{},
		pendingSolves: map[solveKey]int32{},
	}
}

// Observe records whatever in results is not already stored or pending.
func (b *Batcher) Observe(date time.Time, results []Result) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, r := range results {
		name, known := b.pendingUsers[r.NYTUserID]
		if !known {
			name, known = b.names[r.NYTUserID]
		}
		if !known || (r.Name != "" && r.Name != name) {
			b.pendingUsers[r.NYTUserID] = r.Name
		}

		if r.SolveTimeSeconds == nil {
			continue
		}
		k := keyFor(r.NYTUserID, date)
		if _, ok := b.stored[k]; ok {
			continue
		}
		if _, ok := b.pendingSolves[k]; !ok {
			b.pendingSolves[k] = *r.SolveTimeSeconds
		}
	}
}

// Pending reports how many user and solve changes are waiting to be written.
func (b *Batcher) Pending() (users, solves int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.pendingUsers), len(b.pendingSolves)
}

// Flush writes pending changes in one transaction. With nothing pending it
// does not touch the database. On failure the changes stay pending for the
// next Flush.
func (b *Batcher) Flush(ctx context.Context) error {
	b.mu.Lock()
	pendingUsers, pendingSolves := b.pendingUsers, b.pendingSolves
	b.pendingUsers, b.pendingSolves = map[int64]string{}, map[solveKey]int32{}

	// Every user with a solve must be in the upsert to get a database ID.
	users := make(map[int64]string, len(pendingUsers))
	maps.Copy(users, pendingUsers)
	for k := range pendingSolves {
		if _, ok := users[k.NYTUserID]; !ok {
			users[k.NYTUserID] = b.names[k.NYTUserID]
		}
	}
	b.mu.Unlock()

	if len(users) == 0 {
		return nil
	}

	userList := make([]Result, 0, len(users))
	for id, name := range users {
		userList = append(userList, Result{NYTUserID: id, Name: name})
	}
	slices.SortFunc(userList, func(a, b Result) int { return cmp.Compare(a.NYTUserID, b.NYTUserID) })

	solves := make([]solve, 0, len(pendingSolves))
	for k, secs := range pendingSolves {
		solves = append(solves, solve{NYTUserID: k.NYTUserID, Date: k.date(), Seconds: secs})
	}

	err := b.save(ctx, userList, solves)

	b.mu.Lock()
	defer b.mu.Unlock()
	if err != nil {
		// Put the changes back, without clobbering anything observed meanwhile.
		for id, name := range pendingUsers {
			if _, ok := b.pendingUsers[id]; !ok {
				b.pendingUsers[id] = name
			}
		}
		for k, secs := range pendingSolves {
			if _, ok := b.pendingSolves[k]; !ok {
				b.pendingSolves[k] = secs
			}
		}
		return err
	}
	// Mirror the upsert: an empty name never replaces a stored one.
	for id, name := range pendingUsers {
		if _, ok := b.names[id]; !ok || name != "" {
			b.names[id] = name
		}
	}
	for k := range pendingSolves {
		b.stored[k] = struct{}{}
	}

	// Scrapes that ran while the save was in flight may have re-queued what was
	// just written; drop those so the next Flush does not wake the database for
	// nothing.
	for k := range b.pendingSolves {
		if _, ok := b.stored[k]; ok {
			delete(b.pendingSolves, k)
		}
	}
	for id, name := range b.pendingUsers {
		if stored, ok := b.names[id]; ok && (name == "" || name == stored) {
			delete(b.pendingUsers, id)
		}
	}
	return nil
}
