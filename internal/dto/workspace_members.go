package dto

import "time"

type AddWorkspaceMemberRequest struct {
	UserID int64  `json:"user_id"`
	Role   string `json:"role"`
}

type WorkspaceMemberResponse struct {
	WorkspaceID int64     `json:"workspace_id"`
	UserID      int64     `json:"user_id"`
	Role        string    `json:"role"`
	JoinedAt    time.Time `json:"joined_at"`
}

type WorkspaceMemberDetails struct {
	WorkspaceID int64     `json:"workspace_id"`
	UserID      int64     `json:"user_id"`
	Email       string    `json:"email"`
	Fullname    string    `json:"fullname"`
	DisplayName string    `json:"display_name"`
	Role        string    `json:"role"`
	JoinedAt    time.Time `json:"joined_at"`
}

type WorkspaceMemberSummary struct {
	ID          int64     `json:"id"`
	Fullname    string    `json:"fullname"`
	DisplayName string    `json:"display_name"`
	Email       string    `json:"email"`
	Role        string    `json:"role"`
	JoinedAt    time.Time `json:"joined_at"`
}

type RemoveWorkspaceMemberResponse struct {
	WorkspaceID int64 `json:"workspace_id"`
	UserID      int64 `json:"user_id"`
}
