package biblereadings

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrInvalidPayload  = errors.New("payload must be a JSON object no larger than 16 KB")
	ErrAlreadyActive   = errors.New("a reading submission is already pending or approved for this date")
	ErrNotFound        = errors.New("reading submission not found")
	ErrNotReviewable   = errors.New("reading submission is no longer reviewable")
	ErrInvalidPoints   = errors.New("points must be greater than zero")
	ErrRejectNoteEmpty = errors.New("rejection reason is required")
)

const maxPayloadBytes = 16 * 1024

type Repository struct {
	DB *sql.DB
}

func ValidatePayload(payload json.RawMessage) error {
	if len(payload) == 0 || len(payload) > maxPayloadBytes || !json.Valid(payload) {
		return ErrInvalidPayload
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(payload, &object); err != nil || object == nil {
		return ErrInvalidPayload
	}
	return nil
}

func (r *Repository) Submit(userID int64, payload json.RawMessage, now time.Time) (*Submission, error) {
	if r == nil || r.DB == nil {
		return nil, sql.ErrConnDone
	}
	if userID < 1 {
		return nil, errors.New("invalid user id")
	}
	if err := ValidatePayload(payload); err != nil {
		return nil, err
	}
	jakarta, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return nil, err
	}
	readingDate := now.In(jakarta).Format("2006-01-02")
	var item Submission
	err = r.DB.QueryRow(`
		INSERT INTO bible_reading_submissions (user_id, reading_date, payload)
		VALUES ($1, $2, $3)
		RETURNING id, uuid, user_id, reading_date, payload, status, points_awarded, submitted_at`,
		userID, readingDate, []byte(payload)).Scan(
		&item.ID, &item.UUID, &item.UserID, &item.ReadingDate, &item.Payload, &item.Status, &item.PointsAwarded, &item.SubmittedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "uq_bible_reading_one_active_per_user_day" {
			return nil, ErrAlreadyActive
		}
		return nil, err
	}
	return &item, nil
}

func (r *Repository) List(userID *int64, status string, limit, offset int) ([]Submission, error) {
	if r == nil || r.DB == nil {
		return nil, sql.ErrConnDone
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	query := `SELECT s.id, s.uuid, s.user_id, u.email, COALESCE(c.full_name, ''), s.reading_date,
		 s.payload, s.status, s.reviewer_note, s.reviewed_by, s.reviewed_at, s.points_awarded, s.submitted_at
		 FROM bible_reading_submissions s
		 JOIN users u ON u.id = s.user_id
		 LEFT JOIN contacts c ON c.id = u.contact_id
		 WHERE 1=1`
	args := []interface{}{}
	if userID != nil {
		args = append(args, *userID)
		query += fmt.Sprintf(" AND s.user_id = $%d", len(args))
	}
	if status != "" {
		args = append(args, status)
		query += fmt.Sprintf(" AND s.status = $%d", len(args))
	}
	args = append(args, limit, offset)
	query += fmt.Sprintf(" ORDER BY s.reading_date DESC, s.submitted_at DESC LIMIT $%d OFFSET $%d", len(args)-1, len(args))
	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]Submission, 0)
	for rows.Next() {
		var item Submission
		var reviewerID sql.NullInt64
		var reviewerNote sql.NullString
		var reviewedAt sql.NullTime
		if err := rows.Scan(&item.ID, &item.UUID, &item.UserID, &item.UserEmail, &item.FullName, &item.ReadingDate,
			&item.Payload, &item.Status, &reviewerNote, &reviewerID, &reviewedAt, &item.PointsAwarded, &item.SubmittedAt); err != nil {
			return nil, err
		}
		if reviewerID.Valid {
			item.ReviewedBy = &reviewerID.Int64
		}
		if reviewerNote.Valid {
			item.ReviewerNote = &reviewerNote.String
		}
		if reviewedAt.Valid {
			item.ReviewedAt = &reviewedAt.Time
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) Approve(uuid string, reviewerID, points int64) (*Submission, error) {
	if r == nil || r.DB == nil {
		return nil, sql.ErrConnDone
	}
	if points < 1 {
		return nil, ErrInvalidPoints
	}
	tx, err := r.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var userID int64
	var status string
	err = tx.QueryRow(`SELECT user_id, status FROM bible_reading_submissions WHERE uuid=$1 FOR UPDATE`, uuid).Scan(&userID, &status)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if status != "submitted" {
		return nil, ErrNotReviewable
	}

	var balance int64
	err = tx.QueryRow(`UPDATE users SET points_balance=points_balance+$1, updated_at=NOW() WHERE id=$2 RETURNING points_balance`, points, userID).Scan(&balance)
	if err != nil {
		return nil, err
	}
	referenceID := "bible-reading:" + uuid
	if _, err = tx.Exec(`INSERT INTO user_points_ledger (user_id, change_amount, balance_after, reason, reference_id, created_at)
		VALUES ($1,$2,$3,'daily_bible_reading_approved',$4,NOW())`, userID, points, balance, referenceID); err != nil {
		return nil, err
	}

	item, err := scanReviewedSubmission(tx.QueryRow(`UPDATE bible_reading_submissions
		SET status='approved', reviewed_by=$1, reviewed_at=NOW(), points_awarded=$2
		WHERE uuid=$3 AND status='submitted'
		RETURNING id, uuid, user_id, reading_date, payload, status, reviewer_note, reviewed_by, reviewed_at, points_awarded, submitted_at`, reviewerID, points, uuid))
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return item, nil
}

func (r *Repository) Reject(uuid string, reviewerID int64, note string) (*Submission, error) {
	if r == nil || r.DB == nil {
		return nil, sql.ErrConnDone
	}
	note = strings.TrimSpace(note)
	if note == "" {
		return nil, ErrRejectNoteEmpty
	}
	item, err := scanReviewedSubmission(r.DB.QueryRow(`UPDATE bible_reading_submissions
		SET status='rejected', reviewer_note=$1, reviewed_by=$2, reviewed_at=NOW()
		WHERE uuid=$3 AND status='submitted'
		RETURNING id, uuid, user_id, reading_date, payload, status, reviewer_note, reviewed_by, reviewed_at, points_awarded, submitted_at`, note, reviewerID, uuid))
	if errors.Is(err, sql.ErrNoRows) {
		var exists bool
		if lookupErr := r.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM bible_reading_submissions WHERE uuid=$1)`, uuid).Scan(&exists); lookupErr != nil {
			return nil, lookupErr
		}
		if !exists {
			return nil, ErrNotFound
		}
		return nil, ErrNotReviewable
	}
	return item, err
}

type rowScanner interface {
	Scan(...any) error
}

func scanReviewedSubmission(row rowScanner) (*Submission, error) {
	var item Submission
	var reviewerNote sql.NullString
	var reviewerID sql.NullInt64
	var reviewedAt sql.NullTime
	err := row.Scan(&item.ID, &item.UUID, &item.UserID, &item.ReadingDate, &item.Payload, &item.Status,
		&reviewerNote, &reviewerID, &reviewedAt, &item.PointsAwarded, &item.SubmittedAt)
	if err != nil {
		return nil, err
	}
	if reviewerNote.Valid {
		item.ReviewerNote = &reviewerNote.String
	}
	if reviewerID.Valid {
		item.ReviewedBy = &reviewerID.Int64
	}
	if reviewedAt.Valid {
		item.ReviewedAt = &reviewedAt.Time
	}
	return &item, nil
}
