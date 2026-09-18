package bible

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

const defaultCSVPath = "docs/alkitab/tb/tb.csv"

type seedBook struct {
	Code      string
	Name      string
	Testament string
	Order     int
}

var seedBooks = []seedBook{
	{Code: "Kej", Name: "Kejadian", Testament: "PL", Order: 1},
	{Code: "Kel", Name: "Keluaran", Testament: "PL", Order: 2},
	{Code: "Ima", Name: "Imamat", Testament: "PL", Order: 3},
	{Code: "Bil", Name: "Bilangan", Testament: "PL", Order: 4},
	{Code: "Ula", Name: "Ulangan", Testament: "PL", Order: 5},
	{Code: "Yos", Name: "Yosua", Testament: "PL", Order: 6},
	{Code: "Hak", Name: "Hakim-hakim", Testament: "PL", Order: 7},
	{Code: "Rut", Name: "Rut", Testament: "PL", Order: 8},
	{Code: "1Sa", Name: "1 Samuel", Testament: "PL", Order: 9},
	{Code: "2Sa", Name: "2 Samuel", Testament: "PL", Order: 10},
	{Code: "1Ra", Name: "1 Raja-raja", Testament: "PL", Order: 11},
	{Code: "2Ra", Name: "2 Raja-raja", Testament: "PL", Order: 12},
	{Code: "1Ta", Name: "1 Tawarikh", Testament: "PL", Order: 13},
	{Code: "2Ta", Name: "2 Tawarikh", Testament: "PL", Order: 14},
	{Code: "Ezr", Name: "Ezra", Testament: "PL", Order: 15},
	{Code: "Neh", Name: "Nehemia", Testament: "PL", Order: 16},
	{Code: "Est", Name: "Ester", Testament: "PL", Order: 17},
	{Code: "Ayb", Name: "Ayub", Testament: "PL", Order: 18},
	{Code: "Mzm", Name: "Mazmur", Testament: "PL", Order: 19},
	{Code: "Ams", Name: "Amsal", Testament: "PL", Order: 20},
	{Code: "Pkh", Name: "Pengkhotbah", Testament: "PL", Order: 21},
	{Code: "Kid", Name: "Kidung Agung", Testament: "PL", Order: 22},
	{Code: "Yes", Name: "Yesaya", Testament: "PL", Order: 23},
	{Code: "Yer", Name: "Yeremia", Testament: "PL", Order: 24},
	{Code: "Rat", Name: "Ratapan", Testament: "PL", Order: 25},
	{Code: "Yeh", Name: "Yehezkiel", Testament: "PL", Order: 26},
	{Code: "Dan", Name: "Daniel", Testament: "PL", Order: 27},
	{Code: "Hos", Name: "Hosea", Testament: "PL", Order: 28},
	{Code: "Yoe", Name: "Yoel", Testament: "PL", Order: 29},
	{Code: "Amo", Name: "Amos", Testament: "PL", Order: 30},
	{Code: "Oba", Name: "Obaja", Testament: "PL", Order: 31},
	{Code: "Yun", Name: "Yunus", Testament: "PL", Order: 32},
	{Code: "Mik", Name: "Mikha", Testament: "PL", Order: 33},
	{Code: "Nah", Name: "Nahum", Testament: "PL", Order: 34},
	{Code: "Hab", Name: "Habakuk", Testament: "PL", Order: 35},
	{Code: "Zef", Name: "Zefanya", Testament: "PL", Order: 36},
	{Code: "Hag", Name: "Hagai", Testament: "PL", Order: 37},
	{Code: "Zak", Name: "Zakharia", Testament: "PL", Order: 38},
	{Code: "Mal", Name: "Maleakhi", Testament: "PL", Order: 39},
	{Code: "Mat", Name: "Matius", Testament: "PB", Order: 40},
	{Code: "Mrk", Name: "Markus", Testament: "PB", Order: 41},
	{Code: "Luk", Name: "Lukas", Testament: "PB", Order: 42},
	{Code: "Yoh", Name: "Yohanes", Testament: "PB", Order: 43},
	{Code: "Kis", Name: "Kisah Para Rasul", Testament: "PB", Order: 44},
	{Code: "Rom", Name: "Roma", Testament: "PB", Order: 45},
	{Code: "1Ko", Name: "1 Korintus", Testament: "PB", Order: 46},
	{Code: "2Ko", Name: "2 Korintus", Testament: "PB", Order: 47},
	{Code: "Gal", Name: "Galatia", Testament: "PB", Order: 48},
	{Code: "Efe", Name: "Efesus", Testament: "PB", Order: 49},
	{Code: "Flp", Name: "Filipi", Testament: "PB", Order: 50},
	{Code: "Kol", Name: "Kolose", Testament: "PB", Order: 51},
	{Code: "1Te", Name: "1 Tesalonika", Testament: "PB", Order: 52},
	{Code: "2Te", Name: "2 Tesalonika", Testament: "PB", Order: 53},
	{Code: "1Ti", Name: "1 Timotius", Testament: "PB", Order: 54},
	{Code: "2Ti", Name: "2 Timotius", Testament: "PB", Order: 55},
	{Code: "Tit", Name: "Titus", Testament: "PB", Order: 56},
	{Code: "Flm", Name: "Filemon", Testament: "PB", Order: 57},
	{Code: "Ibr", Name: "Ibrani", Testament: "PB", Order: 58},
	{Code: "Yak", Name: "Yakobus", Testament: "PB", Order: 59},
	{Code: "1Pt", Name: "1 Petrus", Testament: "PB", Order: 60},
	{Code: "2Pt", Name: "2 Petrus", Testament: "PB", Order: 61},
	{Code: "1Yo", Name: "1 Yohanes", Testament: "PB", Order: 62},
	{Code: "2Yo", Name: "2 Yohanes", Testament: "PB", Order: 63},
	{Code: "3Yo", Name: "3 Yohanes", Testament: "PB", Order: 64},
	{Code: "Yud", Name: "Yudas", Testament: "PB", Order: 65},
	{Code: "Why", Name: "Wahyu", Testament: "PB", Order: 66},
}

func Seed(db *sql.DB, csvPath string) error {
	if db == nil {
		return sql.ErrConnDone
	}
	if strings.TrimSpace(csvPath) == "" {
		csvPath = defaultCSVPath
	}

	file, err := os.Open(csvPath)
	if err != nil {
		return fmt.Errorf("open Bible CSV: %w", err)
	}
	defer file.Close()

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin Bible seed transaction: %w", err)
	}
	rollback := func() { _ = tx.Rollback() }

	var versionID int64
	err = tx.QueryRow(`INSERT INTO bible_versions (code, name)
		VALUES ('tb', 'Terjemahan Baru')
		ON CONFLICT (code) DO UPDATE SET name = EXCLUDED.name, updated_at = NOW()
		RETURNING id`).Scan(&versionID)
	if err != nil {
		rollback()
		return fmt.Errorf("seed TB version: %w", err)
	}

	bookIDs := make(map[string]int64, len(seedBooks))
	for _, book := range seedBooks {
		var id int64
		err := tx.QueryRow(`INSERT INTO bible_books (code, name, testament, book_order)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (code) DO UPDATE SET name = EXCLUDED.name, testament = EXCLUDED.testament, book_order = EXCLUDED.book_order
			RETURNING id`, book.Code, book.Name, book.Testament, book.Order).Scan(&id)
		if err != nil {
			rollback()
			return fmt.Errorf("seed Bible book %s: %w", book.Code, err)
		}
		bookIDs[book.Code] = id
	}

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = 5
	header, err := reader.Read()
	if err != nil {
		rollback()
		return fmt.Errorf("read Bible CSV header: %w", err)
	}
	if len(header) != 5 || header[0] != "id" || header[1] != "kitab" || header[2] != "pasal" || header[3] != "ayat" || header[4] != "firman" {
		rollback()
		return fmt.Errorf("invalid Bible CSV header")
	}

	statement, err := tx.Prepare(`INSERT INTO bible_verses (version_id, book_id, chapter, verse, text)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (version_id, book_id, chapter, verse) DO UPDATE SET text = EXCLUDED.text`)
	if err != nil {
		rollback()
		return fmt.Errorf("prepare Bible verse insert: %w", err)
	}
	defer statement.Close()

	rowNumber := 1
	for {
		record, readErr := reader.Read()
		if readErr == io.EOF {
			break
		}
		rowNumber++
		if readErr != nil {
			rollback()
			return fmt.Errorf("read Bible CSV row %d: %w", rowNumber, readErr)
		}

		bookCode := strings.TrimSpace(record[1])
		bookID, ok := bookIDs[bookCode]
		if !ok {
			rollback()
			return fmt.Errorf("Bible CSV row %d: unknown book code %q", rowNumber, bookCode)
		}
		chapter, parseErr := strconv.Atoi(strings.TrimSpace(record[2]))
		if parseErr != nil || chapter < 1 {
			rollback()
			return fmt.Errorf("Bible CSV row %d: invalid chapter %q", rowNumber, record[2])
		}
		verse, parseErr := strconv.Atoi(strings.TrimSpace(record[3]))
		if parseErr != nil || verse < 1 {
			rollback()
			return fmt.Errorf("Bible CSV row %d: invalid verse %q", rowNumber, record[3])
		}
		if strings.TrimSpace(record[4]) == "" {
			rollback()
			return fmt.Errorf("Bible CSV row %d: verse text is empty", rowNumber)
		}
		if _, err := statement.Exec(versionID, bookID, chapter, verse, record[4]); err != nil {
			rollback()
			return fmt.Errorf("insert Bible CSV row %d: %w", rowNumber, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit Bible seed: %w", err)
	}
	return nil
}
