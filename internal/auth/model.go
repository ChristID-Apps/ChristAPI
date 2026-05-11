package auth

import "time"

type User struct {
	ID            int64      `json:"id"`
	UUID          string     `json:"uuid"`
	Email         string     `json:"email"`
	Password      string     `json:"-"`
	PointsBalance int64      `json:"points_balance"`
	RoleID        *int64     `json:"role_id"`
	ContactID     *int64     `json:"contact_id"`
	IsActive      bool       `json:"is_active"`
	LastLoginAt   *time.Time `json:"last_login_at"`
	CreatedAt     *time.Time `json:"created_at"`
	UpdatedAt     *time.Time `json:"updated_at"`
	SiteID        *int64     `json:"site_id"`
}
