package auth

import "github.com/gofiber/fiber/v2"

var service = AuthService{
    Repo: AuthRepository{},
}

func Login(c *fiber.Ctx) error {
    type Request struct {
        Username string `json:"username"`
        Password string `json:"password"`
    }

    req := new(Request)

    if err := c.BodyParser(req); err != nil {
        return c.Status(400).JSON(fiber.Map{
            "error": "invalid request",
        })
    }

    token, err := service.Login(req.Username, req.Password)
    if err != nil {
        return c.Status(401).JSON(fiber.Map{
            "error": err.Error(),
        })
    }

    return c.JSON(fiber.Map{
        "token": token,
    })
}