// Command scrapeall scrapes every puzzle date once, from scrape.FirstDate (or
// the YYYY-MM-DD date given as the only argument) through today.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/theodoretliu/crossword-leaderboard/internal/scrape"
)

// concurrency bounds simultaneous requests to NYT and open transactions.
const concurrency = 8

func main() {
	start := scrape.FirstDate
	if len(os.Args) == 2 {
		var err error
		if start, err = time.Parse(time.DateOnly, os.Args[1]); err != nil {
			log.Fatalf("parsing start date: %v", err)
		}
	}

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

	var (
		wg     sync.WaitGroup
		failed atomic.Int64
		slots  = make(chan struct{}, concurrency)
	)
	today := scrape.StartOfDay(time.Now())

dates:
	for date := start; !date.After(today); date = date.AddDate(0, 0, 1) {
		select {
		case slots <- struct{}{}:
		case <-ctx.Done():
			break dates
		}

		wg.Add(1)
		go func(date time.Time) {
			defer func() { <-slots; wg.Done() }()
			if err := s.ScrapeDate(ctx, date); err != nil {
				log.Print(err)
				failed.Add(1)
			}
		}(date)
	}
	wg.Wait()

	if n := failed.Load(); n > 0 {
		log.Printf("%d dates failed", n)
		stop()
		os.Exit(1)
	}
}
