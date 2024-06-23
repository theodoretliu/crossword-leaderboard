package scrape

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/exp/slices"
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

func insertResults(tx pgx.Tx, apiResults []result, userIds []int64, date time.Time) error {
	fmt.Println("inserting results for ", date)
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

	// if there are no results yet
	if len(values) == 0 {
		fmt.Println("no results for date", date)
		return nil
	}

	queryString := fmt.Sprintf("INSERT INTO times (user_id, time_in_seconds, date) VALUES %s ON CONFLICT DO NOTHING", strings.Join(valueParens, ","))
	_, err := tx.Exec(context.Background(), queryString, values...)
	if err != nil {
		return err
	}

	return nil
}

func DbActionsForDate(pool *pgxpool.Pool, date time.Time) error {
	tx, err := pool.Begin(context.Background())
	if err != nil {
		return err
	}

	var cookie string
	err = tx.QueryRow(context.Background(), "SELECT value FROM cookies LIMIT 1").Scan(&cookie)
	if err != nil {
		return err
	}

	apiResults, err := getResultsForDate(date, cookie)
	if err != nil {
		return err
	}

	// sort users for stability and to avoid deadlocking
	slices.SortFunc(apiResults, func(a, b result) int {
		return int(a.UserID - b.UserID)
	})

	userIds, err := updateUsers(tx, apiResults)
	if err != nil {
		return err
	}

	err = insertResults(tx, apiResults, userIds, date)
	if err != nil {
		return err
	}

	return tx.Commit(context.Background())
}
