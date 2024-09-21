package main

import (
	"context"
	"sort"
	"time"

	"github.com/jackc/pgx/v5"
)

type stats struct {
	Average      float64
	Median       float64
	Best         int64
	Worst        int64
	NumCompleted int64
}

type dateElo struct {
	Date time.Time
	Elo  float64
}

type timeStruct struct {
	TimeInSeconds int64     `json:"t"`
	Date          time.Time `json:"d"`
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
	AllTimes      []timeStruct
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

	allTimes, err := pgx.CollectRows(rows, pgx.RowToStructByPos[timeStruct])

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

	allWeekdayTimes := make([]int64, 0)
	allSaturdayTimes := make([]int64, 0)

	for _, timeStruct := range allTimes {
		var timeInSeconds = timeStruct.TimeInSeconds
		var date = timeStruct.Date

		date = date.UTC()

		// at the moment, times less than 3 seconds are implausible
		if timeInSeconds < 3 {
			continue
		}

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

			allSaturdayTimes = append(allSaturdayTimes, timeInSeconds)

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

			allWeekdayTimes = append(allWeekdayTimes, timeInSeconds)

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

	// computing weekday median
	weekdayMedian := float64(0)
	sort.Slice(allWeekdayTimes, func(i, j int) bool { return allWeekdayTimes[i] < allWeekdayTimes[j] })
	if len(allWeekdayTimes) > 0 {
		if len(allWeekdayTimes)%2 == 0 {
			weekdayMedian = float64(allWeekdayTimes[len(allWeekdayTimes)/2-1]+allWeekdayTimes[len(allWeekdayTimes)/2]) / 2
		} else {
			weekdayMedian = float64(allWeekdayTimes[len(allWeekdayTimes)/2])

		}
	}
	miniStats := stats{Median: weekdayMedian}

	miniStats.NumCompleted = miniCount
	if miniCount > 0 {
		miniStats.Average = float64(totalMiniTimes) / float64(miniCount)
	}
	if bestMiniTime != nil {
		miniStats.Best = *bestMiniTime
		miniStats.Worst = *worstMiniTime
	}

	// computing saturday median
	saturdayMedian := float64(0)
	sort.Slice(allSaturdayTimes, func(i, j int) bool { return allSaturdayTimes[i] < allSaturdayTimes[j] })
	if len(allSaturdayTimes) > 0 {
		if len(allSaturdayTimes)%2 == 0 {
			saturdayMedian = float64(allSaturdayTimes[len(allSaturdayTimes)/2-1]+allSaturdayTimes[len(allSaturdayTimes)/2]) / 2
		} else {
			saturdayMedian = float64(allSaturdayTimes[len(allSaturdayTimes)/2])
		}
	}
	saturdayStats := stats{Median: saturdayMedian}

	saturdayStats.NumCompleted = saturdayCount
	if saturdayCount > 0 {
		saturdayStats.Average = float64(totalSaturdayTimes) / float64(saturdayCount)
	}
	if bestSaturdayTime != nil {
		saturdayStats.Best = *bestSaturdayTime
		saturdayStats.Worst = *worstSaturdayTime
	}

	// computing overall statistics
	allFloatTimes := make([]float64, 0)

	for _, time := range allWeekdayTimes {
		allFloatTimes = append(allFloatTimes, float64(time))
	}
	for _, time := range allSaturdayTimes {
		allFloatTimes = append(allFloatTimes, float64(time)/49*25)
	}

	overallMedian := float64(0)
	sort.Slice(allFloatTimes, func(i, j int) bool { return allFloatTimes[i] < allFloatTimes[j] })

	if len(allFloatTimes) > 0 {
		if len(allFloatTimes)%2 == 0 {
			overallMedian = (allFloatTimes[len(allFloatTimes)/2-1] + allFloatTimes[len(allFloatTimes)/2]) / 2
		} else {
			overallMedian = allFloatTimes[len(allFloatTimes)/2]
		}
	}

	overallStats := stats{Median: overallMedian}

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
			AllTimes:      allTimes,
		}
	}

	return UserResponse{
		Username:      username,
		MiniStats:     miniStats,
		SaturdayStats: saturdayStats,
		OverallStats:  overallStats,
		LongestStreak: longestStreak,
		CurrentStreak: currentStreak,
		EloHistory:    []dateElo{},
		AllTimes:      allTimes,
	}
}
