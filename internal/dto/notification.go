package dto

import "time"

type CreateNotificationRequest struct {
	RecepientID      int64  `json:"recepient_id"`
	SenderID         *int64 `json:"sender_id,omitempty"`
	NotificationType string `json:"notification_type"`
	Title            string `json:"title"`
	Message          string `json:"message"`
	ProjectID        *int64 `json:"project_id,omitempty"`
	IssueID          *int64 `json:"issue_id,omitempty"`
	WorkspaceID      *int64 `json:"workspace_id,omitempty"`
}

type NotificationResponse struct {
	ID               int64      `json:"id"`
	RecepientID      int64      `json:"recepient_id"`
	SenderID         *int64     `json:"sender_id,omitempty"`
	NotificationType string     `json:"notification_type"`
	Title            string     `json:"title"`
	Message          string     `json:"message"`
	ProjectID        *int64     `json:"project_id,omitempty"`
	IssueID          *int64     `json:"issue_id,omitempty"`
	WorkspaceID      *int64     `json:"workspace_id,omitempty"`
	IsRead           bool       `json:"is_read"`
	ReadAt           *time.Time `json:"read_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type UnreadCountResponse struct {
	UnreadCount int64 `json:"unread_count"`
}
