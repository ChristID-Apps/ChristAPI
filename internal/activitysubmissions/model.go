package activitysubmissions

import "time"

type Submission struct {
	ID            int64      `json:"id"`
	UUID          string     `json:"uuid"`
	ActivityID    int64      `json:"activity_id"`
	ActivityUUID  string     `json:"activity_uuid"`
	ActivityTitle string     `json:"activity_title"`
	UserID        int64      `json:"user_id"`
	UserEmail     string     `json:"user_email,omitempty"`
	AnswerText    *string    `json:"answer_text,omitempty"`
	Status        string     `json:"status"`
	ReviewerNote  *string    `json:"reviewer_note,omitempty"`
	SubmittedAt   time.Time  `json:"submitted_at"`
	ReviewedAt    *time.Time `json:"reviewed_at,omitempty"`
}

type SubmitRequest struct {
	AnswerText *string `json:"answer_text"`
}
type ReviewRequest struct {
	Note *string `json:"note"`
}
