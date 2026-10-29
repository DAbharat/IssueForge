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
SELECT p.id, p.workspace_id, p.lead_id, p.name, p.description, p.created_at, p.updated_at,
    u.fullname as lead_name
FROM projects p
JOIN users u ON p.lead_id = u.id
WHERE p.id = $1 AND p.deleted_at IS NULL;


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
SELECT p.id, p.workspace_id, p.lead_id, p.name, p.description, p.created_at, p.updated_at,
        u.fullname as lead_name
FROM projects p
JOIN users u ON p.lead_id = u.id
WHERE p.workspace_id = $1 AND p.deleted_at IS NULL
    AND (sqlc.narg(lead_id)::BIGINT IS NULL OR p.lead_id = sqlc.narg(lead_id)::BIGINT)
ORDER BY p.created_at DESC;


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


-- name: GetProjectByIDIncludingDeleted :one
SELECT p.id, p.workspace_id, p.lead_id, p.name, p.description, p.created_at, p.updated_at, p.deleted_at,
    u.fullname as lead_name, w.deleted_at as workspace_deleted_at
FROM projects p
JOIN users u ON p.lead_id = u.id
JOIN workspaces w ON p.workspace_id = w.id
WHERE p.id = $1;


-- name: RestoreProject :one
UPDATE projects
SET deleted_at = NULL
WHERE id = $1 AND deleted_at IS NOT NULL
RETURNING id, workspace_id, lead_id, name, description, created_at, updated_at, deleted_at;


-- name: GetProjectDailyActivityStats :many
WITH all_project_events AS (
    SELECT ia.created_at::date AS event_date, ia.activity_type::text AS event_type
    FROM issue_activities ia
    JOIN issues i ON ia.issue_id = i.id
    WHERE i.project_id = sqlc.arg(project_id)
      AND ia.created_at >= sqlc.arg(start_date)::timestamptz 
      AND ia.created_at <= sqlc.arg(end_date)::timestamptz

    UNION ALL

    SELECT pa.created_at::date AS event_date, pa.activity_type::text AS event_type
    FROM project_activities pa
    WHERE pa.project_id = sqlc.arg(project_id)
      AND pa.created_at >= sqlc.arg(start_date)::timestamptz 
      AND pa.created_at <= sqlc.arg(end_date)::timestamptz
)
SELECT event_date, COUNT(*)::BIGINT AS total_activities
FROM all_project_events
GROUP BY event_date
ORDER BY event_date ASC;


-- name: GetProjectIssueVelocityStats :many
SELECT d.date_bucket::date AS date,
COUNT(DISTINCT i.id) FILTER (WHERE i.created_at::date = d.date_bucket::date) AS issues_created,
COUNT(DISTINCT ia.id) FILTER (
    WHERE ia.activity_type = 'ISSUE_STATUS_CHANGED'
    AND ia.new_value = 'DONE'
    AND ia.created_at::date = d.date_bucket::date
) AS issues_completed
FROM generate_series(sqlc.arg(start_date)::timestamptz, sqlc.arg(end_date)::timestamptz, '1 day'::interval) d(date_bucket)
LEFT JOIN issues i
ON i.project_id = sqlc.arg(project_id)
AND i.deleted_at IS NULL
AND i.created_at::date = d.date_bucket::date
LEFT JOIN issue_activities ia
ON ia.created_at::date = d.date_bucket::date
AND ia.issue_id IN (SELECT id FROM issues WHERE project_id = sqlc.arg(project_id))
GROUP BY d.date_bucket
ORDER BY d.date_bucket ASC;


-- name: GetProjectActivityTypeBreakdown :many
SELECT activity_type::text, COUNT(*)::BIGINT AS count
FROM (
    SELECT ia.activity_type::text AS activity_type
    FROM issue_activities ia
    JOIN issues i ON ia.issue_id = i.id
    WHERE i.project_id = sqlc.arg(project_id)
      AND ia.created_at >= sqlc.arg(start_date)::timestamptz

    UNION ALL

    SELECT pa.activity_type::text AS activity_type
    FROM project_activities pa
    WHERE pa.project_id = sqlc.arg(project_id)
      AND pa.created_at >= sqlc.arg(start_date)::timestamptz
) events
GROUP BY activity_type
ORDER BY count DESC;