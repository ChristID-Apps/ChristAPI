package auth

import (
	"christ-api/internal/contacts"

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
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid request",
		})
	}

	token, err := service.Login(req.Email, req.Password, req.SiteID)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"token": token,
	})
}

func Register(c *fiber.Ctx) error {
	type Request struct {
		FullName      string  `json:"full_name"`
		Phone         *string `json:"phone"`
		Address       *string `json:"address"`
		ContactSiteID *int64  `json:"contact_site_id"`
		Email         string  `json:"email"`
		Password      string  `json:"password"`
		SiteID        *int64  `json:"site_id"`
	}

	req := new(Request)
	if err := c.BodyParser(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}

	if req.FullName == "" || req.Email == "" || req.Password == "" {
		return c.Status(400).JSON(fiber.Map{"error": "full_name, email and password are required"})
	}

	token, user, contact, err := service.RegisterWithContact(req.FullName, req.Phone, req.Address, req.ContactSiteID, req.Email, req.Password, req.SiteID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
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

	return c.Status(201).JSON(resp)
}
