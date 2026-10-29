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


-- name: GetWorkspaceDailyActivityStats :many
WITH all_workspace_events AS (
    SELECT ia.created_at::date AS event_date, p.id AS project_id
    FROM issue_activities ia
    JOIN issues i ON ia.issue_id = i.id
    JOIN projects p ON i.project_id = p.id
    WHERE p.workspace_id = sqlc.arg(workspace_id)
      AND p.deleted_at IS NULL
      AND ia.created_at >= sqlc.arg(start_date)::timestamptz 
      AND ia.created_at <= sqlc.arg(end_date)::timestamptz

    UNION ALL

    SELECT pa.created_at::date AS event_date, p.id AS project_id
    FROM project_activities pa
    JOIN projects p ON pa.project_id = p.id
    WHERE p.workspace_id = sqlc.arg(workspace_id)
      AND p.deleted_at IS NULL
      AND pa.created_at >= sqlc.arg(start_date)::timestamptz 
      AND pa.created_at <= sqlc.arg(end_date)::timestamptz
)
SELECT event_date, COUNT(*)::BIGINT AS total_activities
FROM all_workspace_events
GROUP BY event_date
ORDER BY event_date ASC;


-- name: GetWorkspaceIssueVelocityStats :many
SELECT d.date_bucket::date AS date,
COUNT(DISTINCT i.id) FILTER (WHERE i.created_at::date = d.date_bucket::date) AS issues_created,
COUNT(DISTINCT ia.id) FILTER (
    WHERE ia.activity_type = 'ISSUE_STATUS_CHANGED'
    AND ia.new_value = 'DONE'
    AND ia.created_at::date = d.date_bucket::date
) AS issues_completed
FROM generate_series(sqlc.arg(start_date)::timestamptz, sqlc.arg(end_date)::timestamptz, '1 day'::interval) d(date_bucket)
LEFT JOIN (
    SELECT iss.id, iss.created_at
    FROM issues iss
    JOIN projects p ON iss.project_id = p.id
    WHERE p.workspace_id = sqlc.arg(workspace_id)
      AND p.deleted_at IS NULL
      AND iss.deleted_at IS NULL
) i ON i.created_at::date = d.date_bucket::date
LEFT JOIN (
    SELECT act.id, act.activity_type, act.new_value, act.created_at
    FROM issue_activities act
    JOIN issues iss ON act.issue_id = iss.id
    JOIN projects p ON iss.project_id = p.id
    WHERE p.workspace_id = sqlc.arg(workspace_id)
      AND p.deleted_at IS NULL
) ia ON ia.created_at::date = d.date_bucket::date
GROUP BY d.date_bucket
ORDER BY d.date_bucket ASC;


-- name: GetWorkspaceActivityTypeBreakdown :many
SELECT activity_type::text, COUNT(*)::BIGINT AS count
FROM (
    SELECT ia.activity_type::text AS activity_type
    FROM issue_activities ia
    JOIN issues i ON ia.issue_id = i.id
    JOIN projects p ON i.project_id = p.id
    WHERE p.workspace_id = sqlc.arg(workspace_id) 
      AND p.deleted_at IS NULL
      AND ia.created_at >= sqlc.arg(start_date)::timestamptz

    UNION ALL

    SELECT pa.activity_type::text AS activity_type
    FROM project_activities pa
    JOIN projects p ON pa.project_id = p.id
    WHERE p.workspace_id = sqlc.arg(workspace_id) 
      AND p.deleted_at IS NULL
      AND pa.created_at >= sqlc.arg(start_date)::timestamptz
) events
GROUP BY activity_type
ORDER BY count DESC;