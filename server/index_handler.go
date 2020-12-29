package main

import (
	"database/sql"
	"log"
	"time"
)

type userInfo struct {
	Username     string
	WeeksTimes   []int32
	WeeksAverage int32
}

func GetDaysOfTheWeek() []time.Time {
	now := time.Now().UTC()
	dayOfWeek := now.Weekday()

	duration := time.Duration(24*dayOfWeek) * time.Hour

	startOfWeek := now.Add(-duration).Truncate(24 * time.Hour)

	var daysOfTheWeek []time.Time

	for i := 0; i < 7; i++ {
		daysOfTheWeek = append(daysOfTheWeek, startOfWeek)
		startOfWeek = startOfWeek.AddDate(0, 0, 1)
	}

	return daysOfTheWeek
}

func NewIndexHandler() []userInfo {
	daysOfTheWeek := GetDaysOfTheWeek()
	query := `
select username, time_in_seconds, date from 
	(select * from users) as A 
	left join 
	(select * from times where date >= $1 AND date <= $2) as B 
	on A.id = B.user_id 
	order by A.id, B.date;
	`

	rows, err := db.Query(query, daysOfTheWeek[0], daysOfTheWeek[6])

	if err != nil {
		log.Fatal(err)
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
			log.Fatal(err)
		}

		if user != currentUser {
			result = append(result, userInfo{user, []int32{}, 0})
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
	}

	return result
}
