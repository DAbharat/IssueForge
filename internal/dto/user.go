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
}

type LoginUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginUserResponse struct {
	AccessToken  string     `json:"access_token"`
	RefreshToken string     `json:"refresh_token"`
	User         MeResponse `json:"user"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type WorkspaceSummary struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

type MeResponse struct {
	ID          int64              `json:"id"`
	DisplayName string             `json:"display_name"`
	Fullname    string             `json:"fullname"`
	Email       string             `json:"email"`
	Workspaces  []WorkspaceSummary `json:"workspaces"`
}
