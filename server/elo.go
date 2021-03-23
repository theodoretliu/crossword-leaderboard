package main

import (
	"math"
	"time"
)

var k float64 = 30

var ops int64 = 0

func getAllDatesBeforeDate(date time.Time) ([]time.Time, error) {
	rows, err := db.Query("SELECT DISTINCT(date) FROM times WHERE date(date) <= date(?) ORDER BY date(date);", date)

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
	userId        int64
	username      string
	timeInSeconds int64
}

func getUsersForDate(date time.Time) ([]userAndTime, error) {
	rows, err := db.Query(`SELECT users.id, users.username, times.time_in_seconds
		FROM times JOIN users ON users.id = times.user_id
		WHERE date(times.date) = date(?)
		ORDER BY times.time_in_seconds ASC`, date)

	if err != nil {
		return []userAndTime{}, err
	}

	out := []userAndTime{}

	for rows.Next() {
		var info userAndTime

		err = rows.Scan(&info.userId, &info.username, &info.timeInSeconds)

		if err != nil {
			return []userAndTime{}, err
		}

		out = append(out, info)
	}

	return out, nil
}

func getAllUserIds() ([]int64, error) {
	rows, err := db.Query("SELECT DISTINCT id FROM users")

	if err != nil {
		return nil, err
	}

	ids := []int64{}
	for rows.Next() {
		var userId int64

		err = rows.Scan(&userId)

		if err != nil {
			return nil, err
		}

		ids = append(ids, userId)
	}

	return ids, nil
}

func computeElo() error {
	allDates, err := getAllDatesBeforeDate(time.Now().UTC().Truncate(24 * time.Hour))

	if err != nil {
		return err
	}

	allUserIds, err := getAllUserIds()

	if err != nil {
		return err
	}

	userIdElos := map[int64]float64{}

	for _, ids := range allUserIds {
		userIdElos[ids] = 1000.0
	}

	for _, date := range allDates {
		arr, err := getUsersForDate(date)

		if err != nil {
			return err
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

				userUpdate, opponentUpdate := eloUpdate(userIdElos[user.userId], userIdElos[opponent.userId], userActual, opponentActual)

				userIdElos[user.userId] = userUpdate
				userIdElos[opponent.userId] = opponentUpdate
			}
		}

		for userId, elo := range userIdElos {
			_, err = db.Exec(`INSERT INTO elos (user_id, elo, date)
				VALUES ($1, $2, date($3))
				ON CONFLICT (user_id, date)
					DO UPDATE SET elo = $2;`, userId, elo, date)

			if err != nil {
				return err
			}
		}
	}

	return nil
}

func setElosInDb() error {
	return computeElo()
}

func getEloForUserIdDate(userId int64, date time.Time) (float64, error) {
	row := db.QueryRow(`SELECT elo FROM elos
		WHERE user_id = ? AND date(date) <= date(?)
		ORDER BY date(date) DESC
		LIMIT 1`, userId, date)

	var elo float64

	err := row.Scan(&elo)
	if err != nil {
		return 0.0, err
	}

	return elo, nil
}

func getEloHistory(userId int64) ([]dateElo, error) {
	rows, err := db.Query(`SELECT date, elo FROM elos WHERE user_id = ? ORDER BY date(date) DESC;`, userId)
	if err != nil {
		return nil, err
	}

	dateElos := []dateElo{}

	for rows.Next() {
		scan := dateElo{}

		err = rows.Scan(&scan.Date, &scan.Elo)
		if err != nil {
			return nil, err
		}

		dateElos = append(dateElos, scan)
	}

	return dateElos, nil
}

func getElosForDate(date time.Time) (map[int64]float64, error) {
	rows, err := db.Query(`SELECT user_id, elo FROM elos WHERE date(date) = date(?);`, date)
	if err != nil {
		return nil, err
	}

	elos := map[int64]float64{}

	for rows.Next() {
		var userId int64
		var elo float64

		err = rows.Scan(&userId, &elo)
		if err != nil {
			return nil, err
		}

		elos[userId] = elo
	}

	return elos, nil
}

func getPeakElo(userId int64) (float64, error) {
	row := db.QueryRow(`SELECT MAX(elo) FROM elos WHERE user_id = ?`, userId)
	var maxElo float64

	err := row.Scan(&maxElo)
	if err != nil {
		return 0, err
	}

	return maxElo, nil
}

func getCurrentElo(userId int64) (float64, error) {
	return getEloForUserIdDate(userId, time.Now())
}
