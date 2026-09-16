package sites

type SiteService struct {
	Repo *SiteRepository
}

func (s *SiteService) List() ([]Site, error) { return s.Repo.GetAll() }
func (s *SiteService) Create(name string, address *string) (*Site, error) {
	return s.Repo.Create(name, address)
}
func (s *SiteService) Update(uuid string, req *UpdateSiteRequest) (*Site, error) {
	return s.Repo.Update(uuid, req)
}
