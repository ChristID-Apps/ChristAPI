package rewards

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrInsufficientPoints = errors.New("insufficient points")
	ErrOutOfStock         = errors.New("reward is out of stock")
	ErrAlreadyPending     = errors.New("you already have a pending redemption for this reward")
	ErrInvalidStatus      = errors.New("redemption is no longer pending")
	ErrNoteRequired       = errors.New("rejection note is required")
	ErrInvalidCode        = errors.New("invalid redemption code")
)

type Repository struct{ DB *sql.DB }

func (r *Repository) List(activeOnly bool) ([]Reward, error) {
	query := `SELECT id, uuid, name, description, image_url, points_required, stock, status, created_by, created_at, updated_at FROM rewards`
	if activeOnly {
		query += ` WHERE status = 'active' AND stock > 0`
	}
	query += ` ORDER BY created_at DESC`
	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []Reward
	for rows.Next() {
		var item Reward
		if err := rows.Scan(&item.ID, &item.UUID, &item.Name, &item.Description, &item.ImageURL, &item.PointsRequired, &item.Stock, &item.Status, &item.CreatedBy, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) Create(req *CreateRewardRequest, createdBy int64) (*Reward, error) {
	var item Reward
	err := r.DB.QueryRow(`
		INSERT INTO rewards (name, description, points_required, stock, created_by)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, uuid, name, description, image_url, points_required, stock, status, created_by, created_at, updated_at`,
		req.Name, req.Description, req.PointsRequired, req.Stock, createdBy).Scan(
		&item.ID, &item.UUID, &item.Name, &item.Description, &item.ImageURL, &item.PointsRequired, &item.Stock, &item.Status, &item.CreatedBy, &item.CreatedAt, &item.UpdatedAt)
	return &item, err
}

func (r *Repository) Update(uuid string, req *UpdateRewardRequest) (*Reward, error) {
	set := []string{"updated_at = NOW()"}
	args := []interface{}{uuid}
	add := func(column string, value interface{}) {
		set = append(set, fmt.Sprintf("%s = $%d", column, len(args)+1))
		args = append(args, value)
	}
	if req.Name != nil {
		add("name", *req.Name)
	}
	if req.Description != nil {
		add("description", *req.Description)
	}
	if req.PointsRequired != nil {
		add("points_required", *req.PointsRequired)
	}
	if req.Stock != nil {
		add("stock", *req.Stock)
	}
	if req.Status != nil {
		add("status", *req.Status)
	}
	var item Reward
	query := fmt.Sprintf(`UPDATE rewards SET %s WHERE uuid = $1 RETURNING id, uuid, name, description, image_url, points_required, stock, status, created_by, created_at, updated_at`, strings.Join(set, ", "))
	err := r.DB.QueryRow(query, args...).Scan(&item.ID, &item.UUID, &item.Name, &item.Description, &item.ImageURL, &item.PointsRequired, &item.Stock, &item.Status, &item.CreatedBy, &item.CreatedAt, &item.UpdatedAt)
	return &item, err
}

func (r *Repository) UpdateImage(uuid, imageURL string) error {
	if r == nil || r.DB == nil {
		return sql.ErrConnDone
	}
	result, err := r.DB.Exec(`UPDATE rewards SET image_url = $1, updated_at = NOW() WHERE uuid = $2`, imageURL, uuid)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *Repository) Redeem(userID int64, rewardUUID string) (*Redemption, error) {
	tx, err := r.DB.Begin()
	if err != nil {
		return nil, err
	}
	rollback := func() { _ = tx.Rollback() }
	var rewardID int64
	var rewardName string
	var cost int64
	var stock int
	err = tx.QueryRow(`SELECT id, name, points_required, stock FROM rewards WHERE uuid = $1 AND status = 'active' FOR UPDATE`, rewardUUID).Scan(&rewardID, &rewardName, &cost, &stock)
	if err != nil {
		rollback()
		return nil, err
	}
	if stock < 1 {
		rollback()
		return nil, ErrOutOfStock
	}
	var pending bool
	if err = tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM reward_redemptions WHERE user_id = $1 AND reward_id = $2 AND status = 'pending')`, userID, rewardID).Scan(&pending); err != nil {
		rollback()
		return nil, err
	}
	if pending {
		rollback()
		return nil, ErrAlreadyPending
	}
	var balance int64
	if err = tx.QueryRow(`SELECT points_balance FROM users WHERE id = $1 FOR UPDATE`, userID).Scan(&balance); err != nil {
		rollback()
		return nil, err
	}
	if balance < cost {
		rollback()
		return nil, ErrInsufficientPoints
	}
	if _, err = tx.Exec(`UPDATE rewards SET stock = stock - 1, updated_at = NOW() WHERE id = $1`, rewardID); err != nil {
		rollback()
		return nil, err
	}
	newBalance := balance - cost
	if _, err = tx.Exec(`UPDATE users SET points_balance = $1, updated_at = NOW() WHERE id = $2`, newBalance, userID); err != nil {
		rollback()
		return nil, err
	}
	ref := fmt.Sprintf("reward_redemption:%s", rewardUUID)
	if _, err = tx.Exec(`INSERT INTO user_points_ledger (user_id, change_amount, balance_after, reason, reference_id) VALUES ($1, $2, $3, $4, $5)`, userID, -cost, newBalance, "reward_redemption_hold", ref); err != nil {
		rollback()
		return nil, err
	}
	var item Redemption
	err = tx.QueryRow(`
		INSERT INTO reward_redemptions (reward_id, user_id, points_required)
		VALUES ($1, $2, $3)
		RETURNING id, uuid, reward_id, user_id, points_required, status, created_at, updated_at`, rewardID, userID, cost).Scan(
		&item.ID, &item.UUID, &item.RewardID, &item.UserID, &item.PointsRequired, &item.Status, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		rollback()
		return nil, err
	}
	item.RewardName = rewardName
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *Repository) ListRedemptions(userID *int64, status string) ([]Redemption, error) {
	query := `SELECT rr.id, rr.uuid, rr.reward_id, rw.name, rr.user_id, u.email, rr.points_required, rr.status, rr.admin_note, rr.user_code, rr.admin_code, rr.approved_at, rr.rejected_at, rr.created_at, rr.updated_at FROM reward_redemptions rr JOIN rewards rw ON rw.id = rr.reward_id JOIN users u ON u.id = rr.user_id WHERE 1=1`
	args := []interface{}{}
	if userID != nil {
		query += fmt.Sprintf(" AND rr.user_id = $%d", len(args)+1)
		args = append(args, *userID)
	}
	if status != "" {
		query += fmt.Sprintf(" AND rr.status = $%d", len(args)+1)
		args = append(args, status)
	}
	query += ` ORDER BY rr.created_at DESC`
	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRedemptions(rows)
}

func scanRedemptions(rows *sql.Rows) ([]Redemption, error) {
	var result []Redemption
	for rows.Next() {
		var item Redemption
		if err := rows.Scan(&item.ID, &item.UUID, &item.RewardID, &item.RewardName, &item.UserID, &item.UserEmail, &item.PointsRequired, &item.Status, &item.AdminNote, &item.UserCode, &item.AdminCode, &item.ApprovedAt, &item.RejectedAt, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) Decide(uuid string, approve bool, note *string) (*Redemption, error) {
	if !approve && (note == nil || strings.TrimSpace(*note) == "") {
		return nil, ErrNoteRequired
	}
	tx, err := r.DB.Begin()
	if err != nil {
		return nil, err
	}
	rollback := func() { _ = tx.Rollback() }
	var id, userID, rewardID int64
	var cost int64
	var status string
	var rewardName string
	if err = tx.QueryRow(`SELECT rr.id, rr.user_id, rr.reward_id, rr.points_required, rr.status, rw.name FROM reward_redemptions rr JOIN rewards rw ON rw.id = rr.reward_id WHERE rr.uuid = $1 FOR UPDATE`, uuid).Scan(&id, &userID, &rewardID, &cost, &status, &rewardName); err != nil {
		rollback()
		return nil, err
	}
	if status != "pending" {
		rollback()
		return nil, ErrInvalidStatus
	}
	var userCode, adminCode *string
	if approve {
		uc, err := biblicalCode("USER")
		if err != nil {
			rollback()
			return nil, err
		}
		ac, err := biblicalCode("ADMIN")
		if err != nil {
			rollback()
			return nil, err
		}
		userCode, adminCode = &uc, &ac
		_, err = tx.Exec(`UPDATE reward_redemptions SET status = 'approved', admin_note = $1, user_code = $2, admin_code = $3, approved_at = NOW(), updated_at = NOW() WHERE id = $4`, note, uc, ac, id)
	} else {
		_, err = tx.Exec(`UPDATE reward_redemptions SET status = 'rejected', admin_note = $1, rejected_at = NOW(), updated_at = NOW() WHERE id = $2`, strings.TrimSpace(*note), id)
		if err == nil {
			_, err = tx.Exec(`UPDATE rewards SET stock = stock + 1, updated_at = NOW() WHERE id = $1`, rewardID)
		}
		if err == nil {
			var balance int64
			err = tx.QueryRow(`SELECT points_balance FROM users WHERE id = $1 FOR UPDATE`, userID).Scan(&balance)
			if err == nil {
				newBalance := balance + cost
				_, err = tx.Exec(`UPDATE users SET points_balance = $1, updated_at = NOW() WHERE id = $2`, newBalance, userID)
				if err == nil {
					ref := fmt.Sprintf("reward_redemption:%s", uuid)
					_, err = tx.Exec(`INSERT INTO user_points_ledger (user_id, change_amount, balance_after, reason, reference_id) VALUES ($1, $2, $3, $4, $5)`, userID, cost, newBalance, "reward_redemption_refund", ref)
				}
			}
		}
	}
	if err != nil {
		rollback()
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return &Redemption{ID: id, UUID: uuid, RewardID: rewardID, RewardName: rewardName, UserID: userID, PointsRequired: cost, Status: map[bool]string{true: "approved", false: "rejected"}[approve], AdminNote: note, UserCode: userCode, AdminCode: adminCode}, nil
}

func (r *Repository) Complete(uuid, userCode, adminCode string) (*Redemption, error) {
	tx, err := r.DB.Begin()
	if err != nil {
		return nil, err
	}
	rollback := func() { _ = tx.Rollback() }
	var item Redemption
	err = tx.QueryRow(`SELECT rr.id, rr.uuid, rr.reward_id, rw.name, rr.user_id, rr.points_required, rr.status, rr.user_code, rr.admin_code FROM reward_redemptions rr JOIN rewards rw ON rw.id = rr.reward_id WHERE rr.uuid = $1 FOR UPDATE`, uuid).Scan(&item.ID, &item.UUID, &item.RewardID, &item.RewardName, &item.UserID, &item.PointsRequired, &item.Status, &item.UserCode, &item.AdminCode)
	if err != nil {
		rollback()
		return nil, err
	}
	if item.Status != "approved" {
		rollback()
		return nil, ErrInvalidStatus
	}
	if item.UserCode == nil || item.AdminCode == nil || item.UserCode != nil && *item.UserCode != userCode || item.AdminCode != nil && *item.AdminCode != adminCode {
		rollback()
		return nil, ErrInvalidCode
	}
	if _, err = tx.Exec(`UPDATE reward_redemptions SET status = 'completed', updated_at = NOW() WHERE id = $1`, item.ID); err != nil {
		rollback()
		return nil, err
	}
	item.Status = "completed"
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return &item, nil
}

func biblicalCode(prefix string) (string, error) {
	verses := []string{"PS23", "JHN316", "ROM828", "MAT516", "PHP413"}
	buf := make([]byte, 5)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%s-%s", prefix, verses[int(buf[0])%len(verses)], strings.ToUpper(hex.EncodeToString(buf[1:]))), nil
}
