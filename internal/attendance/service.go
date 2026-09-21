package attendance

import (
	"errors"
	"time"
)

type Service struct {
	Repo *Repository
}

func (s *Service) CheckIn(userID int64, siteID *int64, activityID *int64, notes *string) (*AttendanceRecord, error) {
	if s == nil || s.Repo == nil {
		return nil, errors.New("attendance repository is not initialized")
	}
	return s.Repo.CheckIn(userID, siteID, activityID, notes, time.Now())
}

func (s *Service) GetMyHistory(userID int64, startDate, endDate string, limit, offset int) ([]AttendanceRecord, error) {
	if s == nil || s.Repo == nil {
		return nil, errors.New("attendance repository is not initialized")
	}
	return s.Repo.GetMyHistory(userID, startDate, endDate, limit, offset)
}

func (s *Service) GetMySummary(userID int64, startDate, endDate string) (*AttendanceSummary, error) {
	if s == nil || s.Repo == nil {
		return nil, errors.New("attendance repository is not initialized")
	}
	return s.Repo.GetMySummary(userID, startDate, endDate)
}

func (s *Service) GetAdminReport(date string, siteID *int64, activityID *int64, limit, offset int) ([]AttendanceReportItem, error) {
	if s == nil || s.Repo == nil {
		return nil, errors.New("attendance repository is not initialized")
	}
	return s.Repo.GetAdminReport(date, siteID, activityID, limit, offset)
}

func (s *Service) GetAdminSummary(date string, siteID *int64) (*AttendanceReportSummary, error) {
	if s == nil || s.Repo == nil {
		return nil, errors.New("attendance repository is not initialized")
	}
	return s.Repo.GetAdminSummary(date, siteID)
}
