package queue

type AttachmentJob struct {
	IssueID      int64  `json:"issue_id"`
	CommentID    *int64 `json:"comment_id"`
	UserID       int64  `json:"user_id"`
	FileURL      string `json:"file_url"`
	FilePublicID string `json:"file_public_id"`
}

type AttachmentDeleteJob struct {
	AttachmentID int64  `json:"attachment_id"`
	IssueID      int64  `json:"issue_id"`
	UserID       int64  `json:"user_id"`
	FilePublicID string `json:"file_public_id"`
	ResourceType string `json:"resource_type"`
}
