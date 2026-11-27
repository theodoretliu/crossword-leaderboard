// Command scraper keeps every puzzle date's leaderboard up to date, re-scraping
// recent dates often and old dates rarely (see scrape.Interval).
//
// The database is Neon's free tier, which bills for the minutes the database is
// awake and suspends it after 5 idle minutes. So scrapes only update an
// in-memory scrape.Batcher, changes are written at most once per flushInterval,
// and the cookie is cached; between flushes the database is left to sleep.
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

// flushInterval bounds both how often the database is woken and how stale the
// leaderboard can be.
const flushInterval = 30 * time.Minute

type scraper struct {
	fetcher *scrape.GraphQLFetcher
	cookies *scrape.CookieCache
	batcher *scrape.Batcher
}

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

	// Load everything needed from the database up front, in one burst.
	cookies := &scrape.CookieCache{Load: scrape.DBCookie(pool), MinReload: flushInterval}
	if _, err := cookies.Get(ctx); err != nil {
		log.Fatal(err)
	}
	batcher, err := scrape.NewBatcher(ctx, pool)
	if err != nil {
		log.Fatal(err)
	}
	s := &scraper{
		fetcher: &scrape.GraphQLFetcher{Cookie: cookies.Get},
		cookies: cookies,
		batcher: batcher,
	}

	var wg sync.WaitGroup
	watch := func(date time.Time) {
		wg.Go(func() {
			s.watchDate(ctx, date)
		})
	}

	today := scrape.StartOfDay(time.Now())
	for date := scrape.FirstDate; !date.After(today); date = date.AddDate(0, 0, 1) {
		watch(date)
	}

	wg.Go(func() {
		s.flushLoop(ctx)
	})

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
func (s *scraper) watchDate(ctx context.Context, date time.Time) {
	for sleep(ctx, scrape.Interval(time.Since(date))) {
		s.scrapeOnce(ctx, date)
	}
}

// scrapeOnce fetches date into the batcher and reports any error or panic to
// Sentry, so that one bad date cannot take down the watchers for every other
// date.
func (s *scraper) scrapeOnce(ctx context.Context, date time.Time) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("panic scraping %s: %v", date.Format(time.DateOnly), r)
			sentry.CurrentHub().Recover(r)
		}
	}()

	results, err := s.fetcher.Fetch(ctx, date)
	if err != nil {
		if ctx.Err() == nil {
			// The cookie may have expired; allow a (rate-limited) reload.
			s.cookies.Invalidate()
			log.Printf("fetching %s: %v", date.Format(time.DateOnly), err)
			sentry.CaptureException(err)
		}
		return
	}
	s.batcher.Observe(date, results)
}

// flushLoop writes pending changes every flushInterval, and once more on
// shutdown so nothing observed is lost.
func (s *scraper) flushLoop(ctx context.Context) {
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.flush(ctx)
		case <-ctx.Done():
			final, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			s.flush(final)
			cancel()
			return
		}
	}
}

func (s *scraper) flush(ctx context.Context) {
	users, solves := s.batcher.Pending()
	if users == 0 && solves == 0 {
		return
	}
	if err := s.batcher.Flush(ctx); err != nil {
		log.Printf("flushing %d users, %d solves: %v", users, solves, err)
		sentry.CaptureException(err)
		return
	}
	log.Printf("flushed %d users, %d solves", users, solves)
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
