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
RETURNING id, workspace_id, lead_id, name, description, created_at;


-- name: ListProjectsByWorkspace :many
SELECT id, workspace_id, lead_id, name, description, created_at, updated_at
FROM projects
WHERE workspace_id = $1
ORDER BY created_at DESC;


-- name: ListProjectsByLead :many
SELECT id, workspace_id, lead_id, name, description, created_at, updated_at
FROM projects
WHERE lead_id = $1
ORDER BY created_at DESC;