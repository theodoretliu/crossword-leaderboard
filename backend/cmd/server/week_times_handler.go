package main

import "time"

func WeekTimesHandler(year, month, day int) weeksInfo {
	date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)

	return getWeeksInfo(date, true)
}
