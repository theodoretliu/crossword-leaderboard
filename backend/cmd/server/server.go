package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

var db *sql.DB
var pool *pgxpool.Pool

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

	pool, err = pgxpool.New(context.Background(), os.Getenv("DB_URL"))
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	r := gin.Default()

	r.Use(sentrygin.New(sentrygin.Options{Repanic: true}))

	pprof.Register(r)

	r.Use(cors.Default())

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	r.GET("/new", func(c *gin.Context) {
		c.JSON(http.StatusOK, NewIndexHandler())
	})

	r.GET("/feature_flag", func(c *gin.Context) {
		c.JSON(http.StatusOK, struct{ Status bool }{Status: false})
		// flag, ok := c.GetQuery("flag")

		// if !ok {
		// 	panic(fmt.Errorf("flag was not provided"))
		// }

		// flagValue, err := GetFeatureFlag(flag)

		// if err != nil {
		// 	panic(err)
		// }

		// c.JSON(http.StatusOK, struct {
		// 	Status bool
		// }{Status: flagValue})
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

	r.GET("/leaderboard", func(c *gin.Context) {
		c.JSON(http.StatusOK, LeaderboardHandler())
	})

	r.Run(":" + port)
}
