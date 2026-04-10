package news

type NewsService struct {
	Repo NewsRepository
}

func (s *NewsService) List(filter NewsFilter) ([]News, error) {
	return s.Repo.List(filter)
}

func (s *NewsService) Create(n *News) (*News, error) {
	return s.Repo.Create(n)
}

func (s *NewsService) Update(n *News) error {
	return s.Repo.Update(n)
}

func (s *NewsService) Delete(uuid string) error {
	return s.Repo.SoftDelete(uuid)
}
