package auth

import (
	"database/sql"
)

type AuthRepository struct {
	DB *sql.DB
}

func (r *AuthRepository) FindByEmail(email string) (*User, error) {
	var user User
	if r == nil || r.DB == nil {
		return nil, sql.ErrConnDone
	}

	query := `SELECT id, uuid, email, password_hash, role_id, contact_id, is_active, last_login_at, created_at, updated_at, site_id FROM users WHERE email = $1 LIMIT 1`
	row := r.DB.QueryRow(query, email)
	var roleID sql.NullInt64
	var contactID sql.NullInt64
	var siteID sql.NullInt64
	var lastLogin sql.NullTime
	var createdAt sql.NullTime
	var updatedAt sql.NullTime
	err := row.Scan(&user.ID, &user.UUID, &user.Email, &user.Password, &roleID, &contactID, &user.IsActive, &lastLogin, &createdAt, &updatedAt, &siteID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if roleID.Valid {
		v := roleID.Int64
		user.RoleID = &v
	}
	if contactID.Valid {
		v := contactID.Int64
		user.ContactID = &v
	}
	if siteID.Valid {
		v := siteID.Int64
		user.SiteID = &v
	}
	if lastLogin.Valid {
		user.LastLoginAt = &lastLogin.Time
	}
	if createdAt.Valid {
		user.CreatedAt = &createdAt.Time
	}
	if updatedAt.Valid {
		user.UpdatedAt = &updatedAt.Time
	}

	return &user, nil
}

func (r *AuthRepository) CreateUser(email, passwordHash string, siteID, contactID *int64) (*User, error) {
	if r == nil || r.DB == nil {
		return nil, sql.ErrConnDone
	}

	var user User
	query := `INSERT INTO users (email, password_hash, site_id, contact_id, is_active, created_at, updated_at) VALUES ($1, $2, $3, $4, TRUE, NOW(), NOW()) RETURNING id, uuid, email, password_hash, role_id, contact_id, is_active, last_login_at, created_at, updated_at, site_id`
	row := r.DB.QueryRow(query, email, passwordHash, siteID, contactID)

	var roleIDN sql.NullInt64
	var contactIDN sql.NullInt64
	var siteIDN sql.NullInt64
	var lastLoginN sql.NullTime
	var createdAt sql.NullTime
	var updatedAt sql.NullTime

	err := row.Scan(&user.ID, &user.UUID, &user.Email, &user.Password, &roleIDN, &contactIDN, &user.IsActive, &lastLoginN, &createdAt, &updatedAt, &siteIDN)
	if err != nil {
		return nil, err
	}

	if roleIDN.Valid {
		v := roleIDN.Int64
		user.RoleID = &v
	}
	if contactIDN.Valid {
		v := contactIDN.Int64
		user.ContactID = &v
	}
	if siteIDN.Valid {
		v := siteIDN.Int64
		user.SiteID = &v
	}
	if lastLoginN.Valid {
		user.LastLoginAt = &lastLoginN.Time
	}
	if createdAt.Valid {
		user.CreatedAt = &createdAt.Time
	}
	if updatedAt.Valid {
		user.UpdatedAt = &updatedAt.Time
	}

	return &user, nil
}

func (r *AuthRepository) UpdateLastLoginAndSite(userID int64, siteID *int64) error {
	if r == nil || r.DB == nil {
		return sql.ErrConnDone
	}

	// if siteID is nil, keep existing site_id
	query := `UPDATE users SET last_login_at = NOW(), site_id = COALESCE($2, site_id) WHERE id = $1`
	_, err := r.DB.Exec(query, userID, siteID)
	return err
}
