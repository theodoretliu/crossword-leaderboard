-- name: ListUsers :many
SELECT id, name FROM users;

-- name: GetUserName :one
SELECT name FROM users WHERE id = $1;

-- name: ListUserTimes :many
SELECT time_in_seconds, date FROM times
WHERE user_id = $1
ORDER BY date;

-- name: ListTimesBetween :many
SELECT u.id, u.name, t.time_in_seconds, t.date
FROM users AS u
JOIN times AS t ON t.user_id = u.id
WHERE t.date >= sqlc.arg(start_date) AND t.date <= sqlc.arg(end_date)
ORDER BY u.id, t.date;

-- name: ListWorstTimesBetween :many
SELECT max(time_in_seconds)::integer AS worst_time, date
FROM times
WHERE date >= sqlc.arg(start_date) AND date <= sqlc.arg(end_date)
GROUP BY date
ORDER BY date;

-- GetDayOfWeekSpeed returns, per ISO weekday (1 = Monday, 7 = Sunday), the
-- user's median solve time and where it falls among every other player's solves
-- on that weekday: percentile is the share of those solves slower than the
-- median (ties count half), 50 when there are none. Times under 3 seconds are
-- skipped.
-- name: GetDayOfWeekSpeed :many
WITH valid AS (
    SELECT user_id, EXTRACT(ISODOW FROM date)::integer AS weekday, time_in_seconds
    FROM times
    WHERE time_in_seconds >= 3
), mine AS (
    SELECT
        weekday,
        PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY time_in_seconds) AS median_time,
        COUNT(*) AS num_solves
    FROM valid
    WHERE user_id = sqlc.arg(user_id)::bigint
    GROUP BY weekday
)
SELECT
    mine.weekday::integer AS weekday,
    mine.median_time::float8 AS median_time,
    mine.num_solves::integer AS num_solves,
    COALESCE(
        100 * (
            COUNT(*) FILTER (WHERE others.time_in_seconds > mine.median_time)
            + 0.5 * COUNT(*) FILTER (WHERE others.time_in_seconds = mine.median_time)
        ) / NULLIF(COUNT(others.user_id), 0),
        50
    )::float8 AS percentile
FROM mine
LEFT JOIN valid AS others
    ON others.weekday = mine.weekday
    AND others.user_id <> sqlc.arg(user_id)::bigint
GROUP BY 1, 2, 3
ORDER BY 1;
