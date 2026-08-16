-- name: CreateWorkspaceInvitation :one
INSERT INTO workspace_invitations(
    workspace_id, invited_user_id, invited_by
) VALUES (
    $1, $2, $3
)
RETURNING id, workspace_id, invited_user_id, invited_by, status, created_at, responded_at;


-- name: GetWorkspaceInvitation :one
SELECT wi.id, wi.workspace_id, wi.invited_user_id, wi.invited_by, wi.status, wi.created_at, wi.responded_at
FROM workspace_invitations wi
JOIN workspaces w ON wi.workspace_id = w.id
WHERE wi.id = $1 AND w.deleted_at IS NULL;


-- name: ListPendingWorkspaceInvitations :many
SELECT wi.id, wi.workspace_id, wi.invited_user_id, wi.invited_by, wi.status, wi.created_at,
    inviter.username AS inviter_username, inviter.fullname AS inviter_fullname,
    w.name AS workspace_name
FROM workspace_invitations wi
JOIN users inviter ON wi.invited_by = inviter.id
JOIN workspaces w ON w.id = wi.workspace_id
WHERE wi.invited_user_id = $1 AND wi.status = 'PENDING' AND w.deleted_at IS NULL
ORDER BY wi.created_at DESC;


-- name: ListPendingWorkspaceInvitationsForWorkspace :many
SELECT wi.id, wi.workspace_id, wi.invited_user_id, wi.invited_by, wi.status, wi.created_at,
    u.username AS invited_username, u.fullname AS invited_fullname
FROM workspace_invitations wi
JOIN users u ON wi.invited_user_id = u.id
JOIN workspaces w ON wi.workspace_id = w.id
WHERE wi.workspace_id = $1 AND wi.status = 'PENDING' AND w.deleted_at IS NULL
ORDER BY wi.created_at DESC;


-- name: AcceptInvitation :one
UPDATE workspace_invitations
SET status = 'ACCEPTED', responded_at = NOW()
WHERE id = $1 AND invited_user_id = $2 AND status = 'PENDING'
RETURNING id, workspace_id, invited_user_id, status, responded_at;


-- name: DeclineInvitation :one
UPDATE workspace_invitations
SET status = 'DECLINED', responded_at = NOW()
WHERE id = $1 AND invited_user_id = $2 AND status = 'PENDING'
RETURNING id, workspace_id, invited_user_id, status, responded_at;


-- name: CancelInvitation :one
UPDATE workspace_invitations
SET status = 'CANCELLED', responded_at = NOW()
WHERE id = $1 AND invited_by = $2 AND status = 'PENDING'
RETURNING id, status, responded_at;