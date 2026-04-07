package role

import (
	"christ-api/pkg/database"
	"database/sql"
)

type RoleRepository struct{}

func (r *RoleRepository) Get(id, siteID *int64) ([]Role, error) {
	if database.DB == nil {
		return nil, sql.ErrConnDone
	}

	var rows *sql.Rows
	var err error
	if id != nil {
		rows, err = database.DB.Query(`SELECT id, name, description, site_id FROM roles WHERE id = $1`, *id)
	} else if siteID != nil {
		rows, err = database.DB.Query(`SELECT id, name, description, site_id FROM roles WHERE site_id = $1`, *siteID)
	} else {
		rows, err = database.DB.Query(`SELECT id, name, description, site_id FROM roles`)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Role
	for rows.Next() {
		var rlt Role
		var desc sql.NullString
		var siteID sql.NullInt64
		if err := rows.Scan(&rlt.ID, &rlt.Name, &desc, &siteID); err != nil {
			return nil, err
		}
		if desc.Valid {
			v := desc.String
			rlt.Description = &v
		}
		if siteID.Valid {
			v := siteID.Int64
			rlt.SiteID = &v
		}
		out = append(out, rlt)
	}
	return out, nil
}

func (r *RoleRepository) Create(name string, description *string, siteID *int64) (*Role, error) {
	if database.DB == nil {
		return nil, sql.ErrConnDone
	}
	query := `INSERT INTO roles (name, description, site_id) VALUES ($1, $2, $3) RETURNING id, name, description, site_id`
	var rl Role
	var desc sql.NullString
	var sID sql.NullInt64
	row := database.DB.QueryRow(query, name, description, siteID)
	if err := row.Scan(&rl.ID, &rl.Name, &desc, &sID); err != nil {
		return nil, err
	}
	if desc.Valid {
		v := desc.String
		rl.Description = &v
	}
	if sID.Valid {
		v := sID.Int64
		rl.SiteID = &v
	}
	return &rl, nil
}

func (r *RoleRepository) Update(id int64, name string, description *string) (*Role, error) {
	if database.DB == nil {
		return nil, sql.ErrConnDone
	}
	query := `UPDATE roles SET name = $1, description = $2 WHERE id = $3 RETURNING id, name, description, site_id`
	var rl Role
	var desc sql.NullString
	var sID sql.NullInt64
	row := database.DB.QueryRow(query, name, description, id)
	if err := row.Scan(&rl.ID, &rl.Name, &desc, &sID); err != nil {
		return nil, err
	}
	if desc.Valid {
		v := desc.String
		rl.Description = &v
	}
	if sID.Valid {
		v := sID.Int64
		rl.SiteID = &v
	}
	return &rl, nil
}
