package model

import (
	"time"
)

// ItemChecklistCompliance merepresentasikan satu item pada checklist kepatuhan manual AdSense.
type ItemChecklistCompliance struct {
	ID             int       `json:"id"`
	Kunci          string    `json:"kunci"`
	Label          string    `json:"label"`
	Selesai        bool      `json:"selesai"`
	DiperbaruiPada time.Time `json:"diperbarui_pada"`
}

// AuditCompliance merepresentasikan satu catatan riwayat audit kepatuhan.
type AuditCompliance struct {
	ID         int                `json:"id"`
	Status     string             `json:"status"` // "aman", "perlu_ditinjau", "berisiko"
	DibuatPada time.Time          `json:"dibuat_pada"`
	Temuan     []TemuanCompliance `json:"temuan,omitempty"`
}

// TemuanCompliance merepresentasikan satu temuan kepatuhan dari hasil audit.
type TemuanCompliance struct {
	ID         int       `json:"id,omitempty"`
	AuditID    int       `json:"audit_id,omitempty"`
	Kode       string    `json:"kode"`
	Tingkat    string    `json:"tingkat"` // "informasi", "peringatan", "risiko tinggi"
	Judul      string    `json:"judul"`
	Pesan      string    `json:"pesan"`
	DibuatPada time.Time `json:"dibuat_pada,omitempty"`
}

// HasilComplianceCheck adalah struktur respons standar untuk GET /api/compliance/check.
// CATATAN: Skor harus SELALU bernilai null sesuai kebijakan Google Publisher Compliance.
type HasilComplianceCheck struct {
	Status           string             `json:"status"` // "aman", "perlu_ditinjau", "berisiko"
	Skor             *int               `json:"skor"`   // Nilai SELALU null (tidak ada skor numerik)
	AuditID          int                `json:"audit_id"`
	WaktuPemeriksaan time.Time          `json:"waktu_pemeriksaan"`
	Ringkasan        RingkasanTemuan    `json:"ringkasan"`
	Temuan           []TemuanCompliance `json:"temuan"`
	Penjelasan       string             `json:"penjelasan"`
}

// RingkasanTemuan menghitung agregat temuan berdasarkan tingkat keparahan.
type RingkasanTemuan struct {
	Total        int `json:"total"`
	Informasi    int `json:"informasi"`
	Peringatan   int `json:"peringatan"`
	RisikoTinggi int `json:"risiko_tinggi"`
}
