package main

import (
	"context"

	"github.com/jackc/pgx/v5"
)

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

var leaderboardQuery string = `
WITH average_times AS (
    SELECT
    user_id,
    CASE 
        WHEN EXTRACT(DOW FROM date) = 6 THEN (time_in_seconds / 49.0) * 25.0
        ELSE time_in_seconds
    END AS normalized_time
    FROM times
), user_percentiles AS (
    SELECT 
    user_id,
    PERCENTILE_CONT(1) WITHIN GROUP (ORDER BY normalized_time) AS percentile_0,
    PERCENTILE_CONT(0.9) WITHIN GROUP (ORDER BY normalized_time) AS percentile_10,
    PERCENTILE_CONT(0.8) WITHIN GROUP (ORDER BY normalized_time) AS percentile_20,
    PERCENTILE_CONT(0.7) WITHIN GROUP (ORDER BY normalized_time) AS percentile_30,
    PERCENTILE_CONT(0.6) WITHIN GROUP (ORDER BY normalized_time) AS percentile_40,
    PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY normalized_time) AS percentile_50,
    PERCENTILE_CONT(0.4) WITHIN GROUP (ORDER BY normalized_time) AS percentile_60,
    PERCENTILE_CONT(0.3) WITHIN GROUP (ORDER BY normalized_time) AS percentile_70,
    PERCENTILE_CONT(0.2) WITHIN GROUP (ORDER BY normalized_time) AS percentile_80,
    PERCENTILE_CONT(0.1) WITHIN GROUP (ORDER BY normalized_time) AS percentile_90,
    PERCENTILE_CONT(0) WITHIN GROUP (ORDER BY normalized_time) AS percentile_100
    FROM average_times
    WHERE normalized_time > 2
    GROUP BY user_id
), all_percentiles AS (
    SELECT 
    -1 AS user_id,
    'Everyone' AS name,
    PERCENTILE_CONT(1) WITHIN GROUP (ORDER BY normalized_time) AS percentile_0,
    PERCENTILE_CONT(0.9) WITHIN GROUP (ORDER BY normalized_time) AS percentile_10,
    PERCENTILE_CONT(0.8) WITHIN GROUP (ORDER BY normalized_time) AS percentile_20,
    PERCENTILE_CONT(0.7) WITHIN GROUP (ORDER BY normalized_time) AS percentile_30,
    PERCENTILE_CONT(0.6) WITHIN GROUP (ORDER BY normalized_time) AS percentile_40,
    PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY normalized_time) AS percentile_50,
    PERCENTILE_CONT(0.4) WITHIN GROUP (ORDER BY normalized_time) AS percentile_60,
    PERCENTILE_CONT(0.3) WITHIN GROUP (ORDER BY normalized_time) AS percentile_70,
    PERCENTILE_CONT(0.2) WITHIN GROUP (ORDER BY normalized_time) AS percentile_80,
    PERCENTILE_CONT(0.1) WITHIN GROUP (ORDER BY normalized_time) AS percentile_90,
    PERCENTILE_CONT(0) WITHIN GROUP (ORDER BY normalized_time) AS percentile_100
    FROM average_times
    WHERE normalized_time > 2
)
SELECT 
    users.id,
    users.name,
    user_percentiles.percentile_0,
    user_percentiles.percentile_10,
    user_percentiles.percentile_20,
    user_percentiles.percentile_30,
    user_percentiles.percentile_40,
    user_percentiles.percentile_50,
    user_percentiles.percentile_60,
    user_percentiles.percentile_70,
    user_percentiles.percentile_80,
    user_percentiles.percentile_90,
    user_percentiles.percentile_100
FROM user_percentiles
JOIN users ON user_percentiles.user_id = users.id
UNION ALL
SELECT * FROM all_percentiles
ORDER BY id ASC;
`

func LeaderboardHandler() []LeaderboardEntry {
	rows, err := pool.Query(context.Background(), leaderboardQuery)
	if err != nil {
		panic(err)
	}

	resRows, err := pgx.CollectRows(rows, pgx.RowToStructByPos[LeaderboardEntry])

	if err != nil {
		panic(err)
	}

	return resRows
}
