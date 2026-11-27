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

-- InsertTimes records solve times (parallel arrays, any mix of dates), keeping
-- any time already stored for a user and date.
-- name: InsertTimes :exec
INSERT INTO times (user_id, time_in_seconds, date)
SELECT unnest(@user_ids::bigint[]), unnest(@times_in_seconds::integer[]), unnest(@dates::date[])
ON CONFLICT (user_id, date) DO NOTHING;

-- ListKnownUsers seeds the scraper's in-memory view of what is stored.
-- name: ListKnownUsers :many
SELECT nyt_user_id::bigint AS nyt_user_id, coalesce(name, '')::text AS name
FROM users
WHERE nyt_user_id IS NOT NULL;

-- ListStoredSolves seeds the scraper's in-memory view of what is stored.
-- name: ListStoredSolves :many
SELECT u.nyt_user_id::bigint AS nyt_user_id, t.date
FROM times AS t
JOIN users AS u ON u.id = t.user_id
WHERE u.nyt_user_id IS NOT NULL;
