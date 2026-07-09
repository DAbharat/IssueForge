-- name: CreateOnboardingUser :one
INSERT INTO users(
    email,
    fullname,
    display_name,
    password_hash
)
VALUES(
    $1, $2, $3, $4
)
RETURNING id, email, fullname, display_name, created_at;


-- name: GetUserForLogin :one
SELECT id, email, display_name, fullname, password_hash
FROM users
WHERE email = $1
LIMIT 1;


-- name: GetUserByID :one
SELECT id, display_name, fullname, email, created_at, updated_at
FROM users
WHERE id=$1
LIMIT 1;