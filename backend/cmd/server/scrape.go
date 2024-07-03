package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

const MinYear = 2014
const MinMonth = 8
const MinDay = 21

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
	onlyUserIds := []any{}
	userIdParens := []string{}

	for _, user := range apiResults {
		valueParens = append(valueParens, fmt.Sprintf("($%d, $%d)", count, count+1))
		count += 2
		values = append(values, user.UserID, user.Name)
		userIdParens = append(userIdParens, fmt.Sprintf("$%d", len(onlyUserIds)+1))
		onlyUserIds = append(onlyUserIds, user.UserID)
	}

	queryString := fmt.Sprintf("INSERT INTO users (nyt_user_id, name) VALUES %s ON CONFLICT (nyt_user_id) DO UPDATE SET name = EXCLUDED.name", strings.Join(valueParens, ","))
	_, err := tx.Exec(context.Background(), queryString, values...)
	if err != nil {
		return nil, err
	}

	retrieveUsersQuery := fmt.Sprintf("SELECT id FROM users WHERE nyt_user_id IN (%s) ORDER BY nyt_user_id ASC", strings.Join(userIdParens, ", "))
	retrievedUserIds, err := tx.Query(
		context.Background(),
		retrieveUsersQuery,
		onlyUserIds...,
	)
	if err != nil {
		return nil, err
	}

	userIds, err := pgx.CollectRows(retrievedUserIds, pgx.RowTo[int64])
	if err != nil {
		return nil, err
	}

	return userIds, nil
}
