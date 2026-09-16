package activities

import "time"

type Activity struct {
	ID                   int64               `json:"id"`
	UUID                 string              `json:"uuid"`
	Title                string              `json:"title"`
	Description          *string             `json:"description,omitempty"`
	ImageURL             *string             `json:"image_url,omitempty"`
	CategoryID           int64               `json:"category_id"`
	CategoryCode         string              `json:"category_code"`
	CategoryName         string              `json:"category_name"`
	ActivityType         string              `json:"activity_type"`
	SiteID               *int64              `json:"site_id,omitempty"`
	CreatedBy            *int64              `json:"created_by,omitempty"`
	Status               string              `json:"status"`
	MaxParticipants      *int                `json:"max_participants,omitempty"`
	RequiresRegistration bool                `json:"requires_registration"`
	StreakEnabled        bool                `json:"streak_enabled"`
	StreakType           *string             `json:"streak_type,omitempty"`
	StreakPoints         int64               `json:"streak_points"`
	Schedule             *ActivitySchedule   `json:"schedule,omitempty"`
	Occurrence           *ActivityOccurrence `json:"occurrence,omitempty"`
	CreatedAt            *time.Time          `json:"created_at,omitempty"`
	UpdatedAt            *time.Time          `json:"updated_at,omitempty"`
}

type ActivitySchedule struct {
	ID            int64   `json:"id"`
	Frequency     string  `json:"frequency"`
	IntervalValue int     `json:"interval_value"`
	DaysOfWeek    []int   `json:"days_of_week,omitempty"`
	DayOfMonth    *int    `json:"day_of_month,omitempty"`
	StartDate     string  `json:"start_date"`
	EndDate       *string `json:"end_date,omitempty"`
	StartTime     *string `json:"start_time,omitempty"`
	EndTime       *string `json:"end_time,omitempty"`
	Timezone      string  `json:"timezone"`
}

type ActivityOccurrence struct {
	ID       int64      `json:"id"`
	StartsAt time.Time  `json:"starts_at"`
	EndsAt   *time.Time `json:"ends_at,omitempty"`
	Location *string    `json:"location,omitempty"`
	Status   string     `json:"status"`
	Notes    *string    `json:"notes,omitempty"`
}

type ActivityFilter struct {
	Search       string
	CategoryID   *int64
	ActivityType string
	Status       string
	SiteID       *int64
	Limit        int
	Offset       int
}
