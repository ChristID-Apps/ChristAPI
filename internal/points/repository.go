package points

import (
	"christ-api/pkg/database"
	"database/sql"
	"errors"
)

type Repository struct{}

func (r *Repository) GetBalance(userID int64, siteID *int64) (int64, error) {
	if database.DB == nil {
		return 0, sql.ErrConnDone
	}

	var balance int64
	var err error
	if siteID != nil {
		err = database.DB.QueryRow(`SELECT points_balance FROM users WHERE id = $1 AND site_id = $2 LIMIT 1`, userID, siteID).Scan(&balance)
	} else {
		err = database.DB.QueryRow(`SELECT points_balance FROM users WHERE id = $1 LIMIT 1`, userID).Scan(&balance)
	}
	if err != nil {
		return 0, err
	}

	return balance, nil
}

func (r *Repository) GetHistory(userID int64, siteID *int64, offset, limit int) ([]LedgerEntry, error) {
	if database.DB == nil {
		return nil, sql.ErrConnDone
	}
	if offset < 0 {
		offset = 0
	}
	if limit < 1 {
		limit = 10
	}

	var rows *sql.Rows
	var err error
	if siteID != nil {
		rows, err = database.DB.Query(`
			SELECT l.id, l.user_id, l.change_amount, l.balance_after, l.reason, l.reference_id, l.created_at
			FROM user_points_ledger l
			JOIN users u ON u.id = l.user_id
			WHERE l.user_id = $1 AND u.site_id = $2
			ORDER BY l.id DESC
			LIMIT $3 OFFSET $4`, userID, siteID, limit, offset)
	} else {
		rows, err = database.DB.Query(`
			SELECT id, user_id, change_amount, balance_after, reason, reference_id, created_at
			FROM user_points_ledger
			WHERE user_id = $1
			ORDER BY id DESC
			LIMIT $2 OFFSET $3`, userID, limit, offset)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []LedgerEntry
	for rows.Next() {
		var e LedgerEntry
		var ref sql.NullString
		var created sql.NullTime
		if err := rows.Scan(&e.ID, &e.UserID, &e.ChangeAmount, &e.BalanceAfter, &e.Reason, &ref, &created); err != nil {
			return nil, err
		}
		if ref.Valid {
			v := ref.String
			e.ReferenceID = &v
		}
		if created.Valid {
			e.CreatedAt = &created.Time
		}
		out = append(out, e)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func (r *Repository) applyDelta(userID, delta int64, reason string, referenceID *string) (*LedgerEntry, error) {
	if database.DB == nil {
		return nil, sql.ErrConnDone
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return nil, err
	}

	rollback := func() {
		_ = tx.Rollback()
	}

	var current int64
	err = tx.QueryRow(`SELECT points_balance FROM users WHERE id = $1 FOR UPDATE`, userID).Scan(&current)
	if err != nil {
		rollback()
		return nil, err
	}

	newBalance := current + delta
	if newBalance < 0 {
		rollback()
		return nil, errors.New("insufficient points")
	}

	if _, err := tx.Exec(`UPDATE users SET points_balance = $1, updated_at = NOW() WHERE id = $2`, newBalance, userID); err != nil {
		rollback()
		return nil, err
	}

	var e LedgerEntry
	var ref sql.NullString
	var created sql.NullTime
	err = tx.QueryRow(`
		INSERT INTO user_points_ledger (user_id, change_amount, balance_after, reason, reference_id, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		RETURNING id, user_id, change_amount, balance_after, reason, reference_id, created_at`,
		userID, delta, newBalance, reason, referenceID,
	).Scan(&e.ID, &e.UserID, &e.ChangeAmount, &e.BalanceAfter, &e.Reason, &ref, &created)
	if err != nil {
		rollback()
		return nil, err
	}

	if ref.Valid {
		v := ref.String
		e.ReferenceID = &v
	}
	if created.Valid {
		e.CreatedAt = &created.Time
	}

	if err := tx.Commit(); err != nil {
		rollback()
		return nil, err
	}

	return &e, nil
}

func (r *Repository) Earn(userID, amount int64, reason string, referenceID *string) (*LedgerEntry, error) {
	return r.applyDelta(userID, amount, reason, referenceID)
}

func (r *Repository) Spend(userID, amount int64, reason string, referenceID *string) (*LedgerEntry, error) {
	return r.applyDelta(userID, -amount, reason, referenceID)
}

func (r *Repository) ListBalances(siteID *int64, offset, limit int) ([]UserBalance, error) {
	if database.DB == nil {
		return nil, sql.ErrConnDone
	}
	if offset < 0 {
		offset = 0
	}
	if limit < 1 {
		limit = 10
	}

	var rows *sql.Rows
	var err error
	if siteID != nil {
		rows, err = database.DB.Query(`
			SELECT id, email, points_balance, site_id
			FROM users
			WHERE site_id = $1
			ORDER BY points_balance DESC, id ASC
			LIMIT $2 OFFSET $3`, siteID, limit, offset)
	} else {
		rows, err = database.DB.Query(`
			SELECT id, email, points_balance, site_id
			FROM users
			ORDER BY points_balance DESC, id ASC
			LIMIT $1 OFFSET $2`, limit, offset)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []UserBalance
	for rows.Next() {
		var u UserBalance
		var site sql.NullInt64
		if err := rows.Scan(&u.UserID, &u.Email, &u.Points, &site); err != nil {
			return nil, err
		}
		if site.Valid {
			v := site.Int64
			u.SiteID = &v
		}
		out = append(out, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}
