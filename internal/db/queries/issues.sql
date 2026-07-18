-- name: CreateIssue :one
INSERT INTO issues(
    project_id,
    created_by,
    assigned_to,
    title,
    description,
    status,
    priority
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING id, project_id, created_by, assigned_to, title, description, status, priority, created_at, updated_at;


-- name: GetIssueByID :one
SELECT i.id, i.project_id, i.created_by, i.assigned_to, i.title, i.description, i.status, i.priority, i.created_at, i.updated_at, creator.display_name AS creator_name, assignee.display_name AS assignee_name
FROM issues i
JOIN users creator ON i.created_by = creator.id
LEFT JOIN users assignee ON i.assigned_to = assignee.id
WHERE i.id = $1;


-- name: ListProjectIssues :many
SELECT i.id, i.project_id, i.title, i.status, i.priority, i.created_at, i.assigned_to, i.created_by, 
       u_creator.display_name AS creator_name,
       u_assignee.display_name AS assignee_name
FROM issues i
INNER JOIN users u_creator ON i.created_by = u_creator.id
LEFT JOIN users u_assignee ON i.assigned_to = u_assignee.id
WHERE i.project_id = $1
ORDER BY i.created_at DESC;


-- name: UpdateIssueDetails :one
UPDATE issues
SET title = $2, description = $3
WHERE id = $1
RETURNING id, project_id, created_by, assigned_to, title, description, status, priority, created_at, updated_at;


-- name: UpdateIssueStatus :one
UPDATE issues
SET status = $2
WHERE id = $1
RETURNING id, project_id, created_by, assigned_to, title, description, status, priority, created_at, updated_at;


-- name: UpdateIssueAssignee :one
UPDATE issues
SET assigned_to = $2
WHERE id = $1
RETURNING id, project_id, created_by, assigned_to, title, description, status, priority, created_at, updated_at;


-- name: UpdateIssuePriority :one
UPDATE issues
SET priority = $2
WHERE id = $1
RETURNING id, project_id, created_by, assigned_to, title, description, status, priority, created_at, updated_at;


-- name: ListAssignedIssues :many
SELECT i.id, i.project_id, i.title, i.status, i.priority, i.created_at, p.name AS project_name
FROM issues i
JOIN projects p ON i.project_id = p.id
WHERE i.assigned_to = $1
ORDER BY i.created_at DESC;


-- name: ListCreatedIssues :many
SELECT i.id, i.project_id, i.title, i.status, i.priority, i.created_at, p.name AS project_name
FROM issues i
JOIN projects p ON i.project_id = p.id
WHERE i.created_by = $1
ORDER BY i.created_at DESC;


-- name: DeleteIssue :one
DELETE FROM issues
WHERE id = $1
RETURNING id;