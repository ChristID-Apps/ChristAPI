package contacts

import "time"

type Contact struct {
	ID              int64      `json:"id"`
	FullName        string     `json:"full_name"`
	Phone           *string    `json:"phone"`
	Address         *string    `json:"address"`
	ProfilePhotoURL *string    `json:"profile_photo_url,omitempty"`
	Email           *string    `json:"email,omitempty"`
	Points          *int64     `json:"points,omitempty"`
	CreatedAt       *time.Time `json:"created_at"`
	UpdatedAt       *time.Time `json:"updated_at"`
	SiteID          *int64     `json:"site_id"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty"`
}

type Profile struct {
	ID              int64   `json:"id"`
	FullName        string  `json:"full_name"`
	Username        *string `json:"username,omitempty"`
	Email           string  `json:"email"`
	Phone           *string `json:"phone"`
	Address         *string `json:"address"`
	ProfilePhotoURL *string `json:"profile_photo_url,omitempty"`
	Role            string  `json:"role"`
	Points          int64   `json:"points"`
	ApprovalStatus  string  `json:"approval_status"`
	IsActive        bool    `json:"is_active"`
	SiteID          *int64  `json:"site_id,omitempty"`
}
