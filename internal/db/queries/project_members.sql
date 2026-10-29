-- name: AddMemberToProject :one
INSERT INTO project_members(
    project_id,
    user_id
)
VALUES(
    $1, $2
)
RETURNING project_id, user_id, joined_at;


-- name: SafeAddMemberToProject :one
INSERT INTO project_members (
    project_id, 
    user_id
)
SELECT $1, $2
WHERE EXISTS(
    SELECT 1 FROM projects p
    WHERE p.id =  $1 AND p.lead_id = $3
)
AND EXISTS(
    SELECT 1 FROM workspace_members wm
    WHERE wm.user_id = $2
    AND wm.workspace_id = (
        SELECT workspace_id
        FROM projects
        WHERE id = $1
    )
)
RETURNING project_id, user_id, joined_at;


-- name: ListProjectMembers :many
SELECT u.id, u.email, u.fullname, u.fullname, u.username, pm.joined_at
FROM project_members pm
JOIN users u ON pm.user_id = u.id
JOIN projects p ON pm.project_id = p.id
JOIN workspace_members wm ON wm.workspace_id = p.workspace_id
AND wm.user_id = pm.user_id
WHERE pm.project_id = $1
ORDER BY pm.joined_at ASC;


-- name: IsProjectMember :one
SELECT EXISTS (
    SELECT 1
    FROM project_members pm
    JOIN projects p ON pm.project_id = p.id
    JOIN workspace_members wm ON wm.workspace_id = p.workspace_id
    AND wm.user_id = pm.user_id
    WHERE pm.project_id = $1 AND pm.user_id = $2
);


-- name: GetProjectMemberActivityStats :many
SELECT 
    u.id AS user_id,
    u.fullname,
    u.username,
    u.email,
    COUNT(ia.id)::BIGINT AS issue_actions_count,
    COUNT(DISTINCT i.id) FILTER (WHERE i.status = 'DONE')::BIGINT AS completed_issues_count
FROM project_members pm
JOIN users u ON pm.user_id = u.id
JOIN projects p ON pm.project_id = p.id
LEFT JOIN (
    SELECT act.id, act.actor_id
    FROM issue_activities act
    JOIN issues iss ON act.issue_id = iss.id
    WHERE iss.project_id = sqlc.arg(project_id)
      AND iss.deleted_at IS NULL
      AND act.created_at >= sqlc.arg(start_date)::timestamptz
) ia ON ia.actor_id = u.id
LEFT JOIN (
    SELECT iss.id, iss.assigned_to, iss.status
    FROM issues iss
    WHERE iss.project_id = sqlc.arg(project_id)
      AND iss.deleted_at IS NULL
) i ON i.assigned_to = u.id
WHERE pm.project_id = sqlc.arg(project_id)
  AND p.deleted_at IS NULL
GROUP BY u.id, u.fullname, u.username, u.email
ORDER BY issue_actions_count DESC
LIMIT sqlc.arg(page_limit);