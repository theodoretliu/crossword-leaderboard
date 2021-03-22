package main

import (
	"time"
)

type stats struct {
	Average      float64
	Best         int64
	Worst        int64
	NumCompleted int64
}

type UserResponse struct {
	Username      string
	MiniStats     stats
	SaturdayStats stats
	OverallStats  stats
}

func UserHandler(userId int) UserResponse {
	row := db.QueryRow(`SELECT username FROM users WHERE id = ?`, userId)

	var username string
	err := row.Scan(&username)
	if err != nil {
		panic(err)
	}

	rows, err := db.Query(`SELECT time_in_seconds, date FROM times WHERE user_id = ? ORDER BY date ASC;`, userId)
	if err != nil {
		panic(err)
	}

	totalSaturdayTimes := int64(0)
	var bestSaturdayTime *int64
	var worstSaturdayTime *int64
	saturdayCount := int64(0)

	totalMiniTimes := int64(0)
	var bestMiniTime *int64
	var worstMiniTime *int64
	miniCount := int64(0)

	for rows.Next() {
		var timeInSeconds int64
		var date time.Time

		err = rows.Scan(&timeInSeconds, &date)

		if err != nil {
			panic(err)
		}

		date = date.UTC()

		if date.Weekday() == time.Saturday {
			totalSaturdayTimes += timeInSeconds
			saturdayCount++

			if bestSaturdayTime == nil || worstSaturdayTime == nil {
				bestSaturdayTime = new(int64)
				worstSaturdayTime = new(int64)
				*bestSaturdayTime = timeInSeconds
				*worstSaturdayTime = timeInSeconds
			} else {
				*bestSaturdayTime = min(*bestSaturdayTime, timeInSeconds)
				*worstSaturdayTime = max(*worstSaturdayTime, timeInSeconds)
			}
		} else {
			totalMiniTimes += timeInSeconds
			miniCount++

			if bestMiniTime == nil || worstMiniTime == nil {
				bestMiniTime = new(int64)
				worstMiniTime = new(int64)
				*bestMiniTime = timeInSeconds
				*worstMiniTime = timeInSeconds
			} else {
				*bestMiniTime = min(*bestMiniTime, timeInSeconds)
				*worstMiniTime = max(*worstMiniTime, timeInSeconds)
			}
		}

	}

	miniStats := stats{}

	miniStats.NumCompleted = miniCount
	if miniCount > 0 {
		miniStats.Average = float64(totalMiniTimes) / float64(miniCount)
	}
	if bestMiniTime != nil {
		miniStats.Best = *bestMiniTime
		miniStats.Worst = *worstMiniTime
	}

	saturdayStats := stats{}

	saturdayStats.NumCompleted = saturdayCount
	if saturdayCount > 0 {
		saturdayStats.Average = float64(totalSaturdayTimes) / float64(saturdayCount)
	}
	if bestSaturdayTime != nil {
		saturdayStats.Best = *bestSaturdayTime
		saturdayStats.Worst = *worstSaturdayTime
	}

	overallStats := stats{}

	overallStats.NumCompleted = miniCount + saturdayCount
	if overallStats.NumCompleted > 0 {
		overallStats.Average = float64(totalMiniTimes+totalSaturdayTimes) / float64(25*miniCount+49*saturdayCount) * float64(25)
	}
	overallStats.Best = min(miniStats.Best, saturdayStats.Best)
	overallStats.Worst = max(miniStats.Worst, saturdayStats.Worst)

	return UserResponse{
		Username:      username,
		MiniStats:     miniStats,
		SaturdayStats: saturdayStats,
		OverallStats:  overallStats,
	}
}