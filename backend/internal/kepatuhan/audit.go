package kepatuhan

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"sivilize-traffic-routing-demo/backend/internal/model"
)

// Kata kunci crawler/reviewer yang dilarang dijadikan target perlakuan khusus
var polaCrawlerReviewer = []string{
	"googlebot",
	"adsbot",
	"mediapartners-google",
	"google-inspectiontool",
	"google reviewer",
	"adsense crawler",
	"sistem review",
	"review iklan",
	"crawler",
	"reviewer",
	"spider",
	"googlebot-image",
	"googlebot-news",
	"googlebot-video",
}

// Kata kunci sumber trafik berisiko (invalid traffic, traffic exchange, PTC)
var polaSumberBerisiko = []string{
	"autosurf",
	"traffic-exchange",
	"trafficexchange",
	"ptc",
	"paid-to-click",
	"paid-to-surf",
	"hitleap",
	"click-exchange",
	"auto-surf",
	"bot-traffic",
	"click-farm",
}

// Kata kunci indikasi manipulasi atau cloaking pada nama aturan
var polaManipulatif = []string{
	"cloak",
	"cloaking",
	"bypass",
	"kelabui",
	"fake",
	"stealth",
	"hidden",
	"tipu",
}

// MenargetkanCrawlerReviewer memeriksa apakah nama aturan, agen pengguna, atau asal rujukan
// secara eksplisit menargetkan crawler atau reviewer iklan.
func MenargetkanCrawlerReviewer(nama, agenPengguna, asalRujukan string) bool {
	gabungan := strings.ToLower(nama + " " + agenPengguna + " " + asalRujukan)
	for _, pola := range polaCrawlerReviewer {
		if strings.Contains(gabungan, pola) {
			return true
		}
	}
	return false
}

// MenargetkanCrawlerAturan memeriksa model.Aturan apakah menargetkan crawler/reviewer.
func MenargetkanCrawlerAturan(a model.Aturan) bool {
	var ua, ref string
	if a.AgenPengguna.Valid {
		ua = a.AgenPengguna.String
	}
	if a.AsalRujukan.Valid {
		ref = a.AsalRujukan.String
	}
	return MenargetkanCrawlerReviewer(a.Nama, ua, ref)
}

// DeteksiSumberTrafikBerisiko mengecek apakah string asal rujukan mengandung domain atau pola berbahaya.
func DeteksiSumberTrafikBerisiko(asalRujukan string) (bool, string) {
	refLower := strings.ToLower(strings.TrimSpace(asalRujukan))
	if refLower == "" {
		return false, ""
	}
	for _, pola := range polaSumberBerisiko {
		if strings.Contains(refLower, pola) {
			return true, pola
		}
	}
	return false, ""
}

// LakukanAuditKepatuhan menjalankan pemeriksaan kepatuhan Google AdSense secara internal
// dan menyimpan catatan audit ke database.
func LakukanAuditKepatuhan(db *sql.DB) (model.HasilComplianceCheck, error) {
	waktuAudit := time.Now()
	var temuan []model.TemuanCompliance

	// 1. Ambil seluruh data tujuan
	rowsTujuan, err := db.Query("SELECT id, nama, url, status FROM tujuan")
	if err != nil {
		return model.HasilComplianceCheck{}, fmt.Errorf("gagal memeriksa tabel tujuan: %w", err)
	}
	defer rowsTujuan.Close()

	petaTujuan := make(map[int]model.Tujuan)
	for rowsTujuan.Next() {
		var t model.Tujuan
		if err := rowsTujuan.Scan(&t.ID, &t.Nama, &t.URL, &t.Status); err != nil {
			return model.HasilComplianceCheck{}, fmt.Errorf("gagal membaca data tujuan: %w", err)
		}
		petaTujuan[t.ID] = t

		// Cek pemeriksaan 2: Tujuan mencurigakan atau tidak valid
		urlClean := strings.ToLower(strings.TrimSpace(t.URL))
		if !strings.HasPrefix(urlClean, "http://") && !strings.HasPrefix(urlClean, "https://") && !strings.HasPrefix(urlClean, "/") {
			temuan = append(temuan, model.TemuanCompliance{
				Kode:    "DESTINATION_SUSPICIOUS",
				Tingkat: "peringatan",
				Judul:   fmt.Sprintf("Tujuan '%s' menggunakan skema URL tidak standar", t.Nama),
				Pesan:   fmt.Sprintf("URL tujuan '%s' tidak menggunakan protokol standar HTTP/HTTPS. Pastikan halaman tujuan sah dan aman bagi pengguna.", t.URL),
			})
		}
	}

	// 2. Ambil seluruh data aturan (termasuk nonaktif untuk audit menyeluruh)
	kueriAturan := `
		SELECT id, nama, negara, perangkat, peramban, agen_pengguna, asal_rujukan, tujuan_id, prioritas, status
		FROM aturan
	`
	rowsAturan, err := db.Query(kueriAturan)
	if err != nil {
		return model.HasilComplianceCheck{}, fmt.Errorf("gagal memeriksa tabel aturan: %w", err)
	}
	defer rowsAturan.Close()

	adaUARouting := false
	adaCrawlerTargeting := false
	adaDifferentialTreatment := false

	for rowsAturan.Next() {
		var a model.Aturan
		var ua, ref sql.NullString
		if err := rowsAturan.Scan(&a.ID, &a.Nama, &a.Negara, &a.Perangkat, &a.Peramban, &ua, &ref, &a.TujuanID, &a.Prioritas, &a.Status); err != nil {
			return model.HasilComplianceCheck{}, fmt.Errorf("gagal membaca data aturan: %w", err)
		}
		a.AgenPengguna = ua
		a.AsalRujukan = ref

		uaStr := ""
		if ua.Valid {
			uaStr = strings.TrimSpace(ua.String)
		}
		refStr := ""
		if ref.Valid {
			refStr = strings.TrimSpace(ref.String)
		}

		// Cek pemeriksaan 1 & 6: Menargetkan crawler/reviewer
		if MenargetkanCrawlerReviewer(a.Nama, uaStr, refStr) {
			adaCrawlerTargeting = true
			adaDifferentialTreatment = true
			temuan = append(temuan, model.TemuanCompliance{
				Kode:    "CRAWLER_TARGETING_DETECTED",
				Tingkat: "risiko tinggi",
				Judul:   fmt.Sprintf("Aturan '%s' menargetkan crawler/reviewer terdeteksi", a.Nama),
				Pesan:   "Aturan ini perlu ditinjau karena membedakan pengunjung berdasarkan identitas crawler/reviewer dapat digunakan untuk menghindari pemeriksaan atau kebijakan platform.",
			})
		}

		// Cek pemeriksaan 4: Sumber trafik berisiko (autosurf, PTC, exchange)
		if berisiko, pola := DeteksiSumberTrafikBerisiko(refStr); berisiko {
			temuan = append(temuan, model.TemuanCompliance{
				Kode:    "RISKY_TRAFFIC_SOURCE",
				Tingkat: "risiko tinggi",
				Judul:   fmt.Sprintf("Sumber traffic berisiko pada aturan '%s'", a.Nama),
				Pesan:   fmt.Sprintf("Asal rujukan '%s' mengandung indikasi sumber traffic berisiko ('%s'). Kebijakan Google AdSense melarang penggunaan traffic exchange, autosurf, dan paid-to-click.", refStr, pola),
			})
		}

		// Cek pemeriksaan 3: Konfigurasi berpotensi menyesatkan atau manipulatif
		namaLower := strings.ToLower(a.Nama)
		for _, m := range polaManipulatif {
			if strings.Contains(namaLower, m) {
				temuan = append(temuan, model.TemuanCompliance{
					Kode:    "REDIRECT_MISLEADING_POTENTIAL",
					Tingkat: "peringatan",
					Judul:   fmt.Sprintf("Penamaan aturan '%s' mengindikasikan upaya cloaking/manipulasi", a.Nama),
					Pesan:   "Nama aturan mengindikasikan pola penyembunyian halaman atau bypass. Hindari penggunaan mekanisme cloaking pada situs beriklan.",
				})
				break
			}
		}

		// Cek apakah tujuan aturan tidak aktif
		if t, ok := petaTujuan[a.TujuanID]; ok {
			if a.Status && !t.Status {
				temuan = append(temuan, model.TemuanCompliance{
					Kode:    "INACTIVE_DESTINATION_REFERENCED",
					Tingkat: "peringatan",
					Judul:   fmt.Sprintf("Aturan aktif '%s' mengarah ke tujuan nonaktif", a.Nama),
					Pesan:   fmt.Sprintf("Tujuan '%s' (ID %d) berstatus nonaktif tetapi masih dirujuk oleh aturan aktif. Ini dapat menyebabkan fallback tak terduga.", t.Nama, t.ID),
				})
			}
		}

		// Cek pemeriksaan 5: Adanya filter User-Agent
		if uaStr != "" && !adaCrawlerTargeting {
			adaUARouting = true
		}
	}

	// Pemeriksaan 5: Jika ada routing berbasis UA secara umum
	if adaUARouting {
		temuan = append(temuan, model.TemuanCompliance{
			Kode:    "UA_ROUTING_REVIEW",
			Tingkat: "peringatan",
			Judul:   "Routing berdasarkan User-Agent perlu ditinjau",
			Pesan:   "Routing berdasarkan User-Agent dapat memiliki penggunaan yang sah, tetapi perlu dipastikan tidak digunakan untuk membedakan crawler/reviewer dari pengguna dengan tujuan menghindari pemeriksaan.",
		})
	}

	// Pemeriksaan 6: Potensi perlakuan diskriminatif terhadap crawler
	if adaDifferentialTreatment {
		temuan = append(temuan, model.TemuanCompliance{
			Kode:    "DIFFERENTIAL_TREATMENT_CRAWLER",
			Tingkat: "risiko tinggi",
			Judul:   "Potensi perlakuan berbeda pada crawler/reviewer",
			Pesan:   "Ditemukan konfigurasi yang berpotensi membedakan perlakuan antara crawler/reviewer dan pengunjung manusia. Hal ini melanggar pedoman Google Publisher mengenai cloaking.",
		})
	}

	// Cek pemeriksaan 7: Checklist manual kepatuhan
	rowsChecklist, err := db.Query("SELECT kunci, label, selesai FROM compliance_checklist")
	if err == nil {
		defer rowsChecklist.Close()
		totalChecklist := 0
		selesaiCount := 0
		for rowsChecklist.Next() {
			totalChecklist++
			var k, l string
			var s bool
			if err := rowsChecklist.Scan(&k, &l, &s); err == nil && s {
				selesaiCount++
			}
		}

		if totalChecklist > 0 && selesaiCount < totalChecklist {
			temuan = append(temuan, model.TemuanCompliance{
				Kode:    "CHECKLIST_PENDING",
				Tingkat: "informasi",
				Judul:   "Checklist kepatuhan internal belum selesai",
				Pesan:   fmt.Sprintf("%d dari %d poin checklist kepatuhan manual belum ditandai oleh administrator. Lakukan verifikasi berkala terhadap operasional situs.", totalChecklist-selesaiCount, totalChecklist),
			})
		}
	}

	// Hitung agregat tingkat temuan dan tentukan status akhir
	ringkasan := model.RingkasanTemuan{
		Total: len(temuan),
	}
	status := "aman"

	for _, t := range temuan {
		switch t.Tingkat {
		case "risiko tinggi":
			ringkasan.RisikoTinggi++
			status = "berisiko"
		case "peringatan":
			ringkasan.Peringatan++
			if status != "berisiko" {
				status = "perlu_ditinjau"
			}
		case "informasi":
			ringkasan.Informasi++
		}
	}

	// Simpan riwayat audit ke tabel compliance_audits
	var auditID int
	err = db.QueryRow(
		"INSERT INTO compliance_audits (status, dibuat_pada) VALUES ($1, $2) RETURNING id",
		status, waktuAudit,
	).Scan(&auditID)
	if err != nil {
		return model.HasilComplianceCheck{}, fmt.Errorf("gagal mencatat compliance_audits: %w", err)
	}

	// Simpan setiap temuan ke tabel compliance_findings
	for i := range temuan {
		temuan[i].AuditID = auditID
		temuan[i].DibuatPada = waktuAudit

		var fid int
		err = db.QueryRow(`
			INSERT INTO compliance_findings (audit_id, kode, tingkat, judul, pesan, dibuat_pada)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id
		`, auditID, temuan[i].Kode, temuan[i].Tingkat, temuan[i].Judul, temuan[i].Pesan, waktuAudit).Scan(&fid)
		if err == nil {
			temuan[i].ID = fid
		}
	}

	hasil := model.HasilComplianceCheck{
		Status:           status,
		Skor:             nil, // Mutlak nil / null sesuai spesifikasi
		AuditID:          auditID,
		WaktuPemeriksaan: waktuAudit,
		Ringkasan:        ringkasan,
		Temuan:           temuan,
		Penjelasan:       "Status ini merupakan pemeriksaan internal berdasarkan dokumentasi kebijakan Google yang tersedia dan bukan keputusan resmi Google.",
	}

	return hasil, nil
}

// AmbilChecklist mengambil seluruh daftar checklist manual dari database.
func AmbilChecklist(db *sql.DB) ([]model.ItemChecklistCompliance, error) {
	kueri := "SELECT id, kunci, label, selesai, diperbarui_pada FROM compliance_checklist ORDER BY id ASC"
	rows, err := db.Query(kueri)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil data checklist: %w", err)
	}
	defer rows.Close()

	items := make([]model.ItemChecklistCompliance, 0)
	for rows.Next() {
		var item model.ItemChecklistCompliance
		if err := rows.Scan(&item.ID, &item.Kunci, &item.Label, &item.Selesai, &item.DiperbaruiPada); err != nil {
			return nil, fmt.Errorf("gagal memproses baris checklist: %w", err)
		}
		items = append(items, item)
	}
	return items, nil
}

// PerbaruiItemChecklist memperbarui status satu item checklist manual.
func PerbaruiItemChecklist(db *sql.DB, kunci string, selesai bool) error {
	kueri := "UPDATE compliance_checklist SET selesai = $1, diperbarui_pada = NOW() WHERE kunci = $2"
	res, err := db.Exec(kueri, selesai, kunci)
	if err != nil {
		return fmt.Errorf("gagal memperbarui checklist: %w", err)
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("item checklist dengan kunci '%s' tidak ditemukan", kunci)
	}
	return nil
}

// AmbilRiwayatAudit mengambil daftar audit terakhir beserta temuan-temuannya.
func AmbilRiwayatAudit(db *sql.DB, limit int) ([]model.AuditCompliance, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	kueriAudit := "SELECT id, status, dibuat_pada FROM compliance_audits ORDER BY dibuat_pada DESC LIMIT $1"
	rows, err := db.Query(kueriAudit, limit)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil riwayat audit: %w", err)
	}
	defer rows.Close()

	var daftar []model.AuditCompliance
	for rows.Next() {
		var a model.AuditCompliance
		if err := rows.Scan(&a.ID, &a.Status, &a.DibuatPada); err != nil {
			return nil, fmt.Errorf("gagal memproses baris audit: %w", err)
		}
		daftar = append(daftar, a)
	}

	// Isi temuan untuk setiap audit
	for i := range daftar {
		kueriFinding := "SELECT id, audit_id, kode, tingkat, judul, pesan, dibuat_pada FROM compliance_findings WHERE audit_id = $1 ORDER BY id ASC"
		fRows, err := db.Query(kueriFinding, daftar[i].ID)
		if err != nil {
			continue
		}
		var findings []model.TemuanCompliance
		for fRows.Next() {
			var f model.TemuanCompliance
			if err := fRows.Scan(&f.ID, &f.AuditID, &f.Kode, &f.Tingkat, &f.Judul, &f.Pesan, &f.DibuatPada); err == nil {
				findings = append(findings, f)
			}
		}
		fRows.Close()
		daftar[i].Temuan = findings
	}

	return daftar, nil
}
