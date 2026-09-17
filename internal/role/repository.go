package role

import (
	"database/sql"
	"fmt"
	"strings"

	"christ-api/internal/role/dto/requests"
)

type RoleRepository struct {
	DB *sql.DB
}

func (r *RoleRepository) Get(id, siteID *int64) ([]Role, error) {
	if r == nil || r.DB == nil {
		return nil, sql.ErrConnDone
	}

	var rows *sql.Rows
	var err error

	// Query based on the provided parameters
	// kalau semisal ga ngasih request id maka ambil semua role yang ada di database
	if id != nil {
		rows, err = r.DB.Query(`SELECT id, name, code, description, site_id FROM roles WHERE id = $1`, *id)

		// nah kalau ngasih request site_id maka ambil semua role yang ada di database berdasarkan site_id
	} else if siteID != nil {
		rows, err = r.DB.Query(`SELECT id, name, code, description, site_id FROM roles WHERE site_id = $1`, *siteID)
		// kalau ga ngasih request id dan site_id maka ambil semua role yang ada di database
	} else {
		rows, err = r.DB.Query(`SELECT id, name, code, description, site_id FROM roles`)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Role
	for rows.Next() {
		var role Role
		var code sql.NullString
		var desc sql.NullString
		var siteID sql.NullInt64
		if err := rows.Scan(
			&role.ID,
			&role.Name,
			&code,
			&desc,
			&siteID); err != nil {
			return nil, err
		}

		if desc.Valid {
			v := desc.String
			role.Description = &v
		}
		if code.Valid {
			role.Code = code.String
		}
		if siteID.Valid {
			v := siteID.Int64
			role.SiteID = &v
		}
		out = append(out, role)
	}
	return out, nil
}

func (r *RoleRepository) Create(name string, code string, description *string, siteID *int64) (*Role, error) {
	if r == nil || r.DB == nil {
		return nil, sql.ErrConnDone
	}

	query := `INSERT INTO roles (name, code, description, site_id) VALUES ($1, $2, $3, $4) RETURNING id, name, code, description, site_id`
	var role Role
	var codeN sql.NullString
	var desc sql.NullString
	var sID sql.NullInt64
	row := r.DB.QueryRow(query, name, code, description, siteID)
	if err := row.Scan(&role.ID, &role.Name, &codeN, &desc, &sID); err != nil {
		return nil, err
	}
	if codeN.Valid {
		role.Code = codeN.String
	}
	if desc.Valid {
		v := desc.String
		role.Description = &v
	}
	if sID.Valid {
		v := sID.Int64
		role.SiteID = &v
	}
	return &role, nil
}

func (r *RoleRepository) Update(id int64, req *requests.UpdateRoleRequest) (*Role, error) {
	if r == nil || r.DB == nil {
		return nil, sql.ErrConnDone
	}

	columns := []string{}
	args := []interface{}{}
	add := func(name string, value interface{}) {
		columns = append(columns, name+fmt.Sprintf("=$%d", len(args)+1))
		switch typed := value.(type) {
		case *string:
			if typed == nil {
				args = append(args, nil)
			} else {
				args = append(args, *typed)
			}
		default:
			args = append(args, value)
		}
	}
	if req.Present["name"] {
		add("name", req.Name)
	}
	if req.Present["code"] {
		add("code", req.Code)
	}
	if req.Present["description"] {
		add("description", req.Description)
	}
	if len(columns) == 0 {
		columns = append(columns, "name=name")
	}
	args = append(args, id)
	query := `UPDATE roles SET ` + strings.Join(columns, ", ") + ` WHERE id = $` + fmt.Sprint(len(args)) + ` RETURNING id, name, code, description, site_id`
	var role Role
	var codeN sql.NullString
	var desc sql.NullString
	var sID sql.NullInt64
	row := r.DB.QueryRow(query, args...)
	if err := row.Scan(&role.ID, &role.Name, &codeN, &desc, &sID); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("role %d not found", id)
		}
		return nil, err
	}
	if codeN.Valid {
		role.Code = codeN.String
	}
	if desc.Valid {
		v := desc.String
		role.Description = &v
	}
	if sID.Valid {
		v := sID.Int64
		role.SiteID = &v
	}
	return &role, nil
}

func (r *RoleRepository) GetByID(id int64) (*Role, error) {
	if r == nil || r.DB == nil {
		return nil, sql.ErrConnDone
	}

	query := `SELECT id, name, code, description, site_id FROM roles WHERE id = $1`
	var role Role
	var code sql.NullString
	var desc sql.NullString
	var sID sql.NullInt64
	row := r.DB.QueryRow(query, id)
	if err := row.Scan(&role.ID, &role.Name, &code, &desc, &sID); err != nil {
		return nil, err
	}
	if code.Valid {
		role.Code = code.String
	}
	if desc.Valid {
		v := desc.String
		role.Description = &v
	}
	if sID.Valid {
		v := sID.Int64
		role.SiteID = &v
	}
	return &role, nil
}
