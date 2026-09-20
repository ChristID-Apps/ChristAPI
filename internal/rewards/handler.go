package rewards

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"christ-api/pkg/response"
	"github.com/gofiber/fiber/v2"
)

type Handler struct{ service *Service }

func NewHandler(repo *Repository) *Handler { return &Handler{service: &Service{Repo: repo}} }

func userID(c *fiber.Ctx) (int64, bool) {
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

func (h *Handler) List(c *fiber.Ctx) error {
	items, err := h.service.List(true)
	if err != nil {
		return response.ErrorDetail(c, 500, "Failed to list rewards", err)
	}
	return response.Success(c, "Rewards retrieved", items)
}

func (h *Handler) AdminList(c *fiber.Ctx) error {
	items, err := h.service.List(false)
	if err != nil {
		return response.ErrorDetail(c, 500, "Failed to list rewards", err)
	}
	return response.Success(c, "Rewards retrieved", items)
}

func (h *Handler) Create(c *fiber.Ctx) error {
	adminID, ok := userID(c)
	if !ok {
		return response.Error(c, 401, "Unauthorized", nil)
	}
	var req CreateRewardRequest
	if err := c.BodyParser(&req); err != nil {
		return response.ErrorDetail(c, 422, "Invalid reward request", err)
	}
	item, err := h.service.Create(&req, adminID)
	if err != nil {
		return response.Error(c, 422, err.Error(), nil)
	}
	return response.Created(c, "Reward created", item)
}

func (h *Handler) Update(c *fiber.Ctx) error {
	var req UpdateRewardRequest
	if err := c.BodyParser(&req); err != nil {
		return response.ErrorDetail(c, 422, "Invalid reward request", err)
	}
	item, err := h.service.Update(c.Params("uuid"), &req)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return response.Error(c, 404, "Reward not found", nil)
		}
		return response.Error(c, 422, err.Error(), nil)
	}
	return response.Success(c, "Reward updated", item)
}

func (h *Handler) UploadImage(c *fiber.Ctx) error {
	uuid := strings.TrimSpace(c.Params("uuid"))
	if uuid == "" {
		return response.Error(c, 422, "reward uuid is required", nil)
	}
	file, err := c.FormFile("image")
	if err != nil {
		return response.ErrorDetail(c, 422, "Reward image is required", err)
	}
	if file.Size > 5*1024*1024 {
		return response.Error(c, 422, "Reward image is too large", fiber.Map{"detail": "maximum image size is 5 MB"})
	}
	extension := strings.ToLower(filepath.Ext(file.Filename))
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true}
	if !allowed[extension] {
		return response.Error(c, 422, "Unsupported reward image format", fiber.Map{"detail": "allowed formats: jpg, jpeg, png, webp"})
	}
	directory := filepath.Join("uploads", "rewards")
	if err := os.MkdirAll(directory, 0755); err != nil {
		return response.ErrorDetail(c, 500, "Failed to prepare reward image directory", err)
	}
	filename := fmt.Sprintf("%s%s", uuid, extension)
	path := filepath.Join(directory, filename)
	if err := c.SaveFile(file, path); err != nil {
		return response.ErrorDetail(c, 500, "Failed to save reward image", err)
	}
	imageURL := "/uploads/rewards/" + filename
	if err := h.service.UpdateImage(uuid, imageURL); err != nil {
		_ = os.Remove(path)
		if errors.Is(err, sql.ErrNoRows) {
			return response.Error(c, 404, "Reward not found", nil)
		}
		return response.ErrorDetail(c, 500, "Failed to attach reward image", err)
	}
	return response.Success(c, "Reward image uploaded", fiber.Map{"image_url": imageURL})
}

func (h *Handler) Redeem(c *fiber.Ctx) error {
	id, ok := userID(c)
	if !ok {
		return response.Error(c, 401, "Unauthorized", nil)
	}
	item, err := h.service.Redeem(id, c.Params("uuid"))
	if err != nil {
		switch {
		case errors.Is(err, ErrInsufficientPoints), errors.Is(err, ErrOutOfStock), errors.Is(err, ErrAlreadyPending):
			return response.Error(c, 409, err.Error(), nil)
		case errors.Is(err, sql.ErrNoRows):
			return response.Error(c, 404, "Reward or user not found", nil)
		default:
			return response.ErrorDetail(c, 500, "Failed to redeem reward", err)
		}
	}
	return response.Created(c, "Reward redemption submitted", userView(*item))
}

func (h *Handler) MyRedemptions(c *fiber.Ctx) error {
	id, ok := userID(c)
	if !ok {
		return response.Error(c, 401, "Unauthorized", nil)
	}
	items, err := h.service.ListRedemptions(&id, c.Query("status"))
	if err != nil {
		return response.Error(c, 422, err.Error(), nil)
	}
	result := make([]Redemption, 0, len(items))
	for _, item := range items {
		result = append(result, userView(item))
	}
	return response.Success(c, "Reward redemptions retrieved", result)
}

func (h *Handler) AdminRedemptions(c *fiber.Ctx) error {
	items, err := h.service.ListRedemptions(nil, c.Query("status"))
	if err != nil {
		return response.Error(c, 422, err.Error(), nil)
	}
	return response.Success(c, "Reward redemptions retrieved", items)
}

func (h *Handler) Approve(c *fiber.Ctx) error { return h.decide(c, true) }

func (h *Handler) Reject(c *fiber.Ctx) error { return h.decide(c, false) }

func (h *Handler) Complete(c *fiber.Ctx) error {
	var req CompleteRequest
	if err := c.BodyParser(&req); err != nil {
		return response.ErrorDetail(c, 422, "Invalid completion request", err)
	}
	item, err := h.service.Complete(c.Params("uuid"), req.UserCode, req.AdminCode)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidCode):
			return response.Error(c, 422, err.Error(), nil)
		case errors.Is(err, ErrInvalidStatus):
			return response.Error(c, 409, err.Error(), nil)
		case errors.Is(err, sql.ErrNoRows):
			return response.Error(c, 404, "Redemption not found", nil)
		default:
			return response.ErrorDetail(c, 500, "Failed to complete redemption", err)
		}
	}
	item.AdminCode = nil
	return response.Success(c, "Reward redemption completed", item)
}

func (h *Handler) decide(c *fiber.Ctx, approve bool) error {
	var req DecisionRequest
	if err := c.BodyParser(&req); err != nil && strings.TrimSpace(string(c.Body())) != "" {
		return response.ErrorDetail(c, 422, "Invalid decision request", err)
	}
	item, err := h.service.Decide(c.Params("uuid"), approve, req.Note)
	if err != nil {
		switch {
		case errors.Is(err, ErrNoteRequired):
			return response.Error(c, 422, err.Error(), nil)
		case errors.Is(err, ErrInvalidStatus):
			return response.Error(c, 409, err.Error(), nil)
		case errors.Is(err, sql.ErrNoRows):
			return response.Error(c, 404, "Redemption not found", nil)
		default:
			return response.ErrorDetail(c, 500, "Failed to update redemption", err)
		}
	}
	return response.Success(c, "Reward redemption updated", item)
}

func userView(item Redemption) Redemption {
	item.AdminCode = nil
	item.UserEmail = ""
	return item
}
