package activitysubmissions

import (
	"database/sql"
	"errors"
	"strings"
)

var (
	ErrNotBibleActivity   = errors.New("activity is not a Bible activity")
	ErrAnswerRequired     = errors.New("reflection answer is required")
	ErrAnswerTooShort     = errors.New("reflection answer is too short")
	ErrAlreadySubmitted   = errors.New("activity has already been submitted")
	ErrInvalidStatus      = errors.New("submission is no longer reviewable")
	ErrReviewNoteRequired = errors.New("rejection note is required")
)

type Repository struct{ DB *sql.DB }

func (r *Repository) Submit(userID int64, activityUUID string, answer *string) (*Submission, error) {
	tx, err := r.DB.Begin()
	if err != nil {
		return nil, err
	}
	rollback := func() { _ = tx.Rollback() }
	var activityID int64
	var title, version, book string
	var startChapter, startVerse, endChapter, endVerse, minLength int
	var requiresReflection bool
	err = tx.QueryRow(`SELECT a.id, a.title, b.version_code, b.book_code, b.start_chapter, b.start_verse, b.end_chapter, b.end_verse, b.requires_reflection, b.reflection_min_length FROM activities a JOIN activity_bible_configs b ON b.activity_id = a.id WHERE a.uuid = $1 AND a.deleted_at IS NULL`, activityUUID).Scan(&activityID, &title, &version, &book, &startChapter, &startVerse, &endChapter, &endVerse, &requiresReflection, &minLength)
	if err != nil {
		rollback()
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotBibleActivity
		}
		return nil, err
	}
	if requiresReflection {
		if answer == nil || strings.TrimSpace(*answer) == "" {
			rollback()
			return nil, ErrAnswerRequired
		}
		if len([]rune(strings.TrimSpace(*answer))) < minLength {
			rollback()
			return nil, ErrAnswerTooShort
		}
	}
	var exists bool
	if err = tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM activity_submissions WHERE activity_id = $1 AND user_id = $2)`, activityID, userID).Scan(&exists); err != nil {
		rollback()
		return nil, err
	}
	if exists {
		rollback()
		return nil, ErrAlreadySubmitted
	}
	var item Submission
	err = tx.QueryRow(`INSERT INTO activity_submissions (activity_id, user_id, answer_text) VALUES ($1,$2,$3) RETURNING id, uuid, activity_id, user_id, answer_text, status, submitted_at`, activityID, userID, answer).Scan(&item.ID, &item.UUID, &item.ActivityID, &item.UserID, &item.AnswerText, &item.Status, &item.SubmittedAt)
	if err != nil {
		rollback()
		return nil, err
	}
	item.ActivityUUID, item.ActivityTitle = activityUUID, title
	_ = version
	_ = book
	_ = startChapter
	_ = startVerse
	_ = endChapter
	_ = endVerse
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *Repository) List(userID *int64, status string) ([]Submission, error) {
	query := `SELECT s.id, s.uuid, s.activity_id, a.uuid, a.title, s.user_id, u.email, s.answer_text, s.status, s.reviewer_note, s.submitted_at, s.reviewed_at FROM activity_submissions s JOIN activities a ON a.id = s.activity_id JOIN users u ON u.id = s.user_id WHERE 1=1`
	args := []interface{}{}
	if userID != nil {
		query += ` AND s.user_id = $1`
		args = append(args, *userID)
	}
	if status != "" {
		query += ` AND s.status = $` + string(rune('1'+len(args)))
		args = append(args, status)
	}
	query += ` ORDER BY s.submitted_at DESC`
	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []Submission
	for rows.Next() {
		var item Submission
		if err := rows.Scan(&item.ID, &item.UUID, &item.ActivityID, &item.ActivityUUID, &item.ActivityTitle, &item.UserID, &item.UserEmail, &item.AnswerText, &item.Status, &item.ReviewerNote, &item.SubmittedAt, &item.ReviewedAt); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) Review(uuid string, reviewerID int64, approve bool, note *string) (*Submission, error) {
	if !approve && (note == nil || strings.TrimSpace(*note) == "") {
		return nil, ErrReviewNoteRequired
	}
	status := "rejected"
	if approve {
		status = "approved"
	}
	var item Submission
	err := r.DB.QueryRow(`UPDATE activity_submissions s SET status = $1, reviewer_note = $2, reviewed_by = $3, reviewed_at = NOW() WHERE s.uuid = $4 AND s.status = 'submitted' RETURNING s.id, s.uuid, s.activity_id, s.user_id, s.answer_text, s.status, s.reviewer_note, s.submitted_at, s.reviewed_at`, status, note, reviewerID, uuid).Scan(&item.ID, &item.UUID, &item.ActivityID, &item.UserID, &item.AnswerText, &item.Status, &item.ReviewerNote, &item.SubmittedAt, &item.ReviewedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrInvalidStatus
	}
	return &item, err
}
