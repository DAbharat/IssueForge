-- name: CreateNotification :one
INSERT INTO notifications(
    recepient_id,
    sender_id,
    notification_type,
    title,
    message,
    project_id,
    issue_id,
    workspace_id
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING id, recepient_id, sender_id, notification_type, title, message, project_id, issue_id, workspace_id, is_read, read_at, created_at, updated_at;


-- name: GetNotificationByID :one
SELECT id, recepient_id, sender_id, notification_type, title, message, project_id, issue_id, workspace_id, is_read, read_at, created_at, updated_at
FROM notifications
WHERE id = $1 AND recepient_id = $2;


-- name: ListUserNotifications :many
SELECT id, recepient_id, sender_id, notification_type, title, message, project_id, issue_id, workspace_id, is_read, read_at, created_at, updated_at
FROM notifications
WHERE recepient_id = $1
ORDER BY created_at DESC
LIMIT $2
OFFSET $3;


-- name: ListUnreadNotifications :many
SELECT id, recepient_id, sender_id, notification_type, title, message, project_id, issue_id, workspace_id, is_read, read_at, created_at, updated_at
FROM notifications
WHERE recepient_id = $1 AND is_read = FALSE
ORDER BY created_at DESC;


-- name: CountUnreadNotifications :one
SELECT COUNT(*) AS unread_count
FROM notifications
WHERE recepient_id = $1 AND is_read = FALSE;


-- name: MarkNotificationAsRead :one
UPDATE notifications
SET is_read = TRUE, read_at = NOW(), updated_at = NOW()
WHERE id = $1 AND recepient_id = $2
RETURNING id, recepient_id, sender_id, notification_type, title, message, project_id, issue_id, workspace_id, is_read, read_at, created_at, updated_at;


-- name: MarkAllNotificationsAsRead :exec
UPDATE notifications
SET is_read = TRUE, read_at = NOW(), updated_at = NOW()
WHERE recepient_id = $1 AND is_read = FALSE;


-- name: DeleteNotification :one
DELETE FROM notifications
WHERE id = $1 AND recepient_id = $2
RETURNING id;


-- name: DeleteAllReadNotifications :exec
DELETE FROM notifications
WHERE recepient_id = $1 AND is_read = TRUE;


-- name: DeleteOldNotifications :exec
DELETE FROM notifications
WHERE created_at < sqlc.arg(cutoff_time)::timestamptz;