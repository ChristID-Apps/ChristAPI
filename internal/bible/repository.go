package bible

import (
	"database/sql"
	"fmt"
	"strings"
)

type BibleRepository struct {
	DB *sql.DB
}

func (r *BibleRepository) ListVersions() ([]BibleVersion, error) {
	if r == nil || r.DB == nil {
		return nil, sql.ErrConnDone
	}
	rows, err := r.DB.Query(`SELECT id, code, name FROM bible_versions ORDER BY code`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []BibleVersion
	for rows.Next() {
		var version BibleVersion
		if err := rows.Scan(&version.ID, &version.Code, &version.Name); err != nil {
			return nil, err
		}
		versions = append(versions, version)
	}
	return versions, rows.Err()
}

func (r *BibleRepository) ListBooks(testament, versionCode string) ([]BibleBook, error) {
	if r == nil || r.DB == nil {
		return nil, sql.ErrConnDone
	}
	query := `SELECT b.id, b.code, b.name, b.testament, b.book_order
		FROM bible_books b
		WHERE ($1 = '' OR b.testament = $1)
		AND ($2 = '' OR EXISTS (
			SELECT 1 FROM bible_verses v
			JOIN bible_versions bv ON bv.id = v.version_id
			WHERE v.book_id = b.id AND bv.code = $2
		))
		ORDER BY b.book_order`
	rows, err := r.DB.Query(query, testament, versionCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []BibleBook
	for rows.Next() {
		var book BibleBook
		if err := rows.Scan(&book.ID, &book.Code, &book.Name, &book.Testament, &book.BookOrder); err != nil {
			return nil, err
		}
		books = append(books, book)
	}
	return books, rows.Err()
}

func (r *BibleRepository) GetChapter(versionCode, bookCode string, chapter int) (*BibleChapter, error) {
	if r == nil || r.DB == nil {
		return nil, sql.ErrConnDone
	}
	query := `SELECT bv.id, bv.code, bv.name, b.id, b.code, b.name, b.testament, b.book_order,
		v.id, v.chapter, v.verse, v.text
		FROM bible_verses v
		JOIN bible_versions bv ON bv.id = v.version_id
		JOIN bible_books b ON b.id = v.book_id
		WHERE bv.code = $1 AND b.code = $2 AND v.chapter = $3
		ORDER BY v.verse`
	rows, err := r.DB.Query(query, versionCode, bookCode, chapter)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result *BibleChapter
	for rows.Next() {
		if result == nil {
			result = &BibleChapter{Verses: []BibleVerse{}}
		}
		var verse BibleVerse
		if err := rows.Scan(
			&result.Version.ID, &result.Version.Code, &result.Version.Name,
			&result.Book.ID, &result.Book.Code, &result.Book.Name, &result.Book.Testament, &result.Book.BookOrder,
			&verse.ID, &verse.Chapter, &verse.Verse, &verse.Text,
		); err != nil {
			return nil, err
		}
		result.Chapter = verse.Chapter
		result.Verses = append(result.Verses, verse)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if result == nil {
		return nil, sql.ErrNoRows
	}
	return result, nil
}

func (r *BibleRepository) GetVerse(versionCode, bookCode string, chapter, verse int) (*BibleSearchResult, error) {
	if r == nil || r.DB == nil {
		return nil, sql.ErrConnDone
	}
	result := new(BibleSearchResult)
	err := r.DB.QueryRow(`SELECT bv.id, bv.code, bv.name, b.id, b.code, b.name, b.testament, b.book_order,
		v.chapter, v.verse, v.text
		FROM bible_verses v
		JOIN bible_versions bv ON bv.id = v.version_id
		JOIN bible_books b ON b.id = v.book_id
		WHERE bv.code = $1 AND b.code = $2 AND v.chapter = $3 AND v.verse = $4`,
		versionCode, bookCode, chapter, verse).Scan(
		&result.Version.ID, &result.Version.Code, &result.Version.Name,
		&result.Book.ID, &result.Book.Code, &result.Book.Name, &result.Book.Testament, &result.Book.BookOrder,
		&result.Chapter, &result.Verse, &result.Text)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *BibleRepository) Search(filter BibleSearchFilter) ([]BibleSearchResult, error) {
	if r == nil || r.DB == nil {
		return nil, sql.ErrConnDone
	}
	if filter.Limit < 1 || filter.Limit > 100 {
		filter.Limit = 20
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	query := `SELECT bv.id, bv.code, bv.name, b.id, b.code, b.name, b.testament, b.book_order,
		v.chapter, v.verse, v.text
		FROM bible_verses v
		JOIN bible_versions bv ON bv.id = v.version_id
		JOIN bible_books b ON b.id = v.book_id
		WHERE (v.text ILIKE $1 OR to_tsvector('simple', v.text) @@ plainto_tsquery('simple', $1))`
	args := []interface{}{"%" + filter.Query + "%"}
	index := 2
	if filter.VersionCode != "" {
		query += fmt.Sprintf(" AND bv.code = $%d", index)
		args = append(args, filter.VersionCode)
		index++
	}
	if filter.BookCode != "" {
		query += fmt.Sprintf(" AND b.code = $%d", index)
		args = append(args, filter.BookCode)
		index++
	}
	if filter.Testament != "" {
		query += fmt.Sprintf(" AND b.testament = $%d", index)
		args = append(args, filter.Testament)
		index++
	}
	query += fmt.Sprintf(" ORDER BY b.book_order, v.chapter, v.verse LIMIT $%d OFFSET $%d", index, index+1)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []BibleSearchResult
	for rows.Next() {
		var result BibleSearchResult
		if err := rows.Scan(
			&result.Version.ID, &result.Version.Code, &result.Version.Name,
			&result.Book.ID, &result.Book.Code, &result.Book.Name, &result.Book.Testament, &result.Book.BookOrder,
			&result.Chapter, &result.Verse, &result.Text,
		); err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, rows.Err()
}

func normalizeCode(value string) string {
	return strings.TrimSpace(value)
}
