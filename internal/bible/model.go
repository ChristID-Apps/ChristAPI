package bible

type Surat struct {
	ID        int64         `json:"id"`
	NamaSurat string        `json:"nama_surat"`
	Singkatan string        `json:"singkatan"`
	Urutan    int64 		`json:"urutan"`
	Testament string        `json:"testament"`
}

type Perikop struct {
	ID      int64   `json:"id"`
	PasalID int64   `json:"pasal_id"`
	Judul   *string `json:"judul"`
	Label   *string `json:"label"`
}

type Pasal struct {
	ID         int64   `json:"id"`
	SuratID    int64   `json:"surat_id"`
	NomorPasal int     `json:"nomor_pasal"`
	Judul      *string `json:"judul"`
}

type Ayat struct {
	ID        int64  `json:"id"`
	PasalID   int64  `json:"pasal_id"`
	PerikopID *int64 `json:"perikop_id"`
	Nomor     int    `json:"nomor"`
	Teks      string `json:"teks"`
}

type PerikopWithAyats struct {
	Perikop
	Ayats []Ayat `json:"ayats"`
}

type PasalDetail struct {
	Pasal
	Perikops            []PerikopWithAyats `json:"perikops"`
	AyatsWithoutPerikop []Ayat             `json:"ayats_without_perikop,omitempty"`
}
