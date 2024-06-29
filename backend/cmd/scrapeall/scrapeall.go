package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/theodoretliu/crossword-leaderboard/internal/scrape"
)

func main() {
	pool, err := pgxpool.New(context.Background(), os.Getenv("DB_URL"))

	if err != nil {
		panic(err)
	}

	minDate := time.Date(scrape.MinYear, scrape.MinMonth, scrape.MinDay, 0, 0, 0, 0, time.UTC)

	if len(os.Args) == 2 {
		inputDate := os.Args[1]

		minDate, err = time.Parse("2006-01-02", inputDate)

		if err != nil {
			fmt.Println("Error parsing date:", err)
			os.Exit(1)
		}
	}

	oneDayDuration := time.Duration(int64(24 * 60 * 60 * 1_000_000_000))
	today := time.Now().UTC().Truncate(oneDayDuration)
	iterDate := minDate

	var wg sync.WaitGroup

	for iterDate.Compare(today) <= 0 {
		wg.Add(1)
		go func(date time.Time) {
			defer wg.Done()
			err := scrape.DbActionsForDate(pool, date)
			if err != nil {
				fmt.Println(err)
			}
		}(iterDate)

		fmt.Println(minDate, iterDate)
		iterDate = iterDate.AddDate(0, 0, 1)
	}

	wg.Wait()
}
