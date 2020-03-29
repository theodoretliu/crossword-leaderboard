package main

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func parseTime(s string) int {
	split := strings.Split(s, ":")

	minutes, err := strconv.Atoi(split[0])
	check(err)
	seconds, err := strconv.Atoi(split[1])
	check(err)

	return minutes*60 + seconds
}

func scrape(db *sql.DB) {
	var cookieString string

	row := db.QueryRow("SELECT cookie FROM cookie LIMIT 1")

	err := row.Scan(&cookieString)
	check(err)

	cookieString = strings.TrimSpace(cookieString)

	header := http.Header{}
	header.Add("Cookie", string(cookieString))

	reader := strings.NewReader("")

	request, err := http.NewRequest("GET", "https://www.nytimes.com/puzzles/leaderboards", reader)
	check(err)

	request.Header = header

	client := http.Client{}
	resp, err := client.Do(request)
	check(err)

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	check(err)

	regex := regexp.MustCompile(`window.data\s*=\s*([^<]*)`)

	matches := regex.FindStringSubmatch(string(body))

	var rawData map[string]interface{}

	json.Unmarshal([]byte(matches[1]), &rawData)

	stringDate := rawData["printDate"].(string)
	date, err := time.Parse("2006-01-02", stringDate)
	check(err)

	for _, personData := range rawData["scoreList"].([]interface{}) {
		assertedPerson := personData.(map[string]interface{})

		row := db.QueryRow("SELECT id FROM users WHERE username = $1", assertedPerson["name"])

		var id int

		err := row.Scan(&id)

		if err != nil {
			if err != sql.ErrNoRows {
				panic(err)
			}

			_, err := db.Exec("INSERT INTO users (name, username) VALUES ($1, $2)", assertedPerson["name"], assertedPerson["name"])
			check(err)

			row := db.QueryRow("SELECT id FROM users WHERE username = $1", assertedPerson["name"])

			err = row.Scan(&id)
			check(err)
		}

		if assertedPerson["solveTime"] != nil {
			solveTime := parseTime(assertedPerson["solveTime"].(string))

			row := db.QueryRow("SELECT EXISTS (SELECT id FROM times WHERE user_id = $1 AND date = $2)", id, date)

			var doesExist bool

			err := row.Scan(&doesExist)
			check(err)

			if !doesExist {
				_, err := db.Exec("INSERT INTO times (user_id, time_in_seconds, date) VALUES ($1, $2, $3)", id, solveTime, date)
				check(err)
			}

		}

	}
}
