package activities

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"christ-api/internal/activities/dto/requests"
)

type Repository struct {
	DB *sql.DB
}

func (r *Repository) ListCategories() ([]Category, error) {
	if r == nil || r.DB == nil {
		return nil, sql.ErrConnDone
	}
	rows, err := r.DB.Query(`SELECT id, code, name, COALESCE(description, '') FROM activity_categories ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var category Category
		if err := rows.Scan(&category.ID, &category.Code, &category.Name, &category.Description); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	return categories, rows.Err()
}

func (r *Repository) List(filter ActivityFilter) ([]Activity, error) {
	if r == nil || r.DB == nil {
		return nil, sql.ErrConnDone
	}
	if filter.Limit < 1 || filter.Limit > 100 {
		filter.Limit = 20
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	query := `
		SELECT a.id, a.uuid, a.title, a.description, a.image_url, a.category_id, ac.code, ac.name,
			a.activity_type, a.site_id, a.created_by, a.status, a.max_participants,
			a.requires_registration, a.streak_enabled, a.streak_type, a.streak_points, a.created_at, a.updated_at,
			s.id, s.frequency, s.interval_value, s.days_of_week, s.day_of_month,
			s.start_date, s.end_date, s.start_time, s.end_time, s.timezone,
			o.id, o.starts_at, o.ends_at, o.location, o.status, o.notes
			, b.version_code, b.book_code, b.start_chapter, b.start_verse, b.end_chapter, b.end_verse,
			 b.requires_reflection, b.reflection_prompt, b.reflection_min_length
		FROM activities a
		JOIN activity_categories ac ON ac.id = a.category_id
		LEFT JOIN activity_schedules s ON s.activity_id = a.id
		LEFT JOIN activity_occurrences o ON o.activity_id = a.id
		LEFT JOIN activity_bible_configs b ON b.activity_id = a.id
		WHERE a.deleted_at IS NULL`
	args := make([]interface{}, 0, 6)
	argIndex := 1
	if filter.Search != "" {
		query += fmt.Sprintf(" AND (a.title ILIKE $%d OR a.description ILIKE $%d)", argIndex, argIndex)
		args = append(args, "%"+filter.Search+"%")
		argIndex++
	}
	if filter.CategoryID != nil {
		query += fmt.Sprintf(" AND a.category_id = $%d", argIndex)
		args = append(args, *filter.CategoryID)
		argIndex++
	}
	if filter.SiteID != nil {
		query += fmt.Sprintf(" AND a.site_id = $%d", argIndex)
		args = append(args, *filter.SiteID)
		argIndex++
	}
	if filter.ActivityType != "" {
		query += fmt.Sprintf(" AND a.activity_type = $%d", argIndex)
		args = append(args, filter.ActivityType)
		argIndex++
	}
	if filter.Status != "" {
		query += fmt.Sprintf(" AND a.status = $%d", argIndex)
		args = append(args, filter.Status)
		argIndex++
	}
	query += fmt.Sprintf(" ORDER BY a.created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Activity
	for rows.Next() {
		activity, err := scanActivity(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *activity)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *Repository) GetByUUID(uuid string) (*Activity, error) {
	items, err := r.List(ActivityFilter{Search: "", Limit: 100})
	if err != nil {
		return nil, err
	}
	for i := range items {
		if items[i].UUID == uuid {
			return &items[i], nil
		}
	}
	return nil, sql.ErrNoRows
}

func (r *Repository) Create(req *requests.CreateActivityRequest, createdBy *int64) (*Activity, error) {
	if r == nil || r.DB == nil {
		return nil, sql.ErrConnDone
	}
	tx, err := r.DB.Begin()
	if err != nil {
		return nil, err
	}
	rollback := func() { _ = tx.Rollback() }

	status := req.Status
	if status == "" {
		status = "draft"
	}
	var id int64
	err = tx.QueryRow(`
		INSERT INTO activities (title, description, category_id, activity_type, site_id, created_by, status, max_participants, requires_registration, streak_enabled, streak_type, streak_points)
		VALUES ($1,$2,$3,$4,COALESCE($5, (
			SELECT u.site_id
			FROM users u
			WHERE u.id = $6
			  AND EXISTS (SELECT 1 FROM sites s WHERE s.id = u.site_id)
		)),$6,$7,$8,$9,$10,$11,$12) RETURNING id`,
		req.Title, req.Description, req.CategoryID, req.ActivityType, req.SiteID, createdBy,
		status, req.MaxParticipants, req.RequiresRegistration, req.StreakEnabled, req.StreakType, req.StreakPoints).Scan(&id)
	if err != nil {
		rollback()
		return nil, err
	}
	if err := validateRelationInput(req); err != nil {
		rollback()
		return nil, err
	}
	if req.Schedule != nil {
		if err := insertSchedule(tx, id, req.Schedule); err != nil {
			rollback()
			return nil, err
		}
	}
	if req.Occurrence != nil {
		if err := insertOccurrence(tx, id, req.Occurrence); err != nil {
			rollback()
			return nil, err
		}
	}
	if req.BibleConfig != nil {
		if _, err := tx.Exec(`INSERT INTO activity_bible_configs (activity_id, version_code, book_code, start_chapter, start_verse, end_chapter, end_verse, requires_reflection, reflection_prompt, reflection_min_length) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`, id, req.BibleConfig.VersionCode, req.BibleConfig.BookCode, req.BibleConfig.StartChapter, req.BibleConfig.StartVerse, req.BibleConfig.EndChapter, req.BibleConfig.EndVerse, req.BibleConfig.RequiresReflection, req.BibleConfig.ReflectionPrompt, req.BibleConfig.ReflectionMinLength); err != nil {
			rollback()
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return r.GetByUUIDFromID(id)
}

func (r *Repository) UpdateImage(uuid, imageURL string) error {
	if r == nil || r.DB == nil {
		return sql.ErrConnDone
	}
	result, err := r.DB.Exec(`UPDATE activities SET image_url=$1, updated_at=NOW() WHERE uuid=$2 AND deleted_at IS NULL`, imageURL, uuid)
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

func (r *Repository) UpsertBibleConfig(uuid string, req *requests.BibleActivityRequest) error {
	var activityID int64
	err := r.DB.QueryRow(`
		INSERT INTO activity_bible_configs (activity_id, version_code, book_code, start_chapter, start_verse, end_chapter, end_verse, requires_reflection, reflection_prompt, reflection_min_length)
		SELECT id, $2, $3, $4, $5, $6, $7, $8, $9, $10 FROM activities WHERE uuid = $1 AND deleted_at IS NULL
		ON CONFLICT (activity_id) DO UPDATE SET version_code = EXCLUDED.version_code, book_code = EXCLUDED.book_code, start_chapter = EXCLUDED.start_chapter, start_verse = EXCLUDED.start_verse, end_chapter = EXCLUDED.end_chapter, end_verse = EXCLUDED.end_verse, requires_reflection = EXCLUDED.requires_reflection, reflection_prompt = EXCLUDED.reflection_prompt, reflection_min_length = EXCLUDED.reflection_min_length
		RETURNING activity_id`, uuid, req.VersionCode, req.BookCode, req.StartChapter, req.StartVerse, req.EndChapter, req.EndVerse, req.RequiresReflection, req.ReflectionPrompt, req.ReflectionMinLength).Scan(&activityID)
	return err
}

func (r *Repository) Update(uuid string, req *requests.UpdateActivityRequest) (*Activity, error) {
	if r == nil || r.DB == nil {
		return nil, sql.ErrConnDone
	}
	tx, err := r.DB.Begin()
	if err != nil {
		return nil, err
	}
	rollback := func() { _ = tx.Rollback() }
	var id int64
	columns := []string{}
	args := []interface{}{}
	add := func(name string, value interface{}) {
		columns = append(columns, name+fmt.Sprintf("=$%d", len(args)+1))
		args = append(args, value)
	}
	if req.Present["title"] {
		add("title", req.Title)
	}
	if req.Present["description"] {
		add("description", req.Description)
	}
	if req.Present["category_id"] {
		add("category_id", req.CategoryID)
	}
	if req.Present["activity_type"] {
		add("activity_type", req.ActivityType)
	}
	if req.Present["site_id"] {
		add("site_id", req.SiteID)
	}
	if req.Present["status"] {
		add("status", req.Status)
	}
	if req.Present["max_participants"] {
		add("max_participants", req.MaxParticipants)
	}
	if req.Present["requires_registration"] {
		add("requires_registration", req.RequiresRegistration)
	}
	if req.Present["streak_enabled"] {
		add("streak_enabled", req.StreakEnabled)
	}
	if req.Present["streak_type"] {
		add("streak_type", req.StreakType)
	}
	if req.Present["streak_points"] {
		add("streak_points", req.StreakPoints)
	}
	if len(columns) == 0 {
		err = tx.QueryRow(`SELECT id FROM activities WHERE uuid=$1 AND deleted_at IS NULL`, uuid).Scan(&id)
	} else {
		args = append(args, uuid)
		query := `UPDATE activities SET ` + strings.Join(columns, ", ") + `, updated_at=NOW() WHERE uuid=$` + fmt.Sprint(len(args)) + ` AND deleted_at IS NULL RETURNING id`
		err = tx.QueryRow(query, args...).Scan(&id)
	}
	if err != nil {
		rollback()
		return nil, err
	}
	if req.Present["schedule"] {
		if _, err := tx.Exec(`DELETE FROM activity_schedules WHERE activity_id=$1`, id); err != nil {
			rollback()
			return nil, err
		}
	}
	if req.Present["occurrence"] {
		if _, err := tx.Exec(`DELETE FROM activity_occurrences WHERE activity_id=$1`, id); err != nil {
			rollback()
			return nil, err
		}
	}
	if req.Schedule != nil {
		if err := insertSchedule(tx, id, req.Schedule); err != nil {
			rollback()
			return nil, err
		}
	}
	if req.Occurrence != nil {
		if err := insertOccurrence(tx, id, req.Occurrence); err != nil {
			rollback()
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return r.GetByUUID(uuid)
}

func (r *Repository) Delete(uuid string) error {
	if r == nil || r.DB == nil {
		return sql.ErrConnDone
	}
	result, err := r.DB.Exec(`UPDATE activities SET deleted_at=NOW(), updated_at=NOW(), status='cancelled' WHERE uuid=$1 AND deleted_at IS NULL`, uuid)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *Repository) GetByUUIDFromID(id int64) (*Activity, error) {
	var uuid string
	if err := r.DB.QueryRow(`SELECT uuid FROM activities WHERE id=$1`, id).Scan(&uuid); err != nil {
		return nil, err
	}
	return r.GetByUUID(uuid)
}

func scanActivity(row interface{ Scan(...interface{}) error }) (*Activity, error) {
	var a Activity
	var description sql.NullString
	var imageURL sql.NullString
	var siteID, createdBy sql.NullInt64
	var maxParticipants sql.NullInt64
	var streakType sql.NullString
	var streakPoints int64
	var streakEnabled bool
	var createdAt, updatedAt sql.NullTime
	var scheduleID sql.NullInt64
	var frequency, scheduleStartDate, timezone sql.NullString
	var intervalValue sql.NullInt64
	var daysJSON []byte
	var dayOfMonth sql.NullInt64
	var endDate, startTime, endTime sql.NullString
	var occurrenceID sql.NullInt64
	var startsAt, endsAt sql.NullTime
	var location, occurrenceStatus, notes sql.NullString
	var bibleVersion, bibleBook, biblePrompt sql.NullString
	var bibleStartChapter, bibleStartVerse, bibleEndChapter, bibleEndVerse, bibleMinLength sql.NullInt64
	var bibleReflection sql.NullBool

	err := row.Scan(&a.ID, &a.UUID, &a.Title, &description, &imageURL, &a.CategoryID, &a.CategoryCode, &a.CategoryName,
		&a.ActivityType, &siteID, &createdBy, &a.Status, &maxParticipants, &a.RequiresRegistration, &streakEnabled, &streakType, &streakPoints, &createdAt, &updatedAt,
		&scheduleID, &frequency, &intervalValue, &daysJSON, &dayOfMonth, &scheduleStartDate, &endDate, &startTime, &endTime, &timezone,
		&occurrenceID, &startsAt, &endsAt, &location, &occurrenceStatus, &notes,
		&bibleVersion, &bibleBook, &bibleStartChapter, &bibleStartVerse, &bibleEndChapter, &bibleEndVerse, &bibleReflection, &biblePrompt, &bibleMinLength)
	if err != nil {
		return nil, err
	}
	if description.Valid {
		a.Description = &description.String
	}
	if imageURL.Valid {
		a.ImageURL = &imageURL.String
	}
	if siteID.Valid {
		v := siteID.Int64
		a.SiteID = &v
	}
	if createdBy.Valid {
		v := createdBy.Int64
		a.CreatedBy = &v
	}
	if maxParticipants.Valid {
		v := int(maxParticipants.Int64)
		a.MaxParticipants = &v
	}
	a.StreakEnabled = streakEnabled
	a.StreakPoints = streakPoints
	if streakType.Valid {
		a.StreakType = &streakType.String
	}
	if createdAt.Valid {
		a.CreatedAt = &createdAt.Time
	}
	if updatedAt.Valid {
		a.UpdatedAt = &updatedAt.Time
	}
	if bibleVersion.Valid {
		a.BibleConfig = &BibleActivityConfig{VersionCode: bibleVersion.String, BookCode: bibleBook.String, StartChapter: int(bibleStartChapter.Int64), StartVerse: int(bibleStartVerse.Int64), EndChapter: int(bibleEndChapter.Int64), EndVerse: int(bibleEndVerse.Int64), RequiresReflection: bibleReflection.Bool, ReflectionPrompt: biblePrompt.String, ReflectionMinLength: int(bibleMinLength.Int64)}
	}
	if scheduleID.Valid {
		s := &ActivitySchedule{ID: scheduleID.Int64, Frequency: frequency.String, IntervalValue: int(intervalValue.Int64), StartDate: scheduleStartDate.String, Timezone: timezone.String}
		if len(daysJSON) > 0 {
			_ = json.Unmarshal(daysJSON, &s.DaysOfWeek)
		}
		if dayOfMonth.Valid {
			v := int(dayOfMonth.Int64)
			s.DayOfMonth = &v
		}
		if endDate.Valid {
			s.EndDate = &endDate.String
		}
		if startTime.Valid {
			s.StartTime = &startTime.String
		}
		if endTime.Valid {
			s.EndTime = &endTime.String
		}
		a.Schedule = s
	}
	if occurrenceID.Valid {
		o := &ActivityOccurrence{ID: occurrenceID.Int64, StartsAt: startsAt.Time, Status: occurrenceStatus.String}
		if endsAt.Valid {
			o.EndsAt = &endsAt.Time
		}
		if location.Valid {
			o.Location = &location.String
		}
		if notes.Valid {
			o.Notes = &notes.String
		}
		a.Occurrence = o
	}
	return &a, nil
}

func validateRelationInput(req *requests.CreateActivityRequest) error {
	if req.ActivityType == "recurring" && req.Schedule == nil {
		return fmt.Errorf("schedule is required for recurring activities")
	}
	if req.ActivityType == "one_time" && req.Occurrence == nil {
		return fmt.Errorf("occurrence is required for one-time activities")
	}
	return nil
}

func insertSchedule(tx *sql.Tx, activityID int64, req *requests.ScheduleRequest) error {
	if req.IntervalValue < 1 {
		req.IntervalValue = 1
	}
	if req.Timezone == "" {
		req.Timezone = "Asia/Jakarta"
	}
	days, err := json.Marshal(req.DaysOfWeek)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`INSERT INTO activity_schedules (activity_id, frequency, interval_value, days_of_week, day_of_month, start_date, end_date, start_time, end_time, timezone) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`, activityID, req.Frequency, req.IntervalValue, days, req.DayOfMonth, req.StartDate, req.EndDate, req.StartTime, req.EndTime, req.Timezone)
	return err
}

func insertOccurrence(tx *sql.Tx, activityID int64, req *requests.OccurrenceRequest) error {
	status := req.Status
	if status == "" {
		status = "scheduled"
	}
	_, err := tx.Exec(`INSERT INTO activity_occurrences (activity_id, starts_at, ends_at, location, status, notes) VALUES ($1,$2,$3,$4,$5,$6)`, activityID, req.StartsAt, req.EndsAt, req.Location, status, req.Notes)
	return err
}

func (r *Repository) Count(filter ActivityFilter) (int, error) {
	if r == nil || r.DB == nil {
		return 0, sql.ErrConnDone
	}
	query := `SELECT COUNT(*) FROM activities a WHERE a.deleted_at IS NULL`
	args := []interface{}{}
	argIndex := 1
	if filter.Search != "" {
		query += fmt.Sprintf(" AND (a.title ILIKE $%d OR a.description ILIKE $%d)", argIndex, argIndex)
		args = append(args, "%"+filter.Search+"%")
		argIndex++
	}
	if filter.CategoryID != nil {
		query += fmt.Sprintf(" AND a.category_id = $%d", argIndex)
		args = append(args, *filter.CategoryID)
		argIndex++
	}
	if filter.SiteID != nil {
		query += fmt.Sprintf(" AND a.site_id = $%d", argIndex)
		args = append(args, *filter.SiteID)
		argIndex++
	}
	if filter.ActivityType != "" {
		query += fmt.Sprintf(" AND a.activity_type = $%d", argIndex)
		args = append(args, filter.ActivityType)
		argIndex++
	}
	if filter.Status != "" {
		query += fmt.Sprintf(" AND a.status = $%d", argIndex)
		args = append(args, filter.Status)
	}
	var count int
	err := r.DB.QueryRow(query, args...).Scan(&count)
	return count, err
}

var _ = strings.Builder{}
