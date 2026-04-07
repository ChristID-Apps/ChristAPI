package sites

import (
	"christ-api/pkg/database"
	"database/sql"
)

type SiteRepository struct{}

func (r *SiteRepository) GetAll() ([]Site, error) {
	if database.DB == nil {
		return nil, sql.ErrConnDone
	}
	rows, err := database.DB.Query(`SELECT uuid, name, address, created_at, updated_at FROM sites`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Site
	for rows.Next() {
		var s Site
		var addr sql.NullString
		var created sql.NullTime
		var updated sql.NullTime
		if err := rows.Scan(&s.UUID, &s.Name, &addr, &created, &updated); err != nil {
			return nil, err
		}
		if addr.Valid {
			v := addr.String
			s.Address = &v
		}
		if created.Valid {
			s.CreatedAt = &created.Time
		}
		if updated.Valid {
			s.UpdatedAt = &updated.Time
		}
		out = append(out, s)
	}
	return out, nil
}

func (r *SiteRepository) Create(name string, address *string) (*Site, error) {
	if database.DB == nil {
		return nil, sql.ErrConnDone
	}
	query := `INSERT INTO sites (name, address, created_at, updated_at) VALUES ($1, $2, NOW(), NOW()) RETURNING uuid, name, address, created_at, updated_at`
	var s Site
	var addr sql.NullString
	var created sql.NullTime
	var updated sql.NullTime
	row := database.DB.QueryRow(query, name, address)
	if err := row.Scan(&s.UUID, &s.Name, &addr, &created, &updated); err != nil {
		return nil, err
	}
	if addr.Valid {
		v := addr.String
		s.Address = &v
	}
	if created.Valid {
		s.CreatedAt = &created.Time
	}
	if updated.Valid {
		s.UpdatedAt = &updated.Time
	}
	return &s, nil
}

func (r *SiteRepository) Update(uuid string, name string, address *string) (*Site, error) {
	if database.DB == nil {
		return nil, sql.ErrConnDone
	}
	query := `UPDATE sites SET name = $1, address = $2, updated_at = NOW() WHERE uuid = $3 RETURNING uuid, name, address, created_at, updated_at`
	var s Site
	var addr sql.NullString
	var created sql.NullTime
	var updated sql.NullTime
	row := database.DB.QueryRow(query, name, address, uuid)
	if err := row.Scan(&s.UUID, &s.Name, &addr, &created, &updated); err != nil {
		return nil, err
	}
	if addr.Valid {
		v := addr.String
		s.Address = &v
	}
	if created.Valid {
		s.CreatedAt = &created.Time
	}
	if updated.Valid {
		s.UpdatedAt = &updated.Time
	}
	return &s, nil
}
