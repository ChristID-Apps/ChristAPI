package points

import "errors"

type Service struct {
	Repo Repository
}

func (s *Service) GetState(userID int64, siteID *int64, offset, limit int) (*PointsState, error) {
	balance, err := s.Repo.GetBalance(userID, siteID)
	if err != nil {
		return nil, err
	}
	history, err := s.Repo.GetHistory(userID, siteID, offset, limit)
	if err != nil {
		return nil, err
	}
	return &PointsState{UserID: userID, Balance: balance, History: history}, nil
}

func (s *Service) Earn(userID, amount int64, reason string, referenceID *string) (*LedgerEntry, error) {
	if amount <= 0 {
		return nil, errors.New("amount must be greater than 0")
	}
	if reason == "" {
		reason = "manual_earn"
	}
	return s.Repo.Earn(userID, amount, reason, referenceID)
}

func (s *Service) Spend(userID, amount int64, reason string, referenceID *string) (*LedgerEntry, error) {
	if amount <= 0 {
		return nil, errors.New("amount must be greater than 0")
	}
	if reason == "" {
		reason = "manual_spend"
	}
	return s.Repo.Spend(userID, amount, reason, referenceID)
}

func (s *Service) ListBalances(siteID *int64, offset, limit int) ([]UserBalance, error) {
	return s.Repo.ListBalances(siteID, offset, limit)
}
