package response

import "github.com/gofiber/fiber/v2"

type ApiResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
}

func Success(c *fiber.Ctx, message string, data interface{}) error {
	return c.Status(200).JSON(ApiResponse{Success: true, Message: message, Data: data})
}

func Created(c *fiber.Ctx, message string, data interface{}) error {
	return c.Status(201).JSON(ApiResponse{Success: true, Message: message, Data: data})
}

func Error(c *fiber.Ctx, status int, message string, errors interface{}) error {
	return c.Status(status).JSON(ApiResponse{Success: false, Message: message, Errors: errors})
}

func Paginated(c *fiber.Ctx, message string, data interface{}, meta interface{}) error {
	return c.Status(200).JSON(ApiResponse{Success: true, Message: message, Data: data, Meta: meta})
}
