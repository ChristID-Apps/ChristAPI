package auth

import "github.com/gofiber/fiber/v2"

var service = AuthService{}

// InitService initializes package-level auth service with a repository.
func InitService(repo *AuthRepository) {
	if repo != nil {
		service = AuthService{Repo: *repo}
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
		Email     string `json:"email"`
		Password  string `json:"password"`
		SiteID    *int64 `json:"site_id"`
		ContactID *int64 `json:"contact_id"`
	}

	req := new(Request)

	if err := c.BodyParser(req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid request",
		})
	}

	if req.Email == "" || req.Password == "" {
		return c.Status(400).JSON(fiber.Map{"error": "email and password are required"})
	}

	token, user, err := service.Register(req.Email, req.Password, req.SiteID, req.ContactID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	resp := LoginResponse{
		Token: token,
		User: UserDTO{
			ID:    user.ID,
			Email: user.Email,
		},
	}

	return c.Status(201).JSON(resp)
}
