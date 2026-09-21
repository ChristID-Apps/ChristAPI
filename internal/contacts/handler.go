package contacts

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"christ-api/pkg/response"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service *ContactService
}

func NewHandler(repo *ContactRepository) *Handler {
	return &Handler{service: &ContactService{Repo: repo}}
}

type ContactRequest struct {
	FullName string  `json:"full_name"`
	Phone    *string `json:"phone"`
	Address  *string `json:"address"`
	SiteID   *int64  `json:"site_id"`
}

type UpdateContactRequest struct {
	FullName *string         `json:"full_name"`
	Phone    *string         `json:"phone"`
	Address  *string         `json:"address"`
	SiteID   *int64          `json:"site_id"`
	Present  map[string]bool `json:"-"`
}

type ProfileUpdateRequest struct {
	FullName *string         `json:"full_name"`
	Phone    *string         `json:"phone"`
	Address  *string         `json:"address"`
	Present  map[string]bool `json:"-"`
}

func (r *ProfileUpdateRequest) UnmarshalJSON(data []byte) error {
	type alias ProfileUpdateRequest
	var decoded alias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	decoded.Present = make(map[string]bool, len(fields))
	for field := range fields {
		decoded.Present[field] = true
	}
	*r = ProfileUpdateRequest(decoded)
	return nil
}

func (r *UpdateContactRequest) UnmarshalJSON(data []byte) error {
	type alias UpdateContactRequest
	var decoded alias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	decoded.Present = make(map[string]bool, len(fields))
	for field := range fields {
		decoded.Present[field] = true
	}
	*r = UpdateContactRequest(decoded)
	return nil
}

func (h *Handler) List(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	idStr := c.Params("id")
	if idStr == "" {
		idStr = c.Query("id")
	}
	if idStr != "" {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return response.ErrorDetail(c, 422, "Invalid contact id", err)
		}
		ct, err := h.service.GetByID(id)
		if err != nil {
			if err == sql.ErrNoRows {
				return response.Error(c, 404, "Contact not found", nil)
			}
			return response.ErrorDetail(c, 500, "Failed to retrieve contact", err)
		}
		return response.Success(c, "Contact retrieved", ct)
	}

	out, err := h.service.List(page, limit)
	if err != nil {
		return response.ErrorDetail(c, 500, "Failed to list contacts", err)
	}
	meta := map[string]interface{}{"page": page, "limit": limit}
	return response.Paginated(c, "Contacts retrieved", out, meta)
}

func (h *Handler) Create(c *fiber.Ctx) error {
	r := new(ContactRequest)
	if err := c.BodyParser(r); err != nil {
		return response.ErrorDetail(c, 422, "Invalid contact request", err)
	}
	ct, err := h.service.Create(r.FullName, r.Phone, r.Address, r.SiteID)
	if err != nil {
		return response.ErrorDetail(c, 500, "Failed to create contact", err)
	}
	return response.Created(c, "Contact created", ct)
}

func (h *Handler) Update(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return response.ErrorDetail(c, 422, "Invalid contact id", err)
	}
	r := new(UpdateContactRequest)
	if err := c.BodyParser(r); err != nil {
		return response.ErrorDetail(c, 422, "Invalid contact request", err)
	}
	ct, err := h.service.Update(id, r)
	if err != nil {
		if err == sql.ErrNoRows {
			return response.Error(c, 404, "Contact not found", nil)
		}
		return response.ErrorDetail(c, 500, "Failed to update contact", err)
	}
	return response.Success(c, "Contact updated", ct)
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return response.ErrorDetail(c, 422, "Invalid contact id", err)
	}

	ct, err := h.service.Delete(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return response.Error(c, 404, "Contact not found", nil)
		}
		return response.ErrorDetail(c, 500, "Failed to delete contact", err)
	}
	return response.Success(c, "Contact deleted", ct)
}

func (h *Handler) MyProfile(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(int64)
	if !ok || userID < 1 {
		return response.Error(c, 401, "Invalid user ID", nil)
	}
	profile, err := h.service.Repo.GetProfile(userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return response.Error(c, 404, "Profile not found", nil)
		}
		return response.ErrorDetail(c, 500, "Failed to retrieve profile", err)
	}
	return response.Success(c, "Profile retrieved", profile)
}

func (h *Handler) UpdateMyProfile(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(int64)
	if !ok || userID < 1 {
		return response.Error(c, 401, "Invalid user ID", nil)
	}
	req := new(ProfileUpdateRequest)
	if err := c.BodyParser(req); err != nil {
		return response.ErrorDetail(c, 422, "Invalid profile request", err)
	}
	for field := range req.Present {
		if field != "full_name" && field != "phone" && field != "address" {
			return response.Error(c, 422, "Only full_name, phone, and address can be updated", nil)
		}
	}
	if req.Present["full_name"] && (req.FullName == nil || strings.TrimSpace(*req.FullName) == "") {
		return response.Error(c, 422, "full_name cannot be empty", nil)
	}
	profile, err := h.service.Repo.UpdateProfile(userID, req)
	if err != nil {
		if err == sql.ErrNoRows {
			return response.Error(c, 404, "Profile not found", nil)
		}
		return response.ErrorDetail(c, 500, "Failed to update profile", err)
	}
	return response.Success(c, "Profile updated", profile)
}

func (h *Handler) UploadProfilePhoto(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(int64)
	if !ok || userID < 1 {
		return response.Error(c, 401, "Invalid user ID", nil)
	}
	file, err := c.FormFile("image")
	if err != nil {
		return response.ErrorDetail(c, 422, "Profile photo is required", err)
	}
	if file.Size > 5*1024*1024 {
		return response.Error(c, 422, "Profile photo is too large", fiber.Map{"detail": "maximum image size is 5 MB"})
	}
	extension := strings.ToLower(filepath.Ext(file.Filename))
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true}
	if !allowed[extension] {
		return response.Error(c, 422, "Unsupported profile photo format", fiber.Map{"detail": "allowed formats: jpg, jpeg, png, webp"})
	}
	directory := filepath.Join("uploads", "profiles")
	if err := os.MkdirAll(directory, 0755); err != nil {
		return response.ErrorDetail(c, 500, "Failed to prepare profile photo directory", err)
	}
	filename := fmt.Sprintf("%d%s", userID, extension)
	path := filepath.Join(directory, filename)
	if err := c.SaveFile(file, path); err != nil {
		return response.ErrorDetail(c, 500, "Failed to save profile photo", err)
	}
	profile, err := h.service.Repo.UpdateProfilePhoto(userID, "/uploads/profiles/"+filename)
	if err != nil {
		_ = os.Remove(path)
		if err == sql.ErrNoRows {
			return response.Error(c, 404, "Profile not found", nil)
		}
		return response.ErrorDetail(c, 500, "Failed to attach profile photo", err)
	}
	return response.Success(c, "Profile photo uploaded", profile)
}
