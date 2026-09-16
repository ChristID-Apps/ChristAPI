package news

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"christ-api/pkg/response"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service *NewsService
}

func NewHandler(repo *NewsRepository) *Handler {
	return &Handler{service: &NewsService{Repo: repo}}
}

func (h *Handler) List(c *fiber.Ctx) error {
	var filter NewsFilter

	if v := c.Query("site_id"); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return response.ErrorDetail(c, 422, "Invalid site_id", err)
		}
		filter.SiteID = &id
	}
	if v := c.Query("id"); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return response.ErrorDetail(c, 422, "Invalid id", err)
		}
		filter.ID = &id
	}
	if v := c.Query("search"); v != "" {
		filter.Search = &v
	}
	if v := c.Query("limit"); v != "" {
		lim, err := strconv.Atoi(v)
		if err != nil || lim < 1 {
			return response.ErrorDetail(c, 422, "Invalid limit", err)
		}
		filter.Limit = lim
	}
	if v := c.Query("offset"); v != "" {
		off, err := strconv.Atoi(v)
		if err != nil || off < 0 {
			return response.ErrorDetail(c, 422, "Invalid offset", err)
		}
		filter.Offset = off
	}

	out, err := h.service.List(filter)
	if err != nil {
		return response.ErrorDetail(c, 500, "Failed to list news", err)
	}
	return response.Success(c, "News retrieved", out)
}

func (h *Handler) Create(c *fiber.Ctx) error {
	var req News
	if err := c.BodyParser(&req); err != nil {
		return response.ErrorDetail(c, 422, "Invalid news request", err)
	}
	created, err := h.service.Create(&req)
	if err != nil {
		return response.ErrorDetail(c, 500, "Failed to create news", err)
	}
	return response.Created(c, "News created", created)
}

func (h *Handler) Update(c *fiber.Ctx) error {
	uuid := c.Params("uuid")
	var req NewsUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return response.ErrorDetail(c, 422, "Invalid news request", err)
	}
	if err := h.service.Update(uuid, &req); err != nil {
		return response.ErrorDetail(c, 500, "Failed to update news", err)
	}
	return response.Success(c, "News updated", nil)
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	uuid := c.Params("uuid")
	if uuid == "" {
		return response.Error(c, 422, "uuid required", fiber.Map{"detail": "news uuid path parameter is empty"})
	}
	if err := h.service.Delete(uuid); err != nil {
		return response.ErrorDetail(c, 500, "Failed to delete news", err)
	}
	return response.Success(c, "News deleted", nil)
}

func (h *Handler) UploadImage(c *fiber.Ctx) error {
	uuid := strings.TrimSpace(c.Params("uuid"))
	if uuid == "" {
		return response.Error(c, 422, "news uuid is required", fiber.Map{"detail": "uuid path parameter is empty"})
	}
	file, err := c.FormFile("image")
	if err != nil {
		return response.ErrorDetail(c, 422, "News image is required", err)
	}
	if file.Size > 5*1024*1024 {
		return response.Error(c, 422, "News image is too large", fiber.Map{"detail": "maximum image size is 5 MB"})
	}
	extension := strings.ToLower(filepath.Ext(file.Filename))
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true}
	if !allowed[extension] {
		return response.Error(c, 422, "Unsupported news image format", fiber.Map{"detail": "allowed formats: jpg, jpeg, png, webp"})
	}
	directory := filepath.Join("uploads", "news")
	if err := os.MkdirAll(directory, 0755); err != nil {
		return response.ErrorDetail(c, 500, "Failed to prepare news image directory", err)
	}
	filename := fmt.Sprintf("%s%s", uuid, extension)
	path := filepath.Join(directory, filename)
	if err := c.SaveFile(file, path); err != nil {
		return response.ErrorDetail(c, 500, "Failed to save news image", err)
	}
	imageURL := "/uploads/news/" + filename
	if err := h.service.UpdateImage(uuid, imageURL); err != nil {
		_ = os.Remove(path)
		return response.ErrorDetail(c, 500, "Failed to attach news image", err)
	}
	return response.Success(c, "News image uploaded", fiber.Map{"image_url": imageURL})
}
