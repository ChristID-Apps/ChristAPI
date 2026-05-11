package contacts

type ContactService struct {
	Repo ContactRepository
}

func (s *ContactService) List(page, limit int) ([]Contact, error) { return s.Repo.List(page, limit) }
func (s *ContactService) GetByID(id int64) (*Contact, error)      { return s.Repo.GetByID(id) }
func (s *ContactService) Create(fullName string, phone *string, address *string, siteID *int64) (*Contact, error) {
	return s.Repo.Create(fullName, phone, address, siteID)
}
func (s *ContactService) Update(id int64, fullName string, phone *string, address *string, siteID *int64) (*Contact, error) {
	return s.Repo.Update(id, fullName, phone, address, siteID)
}
func (s *ContactService) Delete(id int64) (*Contact, error) { return s.Repo.SoftDelete(id) }
