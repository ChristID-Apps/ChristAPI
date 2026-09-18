package requests

type GoogleLoginRequest struct {
	IDToken string `json:"id_token"`
	Email   string `json:"email"`
}

type SubmitGoogleUsernameRequest struct {
	IDToken  string  `json:"id_token"`
	Username string  `json:"username"`
	FullName string  `json:"full_name"`
	Phone    *string `json:"phone,omitempty"`
	Address  *string `json:"address,omitempty"`
	SiteID   *int64  `json:"site_id,omitempty"`
}
