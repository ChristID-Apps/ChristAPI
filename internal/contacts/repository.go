package contacts

import (
	"database/sql"
	"fmt"
	"strings"
)

type ContactRepository struct {
	DB *sql.DB
}

const contactSelectColumns = `
	c.id,
	c.full_name,
	c.phone,
	c.address,
	c.profile_photo_url,
	u.email,
	u.points_balance,
	c.created_at,
	c.updated_at,
	c.site_id,
	c.deleted_at`

func scanContactRow(row interface{ Scan(dest ...any) error }) (*Contact, error) {
	var c Contact
	var phone sql.NullString
	var addr sql.NullString
	var profilePhoto sql.NullString
	var email sql.NullString
	var points sql.NullInt64
	var created sql.NullTime
	var updated sql.NullTime
	var siteID sql.NullInt64
	var deleted sql.NullTime
	if err := row.Scan(&c.ID, &c.FullName, &phone, &addr, &profilePhoto, &email, &points, &created, &updated, &siteID, &deleted); err != nil {
		return nil, err
	}
	if phone.Valid {
		v := phone.String
		c.Phone = &v
	}
	if addr.Valid {
		v := addr.String
		c.Address = &v
	}
	if profilePhoto.Valid {
		v := profilePhoto.String
		c.ProfilePhotoURL = &v
	}
	if email.Valid {
		v := email.String
		c.Email = &v
	}
	if points.Valid {
		v := points.Int64
		c.Points = &v
	}
	if created.Valid {
		c.CreatedAt = &created.Time
	}
	if updated.Valid {
		c.UpdatedAt = &updated.Time
	}
	if siteID.Valid {
		v := siteID.Int64
		c.SiteID = &v
	}
	if deleted.Valid {
		c.DeletedAt = &deleted.Time
	}
	return &c, nil
}

func (r *ContactRepository) List(page, limit int) ([]Contact, error) {
	if r == nil || r.DB == nil {
		return nil, sql.ErrConnDone
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit
	rows, err := r.DB.Query(`
		SELECT `+contactSelectColumns+`
		FROM contacts c
		LEFT JOIN users u ON u.contact_id = c.id
		WHERE c.deleted_at IS NULL
		ORDER BY c.id DESC
		LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Contact
	for rows.Next() {
		c, err := scanContactRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *ContactRepository) GetByID(id int64) (*Contact, error) {
	if r == nil || r.DB == nil {
		return nil, sql.ErrConnDone
	}
	row := r.DB.QueryRow(`
		SELECT `+contactSelectColumns+`
		FROM contacts c
		LEFT JOIN users u ON u.contact_id = c.id
		WHERE c.id = $1 AND c.deleted_at IS NULL
		LIMIT 1`, id)
	return scanContactRow(row)
}

func (r *ContactRepository) Create(fullName string, phone *string, address *string, siteID *int64) (*Contact, error) {
	if r == nil || r.DB == nil {
		return nil, sql.ErrConnDone
	}
	query := `
		WITH inserted AS (
			INSERT INTO contacts (full_name, phone, address, created_at, updated_at, site_id)
			VALUES ($1,$2,$3,NOW(),NOW(),$4)
			RETURNING id, full_name, phone, address, profile_photo_url, created_at, updated_at, site_id, deleted_at
		)
		SELECT
			i.id,
			i.full_name,
			i.phone,
			i.address,
			i.profile_photo_url,
			u.email,
			u.points_balance,
			i.created_at,
			i.updated_at,
			i.site_id,
			i.deleted_at
		FROM inserted i
		LEFT JOIN users u ON u.contact_id = i.id`
	row := r.DB.QueryRow(query, fullName, phone, address, siteID)
	c, err := scanContactRow(row)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *ContactRepository) Update(id int64, req *UpdateContactRequest) (*Contact, error) {
	if r == nil || r.DB == nil {
		return nil, sql.ErrConnDone
	}
	columns := []string{}
	args := []interface{}{}
	add := func(name string, value interface{}) {
		columns = append(columns, name+fmt.Sprintf("=$%d", len(args)+1))
		args = append(args, value)
	}
	if req.Present["full_name"] {
		add("full_name", req.FullName)
	}
	if req.Present["phone"] {
		add("phone", req.Phone)
	}
	if req.Present["address"] {
		add("address", req.Address)
	}
	if req.Present["site_id"] {
		add("site_id", req.SiteID)
	}
	if len(columns) == 0 {
		columns = append(columns, "updated_at=NOW()")
	}
	args = append(args, id)
	query := `
		WITH updated AS (
			UPDATE contacts
			SET ` + strings.Join(columns, ", ") + `, updated_at=NOW()
			WHERE id=$` + fmt.Sprint(len(args)) + ` AND deleted_at IS NULL
			RETURNING id, full_name, phone, address, profile_photo_url, created_at, updated_at, site_id, deleted_at
		)
		SELECT
			upt.id,
			upt.full_name,
			upt.phone,
			upt.address,
			upt.profile_photo_url,
			u.email,
			u.points_balance,
			upt.created_at,
			upt.updated_at,
			upt.site_id,
			upt.deleted_at
		FROM updated upt
		LEFT JOIN users u ON u.contact_id = upt.id`
	row := r.DB.QueryRow(query, args...)
	c, err := scanContactRow(row)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *ContactRepository) SoftDelete(id int64) (*Contact, error) {
	if r == nil || r.DB == nil {
		return nil, sql.ErrConnDone
	}
	query := `
		WITH deleted AS (
			UPDATE contacts
			SET deleted_at = NOW(), updated_at = NOW()
			WHERE id = $1 AND deleted_at IS NULL
			RETURNING id, full_name, phone, address, profile_photo_url, created_at, updated_at, site_id, deleted_at
		)
		SELECT
			d.id,
			d.full_name,
			d.phone,
			d.address,
			d.profile_photo_url,
			u.email,
			u.points_balance,
			d.created_at,
			d.updated_at,
			d.site_id,
			d.deleted_at
		FROM deleted d
		LEFT JOIN users u ON u.contact_id = d.id`
	row := r.DB.QueryRow(query, id)
	c, err := scanContactRow(row)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *ContactRepository) GetProfile(userID int64) (*Profile, error) {
	if r == nil || r.DB == nil {
		return nil, sql.ErrConnDone
	}

	query := `
		SELECT u.id, COALESCE(c.full_name, ''), u.username, u.email,
			c.phone, c.address, c.profile_photo_url, COALESCE(ro.name, ''),
			COALESCE(u.points_balance, 0), u.approval_status, u.is_active, u.site_id
		FROM users u
		LEFT JOIN contacts c ON c.id = u.contact_id AND c.deleted_at IS NULL
		LEFT JOIN roles ro ON ro.id = u.role_id
		WHERE u.id = $1
		LIMIT 1`

	var profile Profile
	var username, phone, address, photo sql.NullString
	var siteID sql.NullInt64
	if err := r.DB.QueryRow(query, userID).Scan(
		&profile.ID, &profile.FullName, &username, &profile.Email,
		&phone, &address, &photo, &profile.Role, &profile.Points,
		&profile.ApprovalStatus, &profile.IsActive, &siteID,
	); err != nil {
		return nil, err
	}
	if username.Valid {
		profile.Username = &username.String
	}
	if phone.Valid {
		profile.Phone = &phone.String
	}
	if address.Valid {
		profile.Address = &address.String
	}
	if photo.Valid {
		profile.ProfilePhotoURL = &photo.String
	}
	if siteID.Valid {
		profile.SiteID = &siteID.Int64
	}
	return &profile, nil
}

func (r *ContactRepository) UpdateProfile(userID int64, req *ProfileUpdateRequest) (*Profile, error) {
	if r == nil || r.DB == nil {
		return nil, sql.ErrConnDone
	}

	columns := []string{}
	args := []interface{}{}
	add := func(name string, value interface{}) {
		columns = append(columns, name+fmt.Sprintf("=$%d", len(args)+1))
		args = append(args, value)
	}
	if req.Present["full_name"] {
		add("full_name", req.FullName)
	}
	if req.Present["phone"] {
		add("phone", req.Phone)
	}
	if req.Present["address"] {
		add("address", req.Address)
	}
	if len(columns) == 0 {
		columns = append(columns, "updated_at=NOW()")
	}
	args = append(args, userID)

	query := `UPDATE contacts
		SET ` + strings.Join(columns, ", ") + `, updated_at=NOW()
		WHERE id = (SELECT contact_id FROM users WHERE id = $` + fmt.Sprint(len(args)) + `)
		  AND deleted_at IS NULL`
	result, err := r.DB.Exec(query, args...)
	if err != nil {
		return nil, err
	}
	if affected, err := result.RowsAffected(); err != nil {
		return nil, err
	} else if affected == 0 {
		return nil, sql.ErrNoRows
	}
	return r.GetProfile(userID)
}

func (r *ContactRepository) UpdateProfilePhoto(userID int64, imageURL string) (*Profile, error) {
	if r == nil || r.DB == nil {
		return nil, sql.ErrConnDone
	}
	result, err := r.DB.Exec(`
		UPDATE contacts
		SET profile_photo_url = $1, updated_at = NOW()
		WHERE id = (SELECT contact_id FROM users WHERE id = $2)
		  AND deleted_at IS NULL`, imageURL, userID)
	if err != nil {
		return nil, err
	}
	if affected, err := result.RowsAffected(); err != nil {
		return nil, err
	} else if affected == 0 {
		return nil, sql.ErrNoRows
	}
	return r.GetProfile(userID)
}
