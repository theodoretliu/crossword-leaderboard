package scrape

import (
	"context"
	"time"
)

// Result is one user's outcome for a single Mini puzzle date.
type Result struct {
	NYTUserID int64
	Name      string
	// SolveTimeSeconds is nil when the user has not solved that day's puzzle.
	SolveTimeSeconds *int32
}

// Fetcher retrieves the leaderboard for a single Mini puzzle date.
//
// It is the seam between "how we get the data from NYT" and "what we do with
// it": swap the implementation passed to Scraper to change data sources.
type Fetcher interface {
	Fetch(ctx context.Context, date time.Time) ([]Result, error)
}
