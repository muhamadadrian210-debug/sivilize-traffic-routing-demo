package model

import (
	"database/sql"
	"time"
)

// LogTrafik merepresentasikan satu entri riwayat evaluasi trafik.
type LogTrafik struct {
	ID           int            `json:"id"`
	Negara       sql.NullString `json:"-"`
	Perangkat    sql.NullString `json:"-"`
	Peramban     sql.NullString `json:"-"`
	AgenPengguna sql.NullString `json:"-"`
	AsalRujukan  sql.NullString `json:"-"`
	AturanID     sql.NullInt64  `json:"-"`
	TujuanID     sql.NullInt64  `json:"-"`
	URLTujuan    string         `json:"url_tujuan"`
	Hasil        string         `json:"hasil"` // "aturan" atau "fallback"
	DibuatPada   time.Time      `json:"dibuat_pada"`
}

// LogTrafikJSON adalah representasi JSON dari LogTrafik dengan field nullable yang ramah JSON.
type LogTrafikJSON struct {
	ID           int       `json:"id"`
	Negara       *string   `json:"negara"`
	Perangkat    *string   `json:"perangkat"`
	Peramban     *string   `json:"peramban"`
	AgenPengguna *string   `json:"agen_pengguna,omitempty"`
	AsalRujukan  *string   `json:"asal_rujukan"`
	AturanID     *int      `json:"aturan_id"`
	TujuanID     *int      `json:"tujuan_id"`
	URLTujuan    string    `json:"url_tujuan"`
	Hasil        string    `json:"hasil"`
	DibuatPada   time.Time `json:"dibuat_pada"`
}

// KeJSON mengubah LogTrafik menjadi LogTrafikJSON untuk respons API.
func (l LogTrafik) KeJSON() LogTrafikJSON {
	hasil := LogTrafikJSON{
		ID:         l.ID,
		URLTujuan:  l.URLTujuan,
		Hasil:      l.Hasil,
		DibuatPada: l.DibuatPada,
	}

	if l.Negara.Valid {
		hasil.Negara = &l.Negara.String
	}
	if l.Perangkat.Valid {
		hasil.Perangkat = &l.Perangkat.String
	}
	if l.Peramban.Valid {
		hasil.Peramban = &l.Peramban.String
	}
	if l.AgenPengguna.Valid {
		hasil.AgenPengguna = &l.AgenPengguna.String
	}
	if l.AsalRujukan.Valid {
		hasil.AsalRujukan = &l.AsalRujukan.String
	}
	if l.AturanID.Valid {
		val := int(l.AturanID.Int64)
		hasil.AturanID = &val
	}
	if l.TujuanID.Valid {
		val := int(l.TujuanID.Int64)
		hasil.TujuanID = &val
	}

	return hasil
}
