package activities

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"christ-api/internal/activities/dto/requests"
	"christ-api/pkg/response"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service *Service
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{service: &Service{Repo: repo}}
}

func (h *Handler) List(c *fiber.Ctx) error {
	filter := ActivityFilter{
		Search:       c.Query("search"),
		ActivityType: c.Query("activity_type"),
		Status:       c.Query("status"),
		Limit:        parseIntDefault(c.Query("limit"), 20),
		Offset:       parseIntDefault(c.Query("offset"), 0),
	}
	if value := c.Query("category_id"); value != "" {
		id, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return response.ErrorDetail(c, 422, "Invalid category_id", err)
		}
		filter.CategoryID = &id
	}
	if value := c.Query("site_id"); value != "" {
		id, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return response.ErrorDetail(c, 422, "Invalid site_id", err)
		}
		filter.SiteID = &id
	}
	items, total, err := h.service.List(filter)
	if err != nil {
		return response.ErrorDetail(c, 500, "Failed to list activities", err)
	}
	return response.Paginated(c, "Activities retrieved", items, fiber.Map{"limit": filter.Limit, "offset": filter.Offset, "total": total})
}

func (h *Handler) Categories(c *fiber.Ctx) error {
	categories, err := h.service.ListCategories()
	if err != nil {
		return response.ErrorDetail(c, 500, "Failed to list activity categories", err)
	}
	return response.Success(c, "Activity categories retrieved", categories)
}

func (h *Handler) Get(c *fiber.Ctx) error {
	item, err := h.service.Get(c.Params("uuid"))
	if err != nil {
		if isNotFound(err) {
			return response.ErrorDetail(c, 404, "Activity not found", err)
		}
		return response.ErrorDetail(c, 500, "Failed to retrieve activity", err)
	}
	return response.Success(c, "Activity retrieved", item)
}

func (h *Handler) Create(c *fiber.Ctx) error {
	req := new(requests.CreateActivityRequest)
	if err := c.BodyParser(req); err != nil {
		return response.ErrorDetail(c, 422, "Invalid activity request", err)
	}
	req.SiteID = nil
	var createdBy *int64
	if value, ok := c.Locals("user_id").(int64); ok {
		createdBy = &value
	}
	item, err := h.service.Create(req, createdBy)
	if err != nil {
		return activityError(c, err, "Failed to create activity")
	}
	return response.Created(c, "Activity created", item)
}

func (h *Handler) Update(c *fiber.Ctx) error {
	req := new(requests.UpdateActivityRequest)
	if err := c.BodyParser(req); err != nil {
		return response.ErrorDetail(c, 422, "Invalid activity request", err)
	}
	item, err := h.service.Update(c.Params("uuid"), req)
	if err != nil {
		return activityError(c, err, "Failed to update activity")
	}
	return response.Success(c, "Activity updated", item)
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	if err := h.service.Delete(c.Params("uuid")); err != nil {
		if isNotFound(err) {
			return response.ErrorDetail(c, 404, "Activity not found", err)
		}
		return response.ErrorDetail(c, 500, "Failed to delete activity", err)
	}
	return response.Success(c, "Activity deleted", nil)
}

func (h *Handler) UploadImage(c *fiber.Ctx) error {
	uuid := strings.TrimSpace(c.Params("uuid"))
	if uuid == "" {
		return response.Error(c, 422, "activity uuid is required", fiber.Map{"detail": "uuid path parameter is empty"})
	}

	file, err := c.FormFile("image")
	if err != nil {
		return response.ErrorDetail(c, 422, "Activity image is required", err)
	}
	if file.Size > 5*1024*1024 {
		return response.Error(c, 422, "Activity image is too large", fiber.Map{"detail": "maximum image size is 5 MB"})
	}

	extension := strings.ToLower(filepath.Ext(file.Filename))
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true}
	if !allowed[extension] {
		return response.Error(c, 422, "Unsupported activity image format", fiber.Map{"detail": "allowed formats: jpg, jpeg, png, webp"})
	}

	directory := filepath.Join("uploads", "activities")
	if err := os.MkdirAll(directory, 0755); err != nil {
		return response.ErrorDetail(c, 500, "Failed to prepare activity image directory", err)
	}
	filename := fmt.Sprintf("%s%s", uuid, extension)
	path := filepath.Join(directory, filename)
	if err := c.SaveFile(file, path); err != nil {
		return response.ErrorDetail(c, 500, "Failed to save activity image", err)
	}

	imageURL := "/uploads/activities/" + filename
	if err := h.service.UpdateImage(uuid, imageURL); err != nil {
		_ = os.Remove(path)
		if isNotFound(err) {
			return response.ErrorDetail(c, 404, "Activity not found", err)
		}
		return response.ErrorDetail(c, 500, "Failed to attach activity image", err)
	}
	return response.Success(c, "Activity image uploaded", fiber.Map{"image_url": imageURL})
}

func activityError(c *fiber.Ctx, err error, fallback string) error {
	if isNotFound(err) {
		return response.ErrorDetail(c, 404, "Activity not found", err)
	}
	if err == sql.ErrNoRows {
		return response.ErrorDetail(c, 404, "Activity not found", err)
	}
	if err.Error() == "" {
		return response.ErrorDetail(c, 500, fallback, err)
	}
	if isValidationError(err) {
		return response.ErrorDetail(c, 422, "Activity validation failed", err)
	}
	return response.ErrorDetail(c, 500, fallback, err)
}

func isValidationError(err error) bool {
	messages := []string{"required", "must be", "invalid", "greater than"}
	for _, message := range messages {
		if len(err.Error()) >= len(message) && containsText(err.Error(), message) {
			return true
		}
	}
	return false
}

func containsText(value, target string) bool {
	for i := 0; i+len(target) <= len(value); i++ {
		if value[i:i+len(target)] == target {
			return true
		}
	}
	return false
}

func parseIntDefault(value string, fallback int) int {
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 {
		return fallback
	}
	return parsed
}
