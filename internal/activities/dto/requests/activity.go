package requests

import (
	"encoding/json"
	"time"
)

type ScheduleRequest struct {
	Frequency     string  `json:"frequency"`
	IntervalValue int     `json:"interval_value"`
	DaysOfWeek    []int   `json:"days_of_week"`
	DayOfMonth    *int    `json:"day_of_month"`
	StartDate     string  `json:"start_date"`
	EndDate       *string `json:"end_date"`
	StartTime     *string `json:"start_time"`
	EndTime       *string `json:"end_time"`
	Timezone      string  `json:"timezone"`
}

type OccurrenceRequest struct {
	StartsAt time.Time  `json:"starts_at"`
	EndsAt   *time.Time `json:"ends_at"`
	Location *string    `json:"location"`
	Status   string     `json:"status"`
	Notes    *string    `json:"notes"`
}

type CreateActivityRequest struct {
	Title                string                `json:"title"`
	Description          *string               `json:"description"`
	CategoryID           int64                 `json:"category_id"`
	ActivityType         string                `json:"activity_type"`
	SiteID               *int64                `json:"site_id"`
	Status               string                `json:"status"`
	MaxParticipants      *int                  `json:"max_participants"`
	RequiresRegistration bool                  `json:"requires_registration"`
	StreakEnabled        bool                  `json:"streak_enabled"`
	StreakType           *string               `json:"streak_type"`
	StreakPoints         int64                 `json:"streak_points"`
	Schedule             *ScheduleRequest      `json:"schedule"`
	Occurrence           *OccurrenceRequest    `json:"occurrence"`
	BibleConfig          *BibleActivityRequest `json:"bible_config"`
}

type BibleActivityRequest struct {
	VersionCode         string `json:"version_code"`
	BookCode            string `json:"book_code"`
	StartChapter        int    `json:"start_chapter"`
	StartVerse          int    `json:"start_verse"`
	EndChapter          int    `json:"end_chapter"`
	EndVerse            int    `json:"end_verse"`
	RequiresReflection  bool   `json:"requires_reflection"`
	ReflectionPrompt    string `json:"reflection_prompt"`
	ReflectionMinLength int    `json:"reflection_min_length"`
}

type UpdateActivityRequest struct {
	Title                *string            `json:"title"`
	Description          *string            `json:"description"`
	CategoryID           *int64             `json:"category_id"`
	ActivityType         *string            `json:"activity_type"`
	SiteID               *int64             `json:"site_id"`
	Status               *string            `json:"status"`
	MaxParticipants      *int               `json:"max_participants"`
	RequiresRegistration *bool              `json:"requires_registration"`
	StreakEnabled        *bool              `json:"streak_enabled"`
	StreakType           *string            `json:"streak_type"`
	StreakPoints         *int64             `json:"streak_points"`
	Schedule             *ScheduleRequest   `json:"schedule"`
	Occurrence           *OccurrenceRequest `json:"occurrence"`
	Present              map[string]bool    `json:"-"`
}

func (r *UpdateActivityRequest) UnmarshalJSON(data []byte) error {
	type alias UpdateActivityRequest
	var decoded alias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	decoded.Present = make(map[string]bool, len(fields))
	for field := range fields {
		decoded.Present[field] = true
	}
	*r = UpdateActivityRequest(decoded)
	return nil
}
