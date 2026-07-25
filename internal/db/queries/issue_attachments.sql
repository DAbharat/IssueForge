-- name: CreateAttachment :one
INSERT INTO issue_attachments(
    issue_id, 
    comment_id, 
    uploaded_by, 
    original_name, 
    storage_key, 
    resource_type, 
    mime_type, 
    file_size
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING id, issue_id, comment_id, uploaded_by, original_name, storage_key, resource_type, mime_type, file_size, created_at, deleted_at;


-- name: GetAttachmentByID :one
SELECT ia.id, ia.issue_id, ia.comment_id, ia.uploaded_by, ia.original_name, ia.storage_key, ia.resource_type, ia.mime_type, ia.file_size, ia.created_at, ia.deleted_at,
    uploader.display_name AS uploader_name
FROM issue_attachments ia
JOIN users uploader ON ia.uploaded_by = uploader.id
WHERE ia.id = $1 AND ia.deleted_at IS NULL;


-- name: ListIssueAttachments :many
SELECT ia.id, ia.issue_id, ia.comment_id, ia.uploaded_by, ia.original_name, ia.storage_key, ia.resource_type, ia.mime_type, ia.file_size, ia.created_at,
    uploader.display_name AS uploader_name
FROM issue_attachments ia
JOIN users uploader ON ia.uploaded_by = uploader.id
WHERE ia.issue_id = $1 AND ia.deleted_at IS NULL
ORDER BY ia.created_at ASC;


-- name: ListCommentAttachments :many
SELECT ia.id, ia.issue_id, ia.comment_id, ia.uploaded_by, ia.original_name, ia.storage_key, ia.resource_type, ia.mime_type, ia.file_size, ia.created_at,
    uploader.display_name AS uploader_name
FROM issue_attachments ia
JOIN users uploader ON ia.uploaded_by = uploader.id
WHERE ia.comment_id = $1 AND ia.deleted_at IS NULL
ORDER BY ia.created_at ASC;


-- name: SoftDeleteAttachment :one
UPDATE issue_attachments
SET deleted_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, storage_key, resource_type;