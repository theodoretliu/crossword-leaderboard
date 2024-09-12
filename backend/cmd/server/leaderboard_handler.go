package main

import (
	"context"
	"log"
	"os"
	"sync"

	"github.com/jackc/pgx/v5"
)

type LeaderboardEntry struct {
	ID           int64
	Name         string
	Percentile10 float64
	Percentile20 float64
	Percentile30 float64
	Percentile40 float64
	Percentile50 float64
	Percentile60 float64
	Percentile70 float64
	Percentile80 float64
	Percentile90 float64
}

var leaderboardQuery string

func LeaderboardHandler() []LeaderboardEntry {
	var once sync.Once

	once.Do(func() {
		queryBytes, err := os.ReadFile("./cmd/server/leaderboard_query.sql")
		if err != nil {
			log.Fatalf("Failed to read leaderboard query file: %v", err)
		}
		leaderboardQuery = string(queryBytes)
	})

	rows, err := pool.Query(context.Background(), leaderboardQuery)
	if err != nil {
		panic(err)
	}

	resRows, err := pgx.CollectRows(rows, pgx.RowToStructByPos[LeaderboardEntry])

	if err != nil {
		panic(err)
	}

	return resRows
}
