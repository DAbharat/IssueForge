package dto

type CreateUserRequest struct {
	DisplayName string `json:"display_name"`
	Fullname    string `json:"fullname"`
	Email       string `json:"email"`
	Password    string `json:"password"`
}

type CreateUserResponse struct {
	ID          int64  `json:"id"`
	DisplayName string `json:"display_name"`
	Fullname    string `json:"fullname"`
	Email       string `json:"email"`
	Role        string `json:"role"`
	WorkspaceID *int64 `json:"workspace_id,omitempty"`
}

type LoginUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginUserResponse struct {
	AccessToken string     `json:"access_token"`
	User        MeResponse `json:"user"`
}

type MeResponse struct {
	ID          int64  `json:"id"`
	WorkspaceID *int64 `json:"workspace_id"`
	DisplayName string `json:"display_name"`
	Fullname    string `json:"fullname"`
	Email       string `json:"email"`
	Role        string `json:"role"`
}
