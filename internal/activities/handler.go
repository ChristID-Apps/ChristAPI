package activities

import (
	"database/sql"
	"encoding/json"
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
	req, err := parseCreateActivityRequest(c)
	if err != nil {
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
	req, err := parseUpdateActivityRequest(c)
	if err != nil {
		return response.ErrorDetail(c, 422, "Invalid activity request", err)
	}
	item, err := h.service.Update(c.Params("uuid"), req)
	if err != nil {
		return activityError(c, err, "Failed to update activity")
	}
	return response.Success(c, "Activity updated", item)
}

func formValues(c *fiber.Ctx) (map[string]string, error) {
	values := make(map[string]string)
	contentType := strings.ToLower(string(c.Request().Header.ContentType()))
	if strings.HasPrefix(contentType, "multipart/form-data") {
		form, err := c.MultipartForm()
		if err != nil {
			return nil, err
		}
		for key, items := range form.Value {
			if len(items) > 0 {
				values[key] = items[0]
			}
		}
		return values, nil
	}
	c.Request().PostArgs().VisitAll(func(key, value []byte) {
		values[string(key)] = string(value)
	})
	return values, nil
}

func parseCreateActivityRequest(c *fiber.Ctx) (*requests.CreateActivityRequest, error) {
	contentType := strings.ToLower(string(c.Request().Header.ContentType()))
	if !strings.HasPrefix(contentType, "multipart/form-data") && !strings.HasPrefix(contentType, "application/x-www-form-urlencoded") {
		req := new(requests.CreateActivityRequest)
		return req, c.BodyParser(req)
	}
	values, err := formValues(c)
	if err != nil {
		return nil, err
	}
	req := &requests.CreateActivityRequest{
		Title: values["title"], Description: stringPointer(values, "description"),
		ActivityType: values["activity_type"], Status: values["status"],
		StreakType: stringPointer(values, "streak_type"),
	}
	req.CategoryID, err = int64Value(values, "category_id")
	if err != nil {
		return nil, err
	}
	req.StreakPoints, err = int64Value(values, "streak_points")
	if err != nil && values["streak_points"] != "" {
		return nil, err
	}
	req.MaxParticipants, err = intPointer(values, "max_participants")
	if err != nil {
		return nil, err
	}
	req.RequiresRegistration, err = boolValue(values, "requires_registration")
	if err != nil {
		return nil, err
	}
	req.StreakEnabled, err = boolValue(values, "streak_enabled")
	if err != nil {
		return nil, err
	}
	if values["schedule"] != "" {
		if err := json.Unmarshal([]byte(values["schedule"]), &req.Schedule); err != nil {
			return nil, err
		}
	}
	if values["occurrence"] != "" {
		if err := json.Unmarshal([]byte(values["occurrence"]), &req.Occurrence); err != nil {
			return nil, err
		}
	}
	return req, nil
}

func parseUpdateActivityRequest(c *fiber.Ctx) (*requests.UpdateActivityRequest, error) {
	contentType := strings.ToLower(string(c.Request().Header.ContentType()))
	if !strings.HasPrefix(contentType, "multipart/form-data") && !strings.HasPrefix(contentType, "application/x-www-form-urlencoded") {
		req := new(requests.UpdateActivityRequest)
		return req, c.BodyParser(req)
	}
	values, err := formValues(c)
	if err != nil {
		return nil, err
	}
	req := &requests.UpdateActivityRequest{Present: make(map[string]bool)}
	for key := range values {
		req.Present[key] = true
	}
	req.Title = stringPointer(values, "title")
	req.Description = stringPointer(values, "description")
	req.ActivityType = stringPointer(values, "activity_type")
	if req.Present["site_id"] {
		req.SiteID, err = int64Pointer(values, "site_id")
	}
	if err != nil {
		return nil, err
	}
	req.Status = stringPointer(values, "status")
	req.StreakType = stringPointer(values, "streak_type")
	if req.Present["category_id"] {
		req.CategoryID, err = int64Pointer(values, "category_id")
	}
	if err != nil {
		return nil, err
	}
	if req.Present["streak_points"] {
		req.StreakPoints, err = int64Pointer(values, "streak_points")
	}
	if err != nil {
		return nil, err
	}
	if req.Present["max_participants"] {
		req.MaxParticipants, err = intPointer(values, "max_participants")
	}
	if err != nil {
		return nil, err
	}
	if req.Present["requires_registration"] {
		req.RequiresRegistration, err = boolPointer(values, "requires_registration")
	}
	if err != nil {
		return nil, err
	}
	if req.Present["streak_enabled"] {
		req.StreakEnabled, err = boolPointer(values, "streak_enabled")
	}
	if err != nil {
		return nil, err
	}
	if values["schedule"] != "" {
		if err := json.Unmarshal([]byte(values["schedule"]), &req.Schedule); err != nil {
			return nil, err
		}
	}
	if values["occurrence"] != "" {
		if err := json.Unmarshal([]byte(values["occurrence"]), &req.Occurrence); err != nil {
			return nil, err
		}
	}
	return req, nil
}

func stringPointer(values map[string]string, key string) *string {
	if value, ok := values[key]; ok {
		return &value
	}
	return nil
}

func int64Value(values map[string]string, key string) (int64, error) {
	if values[key] == "" {
		return 0, nil
	}
	return strconv.ParseInt(values[key], 10, 64)
}

func int64Pointer(values map[string]string, key string) (*int64, error) {
	value, err := int64Value(values, key)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func intPointer(values map[string]string, key string) (*int, error) {
	if values[key] == "" || strings.EqualFold(values[key], "null") {
		return nil, nil
	}
	value, err := strconv.Atoi(values[key])
	return &value, err
}

func boolValue(values map[string]string, key string) (bool, error) {
	if values[key] == "" {
		return false, nil
	}
	return strconv.ParseBool(values[key])
}

func boolPointer(values map[string]string, key string) (*bool, error) {
	value, err := boolValue(values, key)
	if err != nil {
		return nil, err
	}
	return &value, nil
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
