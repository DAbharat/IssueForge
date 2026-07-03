-- name: CreateUser :one
INSERT INTO users(
    username,
    fullname,
    email,
    password_hash
)
VALUES(
    $1, $2, $3, $4
)
RETURNING *;

-- name: GetUserByEmail :one
SELECT id, username, fullname, email, created_at, updated_at
FROM users
WHERE email = $1
LIMIT 1;

-- name: GetUserForLogin :one
SELECT id, password_hash
FROM users
WHERE email = $1
   OR username = $1
LIMIT 1;