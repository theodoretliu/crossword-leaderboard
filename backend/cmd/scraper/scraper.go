package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/theodoretliu/crossword-leaderboard/internal/scrape"
)

func hoursToSleepTime(hours float64) int64 {
	return int64(5*hours*hours/576 + 5*hours/24 + 10)
}

func scrapeDate(pool *pgxpool.Pool, date time.Time) {
	for {
		timeSince := time.Since(date)
		time.Sleep(time.Duration(hoursToSleepTime(timeSince.Hours())) * time.Second)
		err := scrape.DbActionsForDate(pool, date)
		if err != nil {
			fmt.Println(err)
			sentry.CaptureException(err)
		}
	}
}

func startScrapers() {
	minDate := time.Date(scrape.MinYear, scrape.MinMonth, scrape.MinDay, 0, 0, 0, 0, time.UTC)
	oneDayDuration := 24 * time.Hour
	today := time.Now().UTC().Truncate(oneDayDuration)
	iterDate := minDate

	for iterDate.Compare(today) <= 0 {
		go scrapeDate(pool, iterDate)
		iterDate = iterDate.AddDate(0, 0, 1)
	}

	for {
		fmt.Println(time.Since(today))
		if time.Since(today) > oneDayDuration {
			fmt.Println("adding a new scraper")
			today = today.AddDate(0, 0, 1)
			fmt.Println("new date: ", today)
			go scrapeDate(pool, today)
		}

		time.Sleep(10 * time.Second)
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
