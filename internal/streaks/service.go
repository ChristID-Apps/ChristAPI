package streaks

import (
	"errors"
	"strings"
	"time"
)

type Service struct {
	Repo *Repository
}

func (s *Service) CheckIn(userID int64, activityUUID string) (*CheckInResult, error) {
	if userID < 1 {
		return nil, errors.New("invalid user id")
	}
	if strings.TrimSpace(activityUUID) == "" {
		return nil, errors.New("activity_uuid is required")
	}
	return s.Repo.CheckIn(userID, activityUUID, time.Now())
}

func (s *Service) List(userID int64) ([]Streak, error) {
	return s.Repo.List(userID)
}
