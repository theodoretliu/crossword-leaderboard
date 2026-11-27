-- name: GetCookie :one
SELECT value FROM cookies LIMIT 1;

-- UpsertUsers inserts or renames users in a single statement. Callers must pass
-- unique NYT ids (a repeated id fails with "cannot affect row a second time"),
-- sorted so that concurrent upserts lock rows in the same order. An empty name
-- never overwrites an existing one.
-- name: UpsertUsers :many
INSERT INTO users (nyt_user_id, name)
SELECT unnest(@nyt_user_ids::bigint[]), unnest(@names::text[])
ON CONFLICT (nyt_user_id) DO UPDATE SET name = COALESCE(NULLIF(EXCLUDED.name, ''), users.name)
RETURNING nyt_user_id, id;

-- InsertTimes records solve times for one date, keeping any time already stored.
-- name: InsertTimes :exec
INSERT INTO times (user_id, time_in_seconds, date)
SELECT unnest(@user_ids::bigint[]), unnest(@times_in_seconds::integer[]), @date::date
ON CONFLICT (user_id, date) DO NOTHING;
