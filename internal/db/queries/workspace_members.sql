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
SELECT wm.workspace_id, wm.user_id, wm.role, wm.joined_at, u.email, u.fullname, u.display_name
FROM workspace_members wm
JOIN users u ON wm.user_id = u.id
JOIN workspaces w ON wm.workspace_id = w.id
WHERE wm.workspace_id = $1
AND wm.user_id = $2
AND w.deleted_at IS NULL;


-- name: ListWorkspaceMembers :many
SELECT u.id, u.fullname, u.display_name, u.email, wm.role, wm.joined_at
FROM workspace_members wm
JOIN users u ON wm.user_id = u.id
JOIN workspaces w ON wm.workspace_id = w.id
WHERE wm.workspace_id = $1 AND w.deleted_at IS NULL
ORDER BY wm.joined_at ASC;


-- name: RemoveWorkspaceMember :one
DELETE FROM workspace_members
WHERE workspace_id = $1
AND user_id = $2
RETURNING workspace_id, user_id;


-- name: IsWorkspaceMember :one
SELECT wm.role
FROM workspace_members wm
JOIN workspaces w ON wm.workspace_id = w.id
WHERE wm.workspace_id = $1
AND wm.user_id = $2
AND w.deleted_at IS NULL;


-- name: ListUserWorkspaces :many
SELECT w.id, w.name, wm.role
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