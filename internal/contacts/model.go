package contacts

import "time"

type Contact struct {
	ID        int64      `json:"id"`
	FullName  string     `json:"full_name"`
	Phone     *string    `json:"phone"`
	Address   *string    `json:"address"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
	SiteID    *int64     `json:"site_id"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}
