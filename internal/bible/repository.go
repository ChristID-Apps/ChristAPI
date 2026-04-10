package bible

import (
	"christ-api/pkg/database"
	"database/sql"
)

type BibleRepository struct{}

func (r *BibleRepository) ListSurat(testament *string) ([]Surat, error) {
	if database.DB == nil {
		return nil, sql.ErrConnDone
	}

	query := `SELECT id, nama_surat, singkatan, urutan, testament FROM surat`
	
	args := []interface{}{}
	if testament != nil {
		query += ` WHERE testament=$1`
		args = append(args, *testament)
	}
	query += ` ORDER BY urutan`

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var out []Surat
	for rows.Next() {
		var s Surat
		if err := rows.Scan(&s.ID, &s.NamaSurat, &s.Singkatan, &s.Urutan, &s.Testament); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, nil
}

func (r *BibleRepository) ListPasalBySurat(suratID int64) ([]Pasal, error) {
	if database.DB == nil {
		return nil, sql.ErrConnDone
	}
	rows, err := database.DB.Query(`SELECT id, surat_id, nomor_pasal, judul FROM pasal WHERE surat_id=$1 ORDER BY nomor_pasal`, suratID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Pasal
	for rows.Next() {
		var p Pasal
		var judul sql.NullString
		if err := rows.Scan(&p.ID, &p.SuratID, &p.NomorPasal, &judul); err != nil {
			return nil, err
		}
		if judul.Valid {
			v := judul.String
			p.Judul = &v
		}
		out = append(out, p)
	}
	return out, nil
}

func (r *BibleRepository) ListAyatByPasal(pasalID int64) ([]Ayat, error) {
	if database.DB == nil {
		return nil, sql.ErrConnDone
	}
	rows, err := database.DB.Query(`SELECT id, pasal_id, perikop_id, nomor, teks FROM ayat WHERE pasal_id=$1 ORDER BY nomor`, pasalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Ayat
	for rows.Next() {
		var a Ayat
		var perikop sql.NullInt64
		if err := rows.Scan(&a.ID, &a.PasalID, &perikop, &a.Nomor, &a.Teks); err != nil {
			return nil, err
		}
		if perikop.Valid {
			v := perikop.Int64
			a.PerikopID = &v
		}
		out = append(out, a)
	}
	return out, nil
}

func (r *BibleRepository) GetAyatByID(id int64) (*Ayat, error) {
	if database.DB == nil {
		return nil, sql.ErrConnDone
	}
	query := `SELECT id, pasal_id, perikop_id, nomor, teks FROM ayat WHERE id=$1`
	var a Ayat
	var perikop sql.NullInt64
	row := database.DB.QueryRow(query, id)
	if err := row.Scan(&a.ID, &a.PasalID, &perikop, &a.Nomor, &a.Teks); err != nil {
		return nil, err
	}
	if perikop.Valid {
		v := perikop.Int64
		a.PerikopID = &v
	}
	return &a, nil
}

func (r *BibleRepository) GetPasalWithContents(pasalID int64) (*PasalDetail, error) {
	if database.DB == nil {
		return nil, sql.ErrConnDone
	}

	// fetch pasal
	var p Pasal
	var judul sql.NullString
	row := database.DB.QueryRow(`SELECT id, surat_id, nomor_pasal, judul FROM pasal WHERE id=$1`, pasalID)
	if err := row.Scan(&p.ID, &p.SuratID, &p.NomorPasal, &judul); err != nil {
		return nil, err
	}
	if judul.Valid {
		v := judul.String
		p.Judul = &v
	}

	// fetch perikops
	perikopRows, err := database.DB.Query(`SELECT id, pasal_id, judul, label FROM perikop WHERE pasal_id=$1 ORDER BY id`, pasalID)
	if err != nil {
		return nil, err
	}
	defer perikopRows.Close()
	var perikops []Perikop
	for perikopRows.Next() {
		var pk Perikop
		var j sql.NullString
		var l sql.NullString
		if err := perikopRows.Scan(&pk.ID, &pk.PasalID, &j, &l); err != nil {
			return nil, err
		}
		if j.Valid {
			v := j.String
			pk.Judul = &v
		}
		if l.Valid {
			v := l.String
			pk.Label = &v
		}
		perikops = append(perikops, pk)
	}

	// fetch ayats for pasal
	ayatRows, err := database.DB.Query(`SELECT id, pasal_id, perikop_id, nomor, teks FROM ayat WHERE pasal_id=$1 ORDER BY nomor`, pasalID)
	if err != nil {
		return nil, err
	}
	defer ayatRows.Close()
	var ayats []Ayat
	for ayatRows.Next() {
		var a Ayat
		var perikop sql.NullInt64
		if err := ayatRows.Scan(&a.ID, &a.PasalID, &perikop, &a.Nomor, &a.Teks); err != nil {
			return nil, err
		}
		if perikop.Valid {
			v := perikop.Int64
			a.PerikopID = &v
		}
		ayats = append(ayats, a)
	}

	// group ayats by perikop id
	ayatsByPerikop := make(map[int64][]Ayat)
	var ayatsNoPerikop []Ayat
	for _, a := range ayats {
		if a.PerikopID == nil {
			ayatsNoPerikop = append(ayatsNoPerikop, a)
			continue
		}
		ayatsByPerikop[*a.PerikopID] = append(ayatsByPerikop[*a.PerikopID], a)
	}

	// assemble PerikopWithAyats
	var perikopWith []PerikopWithAyats
	for _, pk := range perikops {
		pw := PerikopWithAyats{Perikop: pk}
		if v, ok := ayatsByPerikop[pk.ID]; ok {
			pw.Ayats = v
		}
		perikopWith = append(perikopWith, pw)
	}

	out := PasalDetail{
		Pasal:               p,
		Perikops:            perikopWith,
		AyatsWithoutPerikop: ayatsNoPerikop,
	}
	return &out, nil
}

// GetPasalBySuratNomor returns a Pasal by surat_id and nomor_pasal
func (r *BibleRepository) GetPasalBySuratNomor(suratID int64, nomorPasal int64) (Pasal, error) {
	if database.DB == nil {
		return Pasal{}, sql.ErrConnDone
	}
	var p Pasal
	var judul sql.NullString
	row := database.DB.QueryRow(`SELECT id, surat_id, nomor_pasal, judul FROM pasal WHERE surat_id=$1 AND nomor_pasal=$2 LIMIT 1`, suratID, nomorPasal)
	if err := row.Scan(&p.ID, &p.SuratID, &p.NomorPasal, &judul); err != nil {
		return Pasal{}, err
	}
	if judul.Valid {
		v := judul.String
		p.Judul = &v
	}
	return p, nil
}
