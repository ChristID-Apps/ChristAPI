package streaks

import "time"

type CheckInRequest struct {
	ActivityUUID string `json:"activity_uuid"`
}

type CheckInResult struct {
	StreakType       string    `json:"streak_type"`
	ActivityDate     time.Time `json:"activity_date"`
	CurrentStreak    int       `json:"current_streak"`
	LongestStreak    int       `json:"longest_streak"`
	PointsEarned     int64     `json:"points_earned"`
	AlreadyCompleted bool      `json:"already_completed"`
	StreakUpdated    bool      `json:"streak_updated"`
}

type Streak struct {
	StreakType    string     `json:"streak_type"`
	CurrentStreak int        `json:"current_streak"`
	LongestStreak int        `json:"longest_streak"`
	LastDate      *time.Time `json:"last_activity_date,omitempty"`
}
