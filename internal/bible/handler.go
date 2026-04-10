package bible

import (
	"strconv"

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
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(out)
}

func ListPasal(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid id"})
	}
	out, err := service.ListPasalBySurat(id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(out)
}

func ListAyat(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid id"})
	}
	out, err := service.ListAyatByPasal(id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(out)
}

func GetAyat(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid id"})
	}
	out, err := service.GetAyatByID(id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(out)
}

// GetPasalDetail returns a pasal with its perikops and ayats grouped
func GetPasalDetail(c *fiber.Ctx) error {
	// parse book id and chapter number
	bookIDStr := c.Params("book_id")
	bookID, err := strconv.ParseInt(bookIDStr, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid book_id"})
	}
	nomorStr := c.Params("id")
	nomor, err := strconv.ParseInt(nomorStr, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid id"})
	}
	out, err := service.GetPasalWithContentsBySuratNomor(bookID, nomor)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(out)
}
