package sites

type SiteService struct {
	Repo *SiteRepository
}

func (s *SiteService) List() ([]Site, error) { return s.Repo.GetAll() }
func (s *SiteService) Create(name string, address *string, latitude, longitude *float64) (*Site, error) {
	return s.Repo.Create(name, address, latitude, longitude)
}
func (s *SiteService) Update(uuid string, req *UpdateSiteRequest) (*Site, error) {
	return s.Repo.Update(uuid, req)
}
