package dto

import "time"

type CreateWorkspaceInvitationRequest struct {
	UserID int64 `json:"user_id"`
}

type WorkspaceInvitationResponse struct {
	ID            int64      `json:"id"`
	WorkspaceID   int64      `json:"workspace_id"`
	InvitedUserID int64      `json:"invited_user_id"`
	InvitedBy     int64      `json:"invitedd_by"`
	Status        string     `json:"status"`
	CreatedAt     time.Time  `json:"created_at"`
	RespondedAt   *time.Time `json:"responded_at"`
}

type PendingWorkspaceInvitationResponse struct {
	ID              int64     `json:"id"`
	WorkspaceID     int64     `json:"workspace_id"`
	WorkspaceName   string    `json:"workspace_name"`
	InviterUsername string    `json:"invited_username"`
	InviterFullname string    `json:"inviter_fullname"`
	CreatedAt       time.Time `json:"created_at"`
}

//for admin
type WorkspacePendingInvitationResponse struct {
	ID            int64     `json:"id"`
	WorkspaceID   int64     `json:"workspace_id"`
	InvitedUserID int64     `json:"invited_user_id"`
	Username      string    `json:"username"`
	Fullname      string    `json:"fullname"`
	CreatedAt     time.Time `json:"created_at"`
}
