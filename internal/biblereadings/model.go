package biblereadings

import (
	"encoding/json"
	"time"
)

type Submission struct {
	ID            int64           `json:"id"`
	UUID          string          `json:"uuid"`
	UserID        int64           `json:"user_id"`
	UserEmail     string          `json:"user_email,omitempty"`
	FullName      string          `json:"full_name,omitempty"`
	ReadingDate   string          `json:"reading_date"`
	Payload       json.RawMessage `json:"payload"`
	Status        string          `json:"status"`
	ReviewerNote  *string         `json:"reviewer_note,omitempty"`
	ReviewedBy    *int64          `json:"reviewed_by,omitempty"`
	ReviewedAt    *time.Time      `json:"reviewed_at,omitempty"`
	PointsAwarded int64           `json:"points_awarded"`
	SubmittedAt   time.Time       `json:"submitted_at"`
}

type SubmitRequest struct {
	Payload json.RawMessage `json:"payload"`
}

type ApproveRequest struct {
	Points int64 `json:"points"`
}

type RejectRequest struct {
	Note string `json:"note"`
}
