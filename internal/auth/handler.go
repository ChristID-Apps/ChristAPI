package auth

import (
	"christ-api/internal/contacts"
	"christ-api/pkg/response"

	"github.com/gofiber/fiber/v2"
)

var service = AuthService{}

// InitService initializes package-level auth service with a repository.
func InitService(repo *AuthRepository) {
	if repo != nil {
		service = AuthService{Repo: repo}
	}
}

func Login(c *fiber.Ctx) error {
	type Request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		SiteID   *int64 `json:"site_id"`
	}

	req := new(Request)

	if err := c.BodyParser(req); err != nil {
		return response.Error(c, 422, "Invalid request", nil)
	}

	token, user, err := service.Login(req.Email, req.Password, req.SiteID)
	if err != nil {
		return response.Error(c, 401, "Invalid credentials", nil)
	}

	data := LoginDataResponse{User: *user, Token: token}
	return response.Success(c, "Login berhasil", data)
}

func Register(c *fiber.Ctx) error {
	type Request struct {
		FullName      string  `json:"full_name"`
		Phone         *string `json:"phone"`
		Address       *string `json:"address"`
		ContactSiteID *int64  `json:"contact_site_id"`
		Email         string  `json:"email"`
		Password      string  `json:"password"`
		RoleID        *int64  `json:"role_id"`
		SiteID        *int64  `json:"site_id"`
	}

	req := new(Request)
	if err := c.BodyParser(req); err != nil {
		return response.Error(c, 422, "Invalid request", nil)
	}

	validationErrs := make(map[string][]string)
	if req.FullName == "" {
		validationErrs["full_name"] = append(validationErrs["full_name"], "Full name is required")
	}
	if req.Email == "" {
		validationErrs["email"] = append(validationErrs["email"], "Email is required")
	}
	if req.Password == "" {
		validationErrs["password"] = append(validationErrs["password"], "Password is required")
	}
	if len(validationErrs) > 0 {
		return response.Error(c, 422, "Validation failed", validationErrs)
	}

	token, user, contact, err := service.RegisterWithContact(req.FullName, req.Phone, req.Address, req.ContactSiteID, req.Email, req.Password, req.RoleID, req.SiteID)
	if err != nil {
		return response.Error(c, 500, "Failed to register user", nil)
	}

	resp := struct {
		Token   string            `json:"token"`
		User    UserDTO           `json:"user"`
		Contact *contacts.Contact `json:"contact"`
	}{
		Token:   token,
		User:    UserDTO{ID: user.ID, Email: user.Email},
		Contact: contact,
	}

	return response.Created(c, "User registered", resp)
}
