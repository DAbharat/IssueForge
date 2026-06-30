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