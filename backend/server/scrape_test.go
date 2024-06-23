package main

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestScrape(t *testing.T) {
	pool, _ = pgxpool.New(context.Background(), os.Getenv("DB_URL"))

	t.Run("date in past", func(t *testing.T) {
		err := dbActionsForDate(time.Date(2022, 9, 11, 0, 0, 0, 0, time.UTC))
		if err != nil {
			fmt.Println(err)
			t.Fail()
		}
	})

	t.Run("date in far future", func(t *testing.T) {
		err := dbActionsForDate(time.Date(2027, 9, 11, 0, 0, 0, 0, time.UTC))
		if err != nil {
			fmt.Println(err)
			t.Fail()
		}
	})

	t.Run("scrape everything", func(t *testing.T) {
		t.Skip()
		scrapeAllDays()
	})

	t.Run("run scrapers", func(t *testing.T) {
		t.Skip()
		startScrapers()
	})
}
