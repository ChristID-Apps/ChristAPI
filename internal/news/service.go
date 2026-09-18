package news

type NewsService struct {
	Repo *NewsRepository
}

func (s *NewsService) List(filter NewsFilter) ([]News, error) {
	return s.Repo.List(filter)
}

func (s *NewsService) Create(n *News) (*News, error) {
	return s.Repo.Create(n)
}

func (s *NewsService) Update(uuid string, req *NewsUpdateRequest) error {
	return s.Repo.Update(uuid, req)
}

func (s *NewsService) Delete(uuid string) error {
	return s.Repo.SoftDelete(uuid)
}

func (s *NewsService) UpdateImage(uuid, imageURL string) error {
	return s.Repo.UpdateImage(uuid, imageURL)
}
