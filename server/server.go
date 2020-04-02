package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime/pprof"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	_ "github.com/lib/pq"
)

func DaysOfTheWeek() []time.Time {

	now := time.Now().UTC()
	dayOfWeek := now.Weekday()

	duration := time.Duration(24*dayOfWeek) * time.Hour

	startOfWeek := now.Add(-duration).Truncate(24 * time.Hour)

	var daysOfTheWeek []time.Time

	for i := 0; i < 7; i++ {
		daysOfTheWeek = append(daysOfTheWeek, startOfWeek)
		startOfWeek = startOfWeek.AddDate(0, 0, 1)
	}

	return daysOfTheWeek
}

func WeeksTimes(id string) ([]int, error) {

	daysOfTheWeek := DaysOfTheWeek()

	var times []int
	rows, err := db.Query("SELECT time_in_seconds, date FROM times WHERE user_id = $1 AND date >= $2 AND date <= $3 ORDER BY date ASC",
		id, daysOfTheWeek[0], daysOfTheWeek[6])

	if err != nil {
		return nil, err
	}

	i := 0

	for rows.Next() {
		var time_ int
		var date time.Time

		err = rows.Scan(&time_, &date)

		if err != nil {
			return nil, err
		}

		for date.UTC() != daysOfTheWeek[i] {
			i++
			times = append(times, -1)
		}

		times = append(times, time_)
		i++
	}

	for len(times) != 7 {
		times = append(times, -1)
	}

	return times, nil
}

func WeeksWorstTimes(c context.Context) ([]int, error) {
	cache, ok := c.Value("cache").(map[string]interface{})

	if !ok {
		return nil, errors.New("Could not retrieve cache from context")
	}

	weeksWorstTimesInt, found := cache["weeksWorstTimes"]

	if found {
		return weeksWorstTimesInt.([]int), nil
	}

	daysOfTheWeek := DaysOfTheWeek()

	var worstTimes []int

	rows, err := db.Query("SELECT max(time_in_seconds), date FROM times WHERE date >= $1 AND date <= $2 GROUP BY date ORDER BY date ASC", daysOfTheWeek[0], daysOfTheWeek[6])

	if err != nil {
		return nil, err
	}

	i := 0

	for rows.Next() {
		var worstTime int
		var date time.Time

		err = rows.Scan(&worstTime, &date)

		if err != nil {
			return nil, err
		}

		for date.UTC() != daysOfTheWeek[i] {
			i++
			worstTimes = append(worstTimes, -1)
		}

		worstTimes = append(worstTimes, worstTime+1)
		i++
	}

	for len(worstTimes) != 7 {
		worstTimes = append(worstTimes, -1)
	}

	cache["weeksWorstTimes"] = worstTimes

	return worstTimes, nil
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

		"weeklyAverage": &graphql.Field{
			Type:        graphql.Int,
			Description: "The user's weekly average",
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				user, ok := p.Source.(UserGQL)

				if !ok {
					return nil, errors.New("Could not convert to UserGQL struct")
				}

				weeksWorstTimes, err := WeeksWorstTimes(p.Context)
				if err != nil {
					return nil, err
				}

				weeksTimes, err := WeeksTimes(user.ID)
				if err != nil {
					return nil, err
				}

				weights := []int{25, 25, 25, 25, 25, 25, 49}

				total := 0
				totalWeight := 0

				for i := 0; i < 7; i++ {
					if weeksTimes[i] == -1 && weeksWorstTimes[i] != -1 {
						total += weights[i] * (weeksWorstTimes[i] + 1)
					} else if weeksTimes[i] != -1 {
						total += weights[i] * weeksTimes[i]
					}

					if weeksTimes[i] != -1 || weeksWorstTimes[i] != -1 {
						totalWeight += weights[i]
					}
				}

				average := int(float64(total) / float64(totalWeight))

				return average, nil

			},
		},

		"weeksTimes": &graphql.Field{
			Type: graphql.NewList(graphql.Int),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				value, ok := p.Source.(UserGQL)

				if !ok {
					return nil, errors.New("")
				}

				return WeeksTimes(value.ID)
			},
		},
	},
})

type GraphqlPost struct {
	OperationName string
	Variables     map[string]interface{}
	Query         string
}

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		pprof.StopCPUProfile()
		os.Exit(0)
	}()

	err := sentry.Init(sentry.ClientOptions{
		Dsn:              os.Getenv("DSN"),
		AttachStacktrace: true,
	})
	if err != nil {
		panic(err)
	}

	defer sentry.Flush(2 * time.Second)
	defer sentry.Recover()

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

	http.Handle("/", httpHeaderMiddleware(h))

	fmt.Println("Listening on port 8080")
	http.ListenAndServe(":8080", nil)
}

func httpHeaderMiddleware(next *handler.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		ctx := context.WithValue(r.Context(), "cache", make(map[string]interface{}))

		next.ContextHandler(ctx, w, r)
	})
}
