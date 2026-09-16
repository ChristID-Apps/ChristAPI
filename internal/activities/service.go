package activities

import (
	"database/sql"
	"errors"
	"strings"

	"christ-api/internal/activities/dto/requests"
)

type Service struct {
	Repo *Repository
}

func (s *Service) ListCategories() ([]Category, error) {
	return s.Repo.ListCategories()
}

func (s *Service) List(filter ActivityFilter) ([]Activity, int, error) {
	items, err := s.Repo.List(filter)
	if err != nil {
		return nil, 0, err
	}
	count, err := s.Repo.Count(filter)
	return items, count, err
}

func (s *Service) Get(uuid string) (*Activity, error) {
	if strings.TrimSpace(uuid) == "" {
		return nil, errors.New("activity uuid is required")
	}
	return s.Repo.GetByUUID(uuid)
}

func (s *Service) Create(req *requests.CreateActivityRequest, createdBy *int64) (*Activity, error) {
	if err := validateRequest(req); err != nil {
		return nil, err
	}
	return s.Repo.Create(req, createdBy)
}

func (s *Service) Update(uuid string, req *requests.UpdateActivityRequest) (*Activity, error) {
	if err := validateUpdateRequest(req); err != nil {
		return nil, err
	}
	return s.Repo.Update(uuid, req)
}

func validateUpdateRequest(req *requests.UpdateActivityRequest) error {
	if req.Present["title"] && (req.Title == nil || strings.TrimSpace(*req.Title) == "") {
		return errors.New("title cannot be empty")
	}
	if req.Present["category_id"] && (req.CategoryID == nil || *req.CategoryID < 1) {
		return errors.New("category_id must be greater than 0")
	}
	if req.Present["activity_type"] && req.ActivityType != nil && *req.ActivityType != "one_time" && *req.ActivityType != "recurring" {
		return errors.New("activity_type must be one_time or recurring")
	}
	if req.Present["status"] && req.Status != nil && !contains([]string{"draft", "published", "cancelled", "completed"}, *req.Status) {
		return errors.New("invalid activity status")
	}
	if req.Present["max_participants"] && req.MaxParticipants != nil && *req.MaxParticipants < 1 {
		return errors.New("max_participants must be greater than 0")
	}
	if req.Present["streak_points"] && req.StreakPoints != nil && *req.StreakPoints < 0 {
		return errors.New("streak_points cannot be negative")
	}
	return nil
}

func (s *Service) Delete(uuid string) error {
	return s.Repo.Delete(uuid)
}

func (s *Service) UpdateImage(uuid, imageURL string) error {
	return s.Repo.UpdateImage(uuid, imageURL)
}

func validateRequest(req *requests.CreateActivityRequest) error {
	if strings.TrimSpace(req.Title) == "" {
		return errors.New("title is required")
	}
	if req.CategoryID < 1 {
		return errors.New("category_id is required")
	}
	if req.ActivityType != "one_time" && req.ActivityType != "recurring" {
		return errors.New("activity_type must be one_time or recurring")
	}
	if req.Status != "" && !contains([]string{"draft", "published", "cancelled", "completed"}, req.Status) {
		return errors.New("invalid activity status")
	}
	if req.MaxParticipants != nil && *req.MaxParticipants < 1 {
		return errors.New("max_participants must be greater than 0")
	}
	if req.StreakEnabled {
		if req.StreakType == nil || strings.TrimSpace(*req.StreakType) == "" {
			return errors.New("streak_type is required when streak is enabled")
		}
		if req.StreakPoints < 0 {
			return errors.New("streak_points cannot be negative")
		}
	} else if req.StreakPoints != 0 || req.StreakType != nil {
		return errors.New("streak fields require streak_enabled=true")
	}
	if req.ActivityType == "recurring" {
		if req.Schedule == nil {
			return errors.New("schedule is required for recurring activities")
		}
		if !contains([]string{"daily", "weekly", "monthly"}, req.Schedule.Frequency) {
			return errors.New("invalid schedule frequency")
		}
		if strings.TrimSpace(req.Schedule.StartDate) == "" {
			return errors.New("schedule.start_date is required")
		}
	} else if req.Occurrence == nil || req.Occurrence.StartsAt.IsZero() {
		return errors.New("occurrence.starts_at is required for one-time activities")
	}
	return nil
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func isNotFound(err error) bool { return errors.Is(err, sql.ErrNoRows) }
