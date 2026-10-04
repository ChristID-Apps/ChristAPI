package sites

import "time"

type Site struct {
	UUID      string     `json:"uuid"`
	Name      string     `json:"name"`
	Address   *string    `json:"address"`
	Latitude  *float64   `json:"latitude,omitempty"`
	Longitude *float64   `json:"longitude,omitempty"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}
