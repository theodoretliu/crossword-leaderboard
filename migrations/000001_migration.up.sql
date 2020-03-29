BEGIN;
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name TEXT,
    username TEXT
);
CREATE TABLE times (
    id SERIAL PRIMARY KEY,
    user_id INTEGER,
    time_in_seconds INTEGER,
    date DATE
);
COMMIT;
