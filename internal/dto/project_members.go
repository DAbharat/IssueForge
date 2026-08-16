package dto

import "time"

type AddProjectMemberRequest struct {
	UserID int64 `json:"user_id"`
}

type ProjectMemberResponse struct {
	ProjectID int64     `json:"project_id"`
	UserID    int64     `json:"user_id"`
	JoinedAt  time.Time `json:"joined_at"`
}

type ProjectMemberSummary struct {
	ID       int64     `json:"id"`
	Email    string    `json:"email"`
	Fullname string    `json:"fullname"`
	Username string    `json:"username"`
	JoinedAt time.Time `json:"joined_at"`
}
