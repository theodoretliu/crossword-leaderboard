package scrape

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Scraper fetches a date's leaderboard and stores it.
type Scraper struct {
	Pool    *pgxpool.Pool
	Fetcher Fetcher
}

// ScrapeDate fetches the leaderboard for date and upserts its users and times.
func (s *Scraper) ScrapeDate(ctx context.Context, date time.Time) error {
	// Fetch before opening a transaction so a slow NYT request does not hold a
	// pooled connection.
	results, err := s.Fetcher.Fetch(ctx, date)
	if err != nil {
		return fmt.Errorf("fetching %s: %w", date.Format(time.DateOnly), err)
	}

	results = uniqueByUser(results)
	if len(results) == 0 {
		return nil
	}

	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	userIDs, err := upsertUsers(ctx, tx, results)
	if err != nil {
		return fmt.Errorf("upserting users for %s: %w", date.Format(time.DateOnly), err)
	}
	if err := insertTimes(ctx, tx, results, userIDs, date); err != nil {
		return fmt.Errorf("inserting times for %s: %w", date.Format(time.DateOnly), err)
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
func upsertUsers(ctx context.Context, tx pgx.Tx, results []Result) (map[int64]int64, error) {
	args := make([]any, 0, 2*len(results))
	for _, r := range results {
		args = append(args, r.NYTUserID, r.Name)
	}

	// An empty name from the API never overwrites a name we already have.
	query := "INSERT INTO users (nyt_user_id, name) VALUES " + placeholders(len(results), 2) +
		" ON CONFLICT (nyt_user_id) DO UPDATE SET name = COALESCE(NULLIF(EXCLUDED.name, ''), users.name)" +
		" RETURNING nyt_user_id, id"
	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	ids := make(map[int64]int64, len(results))
	var nytID, id int64
	_, err = pgx.ForEachRow(rows, []any{&nytID, &id}, func() error {
		ids[nytID] = id
		return nil
	})
	return ids, err
}

// insertTimes records a solve time for every user who has one.
func insertTimes(ctx context.Context, tx pgx.Tx, results []Result, userIDs map[int64]int64, date time.Time) error {
	day := date.UTC().Format(time.DateOnly)
	var args []any
	for _, r := range results {
		if r.SolveTimeSeconds == nil {
			continue
		}
		id, ok := userIDs[r.NYTUserID]
		if !ok {
			return fmt.Errorf("no database ID for NYT user %d", r.NYTUserID)
		}
		args = append(args, id, *r.SolveTimeSeconds, day)
	}
	if len(args) == 0 {
		return nil
	}

	query := "INSERT INTO times (user_id, time_in_seconds, date) VALUES " +
		placeholders(len(args)/3, 3) + " ON CONFLICT DO NOTHING"
	_, err := tx.Exec(ctx, query, args...)
	return err
}

// placeholders returns "($1, $2), ($3, $4), ..." for rows tuples of cols values.
func placeholders(rows, cols int) string {
	var b strings.Builder
	n := 1
	for row := 0; row < rows; row++ {
		if row > 0 {
			b.WriteString(", ")
		}
		b.WriteByte('(')
		for col := 0; col < cols; col++ {
			if col > 0 {
				b.WriteString(", ")
			}
			fmt.Fprintf(&b, "$%d", n)
			n++
		}
		b.WriteByte(')')
	}
	return b.String()
}
