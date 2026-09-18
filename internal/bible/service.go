package bible

import (
	"errors"
	"fmt"
	"strings"
)

var ErrInvalidBibleInput = errors.New("invalid bible request")

type BibleService struct {
	Repo *BibleRepository
}

func (s *BibleService) ListVersions() ([]BibleVersion, error) {
	return s.Repo.ListVersions()
}

func (s *BibleService) ListBooks(testament, versionCode string) ([]BibleBook, error) {
	testament = strings.ToUpper(strings.TrimSpace(testament))
	if testament != "" && testament != "PL" && testament != "PB" {
		return nil, fmt.Errorf("%w: testament must be PL or PB", ErrInvalidBibleInput)
	}
	return s.Repo.ListBooks(testament, strings.TrimSpace(versionCode))
}

func (s *BibleService) GetChapter(versionCode, bookCode string, chapter int) (*BibleChapter, error) {
	if chapter < 1 {
		return nil, fmt.Errorf("%w: chapter must be greater than 0", ErrInvalidBibleInput)
	}
	return s.Repo.GetChapter(strings.TrimSpace(versionCode), strings.TrimSpace(bookCode), chapter)
}

func (s *BibleService) GetVerse(versionCode, bookCode string, chapter, verse int) (*BibleSearchResult, error) {
	if chapter < 1 || verse < 1 {
		return nil, fmt.Errorf("%w: chapter and verse must be greater than 0", ErrInvalidBibleInput)
	}
	return s.Repo.GetVerse(strings.TrimSpace(versionCode), strings.TrimSpace(bookCode), chapter, verse)
}

func (s *BibleService) Search(filter BibleSearchFilter) ([]BibleSearchResult, error) {
	filter.Query = strings.TrimSpace(filter.Query)
	if filter.Query == "" {
		return nil, fmt.Errorf("%w: query is required", ErrInvalidBibleInput)
	}
	filter.VersionCode = strings.TrimSpace(filter.VersionCode)
	filter.BookCode = strings.TrimSpace(filter.BookCode)
	filter.Testament = strings.ToUpper(strings.TrimSpace(filter.Testament))
	if filter.Testament != "" && filter.Testament != "PL" && filter.Testament != "PB" {
		return nil, fmt.Errorf("%w: testament must be PL or PB", ErrInvalidBibleInput)
	}
	return s.Repo.Search(filter)
}
