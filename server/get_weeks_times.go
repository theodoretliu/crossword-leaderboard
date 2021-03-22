package main

import (
	"database/sql"
	"time"
)

type userInfo struct {
	Username     string
	WeeksTimes   []int32
	WeeksAverage int32
	Elo          float64
}

type weeksInfo struct {
	Users         []userInfo
	DaysOfTheWeek []string
}

func getWeeksInfo(day time.Time, shouldComputeElo bool) weeksInfo {
	firstDayOfWeek := getFirstDayOfWeek(day)
	daysOfTheWeek := getDaysOfTheWeek(firstDayOfWeek)
	query := `
		select username, time_in_seconds, date from
			(select * from users) as A
			left join
			(select * from times where date >= date(?) AND date <= date(?)) as B
			on A.id = B.user_id
			order by A.id, B.date;
	`

	rows, err := db.Query(query, daysOfTheWeek[0], daysOfTheWeek[6])

	if err != nil {
		panic(err)
	}

	result := []userInfo{}

	var currentUser string
	dateIndex := 0

	for rows.Next() {
		var (
			user          string
			timeInSeconds sql.NullInt32
			date          sql.NullTime
		)

		err = rows.Scan(&user, &timeInSeconds, &date)

		if err != nil {
			panic(err)
		}

		if user != currentUser {
			result = append(result, userInfo{user, []int32{}, 0, 1000.0})
			currentUser = user
			dateIndex = 0
		}

		if !timeInSeconds.Valid {
			continue
		}

		for date.Time.UTC() != daysOfTheWeek[dateIndex] {
			result[len(result)-1].WeeksTimes = append(result[len(result)-1].WeeksTimes, -1)
			dateIndex++
		}

		result[len(result)-1].WeeksTimes = append(result[len(result)-1].WeeksTimes, timeInSeconds.Int32)
		dateIndex++
	}

	weeksWorstTimes := GetWeeksWorstTimes(daysOfTheWeek)

	for i := 0; i < len(result); i++ {
		result[i].WeeksAverage = WeeklyAverage(result[i].WeeksTimes, weeksWorstTimes)
	}

	if !shouldComputeElo {
		for i := 0; i < len(result); i++ {
			row := db.QueryRow("SELECT elo FROM users WHERE username = ?", result[i].Username)

			var elo sql.NullFloat64

			err = row.Scan(&elo)

			if err != nil {
				panic(err)
			}

			if elo.Valid {
				result[i].Elo = elo.Float64
			}
		}
	} else {
		elos, err := computeElo(db, daysOfTheWeek[len(daysOfTheWeek)-1])
		if err != nil {
			panic(err)
		}

		for i := 0; i < len(result); i++ {
			result[i].Elo = elos[result[i].Username]
		}
	}

	daysOfTheWeekStrings := []string{}

	for _, day := range daysOfTheWeek {
		daysOfTheWeekStrings = append(daysOfTheWeekStrings, day.Format(time.RFC1123Z))
	}

	return weeksInfo{Users: result, DaysOfTheWeek: daysOfTheWeekStrings}
}
