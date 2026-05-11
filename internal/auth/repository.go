package auth

import (
	"database/sql"

	"christ-api/internal/contacts"
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

func (r *AuthRepository) CreateUser(email, passwordHash string, roleID, siteID, contactID *int64) (*User, error) {
	if r == nil || r.DB == nil {
		return nil, sql.ErrConnDone
	}

	var user User
	query := `INSERT INTO users (email, password_hash, role_id, site_id, contact_id, is_active, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, TRUE, NOW(), NOW()) RETURNING id, uuid, email, password_hash, role_id, contact_id, is_active, last_login_at, created_at, updated_at, site_id`
	row := r.DB.QueryRow(query, email, passwordHash, roleID, siteID, contactID)

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

func (r *AuthRepository) GetLoginUserProfile(userID int64) (*LoginUserResponse, error) {
	if r == nil || r.DB == nil {
		return nil, sql.ErrConnDone
	}

	query := `
		SELECT
			u.id,
			COALESCE(c.full_name, ''),
			u.email,
			COALESCE(ro.name, '')
		FROM users u
		LEFT JOIN contacts c ON c.id = u.contact_id
		LEFT JOIN roles ro ON ro.id = u.role_id
		WHERE u.id = $1
		LIMIT 1`

	var p LoginUserResponse
	row := r.DB.QueryRow(query, userID)
	if err := row.Scan(&p.ID, &p.Name, &p.Email, &p.Role); err != nil {
		return nil, err
	}

	return &p, nil
}

// CreateContactAndUser creates a contact and a user within a single DB transaction.
func (r *AuthRepository) CreateContactAndUser(fullName string, phone *string, address *string, contactSiteID *int64, email, passwordHash string, roleID, userSiteID *int64) (*contacts.Contact, *User, error) {
	if r == nil || r.DB == nil {
		return nil, nil, sql.ErrConnDone
	}

	tx, err := r.DB.Begin()
	if err != nil {
		return nil, nil, err
	}

	// rollback helper
	rollback := func() {
		_ = tx.Rollback()
	}

	// insert contact
	var c contacts.Contact
	contactQuery := `INSERT INTO contacts (full_name, phone, address, created_at, updated_at, site_id) VALUES ($1,$2,$3,NOW(),NOW(),$4) RETURNING id, full_name, phone, address, created_at, updated_at, site_id`
	var phoneN sql.NullString
	var addrN sql.NullString
	var createdN sql.NullTime
	var updatedN sql.NullTime
	var siteIDN sql.NullInt64

	row := tx.QueryRow(contactQuery, fullName, phone, address, contactSiteID)
	if err := row.Scan(&c.ID, &c.FullName, &phoneN, &addrN, &createdN, &updatedN, &siteIDN); err != nil {
		rollback()
		return nil, nil, err
	}
	if phoneN.Valid {
		v := phoneN.String
		c.Phone = &v
	}
	if addrN.Valid {
		v := addrN.String
		c.Address = &v
	}
	if createdN.Valid {
		c.CreatedAt = &createdN.Time
	}
	if updatedN.Valid {
		c.UpdatedAt = &updatedN.Time
	}
	if siteIDN.Valid {
		v := siteIDN.Int64
		c.SiteID = &v
	}

	// insert user with contact_id
	var user User
	userQuery := `INSERT INTO users (email, password_hash, role_id, site_id, contact_id, is_active, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,TRUE,NOW(),NOW()) RETURNING id, uuid, email, password_hash, role_id, contact_id, is_active, last_login_at, created_at, updated_at, site_id`

	var roleIDN sql.NullInt64
	var contactIDN sql.NullInt64
	var siteIDUN sql.NullInt64
	var lastLoginN sql.NullTime
	var createdUN sql.NullTime
	var updatedUN sql.NullTime

	row = tx.QueryRow(userQuery, email, passwordHash, roleID, userSiteID, c.ID)
	if err := row.Scan(&user.ID, &user.UUID, &user.Email, &user.Password, &roleIDN, &contactIDN, &user.IsActive, &lastLoginN, &createdUN, &updatedUN, &siteIDUN); err != nil {
		rollback()
		return nil, nil, err
	}

	if roleIDN.Valid {
		v := roleIDN.Int64
		user.RoleID = &v
	}
	if contactIDN.Valid {
		v := contactIDN.Int64
		user.ContactID = &v
	}
	if siteIDUN.Valid {
		v := siteIDUN.Int64
		user.SiteID = &v
	}
	if lastLoginN.Valid {
		user.LastLoginAt = &lastLoginN.Time
	}
	if createdUN.Valid {
		user.CreatedAt = &createdUN.Time
	}
	if updatedUN.Valid {
		user.UpdatedAt = &updatedUN.Time
	}

	if err := tx.Commit(); err != nil {
		rollback()
		return nil, nil, err
	}

	return &c, &user, nil
}
