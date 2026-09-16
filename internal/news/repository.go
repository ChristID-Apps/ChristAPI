package news

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
)

type NewsRepository struct {
	DB *sql.DB
}

func (r *NewsRepository) List(filter NewsFilter) ([]News, error) {
	if r == nil || r.DB == nil {
		return nil, sql.ErrConnDone
	}

	query := `SELECT n.id, n.uuid, n.title, n.slug, n.image_url, n.excerpt, n.content, n.author_id, n.site_id, n.status, n.is_featured, n.meta, n.published_at, n.views, n.created_at, n.updated_at, n.deleted_at, c.full_name AS author_name FROM news n LEFT JOIN users u ON n.author_id = u.id LEFT JOIN contacts c ON u.contact_id = c.id WHERE n.deleted_at IS NULL`
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

	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []News
	for rows.Next() {
		var n News
		var meta sql.NullString
		var excerpt sql.NullString
		var imageURL sql.NullString
		var authorID sql.NullInt64
		var siteID sql.NullInt64
		var publishedAt sql.NullTime
		var createdAt sql.NullTime
		var updatedAt sql.NullTime
		var deletedAt sql.NullTime
		var authorName sql.NullString

		err := rows.Scan(&n.ID, &n.UUID, &n.Title, &n.Slug, &imageURL, &excerpt, &n.Content, &authorID, &siteID, &n.Status, &n.IsFeatured, &meta, &publishedAt, &n.Views, &createdAt, &updatedAt, &deletedAt, &authorName)
		if err != nil {
			return nil, err
		}
		if excerpt.Valid {
			n.Excerpt = &excerpt.String
		}
		if imageURL.Valid {
			n.ImageURL = &imageURL.String
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
	if r == nil || r.DB == nil {
		return nil, sql.ErrConnDone
	}
	query := `INSERT INTO news (title, slug, image_url, excerpt, content, author_id, site_id, status, is_featured, meta, published_at, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,NOW(),NOW()) RETURNING id, uuid, title, slug, image_url, excerpt, content, author_id, site_id, status, is_featured, meta, published_at, views, created_at, updated_at, deleted_at`

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
	var imageURL sql.NullString
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

	err := r.DB.QueryRow(query, n.Title, n.Slug, n.ImageURL, excerpt, n.Content, authorID, siteID, n.Status, n.IsFeatured, metaStr, n.PublishedAt).Scan(&created.ID, &created.UUID, &created.Title, &created.Slug, &imageURL, &excerpt, &created.Content, &authorID, &siteID, &created.Status, &created.IsFeatured, &metaN, &publishedAt, &created.Views, &createdAt, &updatedAt, &deletedAt)
	if err != nil {
		return nil, err
	}

	if metaN.Valid {
		created.Meta = []byte(metaN.String)
	}
	if imageURL.Valid {
		created.ImageURL = &imageURL.String
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
		row := r.DB.QueryRow(`SELECT c.full_name FROM users u LEFT JOIN contacts c ON u.contact_id = c.id WHERE u.id = $1 LIMIT 1`, *created.AuthorID)
		if err := row.Scan(&authorName); err == nil {
			if authorName.Valid {
				an := authorName.String
				created.AuthorName = &an
			}
		}
	}
	return &created, nil
}

func (r *NewsRepository) Update(uuid string, n *NewsUpdateRequest) error {
	if r == nil || r.DB == nil {
		return sql.ErrConnDone
	}
	columns := []string{}
	args := []interface{}{}
	add := func(name string, value interface{}) {
		columns = append(columns, name+fmt.Sprintf("=$%d", len(args)+1))
		args = append(args, value)
	}
	if n.Present["title"] {
		add("title", n.Title)
	}
	if n.Present["slug"] {
		add("slug", n.Slug)
	}
	if n.Present["image_url"] {
		add("image_url", n.ImageURL)
	}
	if n.Present["excerpt"] {
		add("excerpt", n.Excerpt)
	}
	if n.Present["content"] {
		add("content", n.Content)
	}
	if n.Present["author_id"] {
		add("author_id", n.AuthorID)
	}
	if n.Present["site_id"] {
		add("site_id", n.SiteID)
	}
	if n.Present["status"] {
		add("status", n.Status)
	}
	if n.Present["is_featured"] {
		add("is_featured", n.IsFeatured)
	}
	if n.Present["meta"] {
		add("meta", n.Meta)
	}
	if n.Present["published_at"] {
		add("published_at", n.PublishedAt)
	}
	if len(columns) == 0 {
		columns = append(columns, "uuid=uuid")
	}
	args = append(args, uuid)
	query := `UPDATE news SET ` + strings.Join(columns, ", ") + `, updated_at=NOW() WHERE uuid = $` + fmt.Sprint(len(args))
	_, err := r.DB.Exec(query, args...)
	return err
}

func (r *NewsRepository) UpdateImage(uuid, imageURL string) error {
	if r == nil || r.DB == nil {
		return sql.ErrConnDone
	}
	result, err := r.DB.Exec(`UPDATE news SET image_url=$1, updated_at=NOW() WHERE uuid=$2 AND deleted_at IS NULL`, imageURL, uuid)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *NewsRepository) SoftDelete(uuid string) error {
	if r == nil || r.DB == nil {
		return sql.ErrConnDone
	}
	query := `UPDATE news SET deleted_at = NOW() WHERE uuid = $1`
	_, err := r.DB.Exec(query, uuid)
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
