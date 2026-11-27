-- The production schema, transcribed from the live database's catalog on
-- 2026-11-28. sqlc reads this file to type-check db/queries; it is not a
-- migration and is not applied anywhere. Keep it in sync with production.
-- (The files in /migrations predate the current schema and do not match it.)

CREATE TABLE users (
    id          SERIAL PRIMARY KEY,
    nyt_user_id BIGINT,
    name        TEXT
);
CREATE UNIQUE INDEX users_nyt_user_id_idx ON users (nyt_user_id);

CREATE TABLE times (
    id              SERIAL PRIMARY KEY,
    time_in_seconds INTEGER NOT NULL,
    date            DATE NOT NULL,
    user_id         BIGINT NOT NULL
);
CREATE UNIQUE INDEX times_user_id_date_idx ON times (user_id, date);

CREATE TABLE cookies (
    id    SERIAL PRIMARY KEY,
    value TEXT NOT NULL
);
