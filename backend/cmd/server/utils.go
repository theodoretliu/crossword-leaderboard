package main

import (
	"context"
	"math"
	"time"

	"github.com/theodoretliu/crossword-leaderboard/internal/store"
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

	for range 7 {
		days = append(days, start)
		start = start.AddDate(0, 0, 1)
	}

	return days
}

func GetWeeksWorstTimes(daysOfTheWeek []time.Time) []int32 {
	var worstTimes []int32

	scannedRows, err := queries.ListWorstTimesBetween(context.Background(), store.ListWorstTimesBetweenParams{
		StartDate: daysOfTheWeek[0],
		EndDate:   daysOfTheWeek[6],
	})
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

// deref returns *s, or "" when s is nil (a NULL column).
func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func WeeklyAverage(times []int32, weeksWorstTimes []int32) int32 {
	totalSquares := 0
	totalTime := int32(0)

	for i, time := range times {
		if time != -1 {
			totalTime += time

			if i != 5 {
				totalSquares += 25
			} else {
				totalSquares += 49
			}
		}
	}

	return int32(math.Round(float64(totalTime) / float64(totalSquares) * float64(25)))
}
