package attendance

import (
	"errors"
	"strconv"
	"time"

	"christ-api/internal/location"
	"christ-api/pkg/response"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service *Service
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{service: &Service{Repo: repo}}
}

func currentUserID(c *fiber.Ctx) (int64, bool) {
	v := c.Locals("user_id")
	switch id := v.(type) {
	case int64:
		return id, true
	case int:
		return int64(id), true
	case float64:
		return int64(id), true
	default:
		return 0, false
	}
}

func (h *Handler) CheckIn(c *fiber.Ctx) error {
	userID, ok := currentUserID(c)
	if !ok {
		return response.Error(c, 401, "Unauthorized", nil)
	}

	var req CheckInRequest
	if err := c.BodyParser(&req); err != nil {
		return response.ErrorDetail(c, 422, "Invalid check-in request", err)
	}
	if req.Latitude == nil || req.Longitude == nil || req.AccuracyM == nil {
		return response.Error(c, 422, "Location is required", fiber.Map{"detail": "latitude, longitude, and accuracy_m are required"})
	}

	record, err := h.service.CheckIn(userID, location.Position{
		Latitude: *req.Latitude, Longitude: *req.Longitude, Accuracy: *req.AccuracyM,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrAlreadyCheckedIn):
			return response.Error(c, 409, "You have already checked in today", nil)
		case errors.Is(err, location.ErrInvalidPosition):
			return response.Error(c, 422, "Invalid device location", fiber.Map{"detail": "coordinates and reported accuracy must be valid non-negative values"})
		case errors.Is(err, location.ErrNoNearbyTarget):
			return response.Error(c, 422, "You are not near an attendance location", fiber.Map{"detail": "move within 500 meters of a registered attendance location and try again"})
		case errors.Is(err, location.ErrTargetNotConfigured):
			return response.Error(c, 503, "Attendance location is not configured", fiber.Map{"detail": "set ATTENDANCE_LATITUDE and ATTENDANCE_LONGITUDE in the server environment"})
		default:
			return response.ErrorDetail(c, 500, "Failed to check in", err)
		}
	}

	return response.Success(c, "Check-in successful", record)
}

func parseDateOrEmpty(raw string) (string, error) {
	if raw == "" {
		return "", nil
	}
	if _, err := time.Parse("2006-01-02", raw); err != nil {
		return "", err
	}
	return raw, nil
}

func businessDateNow() string {
	date, err := businessDate(time.Now())
	if err != nil {
		return time.Now().UTC().Format("2006-01-02")
	}
	return date
}

func (h *Handler) GetMyHistory(c *fiber.Ctx) error {
	userID, ok := currentUserID(c)
	if !ok {
		return response.Error(c, 401, "Unauthorized", nil)
	}

	startDate, err := parseDateOrEmpty(c.Query("start_date"))
	if err != nil {
		return response.Error(c, 422, "Invalid start_date", nil)
	}
	endDate, err := parseDateOrEmpty(c.Query("end_date"))
	if err != nil {
		return response.Error(c, 422, "Invalid end_date", nil)
	}
	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil || limit <= 0 {
		limit = 20
	}
	offset, err := strconv.Atoi(c.Query("offset"))
	if err != nil || offset < 0 {
		offset = 0
	}

	records, err := h.service.GetMyHistory(userID, startDate, endDate, limit, offset)
	if err != nil {
		return response.ErrorDetail(c, 500, "Failed to fetch attendance history", err)
	}

	return response.Paginated(c, "Attendance history retrieved", records, fiber.Map{"limit": limit, "offset": offset})
}

func (h *Handler) GetMySummary(c *fiber.Ctx) error {
	userID, ok := currentUserID(c)
	if !ok {
		return response.Error(c, 401, "Unauthorized", nil)
	}

	startDate, err := parseDateOrEmpty(c.Query("start_date"))
	if err != nil {
		return response.Error(c, 422, "Invalid start_date", nil)
	}
	endDate, err := parseDateOrEmpty(c.Query("end_date"))
	if err != nil {
		return response.Error(c, 422, "Invalid end_date", nil)
	}
	summary, err := h.service.GetMySummary(userID, startDate, endDate)
	if err != nil {
		return response.ErrorDetail(c, 500, "Failed to summarize attendance", err)
	}
	return response.Success(c, "Attendance summary retrieved", summary)
}

func (h *Handler) AdminReport(c *fiber.Ctx) error {
	date, err := parseDateOrEmpty(c.Query("date"))
	if err != nil {
		return response.Error(c, 422, "Invalid date", nil)
	}
	if date == "" {
		date = businessDateNow()
	}

	var siteID *int64
	if raw := c.Query("site_id"); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return response.ErrorDetail(c, 422, "Invalid site_id", err)
		}
		siteID = &parsed
	}

	var activityID *int64
	if raw := c.Query("activity_id"); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return response.ErrorDetail(c, 422, "Invalid activity_id", err)
		}
		activityID = &parsed
	}

	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil || limit <= 0 {
		limit = 20
	}
	offset, err := strconv.Atoi(c.Query("offset"))
	if err != nil || offset < 0 {
		offset = 0
	}

	records, err := h.service.GetAdminReport(date, siteID, activityID, limit, offset)
	if err != nil {
		return response.ErrorDetail(c, 500, "Failed to fetch attendance report", err)
	}
	return response.Paginated(c, "Attendance report retrieved", records, fiber.Map{"date": date, "site_id": siteID, "activity_id": activityID, "limit": limit, "offset": offset})
}

func (h *Handler) AdminSummary(c *fiber.Ctx) error {
	date, err := parseDateOrEmpty(c.Query("date"))
	if err != nil {
		return response.Error(c, 422, "Invalid date", nil)
	}
	if date == "" {
		date = businessDateNow()
	}

	var siteID *int64
	if raw := c.Query("site_id"); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return response.ErrorDetail(c, 422, "Invalid site_id", err)
		}
		siteID = &parsed
	}

	summary, err := h.service.GetAdminSummary(date, siteID)
	if err != nil {
		return response.ErrorDetail(c, 500, "Failed to summarize attendance report", err)
	}
	return response.Success(c, "Attendance summary retrieved", summary)
}
