package main

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestScrapeAll(t *testing.T) {
	t.Skip()
	pool, _ = pgxpool.New(context.Background(), os.Getenv("DB_URL"))
	scrapeAllDays()
}
