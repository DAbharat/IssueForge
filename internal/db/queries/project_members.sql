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
SELECT u.id, u.email, u.fullname, u.display_name, pm.joined_at
FROM project_members pm
JOIN users u ON pm.user_id = u.id
WHERE pm.project_id = $1
ORDER BY pm.joined_at ASC;


-- name: IsProjectMember :one
SELECT EXISTS (
    SELECT 1
    FROM project_members
    WHERE project_id = $1 AND user_id = $2
);