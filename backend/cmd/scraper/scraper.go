package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/theodoretliu/crossword-leaderboard/internal/scrape"
)

func startScrapers() {
	minDate := time.Date(scrape.MinYear, scrape.MinMonth, scrape.MinDay, 0, 0, 0, 0, time.UTC)

	for {
		// Sleep until next 10-minute mark
		now := time.Now().UTC()
		nextMark := now.Truncate(10 * time.Minute).Add(10 * time.Minute)
		time.Sleep(time.Until(nextMark))

		now = time.Now().UTC()
		today := now.Truncate(24 * time.Hour)

		// Check each date from minDate to today
		for date := minDate; !date.After(today); date = date.AddDate(0, 0, 1) {
			daysAgo := int(today.Sub(date).Hours() / 24)

			// Quadratic probability falloff: 1/(days+1)²
			probability := 1.0 / float64((daysAgo+1)*(daysAgo+1))

			if rand.Float64() < probability {
				go func(d time.Time) {
					err := scrape.DbActionsForDate(pool, d)
					if err != nil {
						fmt.Println(err)
						sentry.CaptureException(err)
					}
				}(date)
			}
		}
	}
}

var pool *pgxpool.Pool

func main() {
	err := sentry.Init(sentry.ClientOptions{
		Dsn:              os.Getenv("DSN"),
		AttachStacktrace: true,
	})
	if err != nil {
		panic(err)
	}

	pool, err = pgxpool.New(context.Background(), os.Getenv("DB_URL"))

	defer sentry.Recover()

	if err != nil {
		log.Fatal(err)
	}

	startScrapers()
}
