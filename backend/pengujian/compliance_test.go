package pengujian

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	"sivilize-traffic-routing-demo/backend/internal/basisdata"
	"sivilize-traffic-routing-demo/backend/internal/konfigurasi"
	"sivilize-traffic-routing-demo/backend/internal/model"
	"sivilize-traffic-routing-demo/backend/internal/penangan"
)

// persiapkanRouterCompliance menyiapkan mesin router gin dan koneksi database untuk pengujian Batch 8.
func persiapkanRouterCompliance(t *testing.T) (*gin.Engine, *sql.DB) {
	gin.SetMode(gin.TestMode)

	urlDB := os.Getenv("DATABASE_URL")
	if urlDB == "" {
		konf := konfigurasi.MuatKonfigurasi()
		urlDB = konf.UrlDatabase
	}

	db, err := sql.Open("postgres", urlDB)
	if err != nil {
		t.Fatalf("Gagal membuka koneksi database: %v", err)
	}

	if err := db.Ping(); err != nil {
		t.Fatalf("Database tidak dapat dihubungi: %v", err)
	}

	if err := basisdata.JalankanMigrasi(db); err != nil {
		t.Fatalf("Gagal menjalankan migrasi: %v", err)
	}

	r := gin.New()
	api := r.Group("/api")
	{
		api.GET("/tujuan", penangan.DaftarTujuan(db))
		api.POST("/tujuan", penangan.BuatTujuan(db))
		api.DELETE("/tujuan/:id", penangan.HapusTujuan(db))

		api.GET("/aturan", penangan.DaftarAturan(db))
		api.POST("/aturan", penangan.BuatAturan(db))
		api.DELETE("/aturan/:id", penangan.HapusAturan(db))

		api.POST("/periksa", penangan.PeriksaTrafik(db, "/demo/cadangan"))

		api.GET("/compliance/check", penangan.PeriksaCompliance(db))
		api.GET("/compliance/checklist", penangan.AmbilChecklistCompliance(db))
		api.PUT("/compliance/checklist", penangan.UbahChecklistCompliance(db))
		api.GET("/compliance/audits", penangan.RiwayatAuditCompliance(db))
	}

	return r, db
}

// bersihkanDataCompliance membersihkan data uji agar tidak mempengaruhi test lain.
func bersihkanDataCompliance(t *testing.T, db *sql.DB) {
	// Hapus data audit & temuan pengujian
	_, _ = db.Exec("DELETE FROM compliance_findings")
	_, _ = db.Exec("DELETE FROM compliance_audits")
	// Hapus data uji spesifik compliance
	_, _ = db.Exec("DELETE FROM traffic_logs WHERE asal_rujukan LIKE '%compliance%' OR asal_rujukan LIKE '%uji%'")
	_, _ = db.Exec("DELETE FROM aturan WHERE nama LIKE 'Compliance %' OR nama LIKE 'Aturan Browser Firefox' OR nama LIKE 'Target Googlebot%' OR nama LIKE 'Promo Traffic%'")
	_, _ = db.Exec("DELETE FROM tujuan WHERE nama LIKE 'Compliance %' OR nama LIKE 'Mobile Landing' OR nama LIKE 'Normal Page' OR nama LIKE 'Cloaked Page' OR nama LIKE 'Promo Page' OR nama LIKE 'Test Audit' OR nama = 'Halaman Utama'")
}

// TestBatch8Compliance mencakup seluruh 9 skenario pengujian kepatuhan AdSense yang diwajibkan.
func TestBatch8Compliance(t *testing.T) {
	router, db := persiapkanRouterCompliance(t)
	defer db.Close()

	// 1. Konfigurasi normal -> tidak ada temuan risiko tinggi
	t.Run("1_KonfigurasiNormal_TidakAdaTemuanRisikoTinggi", func(t *testing.T) {
		bersihkanDataCompliance(t, db)

		// Buat 1 tujuan valid
		var tujuanID int
		err := db.QueryRow("INSERT INTO tujuan (nama, url, status) VALUES ('Halaman Utama', 'https://example.com', true) RETURNING id").Scan(&tujuanID)
		if err != nil {
			t.Fatalf("Gagal membuat tujuan: %v", err)
		}

		// Tandai semua checklist sebagai selesai untuk skenario normal
		_, err = db.Exec("UPDATE compliance_checklist SET selesai = true")
		if err != nil {
			t.Fatalf("Gagal mengupdate checklist: %v", err)
		}

		req, _ := http.NewRequest(http.MethodGet, "/api/compliance/check", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusOK {
			t.Fatalf("Status respons diharapkan 200, didapat %d", resp.Code)
		}

		var hasil model.HasilComplianceCheck
		if err := json.Unmarshal(resp.Body.Bytes(), &hasil); err != nil {
			t.Fatalf("Gagal membaca respons JSON: %v", err)
		}

		if hasil.Status != "aman" {
			t.Errorf("Status diharapkan 'aman', didapat '%s'", hasil.Status)
		}
		if hasil.Skor != nil {
			t.Errorf("Skor harus selalu null, didapat non-nil")
		}
		if hasil.Ringkasan.RisikoTinggi != 0 {
			t.Errorf("Temuan risiko tinggi diharapkan 0 pada konfigurasi normal, didapat %d", hasil.Ringkasan.RisikoTinggi)
		}
	})

	// 2. Aturan User-Agent biasa -> peringatan bila perlu
	t.Run("2_AturanUserAgentBiasa_PeringatanBilaPerlu", func(t *testing.T) {
		bersihkanDataCompliance(t, db)

		var tujuanID int
		_ = db.QueryRow("INSERT INTO tujuan (nama, url, status) VALUES ('Mobile Landing', 'https://example.com/mobile', true) RETURNING id").Scan(&tujuanID)

		// Buat aturan dengan filter UA umum (misal Firefox atau Mobile)
		_, err := db.Exec(`
			INSERT INTO aturan (nama, agen_pengguna, tujuan_id, prioritas, status)
			VALUES ('Aturan Browser Firefox', 'Firefox', $1, 10, true)
		`, tujuanID)
		if err != nil {
			t.Fatalf("Gagal membuat aturan UA: %v", err)
		}

		req, _ := http.NewRequest(http.MethodGet, "/api/compliance/check", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusOK {
			t.Fatalf("Status diharapkan 200, didapat %d", resp.Code)
		}

		var hasil model.HasilComplianceCheck
		_ = json.Unmarshal(resp.Body.Bytes(), &hasil)

		if hasil.Status != "perlu_ditinjau" {
			t.Errorf("Status diharapkan 'perlu_ditinjau', didapat '%s'", hasil.Status)
		}

		adaUAReview := false
		for _, f := range hasil.Temuan {
			if f.Kode == "UA_ROUTING_REVIEW" {
				adaUAReview = true
				if f.Tingkat != "peringatan" {
					t.Errorf("Tingkat temuan UA_ROUTING_REVIEW harus 'peringatan', didapat '%s'", f.Tingkat)
				}
			}
		}
		if !adaUAReview {
			t.Errorf("Temuan dengan kode 'UA_ROUTING_REVIEW' tidak ditemukan")
		}
	})

	// 3. Aturan yang secara eksplisit menargetkan crawler -> temuan risiko tinggi
	t.Run("3_AturanEksplisitCrawler_TemuanDanTidakDijalankan", func(t *testing.T) {
		bersihkanDataCompliance(t, db)

		var tNormal, tKhusus int
		_ = db.QueryRow("INSERT INTO tujuan (nama, url, status) VALUES ('Normal Page', 'https://example.com/normal', true) RETURNING id").Scan(&tNormal)
		_ = db.QueryRow("INSERT INTO tujuan (nama, url, status) VALUES ('Cloaked Page', 'https://example.com/review', true) RETURNING id").Scan(&tKhusus)

		// Buat aturan via API yang secara eksplisit menargetkan Googlebot
		bodyAturan := `{"nama":"Target Googlebot Crawler","agen_pengguna":"Googlebot","tujuan_id":` + fmt.Sprintf("%d", tKhusus) + `,"prioritas":1}`
		reqCreate, _ := http.NewRequest(http.MethodPost, "/api/aturan", bytes.NewBufferString(bodyAturan))
		reqCreate.Header.Set("Content-Type", "application/json")
		respCreate := httptest.NewRecorder()
		router.ServeHTTP(respCreate, reqCreate)

		if respCreate.Code != http.StatusCreated {
			t.Fatalf("Gagal membuat aturan crawler via API: %d, body: %s", respCreate.Code, respCreate.Body.String())
		}

		// Pastikan respons API memberikan peringatan
		var respJSON map[string]interface{}
		_ = json.Unmarshal(respCreate.Body.Bytes(), &respJSON)
		if _, ok := respJSON["peringatan"]; !ok {
			t.Errorf("Respons pembuatan aturan crawler wajib menyertakan peringatan")
		}

		// Jalankan compliance check
		reqCheck, _ := http.NewRequest(http.MethodGet, "/api/compliance/check", nil)
		respCheck := httptest.NewRecorder()
		router.ServeHTTP(respCheck, reqCheck)

		var hasil model.HasilComplianceCheck
		_ = json.Unmarshal(respCheck.Body.Bytes(), &hasil)

		if hasil.Status != "berisiko" {
			t.Errorf("Status audit kepatuhan diharapkan 'berisiko', didapat '%s'", hasil.Status)
		}

		adaCrawlerTargeting := false
		for _, f := range hasil.Temuan {
			if f.Kode == "CRAWLER_TARGETING_DETECTED" {
				adaCrawlerTargeting = true
				if f.Tingkat != "risiko tinggi" {
					t.Errorf("Tingkat crawler targeting harus 'risiko tinggi', didapat '%s'", f.Tingkat)
				}
			}
		}
		if !adaCrawlerTargeting {
			t.Errorf("Temuan 'CRAWLER_TARGETING_DETECTED' tidak ditemukan")
		}

		// Pastikan aturan tersebut TIDAK dijalankan ketika Googlebot datang (harus fallback)
		bodyPeriksa := `{"agen_pengguna":"Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)"}`
		reqPeriksa, _ := http.NewRequest(http.MethodPost, "/api/periksa", bytes.NewBufferString(bodyPeriksa))
		reqPeriksa.Header.Set("Content-Type", "application/json")
		respPeriksa := httptest.NewRecorder()
		router.ServeHTTP(respPeriksa, reqPeriksa)

		var respPeriksaJSON map[string]interface{}
		_ = json.Unmarshal(respPeriksa.Body.Bytes(), &respPeriksaJSON)
		dataPeriksa := respPeriksaJSON["data"].(map[string]interface{})
		if dataPeriksa["hasil"] != "fallback" {
			t.Errorf("Aturan yang menargetkan crawler tidak boleh dijalankan (harus fallback), didapat '%v'", dataPeriksa["hasil"])
		}
	})

	// 4. Konfigurasi traffic source berisiko -> temuan
	t.Run("4_KonfigurasiTrafficSourceBerisiko_Temuan", func(t *testing.T) {
		bersihkanDataCompliance(t, db)

		var tujuanID int
		_ = db.QueryRow("INSERT INTO tujuan (nama, url, status) VALUES ('Promo Page', 'https://example.com/promo', true) RETURNING id").Scan(&tujuanID)

		// Buat aturan dengan asal rujukan berisiko (traffic-exchange)
		_, err := db.Exec(`
			INSERT INTO aturan (nama, asal_rujukan, tujuan_id, prioritas, status)
			VALUES ('Promo Traffic Exchange', 'https://traffic-exchange-network.com', $1, 20, true)
		`, tujuanID)
		if err != nil {
			t.Fatalf("Gagal membuat aturan traffic source berisiko: %v", err)
		}

		req, _ := http.NewRequest(http.MethodGet, "/api/compliance/check", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		var hasil model.HasilComplianceCheck
		_ = json.Unmarshal(resp.Body.Bytes(), &hasil)

		if hasil.Status != "berisiko" {
			t.Errorf("Status kepatuhan diharapkan 'berisiko' karena sumber traffic exchange, didapat '%s'", hasil.Status)
		}

		adaRiskySource := false
		for _, f := range hasil.Temuan {
			if f.Kode == "RISKY_TRAFFIC_SOURCE" {
				adaRiskySource = true
				if f.Tingkat != "risiko tinggi" {
					t.Errorf("Tingkat temuan sumber berisiko harus 'risiko tinggi', didapat '%s'", f.Tingkat)
				}
			}
		}
		if !adaRiskySource {
			t.Errorf("Temuan 'RISKY_TRAFFIC_SOURCE' tidak ditemukan")
		}
	})

	// 5. Checklist manual tersimpan di database
	t.Run("5_ChecklistTersimpan", func(t *testing.T) {
		// Update satu item checklist
		body := `{"kunci":"no_self_clicks","selesai":true}`
		reqPut, _ := http.NewRequest(http.MethodPut, "/api/compliance/checklist", bytes.NewBufferString(body))
		reqPut.Header.Set("Content-Type", "application/json")
		respPut := httptest.NewRecorder()
		router.ServeHTTP(respPut, reqPut)

		if respPut.Code != http.StatusOK {
			t.Fatalf("PUT checklist gagal: %d, body: %s", respPut.Code, respPut.Body.String())
		}

		// Ambil daftar checklist dan verifikasi
		reqGet, _ := http.NewRequest(http.MethodGet, "/api/compliance/checklist", nil)
		respGet := httptest.NewRecorder()
		router.ServeHTTP(respGet, reqGet)

		var respData struct {
			Data []model.ItemChecklistCompliance `json:"data"`
		}
		_ = json.Unmarshal(respGet.Body.Bytes(), &respData)

		itemDitemukan := false
		for _, item := range respData.Data {
			if item.Kunci == "no_self_clicks" {
				itemDitemukan = true
				if !item.Selesai {
					t.Errorf("Item checklist 'no_self_clicks' harus bernilai true setelah disimpan")
				}
			}
		}
		if !itemDitemukan {
			t.Errorf("Item checklist 'no_self_clicks' tidak ditemukan di database")
		}
	})

	// 6. Audit tersimpan ke database (compliance_audits dan compliance_findings)
	t.Run("6_AuditTersimpan", func(t *testing.T) {
		bersihkanDataCompliance(t, db)

		var tujuanID int
		_ = db.QueryRow("INSERT INTO tujuan (nama, url, status) VALUES ('Test Audit', 'https://example.com', true) RETURNING id").Scan(&tujuanID)

		// Panggil endpoint check
		req, _ := http.NewRequest(http.MethodGet, "/api/compliance/check", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		var hasil model.HasilComplianceCheck
		_ = json.Unmarshal(resp.Body.Bytes(), &hasil)

		if hasil.AuditID <= 0 {
			t.Fatalf("Audit ID harus berupa angka positif, didapat %d", hasil.AuditID)
		}

		// Verifikasi di tabel database
		var dbStatus string
		err := db.QueryRow("SELECT status FROM compliance_audits WHERE id = $1", hasil.AuditID).Scan(&dbStatus)
		if err != nil {
			t.Fatalf("Audit ID %d tidak ditemukan di compliance_audits: %v", hasil.AuditID, err)
		}
		if dbStatus != hasil.Status {
			t.Errorf("Status di database '%s' tidak cocok dengan respon API '%s'", dbStatus, hasil.Status)
		}

		// Verifikasi relasi foreign key di compliance_findings
		var countFindings int
		_ = db.QueryRow("SELECT COUNT(*) FROM compliance_findings WHERE audit_id = $1", hasil.AuditID).Scan(&countFindings)
		if countFindings != len(hasil.Temuan) {
			t.Errorf("Jumlah findings di DB (%d) tidak cocok dengan temuan di response (%d)", countFindings, len(hasil.Temuan))
		}
	})

	// 7. Compliance API mengembalikan struktur yang benar
	t.Run("7_ComplianceAPIMengembalikanStrukturYangBenar", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/compliance/check", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		var rawMap map[string]interface{}
		if err := json.Unmarshal(resp.Body.Bytes(), &rawMap); err != nil {
			t.Fatalf("Respons bukan format JSON yang valid: %v", err)
		}

		// Verifikasi kunci-kunci wajib
		fieldWajib := []string{"status", "skor", "temuan", "ringkasan", "waktu_pemeriksaan", "penjelasan"}
		for _, f := range fieldWajib {
			if _, ada := rawMap[f]; !ada {
				t.Errorf("Field wajib '%s' tidak ditemukan dalam respons compliance check", f)
			}
		}

		// Skor HARUS null
		if rawMap["skor"] != nil {
			t.Errorf("Field 'skor' mutlak harus null, didapat: %v", rawMap["skor"])
		}

		// Status harus salah satu dari: 'aman', 'perlu_ditinjau', 'berisiko'
		statusStr, _ := rawMap["status"].(string)
		if statusStr != "aman" && statusStr != "perlu_ditinjau" && statusStr != "berisiko" {
			t.Errorf("Status '%s' tidak valid. Harus 'aman', 'perlu_ditinjau', atau 'berisiko'", statusStr)
		}
	})

	// 8. Tidak ada secret atau kredensial yang masuk response
	t.Run("8_TidakAdaSecretYangMasukResponse", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/compliance/check", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		bodyText := strings.ToLower(resp.Body.String())
		kataSensitif := []string{"password", "secret", "postgres://", "private_key", "bearer"}
		for _, kata := range kataSensitif {
			if strings.Contains(bodyText, kata) {
				t.Errorf("Respons API mengandung data sensitif '%s'", kata)
			}
		}
	})

	// 9. Tidak ada data sensitif dalam audit log
	t.Run("9_TidakAdaDataSensitifDalamAuditLog", func(t *testing.T) {
		rows, err := db.Query("SELECT kode, judul, pesan FROM compliance_findings LIMIT 50")
		if err != nil {
			t.Fatalf("Gagal query compliance_findings: %v", err)
		}
		defer rows.Close()

		for rows.Next() {
			var kode, judul, pesan string
			_ = rows.Scan(&kode, &judul, &pesan)
			gabungan := strings.ToLower(kode + " " + judul + " " + pesan)
			if strings.Contains(gabungan, "password") || strings.Contains(gabungan, "token") || strings.Contains(gabungan, "bearer") {
				t.Errorf("Ditemukan data sensitif dalam tabel compliance_findings: %s", gabungan)
			}
		}
		if err := rows.Err(); err != nil {
			t.Errorf("Error saat membaca compliance_findings: %v", err)
		}
	})
}

// TestMigrasiBasisDataKosong memverifikasi bahwa migrasi berhasil dijalankan dari basis data yang benar-benar kosong.
func TestMigrasiBasisDataKosong(t *testing.T) {
	urlDB := os.Getenv("DATABASE_URL")
	if urlDB == "" {
		konf := konfigurasi.MuatKonfigurasi()
		urlDB = konf.UrlDatabase
	}

	dbUtama, err := sql.Open("postgres", urlDB)
	if err != nil {
		t.Fatalf("Gagal terhubung ke database utama: %v", err)
	}
	defer dbUtama.Close()

	namaDBTemp := fmt.Sprintf("sivilize_kosong_%d", os.Getpid())
	// Buat database baru yang kosong
	_, err = dbUtama.Exec("CREATE DATABASE " + namaDBTemp)
	if err != nil {
		t.Fatalf("Gagal membuat database kosong '%s': %v", namaDBTemp, err)
	}
	defer func() {
		_, _ = dbUtama.Exec("DROP DATABASE " + namaDBTemp)
	}()

	// Buka koneksi ke database yang baru dibuat (benar-benar kosong)
	urlTemp := strings.Replace(urlDB, "sivilize_trafik", namaDBTemp, 1)
	dbKosong, err := sql.Open("postgres", urlTemp)
	if err != nil {
		t.Fatalf("Gagal terhubung ke database kosong: %v", err)
	}
	defer dbKosong.Close()

	// Jalankan migrasi penuh dari Batch 1 sampai Batch 8
	if err := basisdata.JalankanMigrasi(dbKosong); err != nil {
		t.Fatalf("JalankanMigrasi gagal pada database kosong: %v", err)
	}

	// Verifikasi tabel-tabel utama berhasil dibuat
	tabelWajib := []string{
		"tujuan",
		"aturan",
		"catatan_trafik",
		"traffic_logs",
		"compliance_checklist",
		"compliance_audits",
		"compliance_findings",
	}

	for _, tabel := range tabelWajib {
		var exists bool
		err := dbKosong.QueryRow(`
			SELECT EXISTS (
				SELECT FROM information_schema.tables 
				WHERE table_schema = 'public' AND table_name = $1
			)
		`, tabel).Scan(&exists)
		if err != nil || !exists {
			t.Errorf("Tabel '%s' tidak ditemukan setelah migrasi database kosong (err: %v)", tabel, err)
		}
	}

	// Verifikasi 11 checklist default terisi
	var jumlahChecklist int
	err = dbKosong.QueryRow("SELECT COUNT(*) FROM compliance_checklist").Scan(&jumlahChecklist)
	if err != nil || jumlahChecklist != 11 {
		t.Errorf("Jumlah checklist setelah migrasi harus 11, didapat %d (err: %v)", jumlahChecklist, err)
	}
}
