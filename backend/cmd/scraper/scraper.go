// Command scraper keeps every puzzle date's leaderboard up to date, re-scraping
// recent dates often and old dates rarely (see scrape.Interval).
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/theodoretliu/crossword-leaderboard/internal/scrape"
)

func main() {
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:              os.Getenv("DSN"),
		AttachStacktrace: true,
	}); err != nil {
		log.Fatal(err)
	}
	defer sentry.Flush(2 * time.Second)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := pgxpool.New(ctx, os.Getenv("DB_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	s := &scrape.Scraper{
		Pool:    pool,
		Fetcher: &scrape.GraphQLFetcher{Cookie: scrape.DBCookie(pool)},
	}

	var wg sync.WaitGroup
	watch := func(date time.Time) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			watchDate(ctx, s, date)
		}()
	}

	today := scrape.StartOfDay(time.Now())
	for date := scrape.FirstDate; !date.After(today); date = date.AddDate(0, 0, 1) {
		watch(date)
	}

	// Start watching each new puzzle date as its UTC day begins.
	for {
		tomorrow := today.AddDate(0, 0, 1)
		if !sleep(ctx, time.Until(tomorrow)) {
			break
		}
		today = tomorrow
		log.Printf("watching new date %s", today.Format(time.DateOnly))
		watch(today)
	}

	wg.Wait()
	log.Print("scraper stopped")
}

// watchDate scrapes date until ctx is cancelled, waiting scrape.Interval before
// each attempt. Waiting first (rather than scraping immediately) spreads the
// thousands of dates out after a restart instead of firing them all at once.
func watchDate(ctx context.Context, s *scrape.Scraper, date time.Time) {
	for sleep(ctx, scrape.Interval(time.Since(date))) {
		scrapeOnce(ctx, s, date)
	}
}

// scrapeOnce scrapes date and reports any error or panic to Sentry, so that one
// bad date cannot take down the watchers for every other date.
func scrapeOnce(ctx context.Context, s *scrape.Scraper, date time.Time) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("panic scraping %s: %v", date.Format(time.DateOnly), r)
			sentry.CurrentHub().Recover(r)
		}
	}()

	if err := s.ScrapeDate(ctx, date); err != nil && ctx.Err() == nil {
		log.Print(err)
		sentry.CaptureException(err)
	}
}

// sleep waits for d or until ctx is done, reporting whether the full wait elapsed.
func sleep(ctx context.Context, d time.Duration) bool {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-t.C:
		return true
	}
}
