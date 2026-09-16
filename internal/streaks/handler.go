package streaks

import (
	"errors"
	"strings"

	"christ-api/pkg/response"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service *Service
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{service: &Service{Repo: repo}}
}

func (h *Handler) CheckIn(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(int64)
	if !ok || userID < 1 {
		return response.Error(c, 401, "Unauthorized", nil)
	}
	request := new(CheckInRequest)
	if err := c.BodyParser(request); err != nil {
		return response.ErrorDetail(c, 422, "Invalid streak check-in request", err)
	}
	result, err := h.service.CheckIn(userID, request.ActivityUUID)
	if err != nil {
		switch {
		case errors.Is(err, ErrActivityNotEligible):
			return response.ErrorDetail(c, 422, "Streak check-in validation failed", err)
		case errors.Is(err, ErrInvalidActivity):
			return response.Error(c, 404, err.Error(), nil)
		case strings.Contains(err.Error(), "required"):
			return response.Error(c, 422, err.Error(), nil)
		default:
			return response.ErrorDetail(c, 500, "Failed to check in streak", err)
		}
	}
	if result == nil {
		return response.Error(c, 500, "Failed to check in streak", fiber.Map{"detail": "streak service returned an empty result without an error"})
	}
	message := "Streak checked in"
	if result.AlreadyCompleted {
		message = "Activity already completed for this period"
	} else if !result.StreakUpdated {
		message = "Points added; streak already counted for today"
	}
	return response.Success(c, message, result)
}

func (h *Handler) List(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(int64)
	if !ok || userID < 1 {
		return response.Error(c, 401, "Unauthorized", nil)
	}
	items, err := h.service.List(userID)
	if err != nil {
		return response.ErrorDetail(c, 500, "Failed to list streaks", err)
	}
	return response.Success(c, "Streaks retrieved", items)
}
