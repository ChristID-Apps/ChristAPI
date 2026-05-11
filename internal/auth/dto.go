package auth

type LoginResponse struct {
	Token string  `json:"token"`
	User  UserDTO `json:"user"`
}

type LoginUserResponse struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	Points int64  `json:"points"`
}

type LoginDataResponse struct {
	User  LoginUserResponse `json:"user"`
	Token string            `json:"token"`
}

type LoginSuccessResponse struct {
	Success bool              `json:"success"`
	Message string            `json:"message"`
	Data    LoginDataResponse `json:"data"`
}

type UserDTO struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
}
