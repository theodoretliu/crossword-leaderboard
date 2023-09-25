package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/mattn/go-sqlite3"
	"github.com/newrelic/go-agent/v3/integrations/nrgin"
	"github.com/newrelic/go-agent/v3/newrelic"
)

var db *sql.DB
var pgxPool *pgxpool.Pool

const defaultPort = "8080"

func main() {
	err := sentry.Init(sentry.ClientOptions{
		Dsn:              os.Getenv("DSN"),
		AttachStacktrace: true,
	})
	if err != nil {
		panic(err)
	}

	defer sentry.Recover()

	db, err = sql.Open("sqlite3", os.Getenv("DB_URL"))
	// if err != nil {
	// 	log.Fatal(err)
	// }

	pgxPool, err = pgxpool.New(context.Background(), os.Getenv("DB_URL"))
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
			err := dbActionsForDate(pgxPool, time.Now())
			if err != nil {
				sentry.CaptureException(err)
			}
			time.Sleep(10 * time.Second)
		}
	}()

	// go func() {
	// 	for {
	// 		err := setElosInDb()
	// 		if err != nil {
	// 			sentry.CaptureException(err)
	// 		}
	// 		time.Sleep(10 * time.Second)
	// 	}
	// }()

	r := gin.Default()

	r.Use(sentrygin.New(sentrygin.Options{Repanic: true}))

	pprof.Register(r)

	r.Use(cors.Default())

	r.Use(nrgin.Middleware(app))

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	r.GET("/new", func(c *gin.Context) {
		c.JSON(http.StatusOK, NewIndexHandler())
	})

	r.GET("/feature_flag", func(c *gin.Context) {
		flag, ok := c.GetQuery("flag")

		if !ok {
			panic(fmt.Errorf("flag was not provided"))
		}

		flagValue, err := GetFeatureFlag(flag)

		if err != nil {
			panic(err)
		}

		c.JSON(http.StatusOK, struct {
			Status bool
		}{Status: flagValue})
	})

	r.GET("/all_users", func(c *gin.Context) {
		c.JSON(http.StatusOK, AllUsersHandler())
	})

	r.GET("/users/:userId", func(c *gin.Context) {
		userId, err := strconv.ParseInt(c.Param("userId"), 10, 64)
		if err != nil {
			panic(err)
		}

		c.JSON(http.StatusOK, UserHandler(userId))
	})

	r.GET("/week/:year/:month/:day", func(c *gin.Context) {
		year, err := strconv.Atoi(c.Param("year"))
		if err != nil {
			panic(err)
		}

		month, err := strconv.Atoi(c.Param("month"))
		if err != nil {
			panic(err)
		}

		day, err := strconv.Atoi(c.Param("day"))
		if err != nil {
			panic(err)
		}

		c.JSON(http.StatusOK, WeekTimesHandler(year, month, day))
	})

	r.Run(":" + port)
}
