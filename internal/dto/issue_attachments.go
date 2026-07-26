package dto

import "time"

type AttachmentResponse struct {
	ID           int64     `json:"id"`
	IssueID      int64     `json:"issue_id"`
	CommentID    *int64    `json:"comment_id"`
	UploadedBy   int64     `json:"uploaded_by"`
	UploaderName string    `json:"uploader_name"`
	OriginalName string    `json:"original_name"`
	MimeType     string    `json:"mime_type"`
	ResourceType string    `json:"resource_type"`
	FileSize     int64     `json:"file_size"`
	CreatedAt    time.Time `json:"created_at"`
}
