package news

import (
	"encoding/json"
	"time"
)

type NewsUpdateRequest struct {
	Title       *string         `json:"title"`
	Slug        *string         `json:"slug"`
	ImageURL    *string         `json:"image_url"`
	Excerpt     *string         `json:"excerpt"`
	Content     *string         `json:"content"`
	AuthorID    *int64          `json:"author_id"`
	SiteID      *int64          `json:"site_id"`
	Status      *string         `json:"status"`
	IsFeatured  *bool           `json:"is_featured"`
	Meta        json.RawMessage `json:"meta"`
	PublishedAt *time.Time      `json:"published_at"`
	Present     map[string]bool `json:"-"`
}

func (r *NewsUpdateRequest) UnmarshalJSON(data []byte) error {
	type alias NewsUpdateRequest
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
	*r = NewsUpdateRequest(decoded)
	return nil
}

type News struct {
	ID          int64           `json:"id"`
	UUID        string          `json:"uuid"`
	Title       string          `json:"title"`
	Slug        string          `json:"slug"`
	ImageURL    *string         `json:"image_url,omitempty"`
	Excerpt     *string         `json:"excerpt,omitempty"`
	Content     string          `json:"content"`
	AuthorID    *int64          `json:"author_id,omitempty"`
	AuthorName  *string         `json:"author_name,omitempty"`
	SiteID      *int64          `json:"site_id,omitempty"`
	Status      string          `json:"status"`
	IsFeatured  bool            `json:"is_featured"`
	Meta        json.RawMessage `json:"meta,omitempty"`
	PublishedAt *time.Time      `json:"published_at,omitempty"`
	Views       int64           `json:"views"`
	CreatedAt   *time.Time      `json:"created_at,omitempty"`
	UpdatedAt   *time.Time      `json:"updated_at,omitempty"`
	DeletedAt   *time.Time      `json:"deleted_at,omitempty"`
}

type NewsFilter struct {
	SiteID *int64
	ID     *int64
	Search *string
	Limit  int
	Offset int
}
