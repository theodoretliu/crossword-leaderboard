package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/jackc/pgx/v5"
)

func parseTime(s string) (int, error) {
	split := strings.Split(s, ":")

	minutes, err := strconv.Atoi(split[0])

	if err != nil {
		return 0, err
	}

	seconds, err := strconv.Atoi(split[1])
	if err != nil {
		return 0, err
	}

	return minutes*60 + seconds, nil
}

type result struct {
	UserID int64
	Name   string
	Score  *struct {
		SecondsSpentSolving int32
	}
}

type nytResponse struct {
	Data []result
}

func getResultsForDate(date time.Time, cookie string) ([]result, error) {
	formattedDate := date.UTC().Format("2006-01-02")
	requestString := fmt.Sprintf("https://www.nytimes.com/svc/crosswords/v6/leaderboard/mini/%s.json", formattedDate)

	header := http.Header{}
	header.Add("Cookie", cookie)
	reader := strings.NewReader("")
	request, err := http.NewRequest("GET", requestString, reader)
	if err != nil {
		return nil, err
	}

	request.Header = header

	client := http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var rawData nytResponse
	json.Unmarshal(body, &rawData)

	return rawData.Data, nil
}

func updateUsers(tx pgx.Tx, apiResults []result) ([]int64, error) {
	values := []interface{}{}
	valueParens := []string{}
	count := int32(1)

	for _, user := range apiResults {
		valueParens = append(valueParens, fmt.Sprintf("($%d, $%d)", count, count+1))
		count += 2
		values = append(values, user.UserID, user.Name)
	}

	queryString := fmt.Sprintf("INSERT INTO users (nyt_user_id, name) VALUES %s ON CONFLICT (nyt_user_id) DO UPDATE SET name = name", strings.Join(valueParens, ","))
	rows, err := tx.Query(context.Background(), queryString, values...)
	if err != nil {
		return nil, err
	}

	userIds, err := pgx.CollectRows(rows, pgx.RowTo[int64])

	return userIds, nil
}

func insertResults(tx pgx.Tx, apiResults []result, userIds []int64, date time.Time) error {
	valueParens := []string{}
	values := []interface{}{}
	count := int32(1)

	for i, user := range apiResults {
		if user.Score != nil {
			userId := userIds[i]
			valueParens = append(valueParens, fmt.Sprintf("($%d, $%d, $%d)", count, count+1, count+2))
			values = append(values, userId, user.Score.SecondsSpentSolving, date.UTC().Format("2006-01-02"))
			count += 3
		}
	}
	queryString := fmt.Sprintf("INSERT INTO times (user_id, time_in_seconds, date) VALUES %s ON CONFLICT DO NOTHING", strings.join(valueParens, ","))
	_, err := tx.Exec(context.Background(), queryString, values...)
	if err != nil {
		return err
	}

	return nil
}

func dbActionsForDate(db pgx.Conn, date time.Time, cookie string) error {
	apiResults, err := getResultsForDate(date, cookie)
	if err != nil {
		return err
	}

	tx, err := db.Begin(context.Background())
	if err != nil {
		return err
	}

	userIds, err := updateUsers(tx, apiResults)
	if err != nil {
		return err
	}

	err = insertResults(tx, apiResults, userIds, date)
	if err != nil {
		return err
	}

	return nil
}

func scrape(db *sql.DB) error {
	defer sentry.Flush(2 * time.Second)
	defer sentry.Recover()
	var cookieString string

	row := db.QueryRow("SELECT cookie FROM cookie LIMIT 1")

	err := row.Scan(&cookieString)
	if err != nil {
		return err
	}

	cookieString = strings.TrimSpace(cookieString)

	header := http.Header{}
	header.Add("Cookie", string(cookieString))

	reader := strings.NewReader("")

	request, err := http.NewRequest("GET", "https://www.nytimes.com/puzzles/leaderboards", reader)
	if err != nil {
		return err
	}

	request.Header = header

	client := http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var rawData map[string]interface{}

	json.Unmarshal(body, &rawData)

	stringDate, ok := rawData["printDate"].(string)
	if !ok {
		return errors.New("Could not find printDate")
	}

	date, err := time.Parse("2006-01-02", stringDate)
	if err != nil {
		return err
	}

	for _, personData := range rawData["scoreList"].([]interface{}) {
		assertedPerson := personData.(map[string]interface{})

		row := db.QueryRow("SELECT id FROM users WHERE username = $1", assertedPerson["name"])

		var id int

		err := row.Scan(&id)

		if err != nil {
			if err != sql.ErrNoRows {
				return err
			}

			_, err := db.Exec("INSERT INTO users (name, username) VALUES ($1, $2)", assertedPerson["name"], assertedPerson["name"])
			if err != nil {
				return err
			}

			row := db.QueryRow("SELECT id FROM users WHERE username = $1", assertedPerson["name"])

			err = row.Scan(&id)
			if err != nil {
				return err
			}
		}

		if assertedPerson["solveTime"] != nil {
			solveTime, err := parseTime(assertedPerson["solveTime"].(string))

			if err != nil {
				return err
			}

			row := db.QueryRow("SELECT EXISTS (SELECT id FROM times WHERE user_id = $1 AND date(date) = date($2))", id, date)

			var doesExist bool

			err = row.Scan(&doesExist)
			if err != nil {
				return err
			}

			if !doesExist {
				_, err := db.Exec("INSERT INTO times (user_id, time_in_seconds, date) VALUES ($1, $2, date($3))", id, solveTime, date)
				if err != nil {
					return err
				}
			}

		}

	}

	return nil
}
