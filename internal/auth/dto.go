package auth

type LoginResponse struct {
	Token string  `json:"token"`
	User  UserDTO `json:"user"`
}

type UserDTO struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
}
