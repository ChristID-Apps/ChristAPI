package bible

type BibleService struct {
	Repo BibleRepository
}

func (s *BibleService) ListSurat(testament *string) ([]Surat, error) {
	return s.Repo.ListSurat(testament)
}
func (s *BibleService) ListPasalBySurat(suratID int64) ([]Pasal, error) {
	return s.Repo.ListPasalBySurat(suratID)
}
func (s *BibleService) ListAyatByPasal(pasalID int64) ([]Ayat, error) {
	return s.Repo.ListAyatByPasal(pasalID)
}
func (s *BibleService) GetAyatByID(id int64) (*Ayat, error) { return s.Repo.GetAyatByID(id) }

func (s *BibleService) GetPasalWithContents(pasalID int64) (*PasalDetail, error) {
	return s.Repo.GetPasalWithContents(pasalID)
}
