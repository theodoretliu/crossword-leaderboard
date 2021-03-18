package main

import (
	"database/sql"
	"testing"
)

func TestAllDates(t *testing.T) {
	db, err := sql.Open("sqlite3", "production.sqlite3")

	if err != nil {
		panic("hello")
	}

	getAllDates(db)

}

func TestComputeElo(t *testing.T) {
	db, err := sql.Open("sqlite3", "production.sqlite3")

	if err != nil {
		panic("hello")
	}

	computeElo(db)
}

// func TestEloUpdate(t *testing.T) {
// 	r1, r2 := 1200., 1000.

// 	fmt.Println(r1, r2)
// 	for i := 0; i < 10000; i++ {
// 		r1, r2 = eloUpdate(r1, r2, 1.0, 0.0)

// 		fmt.Println(r1, r2)
// 	}

// 	// fmt.Println(eloUpdate(1200., 1000., 1.0, 0.0))
// }
