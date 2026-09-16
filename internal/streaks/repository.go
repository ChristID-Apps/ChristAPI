package streaks

import (
	"database/sql"
	"errors"
	"strings"
	"time"
)

type Repository struct {
	DB *sql.DB
}

var (
	ErrActivityNotEligible = errors.New("activity is not configured for streaks")
	ErrInvalidActivity     = errors.New("activity not found")
)

func (r *Repository) CheckIn(userID int64, activityUUID string, now time.Time) (*CheckInResult, error) {
	if r == nil || r.DB == nil {
		return nil, sql.ErrConnDone
	}
	if strings.TrimSpace(activityUUID) == "" {
		return nil, ErrInvalidActivity
	}
	location, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return nil, err
	}
	localNow := now.In(location)
	localDate := time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 0, 0, 0, 0, location)
	periodDate := localDate

	tx, err := r.DB.Begin()
	if err != nil {
		return nil, err
	}
	rollback := func() { _ = tx.Rollback() }

	var activityID int64
	var streakType string
	var activityTitle string
	var reward int64
	if err := tx.QueryRow(`SELECT id, title, streak_type, streak_points FROM activities WHERE uuid=$1 AND deleted_at IS NULL AND streak_enabled=true`, activityUUID).Scan(&activityID, &activityTitle, &streakType, &reward); err != nil {
		rollback()
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrActivityNotEligible
		}
		return nil, err
	}
	streakType = "daily_activity"
	var logID int64
	err = tx.QueryRow(`INSERT INTO point_streak_logs (user_id, streak_type, activity_id, activity_date, points_earned) VALUES ($1,$2,$3,$4,$5) ON CONFLICT (user_id, streak_type, activity_date, activity_id) DO NOTHING RETURNING id`, userID, streakType, activityID, periodDate, reward).Scan(&logID)
	if errors.Is(err, sql.ErrNoRows) {
		var current, longest int
		var last sql.NullTime
		err = tx.QueryRow(`SELECT current_streak, longest_streak, last_activity_date FROM point_streaks WHERE user_id=$1 AND streak_type=$2`, userID, streakType).Scan(&current, &longest, &last)
		rollback()
		if errors.Is(err, sql.ErrNoRows) {
			return &CheckInResult{StreakType: streakType, ActivityDate: periodDate, AlreadyCompleted: true, StreakUpdated: false}, nil
		}
		if err != nil {
			return nil, err
		}
		return &CheckInResult{
			StreakType:       streakType,
			ActivityDate:     periodDate,
			CurrentStreak:    current,
			LongestStreak:    longest,
			AlreadyCompleted: true,
			StreakUpdated:    false,
		}, nil
	}
	if err != nil {
		rollback()
		return nil, err
	}
	_ = logID

	_, err = tx.Exec(`INSERT INTO point_streaks (user_id, streak_type, current_streak, longest_streak, last_activity_date) VALUES ($1,$2,0,0,NULL) ON CONFLICT (user_id, streak_type) DO NOTHING`, userID, streakType)
	if err != nil {
		rollback()
		return nil, err
	}

	var current, longest int
	var last sql.NullTime
	err = tx.QueryRow(`SELECT current_streak, longest_streak, last_activity_date FROM point_streaks WHERE user_id=$1 AND streak_type=$2 FOR UPDATE`, userID, streakType).Scan(&current, &longest, &last)
	if err != nil {
		rollback()
		return nil, err
	}

	streakUpdated := true
	if last.Valid {
		lastDate := time.Date(last.Time.Year(), last.Time.Month(), last.Time.Day(), 0, 0, 0, 0, location)
		daysSince := int(periodDate.Sub(lastDate).Hours() / 24)
		if daysSince == 0 {
			streakUpdated = false
		} else {
			switch {
			case daysSince == 1:
				current++
			default:
				current = 1
			}
		}
	} else {
		current = 1
	}
	if streakUpdated && current > longest {
		longest = current
	}

	if streakUpdated {
		_, err = tx.Exec(`UPDATE point_streaks SET current_streak=$1, longest_streak=$2, last_activity_date=$3, updated_at=NOW() WHERE user_id=$4 AND streak_type=$5`, current, longest, periodDate, userID, streakType)
		if err != nil {
			rollback()
			return nil, err
		}
	}

	if reward > 0 {
		var balance int64
		if err := tx.QueryRow(`SELECT points_balance FROM users WHERE id=$1 FOR UPDATE`, userID).Scan(&balance); err != nil {
			rollback()
			return nil, err
		}
		newBalance := balance + reward
		if _, err := tx.Exec(`UPDATE users SET points_balance=$1, updated_at=NOW() WHERE id=$2`, newBalance, userID); err != nil {
			rollback()
			return nil, err
		}
		if _, err := tx.Exec(`INSERT INTO user_points_ledger (user_id, change_amount, balance_after, reason, reference_id) VALUES ($1,$2,$3,$4,$5)`, userID, reward, newBalance, activityTitle, activityUUID); err != nil {
			rollback()
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &CheckInResult{StreakType: streakType, ActivityDate: periodDate, CurrentStreak: current, LongestStreak: longest, PointsEarned: reward, StreakUpdated: streakUpdated}, nil
}

func (r *Repository) List(userID int64) ([]Streak, error) {
	if r == nil || r.DB == nil {
		return nil, sql.ErrConnDone
	}
	rows, err := r.DB.Query(`SELECT streak_type, current_streak, longest_streak, last_activity_date FROM point_streaks WHERE user_id=$1 ORDER BY streak_type`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []Streak
	for rows.Next() {
		var item Streak
		var last sql.NullTime
		if err := rows.Scan(&item.StreakType, &item.CurrentStreak, &item.LongestStreak, &last); err != nil {
			return nil, err
		}
		if last.Valid {
			item.LastDate = &last.Time
		}
		result = append(result, item)
	}
	return result, rows.Err()
}
