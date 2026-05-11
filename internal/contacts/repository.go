package contacts

import (
	"christ-api/pkg/database"
	"database/sql"
)

type ContactRepository struct{}

const contactSelectColumns = `
	c.id,
	c.full_name,
	c.phone,
	c.address,
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
	var email sql.NullString
	var points sql.NullInt64
	var created sql.NullTime
	var updated sql.NullTime
	var siteID sql.NullInt64
	var deleted sql.NullTime
	if err := row.Scan(&c.ID, &c.FullName, &phone, &addr, &email, &points, &created, &updated, &siteID, &deleted); err != nil {
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
	if database.DB == nil {
		return nil, sql.ErrConnDone
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit
	rows, err := database.DB.Query(`
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
	if database.DB == nil {
		return nil, sql.ErrConnDone
	}
	row := database.DB.QueryRow(`
		SELECT `+contactSelectColumns+`
		FROM contacts c
		LEFT JOIN users u ON u.contact_id = c.id
		WHERE c.id = $1 AND c.deleted_at IS NULL
		LIMIT 1`, id)
	return scanContactRow(row)
}

func (r *ContactRepository) Create(fullName string, phone *string, address *string, siteID *int64) (*Contact, error) {
	if database.DB == nil {
		return nil, sql.ErrConnDone
	}
	query := `
		WITH inserted AS (
			INSERT INTO contacts (full_name, phone, address, created_at, updated_at, site_id)
			VALUES ($1,$2,$3,NOW(),NOW(),$4)
			RETURNING id, full_name, phone, address, created_at, updated_at, site_id, deleted_at
		)
		SELECT
			i.id,
			i.full_name,
			i.phone,
			i.address,
			u.email,
			u.points_balance,
			i.created_at,
			i.updated_at,
			i.site_id,
			i.deleted_at
		FROM inserted i
		LEFT JOIN users u ON u.contact_id = i.id`
	row := database.DB.QueryRow(query, fullName, phone, address, siteID)
	c, err := scanContactRow(row)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *ContactRepository) Update(id int64, fullName string, phone *string, address *string, siteID *int64) (*Contact, error) {
	if database.DB == nil {
		return nil, sql.ErrConnDone
	}
	query := `
		WITH updated AS (
			UPDATE contacts
			SET full_name=$1, phone=$2, address=$3, site_id=$4, updated_at=NOW()
			WHERE id=$5 AND deleted_at IS NULL
			RETURNING id, full_name, phone, address, created_at, updated_at, site_id, deleted_at
		)
		SELECT
			upt.id,
			upt.full_name,
			upt.phone,
			upt.address,
			u.email,
			u.points_balance,
			upt.created_at,
			upt.updated_at,
			upt.site_id,
			upt.deleted_at
		FROM updated upt
		LEFT JOIN users u ON u.contact_id = upt.id`
	row := database.DB.QueryRow(query, fullName, phone, address, siteID, id)
	c, err := scanContactRow(row)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *ContactRepository) SoftDelete(id int64) (*Contact, error) {
	if database.DB == nil {
		return nil, sql.ErrConnDone
	}
	query := `
		WITH deleted AS (
			UPDATE contacts
			SET deleted_at = NOW(), updated_at = NOW()
			WHERE id = $1 AND deleted_at IS NULL
			RETURNING id, full_name, phone, address, created_at, updated_at, site_id, deleted_at
		)
		SELECT
			d.id,
			d.full_name,
			d.phone,
			d.address,
			u.email,
			u.points_balance,
			d.created_at,
			d.updated_at,
			d.site_id,
			d.deleted_at
		FROM deleted d
		LEFT JOIN users u ON u.contact_id = d.id`
	row := database.DB.QueryRow(query, id)
	c, err := scanContactRow(row)
	if err != nil {
		return nil, err
	}
	return c, nil
}
