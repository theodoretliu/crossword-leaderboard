package main

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type AllUsersResponse struct {
	Users []string
}

func AllUsersHandler() AllUsersResponse {
	query := `SELECT name FROM users;`

	rows, err := pool.Query(context.Background(), query)
	if err != nil {
		panic(err)
	}

	resRows, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		panic(err)
	}

	return AllUsersResponse{Users: resRows}
}
