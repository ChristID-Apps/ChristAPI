package rewards

import (
	"errors"
	"strings"
)

type Service struct{ Repo *Repository }

func (s *Service) List(activeOnly bool) ([]Reward, error) { return s.Repo.List(activeOnly) }

func (s *Service) Create(req *CreateRewardRequest, createdBy int64) (*Reward, error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, errors.New("name is required")
	}
	if req.PointsRequired < 1 {
		return nil, errors.New("points_required must be greater than 0")
	}
	if req.Stock < 0 {
		return nil, errors.New("stock cannot be negative")
	}
	return s.Repo.Create(req, createdBy)
}

func (s *Service) Update(uuid string, req *UpdateRewardRequest) (*Reward, error) {
	if req.Name != nil && strings.TrimSpace(*req.Name) == "" {
		return nil, errors.New("name cannot be empty")
	}
	if req.PointsRequired != nil && *req.PointsRequired < 1 {
		return nil, errors.New("points_required must be greater than 0")
	}
	if req.Stock != nil && *req.Stock < 0 {
		return nil, errors.New("stock cannot be negative")
	}
	if req.Status != nil && *req.Status != "active" && *req.Status != "inactive" {
		return nil, errors.New("status must be active or inactive")
	}
	return s.Repo.Update(uuid, req)
}

func (s *Service) UpdateImage(uuid, imageURL string) error {
	return s.Repo.UpdateImage(uuid, imageURL)
}

func (s *Service) Redeem(userID int64, rewardUUID string) (*Redemption, error) {
	if userID < 1 || strings.TrimSpace(rewardUUID) == "" {
		return nil, errors.New("reward uuid is required")
	}
	return s.Repo.Redeem(userID, rewardUUID)
}

func (s *Service) ListRedemptions(userID *int64, status string) ([]Redemption, error) {
	if status != "" && status != "pending" && status != "approved" && status != "rejected" && status != "completed" {
		return nil, errors.New("invalid redemption status")
	}
	return s.Repo.ListRedemptions(userID, status)
}

func (s *Service) Decide(uuid string, approve bool, note *string) (*Redemption, error) {
	if strings.TrimSpace(uuid) == "" {
		return nil, errors.New("redemption uuid is required")
	}
	return s.Repo.Decide(uuid, approve, note)
}

func (s *Service) Complete(uuid, userCode, adminCode string) (*Redemption, error) {
	if strings.TrimSpace(uuid) == "" || strings.TrimSpace(userCode) == "" || strings.TrimSpace(adminCode) == "" {
		return nil, errors.New("redemption uuid and both codes are required")
	}
	return s.Repo.Complete(uuid, userCode, adminCode)
}
