package graph

import (
	"database/sql"
	"time"
)

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

func GetWeeksTimes(db *sql.DB, userId string, daysOfTheWeek []time.Time) ([]int, error) {
	rows, err := db.Query("SELECT time_in_seconds, date FROM times WHERE user_id = $1 AND date >= $2 AND date <= $3 ORDER BY date ASC;", userId, daysOfTheWeek[0], daysOfTheWeek[6])

	if err != nil {
		return nil, err
	}

	i := 0

	times := []int{}

	for rows.Next() {
		var time_ int
		var date time.Time

		err := rows.Scan(&time_, &date)

		if err != nil {
			return nil, err
		}

		for date.UTC() != daysOfTheWeek[i] {
			i++
			times = append(times, -1)
		}

		times = append(times, time_)
		i++
	}

	for len(times) != 7 {
		times = append(times, -1)
	}

	return times, nil
}

func GetWeeksWorstTimes(db *sql.DB) ([]int, error) {
	// cache, ok := c.Value("cache").(map[string]interface{})

	// if !ok {
	// 	return nil, errors.New("Could not retrieve cache from context")
	// }

	// weeksWorstTimesInt, found := cache["weeksWorstTimes"]

	// if found {
	// 	return weeksWorstTimesInt.([]int), nil
	// }

	daysOfTheWeek := GetDaysOfTheWeek()

	var worstTimes []int

	rows, err := db.Query("SELECT max(time_in_seconds), date FROM times WHERE date >= $1 AND date <= $2 GROUP BY date ORDER BY date ASC", daysOfTheWeek[0], daysOfTheWeek[6])

	if err != nil {
		return nil, err
	}

	i := 0

	for rows.Next() {
		var worstTime int
		var date time.Time

		err = rows.Scan(&worstTime, &date)

		if err != nil {
			return nil, err
		}

		for date.UTC() != daysOfTheWeek[i] {
			i++
			worstTimes = append(worstTimes, -1)
		}

		worstTimes = append(worstTimes, worstTime)
		i++
	}

	for len(worstTimes) != 7 {
		worstTimes = append(worstTimes, -1)
	}

	// cache["weeksWorstTimes"] = worstTimes

	return worstTimes, nil
}

func CreateNewWorstTimesLoader(db *sql.DB) *WeeksWorstTimesLoader {
	return NewWeeksWorstTimesLoader(WeeksWorstTimesLoaderConfig{
		Fetch: func(keys []interface{}) ([][]int, []error) {

			weeksWorstTimes, err := GetWeeksWorstTimes(db)

			var output [][]int
			var errors []error

			for range keys {
				output = append(output, weeksWorstTimes)
				errors = append(errors, err)
			}

			return output, errors
		},

		Wait: 1 * time.Millisecond,

		MaxBatch: 0,
	})
}

func CreateNewWeeksTimesLoader(db *sql.DB) *WeeksTimesLoader {
	return NewWeeksTimesLoader(WeeksTimesLoaderConfig{
		Fetch: func(keys []string) ([][]int, []error) {
			daysOfTheWeek := GetDaysOfTheWeek()

			output := [][]int{}
			errors := []error{}

			for _, key := range keys {
				times, err := GetWeeksTimes(db, key, daysOfTheWeek)

				output = append(output, times)
				errors = append(errors, err)
			}

			return output, errors
		},

		Wait: 1 * time.Millisecond,

		MaxBatch: 0,
	})
}
