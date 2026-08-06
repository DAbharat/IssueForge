package dto

import "time"

type CreateProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	WorkspaceID int64  `json:"-"`
}

type CreateProjectResponse struct {
	ID          int64     `json:"id"`
	WorkspaceID int64     `json:"workspace_id"`
	LeadID      int64     `json:"lead_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type UpdateProjectDetailsRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type UpdateProjectLeadRequest struct {
	LeadID *int64 `json:"lead_id"`
}

type ProjectResponse struct {
	ID          int64     `json:"id"`
	WorkspaceID int64     `json:"workspace_id"`
	LeadID      int64     `json:"lead_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
