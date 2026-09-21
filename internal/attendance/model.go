package attendance

import "time"

type AttendanceRecord struct {
	ID             int64     `json:"id"`
	UserID         int64     `json:"user_id"`
	SiteID         *int64    `json:"site_id,omitempty"`
	ActivityID     *int64    `json:"activity_id,omitempty"`
	AttendanceDate string    `json:"attendance_date"`
	CheckedInAt    time.Time `json:"checked_in_at"`
	Status         string    `json:"status"`
	PointsEarned   int64     `json:"points_earned"`
	Notes          *string   `json:"notes,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type CheckInRequest struct {
	SiteID     *int64  `json:"site_id"`
	ActivityID *int64  `json:"activity_id,omitempty"`
	Notes      *string `json:"notes"`
}

type AttendanceSummary struct {
	UserID    int64  `json:"user_id"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Present   int    `json:"present"`
	Late      int    `json:"late"`
	Absent    int    `json:"absent"`
	Excused   int    `json:"excused"`
}

type AttendanceReportItem struct {
	ID             int64      `json:"id"`
	UserID         int64      `json:"user_id"`
	FullName       string     `json:"full_name"`
	Email          string     `json:"email"`
	SiteID         *int64     `json:"site_id,omitempty"`
	ActivityID     *int64     `json:"activity_id,omitempty"`
	AttendanceDate string     `json:"attendance_date"`
	CheckedInAt    *time.Time `json:"checked_in_at,omitempty"`
	Status         string     `json:"status"`
	PointsEarned   int64      `json:"points_earned"`
	Notes          *string    `json:"notes,omitempty"`
}

type AttendanceReportSummary struct {
	Date       string `json:"date"`
	SiteID     *int64 `json:"site_id,omitempty"`
	TotalUsers int    `json:"total_users"`
	Present    int    `json:"present"`
	Late       int    `json:"late"`
	Absent     int    `json:"absent"`
	Excused    int    `json:"excused"`
}
