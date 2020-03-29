BEGIN;
CREATE INDEX times_idx ON times (user_id, date);
END;
