ALTER TABLE users
RENAME COLUMN display_name TO username;

ALTER TABLE users
ADD CONSTRAINT users_username_unique UNIQUE (username);

ALTER TABLE users
ADD COLUMN deleted_at TIMESTAMPTZ NULL;