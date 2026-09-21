package activitysubmissions

import (
	"christ-api/pkg/response"
	"database/sql"
	"errors"

	"github.com/gofiber/fiber/v2"
)

type Handler struct{ repo *Repository }

func NewHandler(repo *Repository) *Handler { return &Handler{repo: repo} }
func currentUserID(c *fiber.Ctx) (int64, bool) {
	switch v := c.Locals("user_id").(type) {
	case int64:
		return v, v > 0
	case int:
		return int64(v), v > 0
	case float64:
		return int64(v), v > 0
	default:
		return 0, false
	}
}

func (h *Handler) Submit(c *fiber.Ctx) error {
	id, ok := currentUserID(c)
	if !ok {
		return response.Error(c, 401, "Unauthorized", nil)
	}
	var req SubmitRequest
	if err := c.BodyParser(&req); err != nil {
		return response.ErrorDetail(c, 422, "Invalid activity submission", err)
	}
	item, err := h.repo.Submit(id, c.Params("uuid"), req.AnswerText)
	if err != nil {
		return submissionError(c, err)
	}
	return response.Created(c, "Activity submission created", item)
}
func (h *Handler) Mine(c *fiber.Ctx) error {
	id, ok := currentUserID(c)
	if !ok {
		return response.Error(c, 401, "Unauthorized", nil)
	}
	items, err := h.repo.List(&id, c.Query("status"))
	if err != nil {
		return response.ErrorDetail(c, 500, "Failed to list submissions", err)
	}
	return response.Success(c, "Activity submissions retrieved", items)
}
func (h *Handler) AdminList(c *fiber.Ctx) error {
	items, err := h.repo.List(nil, c.Query("status"))
	if err != nil {
		return response.ErrorDetail(c, 500, "Failed to list submissions", err)
	}
	return response.Success(c, "Activity submissions retrieved", items)
}
func (h *Handler) Approve(c *fiber.Ctx) error { return h.review(c, true) }
func (h *Handler) Reject(c *fiber.Ctx) error  { return h.review(c, false) }
func (h *Handler) review(c *fiber.Ctx, approve bool) error {
	id, ok := currentUserID(c)
	if !ok {
		return response.Error(c, 401, "Unauthorized", nil)
	}
	var req ReviewRequest
	_ = c.BodyParser(&req)
	item, err := h.repo.Review(c.Params("uuid"), id, approve, req.Note)
	if err != nil {
		return submissionError(c, err)
	}
	return response.Success(c, "Activity submission reviewed", item)
}
func submissionError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, ErrAnswerRequired), errors.Is(err, ErrAnswerTooShort), errors.Is(err, ErrReviewNoteRequired):
		return response.Error(c, 422, err.Error(), nil)
	case errors.Is(err, ErrAlreadySubmitted), errors.Is(err, ErrInvalidStatus):
		return response.Error(c, 409, err.Error(), nil)
	case errors.Is(err, ErrNotBibleActivity), errors.Is(err, sql.ErrNoRows):
		return response.Error(c, 404, err.Error(), nil)
	default:
		return response.ErrorDetail(c, 500, "Activity submission failed", err)
	}
}
