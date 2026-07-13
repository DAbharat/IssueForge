package dto

type CreateWorkspaceRequest struct {
	Name string `json:"name"`
}

type CreateWorkspaceResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type WorkspaceResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}
