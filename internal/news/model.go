package news

import "time"

type News struct {
    ID          int64       `json:"id"`
    UUID        string      `json:"uuid"`
    Title       string      `json:"title"`
    Slug        string      `json:"slug"`
    Excerpt     *string     `json:"excerpt,omitempty"`
    Content     string      `json:"content"`
    AuthorID    *int64      `json:"author_id,omitempty"`
    SiteID      *int64      `json:"site_id,omitempty"`
    Status      string      `json:"status"`
    IsFeatured  bool        `json:"is_featured"`
    Meta        []byte      `json:"meta,omitempty"`
    PublishedAt *time.Time  `json:"published_at,omitempty"`
    Views       int64       `json:"views"`
    CreatedAt   *time.Time  `json:"created_at,omitempty"`
    UpdatedAt   *time.Time  `json:"updated_at,omitempty"`
    DeletedAt   *time.Time  `json:"deleted_at,omitempty"`
}

type NewsFilter struct {
    SiteID *int64
    ID     *int64
    Search *string
    Limit  int
    Offset int
}
