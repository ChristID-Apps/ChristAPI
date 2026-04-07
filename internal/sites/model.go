package sites

import "time"

type Site struct {
	UUID      string     `json:"uuid"`
	Name      string     `json:"name"`
	Address   *string    `json:"address"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}
