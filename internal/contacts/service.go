package contacts

type ContactService struct {
	Repo ContactRepository
}

func (s *ContactService) List() ([]Contact, error) { return s.Repo.GetAll() }
func (s *ContactService) Create(fullName string, phone *string, address *string, siteID *int64) (*Contact, error) {
	return s.Repo.Create(fullName, phone, address, siteID)
}
func (s *ContactService) Update(id int64, fullName string, phone *string, address *string) (*Contact, error) {
	return s.Repo.Update(id, fullName, phone, address)
}
