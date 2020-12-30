package main

import (
	"database/sql"
	"log"
	"math"
	"time"
)

type userInfo struct {
	Username     string
	WeeksTimes   []int32
	WeeksAverage int32
}

type indexResponse struct {
	Users         []userInfo
	DaysOfTheWeek []string
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

func NewIndexHandler() indexResponse {
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

	weeksWorstTimes := GetWeeksWorstTimes(daysOfTheWeek)

	for i := 0; i < len(result); i++ {
		result[i].WeeksAverage = WeeklyAverage(result[i].WeeksTimes, weeksWorstTimes)
	}

	daysOfTheWeekStrings := []string{}

	for _, day := range daysOfTheWeek {
		daysOfTheWeekStrings = append(daysOfTheWeekStrings, day.Format(time.RFC1123Z))

	}
	return indexResponse{Users: result, DaysOfTheWeek: daysOfTheWeekStrings}
}
