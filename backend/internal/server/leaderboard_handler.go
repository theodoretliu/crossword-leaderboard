package server

import "context"

type LeaderboardEntry struct {
	ID            int64
	Name          string
	Percentile0   float64
	Percentile10  float64
	Percentile20  float64
	Percentile30  float64
	Percentile40  float64
	Percentile50  float64
	Percentile60  float64
	Percentile70  float64
	Percentile80  float64
	Percentile90  float64
	Percentile100 float64
}

func LeaderboardHandler() []LeaderboardEntry {
	rows, err := queries.GetLeaderboard(context.Background())
	if err != nil {
		panic(err)
	}

	entries := make([]LeaderboardEntry, len(rows))
	for i, r := range rows {
		entries[i] = LeaderboardEntry{
			ID:            int64(r.ID),
			Name:          deref(r.Name),
			Percentile0:   r.Percentile0,
			Percentile10:  r.Percentile10,
			Percentile20:  r.Percentile20,
			Percentile30:  r.Percentile30,
			Percentile40:  r.Percentile40,
			Percentile50:  r.Percentile50,
			Percentile60:  r.Percentile60,
			Percentile70:  r.Percentile70,
			Percentile80:  r.Percentile80,
			Percentile90:  r.Percentile90,
			Percentile100: r.Percentile100,
		}
	}
	return entries
}
