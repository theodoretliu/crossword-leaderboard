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
