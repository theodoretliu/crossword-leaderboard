package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func DaysOfTheWeek() []time.Time {

	now := time.Now()
	dayOfWeek := now.Weekday()

	duration := time.Duration(24*dayOfWeek) * time.Hour

	startOfWeek := now.Add(-duration).Truncate(24 * time.Hour).UTC()

	var daysOfTheWeek []time.Time

	for i := 0; i < 7; i++ {
		daysOfTheWeek = append(daysOfTheWeek, startOfWeek)
		startOfWeek = startOfWeek.AddDate(0, 0, 1)
	}

	return daysOfTheWeek
}

var db *sql.DB

type UserGQL struct {
	ID       string
	Name     string
	Username string
}

var userType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "User",
	Description: "A user of the service",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type:        graphql.ID,
			Description: "The id of the user",
		},

		"name": &graphql.Field{
			Type:        graphql.String,
			Description: "The name of the user",
		},

		"username": &graphql.Field{
			Type:        graphql.String,
			Description: "The user's username",
		},

		"weeksTimes": &graphql.Field{
			Type: graphql.NewList(graphql.Int),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				value, ok := p.Source.(UserGQL)

				if !ok {
					return nil, errors.New("")
				}

				daysOfTheWeek := DaysOfTheWeek()

				var times []int

				for _, day := range daysOfTheWeek {
					row := db.QueryRow("SELECT time_in_seconds FROM times WHERE user_id = $1 AND date = $2 LIMIT 1", value.ID, day)

					var time int

					err := row.Scan(&time)

					if err == sql.ErrNoRows {
						times = append(times, -1)
					} else if err != nil {
						return nil, err
					} else {
						times = append(times, time)
					}

				}

				return times, nil
			},
		},
	},
})

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	err = sentry.Init(sentry.ClientOptions{Dsn: os.Getenv("DSN")})
	if err != nil {
		panic(err)
	}

	tryDb, err := sql.Open("postgres", os.Getenv("DB_URL"))
	if err != nil {
		panic(err)
	}

	db = tryDb

	defer db.Close()

	go func() {
		for {
			err := scrape(db)
			if err != nil {
				sentry.CaptureException(err)
			}
			time.Sleep(10 * time.Second)
		}
	}()

	fields := graphql.Fields{

		"user": &graphql.Field{
			Type: userType,
			Args: graphql.FieldConfigArgument{
				"username": &graphql.ArgumentConfig{
					Type:        graphql.String,
					Description: "The user's username",
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				username := p.Args["username"]

				row := db.QueryRow("SELECT username, id, name FROM users WHERE username = $1", username)

				var userInfo UserGQL

				err := row.Scan(&userInfo.Username, &userInfo.ID, &userInfo.Name)

				if err != nil {
					return nil, err
				}

				return userInfo, nil
			},
		},

		"users": &graphql.Field{
			Type:        graphql.NewList(userType),
			Description: "All users",
			Resolve: func(P graphql.ResolveParams) (interface{}, error) {

				rows, err := db.Query("SELECT id, username, name FROM users")
				if err != nil {
					return nil, err
				}

				res := []UserGQL{}

				for rows.Next() {
					var tmp UserGQL

					rows.Scan(&tmp.ID, &tmp.Username, &tmp.Name)

					res = append(res, tmp)
				}

				return res, nil
			},
		},

		"daysOfTheWeek": &graphql.Field{
			Type:        graphql.NewList(graphql.DateTime),
			Description: "The days of the current week",
			Resolve: func(P graphql.ResolveParams) (interface{}, error) {

				return DaysOfTheWeek(), nil
			},
		},
	}

	mutationFields := graphql.Fields{
		"updateCookie": &graphql.Field{
			Type: graphql.String,
			Args: graphql.FieldConfigArgument{
				"cookie": &graphql.ArgumentConfig{
					Type:        graphql.String,
					Description: "The new cookie",
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				newCookie, ok := params.Args["cookie"].(string)

				if !ok {
					return nil, errors.New("")
				}

				_, err := db.Exec("UPDATE cookie SET cookie = $1", newCookie)

				if err != nil {
					return nil, err
				}

				return newCookie, nil
			},
		},
	}

	rootQuery := graphql.ObjectConfig{Name: "RootQuery", Fields: fields}
	mutationQuery := graphql.ObjectConfig{Name: "MutationQuery", Fields: mutationFields}

	schemaConfig := graphql.SchemaConfig{Query: graphql.NewObject(rootQuery),
		Mutation: graphql.NewObject(mutationQuery)}
	schema, err := graphql.NewSchema(schemaConfig)

	if err != nil {
		log.Fatalf("failed to create new schema, error %v", err)
	}

	h := handler.New(&handler.Config{
		Schema:   &schema,
		Pretty:   true,
		GraphiQL: true,
	})

	http.Handle("/graphql", h)
	fmt.Println("Listening on port 8080")
	http.ListenAndServe(":8080", nil)
}
