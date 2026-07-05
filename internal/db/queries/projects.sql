-- name: CreateProject :one
INSERT INTO projects(
    owner_id,
    name,
    description
)
VALUES(
    $1, $2, $3
)
RETURNING *;