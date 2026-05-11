package bible

import (
	"strconv"

	"christ-api/pkg/response"
	"github.com/gofiber/fiber/v2"
)

var service = BibleService{Repo: BibleRepository{}}

func ListSurat(c *fiber.Ctx) error {
	t := c.Query("testament", "")
	var testament *string
	if t != "" {
		testament = &t
	}

	out, err := service.ListSurat(testament)
	if err != nil {
		return response.Error(c, 500, "Failed to retrieve surat list", nil)
	}
	return response.Success(c, "Surat list retrieved", out)
}

func ListPasal(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return response.Error(c, 422, "Invalid id", nil)
	}
	out, err := service.ListPasalBySurat(id)
	if err != nil {
		return response.Error(c, 500, "Failed to retrieve pasal list", nil)
	}
	return response.Success(c, "Pasal list retrieved", out)
}

func ListAyat(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return response.Error(c, 422, "Invalid id", nil)
	}
	out, err := service.ListAyatByPasal(id)
	if err != nil {
		return response.Error(c, 500, "Failed to retrieve ayat list", nil)
	}
	return response.Success(c, "Ayat list retrieved", out)
}

func GetAyat(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return response.Error(c, 422, "Invalid id", nil)
	}
	out, err := service.GetAyatByID(id)
	if err != nil {
		return response.Error(c, 500, "Failed to retrieve ayat", nil)
	}
	return response.Success(c, "Ayat retrieved", out)
}

// GetPasalDetail returns a pasal with its perikops and ayats grouped
func GetPasalDetail(c *fiber.Ctx) error {
	// parse book id and chapter number
	bookIDStr := c.Params("book_id")
	bookID, err := strconv.ParseInt(bookIDStr, 10, 64)
	if err != nil {
		return response.Error(c, 422, "Invalid book_id", nil)
	}
	nomorStr := c.Params("id")
	nomor, err := strconv.ParseInt(nomorStr, 10, 64)
	if err != nil {
		return response.Error(c, 422, "Invalid id", nil)
	}
	out, err := service.GetPasalWithContentsBySuratNomor(bookID, nomor)
	if err != nil {
		return response.Error(c, 500, "Failed to retrieve pasal detail", nil)
	}
	return response.Success(c, "Pasal detail retrieved", out)
}
