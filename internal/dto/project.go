package dto

type CreateProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CreateProjectResponse struct {
	ID          int64  `json:"id"`
	WorkspaceID int64  `json:"workspace_id"`
	LeadID      int64  `json:"lead_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ProjectResponse struct {
	ID          int64  `json:"id"`
	WorkspaceID int64  `json:"workspace_id"`
	LeadID      int64  `json:"lead_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
