package main

import "context"

type AllUsersResponse struct {
	Users []struct {
		Id   int64
		Name string
	}
}

func AllUsersHandler() AllUsersResponse {
	rows, err := queries.ListUsers(context.Background())
	if err != nil {
		panic(err)
	}

	resp := AllUsersResponse{Users: make([]struct {
		Id   int64
		Name string
	}, len(rows))}
	for i, r := range rows {
		resp.Users[i].Id = int64(r.ID)
		resp.Users[i].Name = deref(r.Name)
	}
	return resp
}
