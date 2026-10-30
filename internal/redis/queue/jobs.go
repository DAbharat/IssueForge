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

type CreateNotificationJob struct {
	NotificationID   int64  `json:"notification_id,omitempty"`
	RecepientID      int64  `json:"recepient_id"`
	SenderID         *int64 `json:"sender_id,omitempty"`
	NotificationType string `json:"notification_type"`
	IssueID          int64  `json:"issue_id,omitempty"`
	ProjectID        int64  `json:"project_id,omitempty"`
	WorkspaceID      int64  `json:"workspace_id,omitempty"`
	Title            string `json:"title"`
	Message          string `json:"message"`
}

type DeleteAllNotificationJob struct {
	NotificationID int64 `json:"notification_id"`
	RecepientID    int64 `json:"recepient_id"`
}
