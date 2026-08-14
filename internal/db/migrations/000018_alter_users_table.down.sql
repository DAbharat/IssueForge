ALTER TABLE users
DROP COLUMN deleted_at;

ALTER TABLE users
DROP CONSTRAINT users_username_unique;

ALTER TABLE users
RENAME COLUMN username TO display_name;