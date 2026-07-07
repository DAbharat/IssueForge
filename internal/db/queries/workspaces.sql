-- name: CreateWorkspace :one
INSERT INTO workspaces(
    name
)
VALUES (
    $1
)
RETURNING id, name, created_at, updated_at;

-- name: GetWorkspaceByID :one
SELECT id, name, created_at, updated_at
FROM workspaces
WHERE id = $1;

-- name: GetWorkspaceByName :one
SELECT id, name, created_at, updated_at
FROM workspaces
WHERE name = $1;