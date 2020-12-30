package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/newrelic/go-agent/v3/integrations/nrgin"
	"github.com/newrelic/go-agent/v3/newrelic"
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

	app, err := newrelic.NewApplication(
		newrelic.ConfigAppName("Crossword Leaderboard"),
		newrelic.ConfigLicense(os.Getenv("NEW_RELIC_LICENSE_KEY")),
	)

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

	r.Use(cors.Default())

	r.Use(nrgin.Middleware(app))

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	r.GET("/new", func(c *gin.Context) {
		c.JSON(http.StatusOK, NewIndexHandler())
	})
	r.GET("/all_users", func(c *gin.Context) {
		c.JSON(http.StatusOK, AllUsersHandler())
	})

	r.Run(":" + port)
}
