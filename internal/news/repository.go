package news

import (
	"christ-api/pkg/database"
	"database/sql"
	"encoding/json"
)

type NewsRepository struct{}

func (r *NewsRepository) List(filter NewsFilter) ([]News, error) {
	if database.DB == nil {
		return nil, sql.ErrConnDone
	}

	query := `SELECT n.id, n.uuid, n.title, n.slug, n.excerpt, n.content, n.author_id, n.site_id, n.status, n.is_featured, n.meta, n.published_at, n.views, n.created_at, n.updated_at, n.deleted_at, c.full_name AS author_name FROM news n LEFT JOIN users u ON n.author_id = u.id LEFT JOIN contacts c ON u.contact_id = c.id WHERE n.deleted_at IS NULL`
	args := []interface{}{}
	idx := 1

	if filter.ID != nil {
		query += ` AND n.id = $` + itoa(idx)
		args = append(args, *filter.ID)
		idx++
	}
	if filter.SiteID != nil {
		query += ` AND n.site_id = $` + itoa(idx)
		args = append(args, *filter.SiteID)
		idx++
	}
	if filter.Search != nil && *filter.Search != "" {
		query += ` AND (title ILIKE $` + itoa(idx) + ` OR content ILIKE $` + itoa(idx) + `)`
		args = append(args, "%"+*filter.Search+"%")
		idx++
	}

	// pagination
	if filter.Limit == 0 {
		filter.Limit = 25
	}
	query += ` ORDER BY n.published_at DESC NULLS LAST, n.created_at DESC LIMIT $` + itoa(idx) + ` OFFSET $` + itoa(idx+1)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []News
	for rows.Next() {
		var n News
		var meta sql.NullString
		var excerpt sql.NullString
		var authorID sql.NullInt64
		var siteID sql.NullInt64
		var publishedAt sql.NullTime
		var createdAt sql.NullTime
		var updatedAt sql.NullTime
		var deletedAt sql.NullTime
		var authorName sql.NullString

		err := rows.Scan(&n.ID, &n.UUID, &n.Title, &n.Slug, &excerpt, &n.Content, &authorID, &siteID, &n.Status, &n.IsFeatured, &meta, &publishedAt, &n.Views, &createdAt, &updatedAt, &deletedAt, &authorName)
		if err != nil {
			return nil, err
		}
		if excerpt.Valid {
			n.Excerpt = &excerpt.String
		}
		if authorID.Valid {
			v := authorID.Int64
			n.AuthorID = &v
		}
		if siteID.Valid {
			v := siteID.Int64
			n.SiteID = &v
		}
		if meta.Valid {
			n.Meta = []byte(meta.String)
		}
		if publishedAt.Valid {
			n.PublishedAt = &publishedAt.Time
		}
		if createdAt.Valid {
			n.CreatedAt = &createdAt.Time
		}
		if updatedAt.Valid {
			n.UpdatedAt = &updatedAt.Time
		}
		if deletedAt.Valid {
			n.DeletedAt = &deletedAt.Time
		}
		if authorName.Valid {
			an := authorName.String
			n.AuthorName = &an
		}
		out = append(out, n)
	}
	return out, nil
}

func (r *NewsRepository) FindByID(id int64) (*News, error) {
	f := NewsFilter{ID: &id, Limit: 1}
	res, err := r.List(f)
	if err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return nil, nil
	}
	return &res[0], nil
}

func (r *NewsRepository) Create(n *News) (*News, error) {
	if database.DB == nil {
		return nil, sql.ErrConnDone
	}
	query := `INSERT INTO news (title, slug, excerpt, content, author_id, site_id, status, is_featured, meta, published_at, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,NOW(),NOW()) RETURNING id, uuid, title, slug, excerpt, content, author_id, site_id, status, is_featured, meta, published_at, views, created_at, updated_at, deleted_at`

	var metaStr interface{}
	if n.Meta != nil {
		var tmp interface{}
		if err := json.Unmarshal(n.Meta, &tmp); err == nil {
			metaStr = n.Meta
		} else {
			metaStr = nil
		}
	}

	var created News
	var metaN sql.NullString
	var excerpt sql.NullString
	var authorID sql.NullInt64
	var siteID sql.NullInt64
	var publishedAt sql.NullTime
	var createdAt sql.NullTime
	var updatedAt sql.NullTime
	var deletedAt sql.NullTime

	if n.Excerpt != nil {
		excerpt = sql.NullString{String: *n.Excerpt, Valid: true}
	}
	if n.AuthorID != nil {
		authorID = sql.NullInt64{Int64: *n.AuthorID, Valid: true}
	}
	if n.SiteID != nil {
		siteID = sql.NullInt64{Int64: *n.SiteID, Valid: true}
	}

	err := database.DB.QueryRow(query, n.Title, n.Slug, excerpt, n.Content, authorID, siteID, n.Status, n.IsFeatured, metaStr, n.PublishedAt).Scan(&created.ID, &created.UUID, &created.Title, &created.Slug, &excerpt, &created.Content, &authorID, &siteID, &created.Status, &created.IsFeatured, &metaN, &publishedAt, &created.Views, &createdAt, &updatedAt, &deletedAt)
	if err != nil {
		return nil, err
	}

	if metaN.Valid {
		created.Meta = []byte(metaN.String)
	}
	if excerpt.Valid {
		created.Excerpt = &excerpt.String
	}
	if authorID.Valid {
		v := authorID.Int64
		created.AuthorID = &v
	}
	if siteID.Valid {
		v := siteID.Int64
		created.SiteID = &v
	}
	if publishedAt.Valid {
		created.PublishedAt = &publishedAt.Time
	}
	if createdAt.Valid {
		created.CreatedAt = &createdAt.Time
	}
	if updatedAt.Valid {
		created.UpdatedAt = &updatedAt.Time
	}
	if deletedAt.Valid {
		created.DeletedAt = &deletedAt.Time
	}

	if created.AuthorID != nil {
		var authorName sql.NullString
		row := database.DB.QueryRow(`SELECT c.full_name FROM users u LEFT JOIN contacts c ON u.contact_id = c.id WHERE u.id = $1 LIMIT 1`, *created.AuthorID)
		if err := row.Scan(&authorName); err == nil {
			if authorName.Valid {
				an := authorName.String
				created.AuthorName = &an
			}
		}
	}
	return &created, nil
}

func (r *NewsRepository) Update(n *News) error {
	if database.DB == nil {
		return sql.ErrConnDone
	}
	query := `UPDATE news SET title=$1, slug=$2, excerpt=$3, content=$4, author_id=$5, site_id=$6, status=$7, is_featured=$8, meta=$9, published_at=$10, updated_at=NOW() WHERE uuid = $11`

	var excerpt sql.NullString
	var authorID sql.NullInt64
	var siteID sql.NullInt64

	if n.Excerpt != nil {
		excerpt = sql.NullString{String: *n.Excerpt, Valid: true}
	}
	if n.AuthorID != nil {
		authorID = sql.NullInt64{Int64: *n.AuthorID, Valid: true}
	}
	if n.SiteID != nil {
		siteID = sql.NullInt64{Int64: *n.SiteID, Valid: true}
	}

	_, err := database.DB.Exec(query, n.Title, n.Slug, excerpt, n.Content, authorID, siteID, n.Status, n.IsFeatured, n.Meta, n.PublishedAt, n.UUID)
	return err
}

func (r *NewsRepository) SoftDelete(uuid string) error {
	if database.DB == nil {
		return sql.ErrConnDone
	}
	query := `UPDATE news SET deleted_at = NOW() WHERE uuid = $1`
	_, err := database.DB.Exec(query, uuid)
	return err
}

// small helper to convert int to string without importing strconv multiple times
func itoa(i int) string {
	// cheap and safe for small ints used here
	switch i {
	case 0:
		return "0"
	case 1:
		return "1"
	case 2:
		return "2"
	case 3:
		return "3"
	case 4:
		return "4"
	case 5:
		return "5"
	case 6:
		return "6"
	case 7:
		return "7"
	case 8:
		return "8"
	case 9:
		return "9"
	case 10:
		return "10"
	case 11:
		return "11"
	default:
		return "0"
	}
}
