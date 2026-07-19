-- name: CreateComment :one
INSERT INTO comments(
    issue_id,
    author_id,
    parent_comment_id,
    content
) VALUES (
    $1, $2, $3, $4
)
RETURNING id, issue_id, author_id, parent_comment_id, content, is_edited, created_at, updated_at;


-- name: GetCommentByID :one
SELECT c.id, c.issue_id, c.author_id, u.display_name as author_name, c.parent_comment_id, c.content, c.is_edited, c.created_at, c.updated_at
FROM comments c
JOIN users u ON c.author_id = u.id
WHERE c.id = $1;


-- name: ListIssueComments :many
SELECT c.id, c.issue_id, c.author_id, u.display_name as author_name, c.parent_comment_id, c.content, c.is_edited, c.created_at, c.updated_at
FROM comments c
JOIN users u
ON c.author_id = u.id
WHERE c.issue_id = $1
ORDER BY c.created_at ASC
LIMIT $2 OFFSET $3;


-- name: UpdateComment :one
UPDATE comments
SET content = $2, is_edited = TRUE, updated_at = NOW()
WHERE id = $1 AND author_id = $3
RETURNING id, issue_id, author_id, parent_comment_id, content, is_edited, created_at, updated_at;


-- name: DeleteComment :one
DELETE FROM comments
WHERE id = $1
RETURNING id;

