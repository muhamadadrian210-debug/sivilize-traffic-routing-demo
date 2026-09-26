package aturan

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"

	"sivilize-traffic-routing-demo/backend/internal/kepatuhan"
	"sivilize-traffic-routing-demo/backend/internal/model"
)

// HasilRouting menyimpan keputusan akhir dari evaluasi traffic routing.
type HasilRouting struct {
	Cocok     bool   `json:"cocok"`
	AturanID  *int   `json:"aturan_id"`
	TujuanID  *int   `json:"tujuan_id"`
	URLTujuan string `json:"url_tujuan"`
	Hasil     string `json:"hasil"` // "aturan" atau "fallback"
	Alasan    string `json:"alasan"`
}

// CocokkanAturan mengecek apakah satu aturan cocok dengan konteks pengunjung.
// Aturan dianggap cocok jika SEMUA kondisi yang terisi (bukan NULL) sesuai dengan konteks.
// Kolom NULL diartikan sebagai wildcard (tidak membatasi filter).
func CocokkanAturan(a model.Aturan, k KonteksPengunjung) bool {
	// 1. Hanya aturan dengan status aktif yang dievaluasi
	if !a.Status {
		return false
	}

	// 1.1 Kepatuhan Google Publisher (Batch 8):
	// Jangan jalankan aturan yang secara eksplisit menargetkan crawler/reviewer Google/AdSense
	// untuk mencegah mekanisme cloaking, bot evasion, atau manipulasi peninjauan iklan.
	if kepatuhan.MenargetkanCrawlerAturan(a) {
		return false
	}

	// 2. Filter Negara (jika diatur pada aturan)
	if a.Negara.Valid && strings.TrimSpace(a.Negara.String) != "" {
		if NormalisasiNegara(a.Negara.String) != NormalisasiNegara(k.Negara) {
			return false
		}
	}

	// 3. Filter Perangkat (jika diatur pada aturan)
	if a.Perangkat.Valid && strings.TrimSpace(a.Perangkat.String) != "" {
		if NormalisasiPerangkat(a.Perangkat.String) != NormalisasiPerangkat(k.Perangkat) {
			return false
		}
	}

	// 4. Filter Peramban (jika diatur pada aturan)
	if a.Peramban.Valid && strings.TrimSpace(a.Peramban.String) != "" {
		if NormalisasiPeramban(a.Peramban.String) != NormalisasiPeramban(k.Peramban) {
			return false
		}
	}

	// 5. Filter User-Agent (pencocokan substring case-insensitive)
	if a.AgenPengguna.Valid && strings.TrimSpace(a.AgenPengguna.String) != "" {
		polaUA := strings.ToLower(strings.TrimSpace(a.AgenPengguna.String))
		if !strings.Contains(strings.ToLower(k.AgenPengguna), polaUA) {
			return false
		}
	}

	// 6. Filter Referrer / Asal Rujukan (pencocokan substring case-insensitive)
	if a.AsalRujukan.Valid && strings.TrimSpace(a.AsalRujukan.String) != "" {
		polaRef := strings.ToLower(strings.TrimSpace(a.AsalRujukan.String))
		if !strings.Contains(strings.ToLower(k.AsalRujukan), polaRef) {
			return false
		}
	}

	return true
}

// EvaluasiKonteks mencocokkan KonteksPengunjung dengan daftar aturan yang tersedia.
// Jika beberapa aturan cocok, dipilih yang memiliki prioritas tertinggi (angka prioritas terkecil).
// Jika prioritas sama, keputusan dibuat deterministik berdasarkan ID terkecil.
// Jika tidak ada aturan yang cocok atau tujuan tidak valid, mengembalikan fallback.
func EvaluasiKonteks(daftarAturan []model.Aturan, k KonteksPengunjung, urlFallback string) HasilRouting {
	if strings.TrimSpace(urlFallback) == "" {
		urlFallback = "/demo/cadangan"
	}

	cocok := make([]model.Aturan, 0)
	for _, a := range daftarAturan {
		if CocokkanAturan(a, k) {
			cocok = append(cocok, a)
		}
	}

	// Jika tidak ada aturan yang cocok, gunakan fallback
	if len(cocok) == 0 {
		return HasilRouting{
			Cocok:     false,
			AturanID:  nil,
			TujuanID:  nil,
			URLTujuan: urlFallback,
			Hasil:     "fallback",
			Alasan:    "Tidak ada aturan aktif yang cocok",
		}
	}

	// Urutkan aturan yang cocok berdasarkan prioritas (angka kecil = prioritas tinggi).
	// Jika prioritas bernilai sama, gunakan ID aturan terkecil sebagai pemutus deterministik.
	sort.SliceStable(cocok, func(i, j int) bool {
		if cocok[i].Prioritas != cocok[j].Prioritas {
			return cocok[i].Prioritas < cocok[j].Prioritas
		}
		return cocok[i].ID < cocok[j].ID
	})

	terpilih := cocok[0]

	// Pastikan tujuan yang dituju tersedia
	if strings.TrimSpace(terpilih.TujuanURL) == "" {
		return HasilRouting{
			Cocok:     false,
			AturanID:  nil,
			TujuanID:  nil,
			URLTujuan: urlFallback,
			Hasil:     "fallback",
			Alasan:    fmt.Sprintf("Aturan '%s' (ID %d) cocok namun tujuan tidak tersedia", terpilih.Nama, terpilih.ID),
		}
	}

	aturanID := terpilih.ID
	tujuanID := terpilih.TujuanID

	return HasilRouting{
		Cocok:     true,
		AturanID:  &aturanID,
		TujuanID:  &tujuanID,
		URLTujuan: terpilih.TujuanURL,
		Hasil:     "aturan",
		Alasan:    fmt.Sprintf("Aturan '%s' (ID %d) cocok dengan permintaan pengunjung", terpilih.Nama, terpilih.ID),
	}
}

// EvaluasiDenganBasisData memuat aturan aktif langsung dari PostgreSQL dan mengevaluasinya.
func EvaluasiDenganBasisData(db *sql.DB, k KonteksPengunjung, urlFallback string) (HasilRouting, error) {
	kueri := `
		SELECT 
			a.id, a.nama, a.negara, a.perangkat, a.peramban, a.agen_pengguna, a.asal_rujukan,
			a.tujuan_id, t.nama AS tujuan_nama, t.url AS tujuan_url,
			a.prioritas, a.status, a.dibuat_pada, a.diperbarui_pada
		FROM aturan a
		JOIN tujuan t ON a.tujuan_id = t.id
		WHERE a.status = TRUE AND t.status = TRUE
		ORDER BY a.prioritas ASC, a.id ASC
	`

	baris, err := db.Query(kueri)
	if err != nil {
		return HasilRouting{}, fmt.Errorf("gagal mengambil aturan aktif dari database: %w", err)
	}
	defer baris.Close()

	daftar := make([]model.Aturan, 0)
	for baris.Next() {
		var a model.Aturan
		err := baris.Scan(
			&a.ID, &a.Nama, &a.Negara, &a.Perangkat, &a.Peramban, &a.AgenPengguna, &a.AsalRujukan,
			&a.TujuanID, &a.TujuanNama, &a.TujuanURL,
			&a.Prioritas, &a.Status, &a.DibuatPada, &a.DiperbaruiPada,
		)
		if err != nil {
			return HasilRouting{}, fmt.Errorf("gagal memproses baris aturan: %w", err)
		}
		daftar = append(daftar, a)
	}

	if err := baris.Err(); err != nil {
		return HasilRouting{}, fmt.Errorf("kesalahan saat membaca aturan: %w", err)
	}

	return EvaluasiKonteks(daftar, k, urlFallback), nil
}
