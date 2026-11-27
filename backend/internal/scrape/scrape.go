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

	q := store.New(tx)
	userIDs, err := upsertUsers(ctx, q, results)
	if err != nil {
		return fmt.Errorf("upserting users for %s: %w", date.Format(time.DateOnly), err)
	}
	if err := insertTimes(ctx, q, results, userIDs, date); err != nil {
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
// by NYT user ID. results must be unique and sorted by NYT user ID.
func upsertUsers(ctx context.Context, q *store.Queries, results []Result) (map[int64]int64, error) {
	params := store.UpsertUsersParams{
		NytUserIds: make([]int64, len(results)),
		Names:      make([]string, len(results)),
	}
	for i, r := range results {
		params.NytUserIds[i] = r.NYTUserID
		params.Names[i] = r.Name
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

// insertTimes records a solve time for every user who has one.
func insertTimes(ctx context.Context, q *store.Queries, results []Result, userIDs map[int64]int64, date time.Time) error {
	params := store.InsertTimesParams{Date: StartOfDay(date)}
	for _, r := range results {
		if r.SolveTimeSeconds == nil {
			continue
		}
		id, ok := userIDs[r.NYTUserID]
		if !ok {
			return fmt.Errorf("no database ID for NYT user %d", r.NYTUserID)
		}
		params.UserIds = append(params.UserIds, id)
		params.TimesInSeconds = append(params.TimesInSeconds, *r.SolveTimeSeconds)
	}
	if len(params.UserIds) == 0 {
		return nil
	}
	return q.InsertTimes(ctx, params)
}
