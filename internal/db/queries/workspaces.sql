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
WHERE id = $1 AND deleted_at IS NULL;


-- name: GetWorkspaceByName :one
SELECT id, name, created_at, updated_at
FROM workspaces
WHERE name = $1 AND deleted_at IS NULL;


-- name: UpdateWorkspaceName :one
UPDATE workspaces
SET name = sqlc.arg(name)
WHERE id = @id AND deleted_at IS NULL
RETURNING id, name, created_at, updated_at;


-- name: DeleteWorkspace :one
UPDATE workspaces
SET deleted_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING id;


-- name: RestoreDeletedWorkspace :one
UPDATE workspaces
SET deleted_at = NULL
WHERE id = $1 AND deleted_at IS NOT NULL
RETURNING id, name, created_at, updated_at;