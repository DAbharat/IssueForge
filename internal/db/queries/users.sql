-- name: CreateOnboardingUser :one
INSERT INTO users(
    email,
    fullname,
    username,
    password_hash
)
VALUES(
    $1, $2, $3, $4
)
RETURNING id, email, fullname, username, created_at;


-- name: GetUserForLogin :one
SELECT id, email, username, fullname, password_hash
FROM users
WHERE email = $1 AND deleted_at IS NULL
LIMIT 1;


-- name: GetUserByID :one
SELECT id, username, fullname, email, created_at, updated_at
FROM users
WHERE id = $1 AND deleted_at IS NULL
LIMIT 1;


-- name: SearchUserByUsername :many
SELECT id, username, fullname
FROM users
WHERE username ILIKE '%' || sqlc.arg(search) || '%'
  AND deleted_at IS NULL
ORDER BY username
LIMIT 10;


-- name: GetUserByUsername :one
SELECT id, username, fullname, email, created_at, updated_at
FROM users
WHERE username = sqlc.arg(username) AND deleted_at IS NULL
LIMIT 1;


-- name: DeleteUser :one
UPDATE users
SET deleted_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING id;


-- name: GetUserByGoogleSub :one
SELECT id, email, username, fullname, password_hash
FROM users
WHERE google_sub = $1 AND deleted_at IS NULL
LIMIT 1;


-- name: CreateOIDCUser :one
INSERT INTO users(
  email, fullname, username, password_hash, google_sub
) VALUES (
  $1, $2, $3, $4, $5
)
RETURNING id, email, fullname, username, created_at;


-- name: BindGoogleSub :one
UPDATE users
SET google_sub = $2
WHERE id = $1 AND deleted_at IS NULL
RETURNING id;