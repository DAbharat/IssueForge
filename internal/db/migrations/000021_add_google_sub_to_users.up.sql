ALTER TABLE users ADD COLUMN google_sub TEXT;

CREATE UNIQUE INDEX users_google_sub_key
    ON users (google_sub)
    WHERE google_sub IS NOT NULL AND deleted_at IS NULL;