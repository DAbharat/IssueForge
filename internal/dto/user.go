package dto

type CreateUserRequest struct {
	Username string `json:"username"`
	Fullname string `json:"fullname"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreateUserResponse struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Fullname string `json:"fullname"`
	Email    string `json:"email"`
}

type LoginUserRequest struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

type LoginUserResponse struct {
	AccessToken string `json:"access_token"`
}

type MeResponse struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Fullname string `json:"fullname"`
	Email    string `json:"email"`
}
