-- name: CreateLabel :one
INSERT INTO labels(
    project_id,
    name,
    color,
    created_by
) VALUES (
    sqlc.arg(project_id),sqlc.arg(name),sqlc.arg(color),sqlc.arg(created_by)
)
RETURNING id, project_id, name, color, created_by, created_at, updated_at, deleted_at;


-- name: GetLabelByID :one
SELECT id, project_id, name, color, created_by, created_at, updated_at, deleted_at
FROM labels
WHERE id = $1 AND deleted_at IS NULL;


-- name: ListProjectLabels :many
SELECT id, project_id, name, color, created_by, created_at, updated_at, deleted_at
FROM labels
WHERE project_id = $1 AND deleted_at IS NULL
ORDER BY name ASC;


-- name: UpdateLabel :one
UPDATE labels
SET name = COALESCE(sqlc.narg(name), name), 
    color = COALESCE(sqlc.narg(color), color)
WHERE id = @id AND deleted_at IS NULL
RETURNING id, project_id, name, color, created_by, created_at, updated_at, deleted_at;


-- name: DeleteLabel :one
UPDATE labels
SET deleted_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, project_id;


-- name: GetLabelProjectID :one
SELECT project_id
FROM labels
WHERE id = $1
AND deleted_at IS NULL;


-- name: CountProjectLabels :one
SELECT COUNT(*)
FROM labels
WHERE project_id = $1
AND id = ANY($2::bigint[])
AND deleted_at IS NULL;


-- name: AttachLabelsToIssue :exec
INSERT INTO issue_labels(
    issue_id,
    label_id
) SELECT $1, unnest(sqlc.arg(label_id)::bigint[])
ON CONFLICT (issue_id, label_id) DO NOTHING;


-- name: RemoveLabelFromIssue :exec
DELETE FROM issue_labels
WHERE issue_id = $1
AND label_id = $2;


-- name: ListIssueLabels :many
SELECT l.id, l.project_id, l.name, l.color, l.created_by, l.created_at, l.updated_at, l.deleted_at
FROM issue_labels il
JOIN labels l ON il.label_id = l.id
WHERE il.issue_id = $1
AND l.deleted_at IS NULL
ORDER BY l.name;