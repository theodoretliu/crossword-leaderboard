package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"theodoretliu.com/crossword/server/graph"
	"theodoretliu.com/crossword/server/graph/generated"
)

var db *sql.DB

const defaultPort = "8080"

func handlerToGinHandler(h http.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

func main() {
	err := sentry.Init(sentry.ClientOptions{
		Dsn:              os.Getenv("DSN"),
		AttachStacktrace: true,
	})
	if err != nil {
		panic(err)
	}

	defer sentry.Flush(2 * time.Second)
	defer sentry.Recover()

	db, err = sql.Open("postgres", os.Getenv("DB_URL"))

	if err != nil {
		log.Fatal(err)
	}

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

	r := gin.Default()

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{}}))

	r.GET("/", handlerToGinHandler(playground.Handler("GraphQL playground", "/graphql")))
	r.POST("/graphql", handlerToGinHandler(middleware(db, srv)))

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	r.Run(":" + port)
}

func middleware(db *sql.DB, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		ctx := context.WithValue(r.Context(), "db", db)
		ctx = context.WithValue(ctx, "loader", graph.CreateNewWorstTimesLoader(db))
		ctx = context.WithValue(ctx, "weeksTimesLoader", graph.CreateNewWeeksTimesLoader(db))

		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
