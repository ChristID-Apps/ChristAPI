package rewards

import "time"

type Reward struct {
	ID             int64      `json:"id"`
	UUID           string     `json:"uuid"`
	Name           string     `json:"name"`
	Description    *string    `json:"description,omitempty"`
	ImageURL       *string    `json:"image_url,omitempty"`
	PointsRequired int64      `json:"points_required"`
	Stock          int        `json:"stock"`
	Status         string     `json:"status"`
	CreatedBy      *int64     `json:"created_by,omitempty"`
	CreatedAt      *time.Time `json:"created_at,omitempty"`
	UpdatedAt      *time.Time `json:"updated_at,omitempty"`
}

type Redemption struct {
	ID             int64      `json:"id"`
	UUID           string     `json:"uuid"`
	RewardID       int64      `json:"reward_id"`
	RewardName     string     `json:"reward_name"`
	UserID         int64      `json:"user_id"`
	UserEmail      string     `json:"user_email,omitempty"`
	PointsRequired int64      `json:"points_required"`
	Status         string     `json:"status"`
	AdminNote      *string    `json:"admin_note,omitempty"`
	UserCode       *string    `json:"user_code,omitempty"`
	AdminCode      *string    `json:"admin_code,omitempty"`
	ApprovedAt     *time.Time `json:"approved_at,omitempty"`
	RejectedAt     *time.Time `json:"rejected_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type CreateRewardRequest struct {
	Name           string  `json:"name"`
	Description    *string `json:"description"`
	PointsRequired int64   `json:"points_required"`
	Stock          int     `json:"stock"`
}

type UpdateRewardRequest struct {
	Name           *string `json:"name"`
	Description    *string `json:"description"`
	PointsRequired *int64  `json:"points_required"`
	Stock          *int    `json:"stock"`
	Status         *string `json:"status"`
}

type DecisionRequest struct {
	Note *string `json:"note"`
}

type CompleteRequest struct {
	UserCode  string `json:"user_code"`
	AdminCode string `json:"admin_code"`
}
