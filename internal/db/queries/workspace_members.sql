-- name: AddWorkspaceMember :one
INSERT INTO workspace_members(
    workspace_id,
    user_id,
    role
)
VALUES(
    $1, $2, $3
)
RETURNING workspace_id, user_id, role, joined_at;


-- name: GetWorkspaceMember :one
SELECT wm.workspace_id, wm.user_id, wm.role, wm.joined_at, u.email, u.fullname, u.username
FROM workspace_members wm
JOIN users u ON wm.user_id = u.id
JOIN workspaces w ON wm.workspace_id = w.id
WHERE wm.workspace_id = $1
AND wm.user_id = $2
AND w.deleted_at IS NULL;


-- name: ListWorkspaceMembers :many
SELECT u.id, u.fullname, u.username, u.email, wm.role, wm.joined_at
FROM workspace_members wm
JOIN users u ON wm.user_id = u.id
JOIN workspaces w ON wm.workspace_id = w.id
WHERE wm.workspace_id = $1 AND w.deleted_at IS NULL
ORDER BY wm.joined_at ASC;


-- name: RemoveWorkspaceMember :one
WITH deleted_project_members AS (
    DELETE FROM project_members pm
    USING projects p
    WHERE pm.project_id = p.id AND p.workspace_id = $1 AND pm.user_id = $2
),
deleted_workspace_member AS (
    DELETE FROM workspace_members
    WHERE workspace_id = $1 AND user_id = $2
    RETURNING workspace_id, user_id
)
SELECT dwm.workspace_id, dwm.user_id FROM deleted_workspace_member dwm;


-- name: IsWorkspaceMember :one
SELECT wm.role
FROM workspace_members wm
JOIN workspaces w ON wm.workspace_id = w.id
WHERE wm.workspace_id = $1
AND wm.user_id = $2
AND w.deleted_at IS NULL;


-- name: ListUserWorkspaces :many
SELECT w.id, w.name, wm.role, w.created_at
FROM workspace_members wm
JOIN workspaces w ON wm.workspace_id = w.id
WHERE wm.user_id = sqlc.arg(user_id)
AND w.deleted_at IS NULL
AND (
    sqlc.arg(search) = ''
    OR LOWER(w.name) LIKE '%' || LOWER(sqlc.arg(search)) || '%'
)
ORDER BY w.name;


-- name: IsWorkspaceAdminIncludingDeleted :one
SELECT wm.role
FROM workspace_members wm
WHERE wm.workspace_id = $1
AND wm.user_id = $2;


-- name: PromoteMemberToAdmin :one
UPDATE workspace_members wm
SET role = 'ADMIN'
FROM workspaces w
WHERE wm.workspace_id = w.id AND wm.user_id = $1 AND wm.workspace_id = $2 AND wm.role != 'ADMIN' AND w.deleted_at IS NULL
RETURNING wm.workspace_id, wm.user_id, wm.role;