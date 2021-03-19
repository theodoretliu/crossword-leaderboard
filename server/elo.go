package main

import (
	"database/sql"
	"math"
	"time"
)

var k float64 = 30

var ops int64 = 0

func getAllDatesBeforeDate(db *sql.DB, date time.Time) ([]time.Time, error) {
	rows, err := db.Query("SELECT DISTINCT(date) FROM times WHERE date <= date(?) ORDER BY date;", date)

	if err != nil {
		return []time.Time{}, err
	}

	dates := []time.Time{}
	for rows.Next() {
		var date time.Time

		err = rows.Scan(&date)

		if err != nil {
			return []time.Time{}, err
		}

		dates = append(dates, date)
	}

	return dates, nil
}

func eloUpdate(p1rating, p2rating, p1actual, p2actual float64) (float64, float64) {
	p1 := (1.0 / (1.0 + math.Pow(10, (p2rating-p1rating)/400.0)))
	p2 := 1.0 - p1

	// fmt.Println(p1, p2)
	return p1rating + k*(p1actual-p1), p2rating + k*(p2actual-p2)
}

type userAndTime struct {
	username      string
	timeInSeconds int
}

func getUsersForDate(db *sql.DB, date time.Time) ([]userAndTime, error) {
	rows, err := db.Query(`SELECT users.username, times.time_in_seconds
		FROM times JOIN users ON users.id = times.user_id
		WHERE times.date = date(?)
		ORDER BY times.time_in_seconds ASC`, date)

	if err != nil {
		return []userAndTime{}, err
	}

	out := []userAndTime{}

	for rows.Next() {
		var info userAndTime

		err = rows.Scan(&info.username, &info.timeInSeconds)

		if err != nil {
			return []userAndTime{}, err
		}

		out = append(out, info)
	}

	return out, nil
}

func getAllUsernames(db *sql.DB) ([]string, error) {
	rows, err := db.Query("SELECT username FROM users")

	if err != nil {
		return nil, err
	}

	ids := []string{}
	for rows.Next() {
		var username string

		err = rows.Scan(&username)

		if err != nil {
			return nil, err
		}

		ids = append(ids, username)
	}

	return ids, nil
}

func computeElo(db *sql.DB, date time.Time) (map[string]float64, error) {
	allDates, err := getAllDatesBeforeDate(db, date)

	if err != nil {
		return nil, err
	}

	allUsernames, err := getAllUsernames(db)

	if err != nil {
		return nil, err
	}

	usernameElos := map[string]float64{}

	for _, ids := range allUsernames {
		usernameElos[ids] = 1000.0
	}

	for _, date := range allDates {
		arr, err := getUsersForDate(db, date)

		if err != nil {
			return nil, err
		}

		for i := 0; i < len(arr); i++ {
			for j := i + 1; j < len(arr); j++ {
				user := arr[i]
				opponent := arr[j]
				ops += 1
				var userActual float64
				var opponentActual float64

				if user.timeInSeconds < opponent.timeInSeconds {
					userActual = 1.0
					opponentActual = 0.0
				} else if user.timeInSeconds > opponent.timeInSeconds {
					userActual = 0.0
					opponentActual = 1.0
				} else {
					userActual = 0.5
					opponentActual = 0.5
				}

				userUpdate, opponentUpdate := eloUpdate(usernameElos[user.username], usernameElos[opponent.username], userActual, opponentActual)

				usernameElos[user.username] = userUpdate
				usernameElos[opponent.username] = opponentUpdate
			}

		}
	}

	return usernameElos, nil
}

func setElosInDb(db *sql.DB) error {
	today := time.Now().UTC().Truncate(24 * time.Hour)

	elos, err := computeElo(db, today)

	if err != nil {
		return err
	}

	for username, elo := range elos {
		_, err = db.Exec("UPDATE users SET elo = ? WHERE username = ?", elo, username)

		if err != nil {
			return err
		}
	}

	return nil
}
