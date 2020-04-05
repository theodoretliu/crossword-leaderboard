package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
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

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	regex := regexp.MustCompile(`window.data\s*=\s*([^<]*)`)

	matches := regex.FindStringSubmatch(string(body))

	var rawData map[string]interface{}

	json.Unmarshal([]byte(matches[1]), &rawData)

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

			row := db.QueryRow("SELECT EXISTS (SELECT id FROM times WHERE user_id = $1 AND date = $2)", id, date)

			var doesExist bool

			err = row.Scan(&doesExist)
			if err != nil {
				return err
			}

			if !doesExist {
				_, err := db.Exec("INSERT INTO times (user_id, time_in_seconds, date) VALUES ($1, $2, $3)", id, solveTime, date)
				if err != nil {
					return err
				}
			}

		}

	}

	return nil
}
