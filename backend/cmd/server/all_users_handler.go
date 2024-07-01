package main

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type AllUsersResponse struct {
	Users []struct {
		Id   int64
		Name string
	}
}

func AllUsersHandler() AllUsersResponse {
	query := `SELECT id, name FROM users;`

	rows, err := pool.Query(context.Background(), query)
	if err != nil {
		panic(err)
	}

	resRows, err := pgx.CollectRows(rows, pgx.RowToStructByName[struct {
		Id   int64
		Name string
	}])

	if err != nil {
		panic(err)
	}

	return AllUsersResponse{Users: resRows}
}
