package sites

import (
	"database/sql"
	"fmt"
	"strings"
)

type SiteRepository struct {
	DB *sql.DB
}

func (r *SiteRepository) GetAll() ([]Site, error) {
	if r == nil || r.DB == nil {
		return nil, sql.ErrConnDone
	}
	rows, err := r.DB.Query(`SELECT uuid, name, address, latitude, longitude, created_at, updated_at FROM sites`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Site
	for rows.Next() {
		var s Site
		var addr sql.NullString
		var latitude, longitude sql.NullFloat64
		var created sql.NullTime
		var updated sql.NullTime
		if err := rows.Scan(&s.UUID, &s.Name, &addr, &latitude, &longitude, &created, &updated); err != nil {
			return nil, err
		}
		if addr.Valid {
			v := addr.String
			s.Address = &v
		}
		if latitude.Valid {
			v := latitude.Float64
			s.Latitude = &v
		}
		if longitude.Valid {
			v := longitude.Float64
			s.Longitude = &v
		}
		if created.Valid {
			s.CreatedAt = &created.Time
		}
		if updated.Valid {
			s.UpdatedAt = &updated.Time
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *SiteRepository) Create(name string, address *string, latitude, longitude *float64) (*Site, error) {
	if r == nil || r.DB == nil {
		return nil, sql.ErrConnDone
	}
	query := `INSERT INTO sites (name, address, latitude, longitude, created_at, updated_at) VALUES ($1, $2, $3, $4, NOW(), NOW()) RETURNING uuid, name, address, latitude, longitude, created_at, updated_at`
	var s Site
	var addr sql.NullString
	var latitudeN, longitudeN sql.NullFloat64
	var created sql.NullTime
	var updated sql.NullTime
	row := r.DB.QueryRow(query, name, address, latitude, longitude)
	if err := row.Scan(&s.UUID, &s.Name, &addr, &latitudeN, &longitudeN, &created, &updated); err != nil {
		return nil, err
	}
	if addr.Valid {
		v := addr.String
		s.Address = &v
	}
	if latitudeN.Valid {
		v := latitudeN.Float64
		s.Latitude = &v
	}
	if longitudeN.Valid {
		v := longitudeN.Float64
		s.Longitude = &v
	}
	if created.Valid {
		s.CreatedAt = &created.Time
	}
	if updated.Valid {
		s.UpdatedAt = &updated.Time
	}
	return &s, nil
}

func (r *SiteRepository) Update(uuid string, req *UpdateSiteRequest) (*Site, error) {
	if r == nil || r.DB == nil {
		return nil, sql.ErrConnDone
	}
	columns := []string{}
	args := []interface{}{}
	add := func(name string, value interface{}) {
		columns = append(columns, name+fmt.Sprintf("=$%d", len(args)+1))
		args = append(args, value)
	}
	if req.Present["name"] {
		add("name", req.Name)
	}
	if req.Present["address"] {
		add("address", req.Address)
	}
	if req.Present["latitude"] {
		add("latitude", req.Latitude)
	}
	if req.Present["longitude"] {
		add("longitude", req.Longitude)
	}
	if len(columns) == 0 {
		columns = append(columns, "name=name")
	}
	args = append(args, uuid)
	query := `UPDATE sites SET ` + strings.Join(columns, ", ") + `, updated_at = NOW() WHERE uuid = $` + fmt.Sprint(len(args)) + ` RETURNING uuid, name, address, latitude, longitude, created_at, updated_at`
	var s Site
	var addr sql.NullString
	var latitudeN, longitudeN sql.NullFloat64
	var created sql.NullTime
	var updated sql.NullTime
	row := r.DB.QueryRow(query, args...)
	if err := row.Scan(&s.UUID, &s.Name, &addr, &latitudeN, &longitudeN, &created, &updated); err != nil {
		return nil, err
	}
	if addr.Valid {
		v := addr.String
		s.Address = &v
	}
	if latitudeN.Valid {
		v := latitudeN.Float64
		s.Latitude = &v
	}
	if longitudeN.Valid {
		v := longitudeN.Float64
		s.Longitude = &v
	}
	if created.Valid {
		s.CreatedAt = &created.Time
	}
	if updated.Valid {
		s.UpdatedAt = &updated.Time
	}
	return &s, nil
}
