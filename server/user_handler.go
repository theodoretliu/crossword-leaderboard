package main

import (
	"context"
	"time"
)

type stats struct {
	Average      float64
	Best         int64
	Worst        int64
	NumCompleted int64
}

type dateElo struct {
	Date time.Time
	Elo  float64
}

type UserResponse struct {
	Username      string
	MiniStats     stats
	SaturdayStats stats
	OverallStats  stats
	EloHistory    []dateElo
	LongestStreak int64
	CurrentStreak int64
	PeakElo       float64
	CurrentElo    float64
}

func UserHandler(userId int64) UserResponse {
	var username string
	err := pool.QueryRow(context.Background(), "SELECT name FROM users where id = $1", userId).Scan(&username)
	if err != nil {
		panic(err)
	}

	rows, err := pool.Query(
		context.Background(),
		`
			SELECT time_in_seconds, date FROM times WHERE user_id = $1 ORDER BY date(date) ASC;
		`, userId)

	if err != nil {
		panic(err)
	}

	longestStreak := int64(0)
	currentStreak := int64(0)
	var previousDay *time.Time

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

		if previousDay == nil {
			currentStreak = 1
			longestStreak = 1
			previousDay = new(time.Time)
			*previousDay = date
		} else {
			if date.Sub(*previousDay) > 24*time.Hour {
				longestStreak = max(longestStreak, currentStreak)
				currentStreak = 1
			} else {
				currentStreak++
				longestStreak = max(longestStreak, currentStreak)
			}
			*previousDay = date
		}

		// collecting statistics based on the day the of week
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

	eloActive, err := GetFeatureFlag("elos")

	if err != nil {
		panic(err)
	}

	if eloActive {
		eloHistory, err := getEloHistory(userId)
		if err != nil {
			panic(err)
		}

		peakElo, err := getPeakElo(userId)
		if err != nil {
			panic(err)
		}

		currentElo, err := getCurrentElo(userId)
		if err != nil {
			panic(err)
		}

		return UserResponse{
			Username:      username,
			MiniStats:     miniStats,
			SaturdayStats: saturdayStats,
			OverallStats:  overallStats,
			EloHistory:    eloHistory,
			LongestStreak: longestStreak,
			CurrentStreak: currentStreak,
			PeakElo:       peakElo,
			CurrentElo:    currentElo,
		}
	} else {
		return UserResponse{
			Username:      username,
			MiniStats:     miniStats,
			SaturdayStats: saturdayStats,
			OverallStats:  overallStats,
			LongestStreak: longestStreak,
			CurrentStreak: currentStreak,
			EloHistory:    []dateElo{},
		}
	}
}
