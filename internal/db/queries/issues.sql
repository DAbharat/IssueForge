-- name: CreateIssue :one
INSERT INTO issues(
    project_id,
    created_by,
    assigned_to,
    title,
    description,
    status,
    priority,
    due_date
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING id, project_id, created_by, assigned_to, title, description, status, priority, created_at, updated_at, deleted_at, due_date;


-- name: GetIssueByID :one
SELECT i.id, i.project_id, i.created_by, i.assigned_to, i.title, i.description, i.status, i.priority, i.created_at, i.updated_at, i.due_date, creator.display_name AS creator_name, assignee.display_name AS assignee_name, deleted_at
FROM issues i
JOIN users creator ON i.created_by = creator.id
LEFT JOIN users assignee ON i.assigned_to = assignee.id
WHERE i.id = $1 AND i.deleted_at IS NULL;


-- name: ListProjectIssues :many
SELECT i.id, i.project_id, i.title, i.status, i.priority, i.created_at, i.assigned_to, i.created_by, i.deleted_at, i.due_date,
       u_creator.display_name AS creator_name,
       u_assignee.display_name AS assignee_name
FROM issues i
INNER JOIN users u_creator ON i.created_by = u_creator.id
LEFT JOIN users u_assignee ON i.assigned_to = u_assignee.id
WHERE i.project_id = $1 AND i.deleted_at IS NULL
    AND (sqlc.narg(status)::issue_status IS NULL OR i.status = sqlc.narg(status))
    AND (sqlc.narg(priority)::issue_priority IS NULL OR i.priority = sqlc.narg(priority))
    AND (sqlc.narg(assigned_to)::bigint IS NULL OR i.assigned_to = sqlc.narg(assigned_to))
    AND (sqlc.narg(search)::text IS NULL OR i.title ILIKE '%' || sqlc.narg(search) || '%')
ORDER BY i.created_at DESC
LIMIT sqlc.arg(page_limit)
OFFSET sqlc.arg(page_offset);


-- name: UpdateIssueDetails :one
UPDATE issues
SET title = $2, description = $3
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, project_id, created_by, assigned_to, title, description, status, priority, created_at, updated_at, deleted_at, due_date;


-- name: UpdateIssueStatus :one
UPDATE issues
SET status = $2
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, project_id, created_by, assigned_to, title, description, status, priority, created_at, updated_at, deleted_at, due_date;


-- name: UpdateIssueAssignee :one
UPDATE issues
SET assigned_to = $2
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, project_id, created_by, assigned_to, title, description, status, priority, created_at, updated_at, deleted_at, due_date;


-- name: UpdateIssuePriority :one
UPDATE issues
SET priority = $2
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, project_id, created_by, assigned_to, title, description, status, priority, created_at, updated_at, deleted_at, due_date;


-- name: UpdateIssueDueDate :one
UPDATE issues
SET due_date = $2,
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, project_id, created_by, assigned_to, title, description, status, priority, created_at, updated_at, deleted_at, due_date;


-- name: ListAssignedIssues :many
SELECT i.id, i.project_id, i.title, i.status, i.priority, i.created_at, i.deleted_at, i.due_date, p.name AS project_name
FROM issues i
JOIN projects p ON i.project_id = p.id
WHERE i.assigned_to = $1 AND i.deleted_at IS NULL
ORDER BY i.created_at DESC;


-- name: ListCreatedIssues :many
SELECT i.id, i.project_id, i.title, i.status, i.priority, i.created_at, i.deleted_at, i.due_date, p.name AS project_name
FROM issues i
JOIN projects p ON i.project_id = p.id
WHERE i.created_by = $1 AND i.deleted_at IS NULL
ORDER BY i.created_at DESC;


-- name: DeleteIssue :one
UPDATE issues
SET deleted_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING id;


-- name: GetIssueProjectID :one
SELECT project_id
FROM issues
WHERE id = $1 AND deleted_at IS NULL;


-- name: RestoreIssue :one
UPDATE issues
SET deleted_at = NULL
WHERE id = $1 AND deleted_at IS NOT NULL
RETURNING id;