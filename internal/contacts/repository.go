package contacts

import (
	"christ-api/pkg/database"
	"database/sql"
)

type ContactRepository struct{}

func scanContactRow(row interface{ Scan(dest ...any) error }) (*Contact, error) {
	var c Contact
	var phone sql.NullString
	var addr sql.NullString
	var created sql.NullTime
	var updated sql.NullTime
	var siteID sql.NullInt64
	var deleted sql.NullTime
	if err := row.Scan(&c.ID, &c.FullName, &phone, &addr, &created, &updated, &siteID, &deleted); err != nil {
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
	rows, err := database.DB.Query(`SELECT id, full_name, phone, address, created_at, updated_at, site_id, deleted_at FROM contacts WHERE deleted_at IS NULL ORDER BY id DESC LIMIT $1 OFFSET $2`, limit, offset)
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
	row := database.DB.QueryRow(`SELECT id, full_name, phone, address, created_at, updated_at, site_id, deleted_at FROM contacts WHERE id = $1 AND deleted_at IS NULL LIMIT 1`, id)
	return scanContactRow(row)
}

func (r *ContactRepository) Create(fullName string, phone *string, address *string, siteID *int64) (*Contact, error) {
	if database.DB == nil {
		return nil, sql.ErrConnDone
	}
	query := `INSERT INTO contacts (full_name, phone, address, created_at, updated_at, site_id) VALUES ($1,$2,$3,NOW(),NOW(),$4) RETURNING id, full_name, phone, address, created_at, updated_at, site_id, deleted_at`
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
	query := `UPDATE contacts SET full_name=$1, phone=$2, address=$3, site_id=$4, updated_at=NOW() WHERE id=$5 AND deleted_at IS NULL RETURNING id, full_name, phone, address, created_at, updated_at, site_id, deleted_at`
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
	query := `UPDATE contacts SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL RETURNING id, full_name, phone, address, created_at, updated_at, site_id, deleted_at`
	row := database.DB.QueryRow(query, id)
	c, err := scanContactRow(row)
	if err != nil {
		return nil, err
	}
	return c, nil
}
