package model

import (
	"database/sql"
	"time"
)

// Tujuan merepresentasikan satu entri halaman tujuan trafik.
type Tujuan struct {
	ID             int       `json:"id"`
	Nama           string    `json:"nama"`
	URL            string    `json:"url"`
	Status         bool      `json:"status"`
	DibuatPada     time.Time `json:"dibuat_pada"`
	DiperbaruiPada  time.Time `json:"diperbarui_pada"`
}

// Aturan merepresentasikan satu aturan pencocokan trafik.
type Aturan struct {
	ID             int            `json:"id"`
	Nama           string         `json:"nama"`
	Negara         sql.NullString `json:"-"`
	Perangkat      sql.NullString `json:"-"`
	Peramban       sql.NullString `json:"-"`
	AgenPengguna   sql.NullString `json:"-"`
	AsalRujukan    sql.NullString `json:"-"`
	TujuanID       int            `json:"tujuan_id"`
	TujuanNama     string         `json:"tujuan_nama"`
	TujuanURL      string         `json:"tujuan_url"`
	Prioritas      int            `json:"prioritas"`
	Status         bool           `json:"status"`
	DibuatPada     time.Time      `json:"dibuat_pada"`
	DiperbaruiPada  time.Time     `json:"diperbarui_pada"`
}

// AturanJSON adalah representasi JSON dari Aturan dengan field nullable sebagai *string.
type AturanJSON struct {
	ID             int       `json:"id"`
	Nama           string    `json:"nama"`
	Negara         *string   `json:"negara"`
	Perangkat      *string   `json:"perangkat"`
	Peramban       *string   `json:"peramban"`
	AgenPengguna   *string   `json:"agen_pengguna"`
	AsalRujukan    *string   `json:"asal_rujukan"`
	TujuanID       int       `json:"tujuan_id"`
	TujuanNama     string    `json:"tujuan_nama"`
	TujuanURL      string    `json:"tujuan_url"`
	Prioritas      int       `json:"prioritas"`
	Status         bool      `json:"status"`
	DibuatPada     time.Time `json:"dibuat_pada"`
	DiperbaruiPada  time.Time `json:"diperbarui_pada"`
}

// KeJSON mengubah Aturan menjadi AturanJSON untuk respons API.
func (a Aturan) KeJSON() AturanJSON {
	hasil := AturanJSON{
		ID:             a.ID,
		Nama:           a.Nama,
		TujuanID:       a.TujuanID,
		TujuanNama:     a.TujuanNama,
		TujuanURL:      a.TujuanURL,
		Prioritas:      a.Prioritas,
		Status:         a.Status,
		DibuatPada:     a.DibuatPada,
		DiperbaruiPada:  a.DiperbaruiPada,
	}

	if a.Negara.Valid {
		hasil.Negara = &a.Negara.String
	}
	if a.Perangkat.Valid {
		hasil.Perangkat = &a.Perangkat.String
	}
	if a.Peramban.Valid {
		hasil.Peramban = &a.Peramban.String
	}
	if a.AgenPengguna.Valid {
		hasil.AgenPengguna = &a.AgenPengguna.String
	}
	if a.AsalRujukan.Valid {
		hasil.AsalRujukan = &a.AsalRujukan.String
	}

	return hasil
}
