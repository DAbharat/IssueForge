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

-- name: ListProjectsByOwner :many
SELECT * FROM projects
WHERE owner_id = $1
ORDER BY created_at DESC;