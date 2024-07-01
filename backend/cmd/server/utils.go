package main

import (
	"context"
	"math"
	"time"

	"github.com/jackc/pgx/v5"
)

func getFirstDayOfWeek(givenDay time.Time) time.Time {
	year, month, day := givenDay.Date()
	truncated := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)

	for truncated.Weekday() != time.Monday {
		truncated = truncated.AddDate(0, 0, -1)
	}

	return truncated
}

func getDaysOfTheWeek(start time.Time) []time.Time {
	days := []time.Time{}

	for i := 0; i < 7; i++ {
		days = append(days, start)
		start = start.AddDate(0, 0, 1)
	}

	return days
}

func GetWeeksWorstTimes(daysOfTheWeek []time.Time) []int32 {
	var worstTimes []int32

	query := `
		SELECT max(time_in_seconds), date FROM times
		WHERE date >= $1 AND date <= $2
		GROUP BY date(date)
		ORDER BY date(date) ASC
	`

	rows, err := pool.Query(context.Background(), query, daysOfTheWeek[0], daysOfTheWeek[6])

	if err != nil {
		panic(err)
	}

	scannedRows, err := pgx.CollectRows(rows, pgx.RowToStructByPos[struct {
		WorstTime int32
		Date      time.Time
	}])

	if err != nil {
		panic(err)
	}

	i := 0
	for _, row := range scannedRows {
		for row.Date.UTC() != daysOfTheWeek[i] {
			i++
			worstTimes = append(worstTimes, -1)
		}

		worstTimes = append(worstTimes, row.WorstTime)
		i++
	}

	return worstTimes
}

func max(x, y int64) int64 {
	if x > y {
		return x
	}
	return y
}

func min(x, y int64) int64 {
	if x < y {
		return x
	}
	return y
}

func WeeklyAverage(times []int32, weeksWorstTimes []int32) int32 {
	totalSquares := 0
	totalTime := int32(0)
	for i := 0; i < len(weeksWorstTimes); i++ {
		if i < len(times) && times[i] != -1 {
			totalTime += times[i]
		} else if weeksWorstTimes[i] != -1 {
			totalTime += weeksWorstTimes[i] + 1
		}

		if weeksWorstTimes[i] != -1 {
			if i != 6 {
				totalSquares += 25
			} else {
				totalSquares += 49
			}
		}
	}

	return int32(math.Round(float64(totalTime) / float64(totalSquares) * float64(25)))

}
