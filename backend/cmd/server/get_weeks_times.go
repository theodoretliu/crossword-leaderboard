package main

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

type userInfo struct {
	UserId       int64
	Username     string
	WeeksTimes   []int32
	WeeksAverage int32
	Elo          float64
	Qualified    bool
}

type weeksInfo struct {
	Users         []userInfo
	DaysOfTheWeek []string
}

func getWeeksInfo(day time.Time, shouldComputeElo bool) weeksInfo {
	firstDayOfWeek := getFirstDayOfWeek(day)
	daysOfTheWeek := getDaysOfTheWeek(firstDayOfWeek)
	query := `
		select u.id, u.name, t.time_in_seconds, t.date from
		users as u
		inner join
		times as t
		on t.user_id = u.id AND t.date >= $1 AND t.date <= $2
		order by u.id, t.date;
	`

	rows, err := pool.Query(context.Background(), query, daysOfTheWeek[0], daysOfTheWeek[6])

	if err != nil {
		panic(err)
	}

	resRows, err := pgx.CollectRows(rows, pgx.RowToStructByPos[struct {
		Id            int64
		Name          string
		TimeInSeconds int32
		Date          time.Time
	}])

	if err != nil {
		panic(err)
	}

	result := []userInfo{}

	var currentUser string
	dateIndex := 0

	for _, row := range resRows {
		if row.Name != currentUser {
			result = append(result, userInfo{row.Id, row.Name, []int32{}, 0, 1000.0, false})
			currentUser = row.Name
			dateIndex = 0
		}

		for row.Date.UTC() != daysOfTheWeek[dateIndex] {
			result[len(result)-1].WeeksTimes = append(result[len(result)-1].WeeksTimes, -1)
			dateIndex++
		}

		result[len(result)-1].WeeksTimes = append(result[len(result)-1].WeeksTimes, row.TimeInSeconds)
		dateIndex++
	}

	saturdayPossible := time.Now().UTC().After(daysOfTheWeek[5])
	daysElapsed := 0

	today := time.Now().UTC().Round(24 * time.Hour)
	var curDay time.Time

	if daysOfTheWeek[6].After(today) {
		curDay = today
	} else {
		curDay = daysOfTheWeek[6]
	}

	daysElapsed = int((curDay.Sub(daysOfTheWeek[0])).Hours())/24 + 1

	for i := range result {
		user := &result[i]
		numDaysCompleted := 0

		for _, time := range user.WeeksTimes {
			if time != -1 {
				numDaysCompleted++
			}
		}

		if saturdayPossible {
			user.Qualified = len(user.WeeksTimes) > 5 && user.WeeksTimes[5] != -1 && numDaysCompleted >= daysElapsed-3
		} else {
			user.Qualified = numDaysCompleted >= daysElapsed-3
		}
	}

	weeksWorstTimes := GetWeeksWorstTimes(daysOfTheWeek)

	for i := 0; i < len(result); i++ {
		result[i].WeeksAverage = WeeklyAverage(result[i].WeeksTimes, weeksWorstTimes)
	}

	// elosActive, err := GetFeatureFlag("elos")

	// if elosActive {

	// 	elos, err := getElosForDate(daysOfTheWeek[len(daysOfTheWeek)-1])
	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	for i := 0; i < len(result); i++ {
	// 		val, ok := elos[result[i].UserId]
	// 		if !ok {
	// 			result[i].Elo = 1000.0
	// 		} else {
	// 			result[i].Elo = val
	// 		}
	// 	}
	// }

	daysOfTheWeekStrings := []string{}

	for _, day := range daysOfTheWeek {
		daysOfTheWeekStrings = append(daysOfTheWeekStrings, day.Format(time.RFC1123Z))
	}

	return weeksInfo{Users: result, DaysOfTheWeek: daysOfTheWeekStrings}
}
