package attendance

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	"christ-api/internal/location"
)

var (
	ErrAlreadyCheckedIn = errors.New("already checked in for this date")
)

type Repository struct {
	DB *sql.DB
}

func (r *Repository) CheckIn(userID int64, position location.Position, now time.Time) (*AttendanceRecord, error) {
	if r == nil || r.DB == nil {
		return nil, sql.ErrConnDone
	}
	if err := location.ValidatePosition(position); err != nil {
		return nil, err
	}

	target, err := location.ConfiguredAttendanceTarget(os.Getenv("ATTENDANCE_LATITUDE"), os.Getenv("ATTENDANCE_LONGITUDE"))
	if err != nil {
		return nil, err
	}
	match, err := location.NearestTarget(position, []location.Target{target})
	if err != nil {
		return nil, err
	}

	attendanceDate, err := businessDate(now)
	if err != nil {
		return nil, err
	}

	tx, err := r.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	var insertID int64
	var siteID, activityID interface{}
	if match.Target.Type == "site" {
		siteID = match.Target.ID
	} else if match.Target.Type == "activity" {
		activityID = match.Target.ID
	} else if match.Target.Type != "configured" {
		return nil, location.ErrNoNearbyTarget
	}
	err = tx.QueryRow(`
		INSERT INTO attendance_records (
			user_id, site_id, activity_id, attendance_date, checked_in_at, status, points_earned,
			matched_location_type, matched_location_id, matched_location_name, distance_meters, location_accuracy_m, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,'present',10,$6,$7,$8,$9,$10,NOW(),NOW())
		ON CONFLICT (user_id, attendance_date) DO NOTHING
		RETURNING id`,
		userID, siteID, activityID, attendanceDate, now, match.Target.Type, match.Target.ID,
		match.Target.Name, match.Distance, position.Accuracy,
	).Scan(&insertID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrAlreadyCheckedIn
	}
	if err != nil {
		return nil, err
	}

	var currentBalance int64
	err = tx.QueryRow(`SELECT points_balance FROM users WHERE id = $1 FOR UPDATE`, userID).Scan(&currentBalance)
	if err != nil {
		return nil, err
	}

	newBalance := currentBalance + 10
	if _, err = tx.Exec(`UPDATE users SET points_balance = $1, updated_at = NOW() WHERE id = $2`, newBalance, userID); err != nil {
		return nil, err
	}

	referenceID := fmt.Sprintf("attendance:%d:%s", userID, attendanceDate)
	if _, err = tx.Exec(`
		INSERT INTO user_points_ledger (user_id, change_amount, balance_after, reason, reference_id, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())`, userID, 10, newBalance, "attendance_check_in", referenceID); err != nil {
		return nil, err
	}

	var existing AttendanceRecord
	err = tx.QueryRow(`
		SELECT id, user_id, site_id, activity_id, attendance_date, checked_in_at, status, points_earned, notes,
			matched_location_type, matched_location_id, matched_location_name, distance_meters, location_accuracy_m, created_at, updated_at
		FROM attendance_records WHERE id = $1`, insertID,
	).Scan(
		&existing.ID, &existing.UserID, &existing.SiteID, &existing.ActivityID, &existing.AttendanceDate,
		&existing.CheckedInAt, &existing.Status, &existing.PointsEarned, &existing.Notes,
		&existing.MatchedLocationType, &existing.MatchedLocationID, &existing.MatchedLocationName,
		&existing.DistanceMeters, &existing.LocationAccuracyM,
		&existing.CreatedAt, &existing.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return &existing, nil
}

func (r *Repository) GetMyHistory(userID int64, startDate, endDate string, limit, offset int) ([]AttendanceRecord, error) {
	if r == nil || r.DB == nil {
		return nil, sql.ErrConnDone
	}
	if limit < 1 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT id, user_id, site_id, activity_id, attendance_date, checked_in_at, status, points_earned, notes,
			COALESCE(matched_location_type, ''), COALESCE(matched_location_id, 0), COALESCE(matched_location_name, ''),
			COALESCE(distance_meters, 0), COALESCE(location_accuracy_m, 0), created_at, updated_at
		FROM attendance_records
		WHERE user_id = $1
	`
	args := []interface{}{userID}
	idx := 2
	if startDate != "" {
		query += ` AND attendance_date >= $` + fmt.Sprintf("%d", idx)
		args = append(args, startDate)
		idx++
	}
	if endDate != "" {
		query += ` AND attendance_date <= $` + fmt.Sprintf("%d", idx)
		args = append(args, endDate)
		idx++
	}
	query += ` ORDER BY attendance_date DESC LIMIT $` + fmt.Sprintf("%d", idx) + ` OFFSET $` + fmt.Sprintf("%d", idx+1)
	args = append(args, limit, offset)

	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []AttendanceRecord
	for rows.Next() {
		var rec AttendanceRecord
		var notes sql.NullString
		var activityID sql.NullInt64
		if err := rows.Scan(
			&rec.ID, &rec.UserID, &rec.SiteID, &activityID, &rec.AttendanceDate, &rec.CheckedInAt,
			&rec.Status, &rec.PointsEarned, &notes, &rec.MatchedLocationType, &rec.MatchedLocationID,
			&rec.MatchedLocationName, &rec.DistanceMeters, &rec.LocationAccuracyM, &rec.CreatedAt, &rec.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if activityID.Valid {
			v := activityID.Int64
			rec.ActivityID = &v
		}
		if notes.Valid {
			v := notes.String
			rec.Notes = &v
		}
		out = append(out, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *Repository) GetMySummary(userID int64, startDate, endDate string) (*AttendanceSummary, error) {
	if r == nil || r.DB == nil {
		return nil, sql.ErrConnDone
	}
	if startDate == "" {
		startDate = "2000-01-01"
	}
	if endDate == "" {
		endDate = "2100-12-31"
	}

	summary := &AttendanceSummary{UserID: userID, StartDate: startDate, EndDate: endDate}
	query := `
		SELECT
			COUNT(*) FILTER (WHERE status = 'present') AS present,
			COUNT(*) FILTER (WHERE status = 'late') AS late,
			COUNT(*) FILTER (WHERE status = 'absent') AS absent,
			COUNT(*) FILTER (WHERE status = 'excused') AS excused
		FROM attendance_records
		WHERE user_id = $1 AND attendance_date BETWEEN $2 AND $3
	`
	if err := r.DB.QueryRow(query, userID, startDate, endDate).Scan(
		&summary.Present, &summary.Late, &summary.Absent, &summary.Excused,
	); err != nil {
		return nil, err
	}
	return summary, nil
}

func (r *Repository) GetAdminReport(date string, siteID *int64, activityID *int64, limit, offset int) ([]AttendanceReportItem, error) {
	if r == nil || r.DB == nil {
		return nil, sql.ErrConnDone
	}
	if limit < 1 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT ar.id, ar.user_id, COALESCE(c.full_name, ''), u.email, ar.site_id, ar.activity_id, ar.attendance_date,
		       ar.checked_in_at, ar.status, ar.points_earned, ar.notes,
		       COALESCE(ar.matched_location_type, ''), COALESCE(ar.matched_location_id, 0),
		       COALESCE(ar.matched_location_name, ''), COALESCE(ar.distance_meters, 0),
		       COALESCE(ar.location_accuracy_m, 0)
		FROM attendance_records ar
		JOIN users u ON u.id = ar.user_id
		LEFT JOIN contacts c ON c.id = u.contact_id
		WHERE ar.attendance_date = $1`
	args := []interface{}{date}
	idx := 2
	if siteID != nil {
		query += ` AND ar.site_id = $` + fmt.Sprintf("%d", idx)
		args = append(args, *siteID)
		idx++
	}
	if activityID != nil {
		query += ` AND ar.activity_id = $` + fmt.Sprintf("%d", idx)
		args = append(args, *activityID)
		idx++
	}
	query += ` ORDER BY ar.checked_in_at ASC LIMIT $` + fmt.Sprintf("%d", idx) + ` OFFSET $` + fmt.Sprintf("%d", idx+1)
	args = append(args, limit, offset)

	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []AttendanceReportItem
	for rows.Next() {
		var item AttendanceReportItem
		var site sql.NullInt64
		var activity sql.NullInt64
		var checkedIn sql.NullTime
		var notes sql.NullString
		if err := rows.Scan(
			&item.ID, &item.UserID, &item.FullName, &item.Email, &site, &activity, &item.AttendanceDate,
			&checkedIn, &item.Status, &item.PointsEarned, &notes, &item.MatchedLocationType,
			&item.MatchedLocationID, &item.MatchedLocationName, &item.DistanceMeters, &item.LocationAccuracyM,
		); err != nil {
			return nil, err
		}
		if site.Valid {
			v := site.Int64
			item.SiteID = &v
		}
		if activity.Valid {
			v := activity.Int64
			item.ActivityID = &v
		}
		if checkedIn.Valid {
			v := checkedIn.Time
			item.CheckedInAt = &v
		}
		if notes.Valid {
			v := notes.String
			item.Notes = &v
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *Repository) GetAdminSummary(date string, siteID *int64) (*AttendanceReportSummary, error) {
	if r == nil || r.DB == nil {
		return nil, sql.ErrConnDone
	}
	query := `
		SELECT
			COUNT(*) AS total_users,
			COUNT(*) FILTER (WHERE status = 'present') AS present,
			COUNT(*) FILTER (WHERE status = 'late') AS late,
			COUNT(*) FILTER (WHERE status = 'absent') AS absent,
			COUNT(*) FILTER (WHERE status = 'excused') AS excused
		FROM attendance_records
		WHERE attendance_date = $1`
	args := []interface{}{date}
	if siteID != nil {
		query += ` AND site_id = $2`
		args = append(args, *siteID)
	}

	summary := &AttendanceReportSummary{Date: date}
	if siteID != nil {
		summary.SiteID = siteID
	}
	if err := r.DB.QueryRow(query, args...).Scan(
		&summary.TotalUsers, &summary.Present, &summary.Late, &summary.Absent, &summary.Excused,
	); err != nil {
		return nil, err
	}
	return summary, nil
}
