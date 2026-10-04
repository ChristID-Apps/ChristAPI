package biblereadings

import (
	"errors"
	"strconv"
	"time"

	"christ-api/pkg/response"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	repo *Repository
}

func NewHandler(repo *Repository) *Handler { return &Handler{repo: repo} }

func (h *Handler) Submit(c *fiber.Ctx) error {
	userID, ok := currentUserID(c)
	if !ok {
		return response.Error(c, 401, "Unauthorized", nil)
	}
	var req SubmitRequest
	if err := c.BodyParser(&req); err != nil {
		return response.ErrorDetail(c, 422, "Invalid Bible reading submission", err)
	}
	item, err := h.repo.Submit(userID, req.Payload, time.Now())
	if err != nil {
		return submissionError(c, err)
	}
	return response.Created(c, "Bible reading submitted for review", item)
}

func (h *Handler) Mine(c *fiber.Ctx) error {
	userID, ok := currentUserID(c)
	if !ok {
		return response.Error(c, 401, "Unauthorized", nil)
	}
	status := c.Query("status")
	if status != "" && !validStatus(status) {
		return response.Error(c, 422, "Invalid submission status", nil)
	}
	items, err := h.repo.List(&userID, status, parseLimit(c.Query("limit"), 50), parseOffset(c.Query("offset")))
	if err != nil {
		return response.ErrorDetail(c, 500, "Failed to list Bible reading submissions", err)
	}
	return response.Success(c, "Bible reading submissions retrieved", items)
}

func (h *Handler) AdminList(c *fiber.Ctx) error {
	status := c.Query("status")
	if status != "" && !validStatus(status) {
		return response.Error(c, 422, "Invalid submission status", nil)
	}
	items, err := h.repo.List(nil, status, parseLimit(c.Query("limit"), 50), parseOffset(c.Query("offset")))
	if err != nil {
		return response.ErrorDetail(c, 500, "Failed to list Bible reading submissions", err)
	}
	return response.Success(c, "Bible reading submissions retrieved", items)
}

func (h *Handler) Approve(c *fiber.Ctx) error {
	reviewerID, ok := currentUserID(c)
	if !ok {
		return response.Error(c, 401, "Unauthorized", nil)
	}
	var req ApproveRequest
	if err := c.BodyParser(&req); err != nil {
		return response.ErrorDetail(c, 422, "Invalid approval request", err)
	}
	item, err := h.repo.Approve(c.Params("uuid"), reviewerID, req.Points)
	if err != nil {
		return submissionError(c, err)
	}
	return response.Success(c, "Bible reading approved and points awarded", item)
}

func (h *Handler) Reject(c *fiber.Ctx) error {
	reviewerID, ok := currentUserID(c)
	if !ok {
		return response.Error(c, 401, "Unauthorized", nil)
	}
	var req RejectRequest
	if err := c.BodyParser(&req); err != nil {
		return response.ErrorDetail(c, 422, "Invalid rejection request", err)
	}
	item, err := h.repo.Reject(c.Params("uuid"), reviewerID, req.Note)
	if err != nil {
		return submissionError(c, err)
	}
	return response.Success(c, "Bible reading rejected", item)
}

func submissionError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, ErrInvalidPayload), errors.Is(err, ErrInvalidPoints), errors.Is(err, ErrRejectNoteEmpty):
		return response.Error(c, 422, err.Error(), nil)
	case errors.Is(err, ErrAlreadyActive), errors.Is(err, ErrNotReviewable):
		return response.Error(c, 409, err.Error(), nil)
	case errors.Is(err, ErrNotFound):
		return response.Error(c, 404, "Bible reading submission not found", nil)
	default:
		return response.ErrorDetail(c, 500, "Bible reading submission failed", err)
	}
}

func currentUserID(c *fiber.Ctx) (int64, bool) {
	switch value := c.Locals("user_id").(type) {
	case int64:
		return value, value > 0
	case int:
		return int64(value), value > 0
	case float64:
		return int64(value), value > 0
	default:
		return 0, false
	}
}

func validStatus(status string) bool {
	return status == "submitted" || status == "approved" || status == "rejected"
}

func parseLimit(raw string, fallback int) int {
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 {
		return fallback
	}
	if value > 100 {
		return 100
	}
	return value
}

func parseOffset(raw string) int {
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return 0
	}
	return value
}
