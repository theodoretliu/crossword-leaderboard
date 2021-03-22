package main

type AllUsersResponse struct {
	Users []string
}

func AllUsersHandler() AllUsersResponse {
	query := `SELECT username FROM users;`

	res, err := db.Query(query)

	if err != nil {
		panic(err)
	}

	names := []string{}

	for res.Next() {
		var name string

		err = res.Scan(&name)

		if err != nil {
			panic(err)
		}

		names = append(names, name)
	}

	return AllUsersResponse{names}
}
