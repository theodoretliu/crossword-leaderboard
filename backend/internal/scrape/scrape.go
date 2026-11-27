package scrape

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/theodoretliu/crossword-leaderboard/internal/store"
)

// Scraper fetches a date's leaderboard and writes it straight to the database.
// It suits one-off runs such as scrapeall; the long-running scraper uses a
// Batcher instead so the database can sleep between writes.
type Scraper struct {
	Pool    *pgxpool.Pool
	Fetcher Fetcher
}

// ScrapeDate fetches the leaderboard for date and upserts its users and times.
func (s *Scraper) ScrapeDate(ctx context.Context, date time.Time) error {
	results, err := s.Fetcher.Fetch(ctx, date)
	if err != nil {
		return fmt.Errorf("fetching %s: %w", date.Format(time.DateOnly), err)
	}

	users := uniqueByUser(results)
	var solves []solve
	for _, r := range users {
		if r.SolveTimeSeconds != nil {
			solves = append(solves, solve{NYTUserID: r.NYTUserID, Date: StartOfDay(date), Seconds: *r.SolveTimeSeconds})
		}
	}
	if err := save(ctx, s.Pool, users, solves); err != nil {
		return fmt.Errorf("saving %s: %w", date.Format(time.DateOnly), err)
	}
	return nil
}

// solve is one user's time for one puzzle date.
type solve struct {
	NYTUserID int64
	Date      time.Time
	Seconds   int32
}

// save upserts users and inserts solves in a single transaction. users must be
// unique and sorted by NYT user ID (see uniqueByUser) and must include every
// user that appears in solves.
func save(ctx context.Context, pool *pgxpool.Pool, users []Result, solves []solve) error {
	if len(users) == 0 {
		return nil
	}

	// Network I/O happens before this point, so the transaction stays short.
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	q := store.New(tx)

	userIDs, err := upsertUsers(ctx, q, users)
	if err != nil {
		return fmt.Errorf("upserting users: %w", err)
	}
	if err := insertTimes(ctx, q, solves, userIDs); err != nil {
		return fmt.Errorf("inserting times: %w", err)
	}
	return tx.Commit(ctx)
}

// uniqueByUser sorts results by NYT user ID and drops duplicates. A consistent
// order keeps concurrent upserts from deadlocking, and a user appearing twice
// would make the upsert fail ("ON CONFLICT DO UPDATE command cannot affect row a
// second time").
func uniqueByUser(results []Result) []Result {
	results = slices.Clone(results)
	slices.SortStableFunc(results, func(a, b Result) int {
		return cmp.Compare(a.NYTUserID, b.NYTUserID)
	})
	return slices.CompactFunc(results, func(a, b Result) bool {
		return a.NYTUserID == b.NYTUserID
	})
}

// upsertUsers inserts or renames each user and returns their database IDs keyed
// by NYT user ID.
func upsertUsers(ctx context.Context, q *store.Queries, users []Result) (map[int64]int64, error) {
	params := store.UpsertUsersParams{
		NytUserIds: make([]int64, len(users)),
		Names:      make([]string, len(users)),
	}
	for i, u := range users {
		params.NytUserIds[i] = u.NYTUserID
		params.Names[i] = u.Name
	}

	rows, err := q.UpsertUsers(ctx, params)
	if err != nil {
		return nil, err
	}
	ids := make(map[int64]int64, len(rows))
	for _, row := range rows {
		if row.NytUserID != nil {
			ids[*row.NytUserID] = int64(row.ID)
		}
	}
	return ids, nil
}

// insertTimes records solves, keeping any time already stored.
func insertTimes(ctx context.Context, q *store.Queries, solves []solve, userIDs map[int64]int64) error {
	if len(solves) == 0 {
		return nil
	}
	params := store.InsertTimesParams{
		UserIds:        make([]int64, len(solves)),
		TimesInSeconds: make([]int32, len(solves)),
		Dates:          make([]time.Time, len(solves)),
	}
	for i, s := range solves {
		id, ok := userIDs[s.NYTUserID]
		if !ok {
			return fmt.Errorf("no database ID for NYT user %d", s.NYTUserID)
		}
		params.UserIds[i] = id
		params.TimesInSeconds[i] = s.Seconds
		params.Dates[i] = s.Date
	}
	return q.InsertTimes(ctx, params)
}
