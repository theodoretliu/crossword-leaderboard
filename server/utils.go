package main

import (
	"log"
	"math"
	"time"
)

func getFirstDayOfWeek(givenDay time.Time) time.Time {
	year, month, day := givenDay.Date()
	truncated := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)

	for truncated.Weekday() != time.Sunday {
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
	GROUP BY date
	ORDER BY date ASC
	`

	rows, err := db.Query(query, daysOfTheWeek[0], daysOfTheWeek[6])

	if err != nil {
		log.Fatal(err)
	}

	i := 0

	for rows.Next() {
		var worstTime int32
		var date time.Time

		err = rows.Scan(&worstTime, &date)

		if err != nil {
			log.Fatal(err)
		}

		for date.UTC() != daysOfTheWeek[i] {
			i++
			worstTimes = append(worstTimes, -1)
		}

		worstTimes = append(worstTimes, worstTime)
		i++
	}

	return worstTimes
}

func max(x, y int) int {
	if x > y {
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
			totalTime += weeksWorstTimes[i]
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
