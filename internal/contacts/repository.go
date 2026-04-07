package contacts

import (
	"christ-api/pkg/database"
	"database/sql"
)

type ContactRepository struct{}

func (r *ContactRepository) GetAll() ([]Contact, error) {
	if database.DB == nil {
		return nil, sql.ErrConnDone
	}
	rows, err := database.DB.Query(`SELECT id, full_name, phone, address, created_at, updated_at, site_id FROM contacts`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Contact
	for rows.Next() {
		var c Contact
		var phone sql.NullString
		var addr sql.NullString
		var created sql.NullTime
		var updated sql.NullTime
		var siteID sql.NullInt64
		if err := rows.Scan(&c.ID, &c.FullName, &phone, &addr, &created, &updated, &siteID); err != nil {
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
		out = append(out, c)
	}
	return out, nil
}

func (r *ContactRepository) Create(fullName string, phone *string, address *string, siteID *int64) (*Contact, error) {
	if database.DB == nil {
		return nil, sql.ErrConnDone
	}
	query := `INSERT INTO contacts (full_name, phone, address, created_at, updated_at, site_id) VALUES ($1,$2,$3,NOW(),NOW(),$4) RETURNING id, full_name, phone, address, created_at, updated_at, site_id`
	var c Contact
	var phoneN sql.NullString
	var addr sql.NullString
	var created sql.NullTime
	var updated sql.NullTime
	var siteIDn sql.NullInt64
	row := database.DB.QueryRow(query, fullName, phone, address, siteID)
	if err := row.Scan(&c.ID, &c.FullName, &phoneN, &addr, &created, &updated, &siteIDn); err != nil {
		return nil, err
	}
	if phoneN.Valid {
		v := phoneN.String
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
	if siteIDn.Valid {
		v := siteIDn.Int64
		c.SiteID = &v
	}
	return &c, nil
}

func (r *ContactRepository) Update(id int64, fullName string, phone *string, address *string) (*Contact, error) {
	if database.DB == nil {
		return nil, sql.ErrConnDone
	}
	query := `UPDATE contacts SET full_name=$1, phone=$2, address=$3, updated_at=NOW() WHERE id=$4 RETURNING id, full_name, phone, address, created_at, updated_at, site_id`
	var c Contact
	var phoneN sql.NullString
	var addr sql.NullString
	var created sql.NullTime
	var updated sql.NullTime
	var siteIDn sql.NullInt64
	row := database.DB.QueryRow(query, fullName, phone, address, id)
	if err := row.Scan(&c.ID, &c.FullName, &phoneN, &addr, &created, &updated, &siteIDn); err != nil {
		return nil, err
	}
	if phoneN.Valid {
		v := phoneN.String
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
	if siteIDn.Valid {
		v := siteIDn.Int64
		c.SiteID = &v
	}
	return &c, nil
}
