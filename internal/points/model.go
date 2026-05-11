package points

import "time"

type LedgerEntry struct {
	ID           int64      `json:"id"`
	UserID       int64      `json:"UserId"`
	ChangeAmount int64      `json:"ChangeAmount"`
	BalanceAfter int64      `json:"BalanceAfter"`
	Reason       string     `json:"reason"`
	ReferenceID  *string    `json:"ReferenceId,omitempty"`
	CreatedAt    *time.Time `json:"CreatedAt"`
}

type PointsState struct {
	UserID  int64         `json:"UserId"`
	Balance int64         `json:"balance"`
	History []LedgerEntry `json:"history"`
}

type UserBalance struct {
	UserID int64  `json:"UserId"`
	Email  string `json:"email"`
	Points int64  `json:"points"`
	SiteID *int64 `json:"SiteId,omitempty"`
}
