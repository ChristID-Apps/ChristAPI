package bible

import (
	"database/sql"
	"errors"
	"strconv"

	"christ-api/pkg/response"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service *BibleService
}

func NewHandler(repo *BibleRepository) *Handler {
	return &Handler{service: &BibleService{Repo: repo}}
}

func (h *Handler) ListVersions(c *fiber.Ctx) error {
	versions, err := h.service.ListVersions()
	if err != nil {
		return response.ErrorDetail(c, 500, "Failed to list Bible versions", err)
	}
	return response.Success(c, "Bible versions retrieved", versions)
}

func (h *Handler) ListBooks(c *fiber.Ctx) error {
	books, err := h.service.ListBooks(c.Query("testament"), c.Query("version"))
	if err != nil {
		return bibleError(c, err, "Failed to list Bible books")
	}
	return response.Success(c, "Bible books retrieved", books)
}

func (h *Handler) GetChapter(c *fiber.Ctx) error {
	chapter, err := strconv.Atoi(c.Params("chapter"))
	if err != nil {
		return response.ErrorDetail(c, 422, "Invalid chapter", err)
	}
	result, err := h.service.GetChapter(c.Params("version"), c.Params("book"), chapter)
	if err != nil {
		return bibleError(c, err, "Failed to get Bible chapter")
	}
	return response.Success(c, "Bible chapter retrieved", result)
}

func (h *Handler) GetVerse(c *fiber.Ctx) error {
	chapter, err := strconv.Atoi(c.Params("chapter"))
	if err != nil {
		return response.ErrorDetail(c, 422, "Invalid chapter", err)
	}
	verse, err := strconv.Atoi(c.Params("verse"))
	if err != nil {
		return response.ErrorDetail(c, 422, "Invalid verse", err)
	}
	result, err := h.service.GetVerse(c.Params("version"), c.Params("book"), chapter, verse)
	if err != nil {
		return bibleError(c, err, "Failed to get Bible verse")
	}
	return response.Success(c, "Bible verse retrieved", result)
}

func (h *Handler) Search(c *fiber.Ctx) error {
	limit, err := parseQueryInt(c.Query("limit"), 20)
	if err != nil {
		return response.ErrorDetail(c, 422, "Invalid limit", err)
	}
	offset, err := parseQueryInt(c.Query("offset"), 0)
	if err != nil {
		return response.ErrorDetail(c, 422, "Invalid offset", err)
	}
	results, err := h.service.Search(BibleSearchFilter{
		VersionCode: c.Query("version"),
		BookCode:    c.Query("book"),
		Testament:   c.Query("testament"),
		Query:       c.Query("q"),
		Limit:       limit,
		Offset:      offset,
	})
	if err != nil {
		return bibleError(c, err, "Failed to search Bible")
	}
	return response.Paginated(c, "Bible search results retrieved", results, fiber.Map{"limit": limit, "offset": offset})
}

func parseQueryInt(value string, defaultValue int) (int, error) {
	if value == "" {
		return defaultValue, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 {
		if err == nil {
			err = errors.New("value must not be negative")
		}
		return 0, err
	}
	return parsed, nil
}

func bibleError(c *fiber.Ctx, err error, message string) error {
	switch {
	case errors.Is(err, ErrInvalidBibleInput):
		return response.ErrorDetail(c, 422, "Invalid Bible request", err)
	case errors.Is(err, sql.ErrNoRows):
		return response.Error(c, 404, "Bible content not found", nil)
	default:
		return response.ErrorDetail(c, 500, message, err)
	}
}
