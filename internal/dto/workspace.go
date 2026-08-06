package dto

import "time"

type CreateWorkspaceRequest struct {
	Name string `json:"name"`
}

type CreateWorkspaceResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type UpdateWorkspaceRequest struct {
	Name *string `json:"name"`
}

type WorkspaceResponse struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
