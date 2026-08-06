-- name: CreateProject :one
INSERT INTO projects(
    workspace_id,
    lead_id,
    name,
    description
)
VALUES(
    $1, $2, $3, $4
)
RETURNING id, workspace_id, lead_id, name, description, created_at, updated_at, deleted_at;


-- name: GetProjectByID :one
SELECT id, workspace_id, lead_id, name, description, created_at, updated_at
FROM projects
WHERE id = $1 AND deleted_at IS NULL;


-- name: UpdateProjectDetails :one
UPDATE projects
SET name = COALESCE(sqlc.narg(name), name),
    description = COALESCE(sqlc.narg(description), description)
WHERE id = @id AND deleted_at IS NULL
RETURNING id, workspace_id, lead_id, name, description, created_at, updated_at, deleted_at;


-- name: UpdateProjectLead :one
UPDATE projects
SET lead_id = COALESCE(sqlc.narg(lead_id), lead_id)
WHERE id = @id AND deleted_at IS NULL
RETURNING id, workspace_id, lead_id, name, description, created_at, updated_at, deleted_at;


-- name: ListProjectsByLead :many
SELECT id, workspace_id, lead_id, name, description, created_at, updated_at
FROM projects
WHERE workspace_id = $1 AND deleted_at IS NULL
    AND (sqlc.narg(lead_id)::BIGINT IS NULL OR lead_id = sqlc.narg(lead_id)::BIGINT)
ORDER BY created_at DESC;


-- name: DeleteProject :one
UPDATE projects
SET deleted_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, workspace_id, lead_id;


-- name: IsProjectLead :one
SELECT EXISTS (
    SELECT 1
    FROM projects
    WHERE id = $1 AND lead_id = $2 AND deleted_at IS NULL
);