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

// persiapkanRouterBatch5 menyiapkan mesin router Gin lengkap dengan endpoint Batch 1 - 5.
func persiapkanRouterBatch5(t *testing.T) (*gin.Engine, *sql.DB, string) {
	gin.SetMode(gin.TestMode)

	urlDB := os.Getenv("DATABASE_URL")
	urlFallback := "/demo/cadangan"
	if urlDB == "" {
		konf := konfigurasi.MuatKonfigurasi()
		urlDB = konf.UrlDatabase
		urlFallback = konf.UrlFallback
	}

	db, err := sql.Open("postgres", urlDB)
	if err != nil {
		t.Fatalf("Gagal membuka koneksi database: %v", err)
	}

	if err := db.Ping(); err != nil {
		t.Fatalf("Database tidak dapat dihubungi: %v", err)
	}

	// Pastikan migrasi tabel Batch 1-5 sudah berjalan
	if err := basisdata.JalankanMigrasi(db); err != nil {
		t.Fatalf("Gagal menjalankan migrasi: %v", err)
	}

	r := gin.New()
	api := r.Group("/api")
	{
		api.GET("/kesehatan", penangan.PeriksaKesehatan(db))
		api.GET("/tujuan", penangan.DaftarTujuan(db))
		api.GET("/tujuan/:id", penangan.SatuTujuan(db))
		api.POST("/tujuan", penangan.BuatTujuan(db))
		api.PUT("/tujuan/:id", penangan.UbahTujuan(db))
		api.DELETE("/tujuan/:id", penangan.HapusTujuan(db))

		api.GET("/aturan", penangan.DaftarAturan(db))
		api.GET("/aturan/:id", penangan.SatuAturan(db))
		api.POST("/aturan", penangan.BuatAturan(db))
		api.PUT("/aturan/:id", penangan.UbahAturan(db))
		api.PATCH("/aturan/:id/status", penangan.UbahStatusAturan(db))
		api.DELETE("/aturan/:id", penangan.HapusAturan(db))

		// Batch 5
		api.POST("/periksa", penangan.PeriksaTrafik(db, urlFallback))
		api.GET("/log-trafik", penangan.DaftarLogTrafik(db))
		api.GET("/log-trafik/:id", penangan.SatuLogTrafik(db))
	}

	return r, db, urlFallback
}

// TestBatch5EndpointPeriksaDanLog menguji 23 skenario yang diwajibkan untuk Batch 5.
func TestBatch5EndpointPeriksaDanLog(t *testing.T) {
	r, db, _ := persiapkanRouterBatch5(t)
	defer db.Close()

	// Siapkan data tujuan dan aturan khusus untuk pengujian Batch 5
	var idTujuanA, idTujuanB int
	urlTujuanA := fmt.Sprintf("/demo/tujuan-a-b5-%d", os.Getpid())
	urlTujuanB := fmt.Sprintf("/demo/tujuan-b-b5-%d", os.Getpid())

	// Bersihkan data tes jika ada sisa run sebelumnya (urutkan dari foreign key anak terlebih dahulu)
	db.Exec("DELETE FROM traffic_logs WHERE aturan_id IN (SELECT id FROM aturan WHERE nama LIKE 'Aturan B5%')")
	db.Exec("DELETE FROM aturan WHERE nama LIKE 'Aturan B5%'")
	db.Exec("DELETE FROM tujuan WHERE url LIKE '/demo/tujuan-%-b5-%'")

	err := db.QueryRow(`
		INSERT INTO tujuan (nama, url, status)
		VALUES ('Tujuan A Batch 5', $1, true)
		RETURNING id
	`, urlTujuanA).Scan(&idTujuanA)
	if err != nil {
		t.Fatalf("Gagal membuat tujuan A: %v", err)
	}

	err = db.QueryRow(`
		INSERT INTO tujuan (nama, url, status)
		VALUES ('Tujuan B Batch 5', $1, true)
		RETURNING id
	`, urlTujuanB).Scan(&idTujuanB)
	if err != nil {
		t.Fatalf("Gagal membuat tujuan B: %v", err)
	}

	namaAturan1 := fmt.Sprintf("Aturan B5 MY Mobile %d", os.Getpid())
	namaAturan2 := fmt.Sprintf("Aturan B5 MY Mobile Prioritas Tinggi %d", os.Getpid())
	namaAturan3 := fmt.Sprintf("Aturan B5 SG Inactive %d", os.Getpid())

	// Buat Aturan 1: MY + mobile -> Tujuan A (Prioritas 10)
	var idAturan1 int
	err = db.QueryRow(`
		INSERT INTO aturan (nama, negara, perangkat, prioritas, status, tujuan_id)
		VALUES ($1, 'MY', 'mobile', 10, true, $2)
		RETURNING id
	`, namaAturan1, idTujuanA).Scan(&idAturan1)
	if err != nil {
		t.Fatalf("Gagal membuat aturan 1: %v", err)
	}

	// Buat Aturan 2: MY + mobile -> Tujuan B (Prioritas 1 - lebih tinggi dari 10)
	var idAturan2 int
	err = db.QueryRow(`
		INSERT INTO aturan (nama, negara, perangkat, prioritas, status, tujuan_id)
		VALUES ($1, 'MY', 'mobile', 1, true, $2)
		RETURNING id
	`, namaAturan2, idTujuanB).Scan(&idAturan2)
	if err != nil {
		t.Fatalf("Gagal membuat aturan 2: %v", err)
	}

	// Buat Aturan 3: SG + desktop -> Tujuan A (Status FALSE - tidak aktif)
	var idAturan3 int
	err = db.QueryRow(`
		INSERT INTO aturan (nama, negara, perangkat, prioritas, status, tujuan_id)
		VALUES ($1, 'SG', 'desktop', 1, false, $2)
		RETURNING id
	`, namaAturan3, idTujuanA).Scan(&idAturan3)
	if err != nil {
		t.Fatalf("Gagal membuat aturan 3: %v", err)
	}

	// Pastikan ada aturan ID Mobile khusus pengujian B5 (prioritas 5)
	namaAturanID := fmt.Sprintf("Aturan B5 ID Mobile %d", os.Getpid())
	_, _ = db.Exec(`
		INSERT INTO aturan (nama, negara, perangkat, prioritas, status, tujuan_id)
		VALUES ($1, 'ID', 'mobile', 5, true, $2)
		ON CONFLICT DO NOTHING
	`, namaAturanID, idTujuanA)

	t.Cleanup(func() {
		db.Exec("DELETE FROM traffic_logs WHERE aturan_id IN ($1, $2, $3)", idAturan1, idAturan2, idAturan3)
		db.Exec("DELETE FROM traffic_logs WHERE negara IN ('MY', 'SG', 'ZZ', 'US', 'JP', 'ID')")
		db.Exec("DELETE FROM aturan WHERE id IN ($1, $2, $3)", idAturan1, idAturan2, idAturan3)
		db.Exec("DELETE FROM tujuan WHERE id IN ($1, $2)", idTujuanA, idTujuanB)
	})

	// 1. POST /api/periksa dengan rule yang cocok
	t.Run("1_RuleCocok", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"negara":    "MY",
			"perangkat": "mobile",
			"peramban":  "Chrome",
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/periksa", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Diharapkan 200, didapat %d: %s", w.Code, w.Body.String())
		}

		var respon struct {
			Data struct {
				Cocok     bool   `json:"cocok"`
				AturanID  *int   `json:"aturan_id"`
				TujuanID  *int   `json:"tujuan_id"`
				URLTujuan string `json:"url_tujuan"`
				Hasil     string `json:"hasil"`
			} `json:"data"`
		}
		json.Unmarshal(w.Body.Bytes(), &respon)

		if !respon.Data.Cocok || respon.Data.Hasil != "aturan" || respon.Data.AturanID == nil {
			t.Errorf("Harusnya cocok dengan aturan, didapat: %+v", respon.Data)
		}
	})

	// 2. POST /api/periksa tanpa rule → fallback
	t.Run("2_TanpaRuleFallback", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"negara":    "ZZ",
			"perangkat": "desktop",
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/periksa", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Diharapkan 200, didapat %d", w.Code)
		}

		var respon struct {
			Data struct {
				Cocok     bool   `json:"cocok"`
				AturanID  *int   `json:"aturan_id"`
				URLTujuan string `json:"url_tujuan"`
				Hasil     string `json:"hasil"`
			} `json:"data"`
		}
		json.Unmarshal(w.Body.Bytes(), &respon)

		if respon.Data.Cocok || respon.Data.Hasil != "fallback" || respon.Data.AturanID != nil {
			t.Errorf("Harusnya fallback, didapat: %+v", respon.Data)
		}
	})

	// 3. POST /api/periksa dengan rule priority tinggi (Aturan 2 prioritas 1 vs Aturan 1 prioritas 10)
	t.Run("3_RulePriorityTinggiMenang", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"negara":    "MY",
			"perangkat": "mobile",
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/periksa", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		var respon struct {
			Data struct {
				AturanID *int   `json:"aturan_id"`
				TujuanID *int   `json:"tujuan_id"`
				Hasil    string `json:"hasil"`
			} `json:"data"`
		}
		json.Unmarshal(w.Body.Bytes(), &respon)

		if respon.Data.AturanID == nil || *respon.Data.AturanID != idAturan2 {
			var idDapat int
			if respon.Data.AturanID != nil {
				idDapat = *respon.Data.AturanID
			}
			t.Errorf("Aturan 2 (prioritas 1) harus menang atas Aturan 1 (prioritas 10), didapat ID: %d (harapan: %d)", idDapat, idAturan2)
		}
	})

	// 4. Rule inactive tidak digunakan
	t.Run("4_RuleInactiveTidakDigunakan", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"negara":    "SG",
			"perangkat": "desktop",
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/periksa", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		var respon struct {
			Data struct {
				Cocok    bool   `json:"cocok"`
				AturanID *int   `json:"aturan_id"`
				Hasil    string `json:"hasil"`
			} `json:"data"`
		}
		json.Unmarshal(w.Body.Bytes(), &respon)

		// Rule SG tidak aktif, sehingga harus fallback
		if respon.Data.Cocok || respon.Data.Hasil != "fallback" {
			t.Errorf("Rule nonaktif tidak boleh digunakan, harus fallback: %+v", respon.Data)
		}
	})

	// 5. Destination valid
	t.Run("5_DestinationValid", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"negara":    "MY",
			"perangkat": "mobile",
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/periksa", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		var respon struct {
			Data struct {
				URLTujuan string `json:"url_tujuan"`
			} `json:"data"`
		}
		json.Unmarshal(w.Body.Bytes(), &respon)

		if respon.Data.URLTujuan != urlTujuanB {
			t.Errorf("URL tujuan harus sesuai dengan tujuan dari aturan pemenang: %s", respon.Data.URLTujuan)
		}
	})

	// 6. Destination tidak tersedia (tujuan dinonaktifkan / tidak aktif)
	t.Run("6_DestinationTidakTersediaFallback", func(t *testing.T) {
		// Buat tujuan dummy lalu nonaktifkan tujuannya
		var idTujuanNonaktif int
		urlNonaktif := fmt.Sprintf("/demo/tujuan-mati-%d", os.Getpid())
		db.QueryRow("INSERT INTO tujuan (nama, url, status) VALUES ('Tujuan Mati', $1, false) RETURNING id", urlNonaktif).Scan(&idTujuanNonaktif)

		var idAturanMati int
		db.QueryRow("INSERT INTO aturan (nama, negara, prioritas, status, tujuan_id) VALUES ('Aturan Tujuan Mati', 'JP', 1, true, $1) RETURNING id", idTujuanNonaktif).Scan(&idAturanMati)

		defer func() {
			db.Exec("DELETE FROM aturan WHERE id = $1", idAturanMati)
			db.Exec("DELETE FROM tujuan WHERE id = $1", idTujuanNonaktif)
		}()

		body, _ := json.Marshal(map[string]interface{}{
			"negara": "JP",
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/periksa", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		var respon struct {
			Data struct {
				Cocok bool   `json:"cocok"`
				Hasil string `json:"hasil"`
			} `json:"data"`
		}
		json.Unmarshal(w.Body.Bytes(), &respon)

		if respon.Data.Hasil != "fallback" {
			t.Errorf("Tujuan tidak tersedia harus menghasilkan fallback, didapat: %s", respon.Data.Hasil)
		}
	})

	// 7. Input perangkat invalid (harus 400)
	t.Run("7_InputPerangkatInvalid", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"perangkat": "kulkas",
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/periksa", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Diharapkan 400 untuk perangkat tidak valid, didapat %d", w.Code)
		}
	})

	// 8. Input browser invalid (harus 400)
	t.Run("8_InputBrowserInvalid", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"peramban": "BrowserAlien123",
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/periksa", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Diharapkan 400 untuk browser tidak valid, didapat %d", w.Code)
		}
	})

	// 9. Input terlalu panjang (harus 400)
	t.Run("9_InputTerlaluPanjang", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"negara": "INDONESIA_RAYA_MERDEKA", // > 10 karakter
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/periksa", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Diharapkan 400 untuk input negara terlalu panjang, didapat %d", w.Code)
		}
	})

	// 10. GET /api/log-trafik
	t.Run("10_GetLogTrafik", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/api/log-trafik", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Diharapkan 200, didapat %d: %s", w.Code, w.Body.String())
		}

		var respon struct {
			Data       []model.LogTrafikJSON `json:"data"`
			Pagination penangan.InfoPaginasi `json:"pagination"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &respon); err != nil {
			t.Fatalf("Gagal parse response: %v", err)
		}

		if respon.Pagination.Limit != 20 || respon.Pagination.Page != 1 {
			t.Errorf("Metadata paginasi tidak sesuai: %+v", respon.Pagination)
		}
	})

	// 11. Pagination page=1 limit=20
	t.Run("11_PaginasiDefault", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/api/log-trafik?page=1&limit=20", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Diharapkan status 200, didapat %d", w.Code)
		}
	})

	// 12. Pagination limit maksimal 100
	t.Run("12_PaginasiLimitMaksimal100", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/api/log-trafik?limit=500", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Diharapkan status 200, didapat %d", w.Code)
		}

		var respon struct {
			Pagination penangan.InfoPaginasi `json:"pagination"`
		}
		json.Unmarshal(w.Body.Bytes(), &respon)

		if respon.Pagination.Limit != 100 {
			t.Errorf("Limit 500 harus dibatasi maksimal 100, didapat: %d", respon.Pagination.Limit)
		}
	})

	// 13. Page invalid (harus 400)
	t.Run("13_PageInvalid", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/api/log-trafik?page=0", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Diharapkan 400 untuk page=0, didapat %d", w.Code)
		}
	})

	// 14. Log terbaru muncul lebih dahulu (dibuat_pada DESC)
	t.Run("14_LogTerbaruMunculLebihDahulu", func(t *testing.T) {
		// Kirim permintaan baru dengan tanda khusus
		body, _ := json.Marshal(map[string]interface{}{
			"negara":       "ID",
			"asal_rujukan": "https://penanda-terbaru.com",
		})
		wPost := httptest.NewRecorder()
		reqPost, _ := http.NewRequest(http.MethodPost, "/api/periksa", bytes.NewBuffer(body))
		reqPost.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(wPost, reqPost)

		// Ambil log
		wGet := httptest.NewRecorder()
		reqGet, _ := http.NewRequest(http.MethodGet, "/api/log-trafik?limit=5", nil)
		r.ServeHTTP(wGet, reqGet)

		var respon struct {
			Data []model.LogTrafikJSON `json:"data"`
		}
		json.Unmarshal(wGet.Body.Bytes(), &respon)

		if len(respon.Data) == 0 || respon.Data[0].AsalRujukan == nil || *respon.Data[0].AsalRujukan != "https://penanda-terbaru.com" {
			t.Errorf("Log terbaru harus berada pada elemen pertama daftar")
		}
	})

	// 15. GET /api/log-trafik/:id
	var idLogAda int
	t.Run("15_GetSatuLogTrafik", func(t *testing.T) {
		// Ambil 1 log yang ada di database
		err := db.QueryRow("SELECT id FROM traffic_logs ORDER BY id DESC LIMIT 1").Scan(&idLogAda)
		if err != nil {
			t.Fatalf("Gagal mengambil log dari database: %v", err)
		}

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/log-trafik/%d", idLogAda), nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Diharapkan status 200, didapat %d", w.Code)
		}

		var respon struct {
			Data model.LogTrafikJSON `json:"data"`
		}
		json.Unmarshal(w.Body.Bytes(), &respon)

		if respon.Data.ID != idLogAda {
			t.Errorf("ID log tidak sesuai: %d vs %d", respon.Data.ID, idLogAda)
		}
	})

	// 16. ID log tidak ditemukan → 404
	t.Run("16_IDLogTidakDitemukan404", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/api/log-trafik/99999999", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Diharapkan status 404, didapat %d", w.Code)
		}
	})

	// 17. Traffic log tercatat setelah /api/periksa
	t.Run("17_TrafficLogTercatatSetelahPeriksa", func(t *testing.T) {
		var countSebelum int
		db.QueryRow("SELECT COUNT(*) FROM traffic_logs").Scan(&countSebelum)

		body, _ := json.Marshal(map[string]interface{}{
			"negara": "ID",
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/periksa", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		var countSesudah int
		db.QueryRow("SELECT COUNT(*) FROM traffic_logs").Scan(&countSesudah)

		if countSesudah != countSebelum+1 {
			t.Errorf("Jumlah log harus bertambah 1: %d -> %d", countSebelum, countSesudah)
		}
	})

	// 18. Fallback tercatat sebagai hasil "fallback"
	t.Run("18_FallbackTercatatSebagaiHasilFallback", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"negara":       "ZZ",
			"asal_rujukan": "https://uji-fallback.com",
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/periksa", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		var hasil string
		err := db.QueryRow("SELECT hasil FROM traffic_logs WHERE asal_rujukan = 'https://uji-fallback.com' ORDER BY id DESC LIMIT 1").Scan(&hasil)
		if err != nil || hasil != "fallback" {
			t.Errorf("Hasil log harus 'fallback', didapat: %s (err: %v)", hasil, err)
		}
	})

	// 19. Rule match tercatat sebagai hasil "aturan"
	t.Run("19_RuleMatchTercatatSebagaiHasilAturan", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"negara":       "ID",
			"perangkat":    "mobile",
			"asal_rujukan": "https://uji-aturan.com",
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/periksa", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		var hasil string
		err := db.QueryRow("SELECT hasil FROM traffic_logs WHERE asal_rujukan = 'https://uji-aturan.com' ORDER BY id DESC LIMIT 1").Scan(&hasil)
		if err != nil || hasil != "aturan" {
			t.Errorf("Hasil log harus 'aturan', didapat: %s (err: %v)", hasil, err)
		}
	})

	// 20. aturan_id NULL untuk fallback
	t.Run("20_AturanIDNullUntukFallback", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"negara":       "ZZ",
			"asal_rujukan": "https://uji-null-aturan.com",
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/periksa", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		var aturanID sql.NullInt64
		err := db.QueryRow("SELECT aturan_id FROM traffic_logs WHERE asal_rujukan = 'https://uji-null-aturan.com' ORDER BY id DESC LIMIT 1").Scan(&aturanID)
		if err != nil || aturanID.Valid {
			t.Errorf("aturan_id harus NULL pada hasil fallback, didapat valid=%v", aturanID.Valid)
		}
	})

	// 21. SQL injection input tidak menyebabkan query rusak
	t.Run("21_SQLInjectionAman", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"negara":       "ID",
			"asal_rujukan": "'; DROP TABLE traffic_logs; --",
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/periksa", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Harus tetap 200 aman dan di-escape oleh parameter query, didapat: %d", w.Code)
		}

		// Pastikan tabel traffic_logs tidak ter-drop
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM traffic_logs").Scan(&count)
		if err != nil {
			t.Errorf("Tabel traffic_logs harus tetap ada dan aman: %v", err)
		}
	})

	// 22. API tidak mengembalikan credential/token/session
	t.Run("22_TidakMengembalikanKredensial", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/api/log-trafik", nil)
		r.ServeHTTP(w, req)

		bodyTeks := strings.ToLower(w.Body.String())
		kataKunciBahaya := []string{"password", "token", "session", "authorization", "secret"}
		for _, kata := range kataKunciBahaya {
			if strings.Contains(bodyTeks, kata) {
				t.Errorf("Respons API tidak boleh membocorkan informasi sensitif: '%s'", kata)
			}
		}
	})

	// 23. Error database ditangani dengan response 500 yang aman
	t.Run("23_ErrorDatabaseDitanganiAman", func(t *testing.T) {
		// Siapkan db tertutup untuk mensimulasikan error database
		dbTutup, _ := sql.Open("postgres", "postgres://user:pass@localhost:5432/db_palsu?sslmode=disable")
		dbTutup.Close() // Tutup koneksi agar query pasti gagal

		rErr := gin.New()
		rErr.POST("/api/periksa", penangan.PeriksaTrafik(dbTutup, "/demo/cadangan"))

		body, _ := json.Marshal(map[string]interface{}{"negara": "ID"})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/periksa", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rErr.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Diharapkan status 500 saat database error, didapat: %d", w.Code)
		}

		var respon map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &respon)
		if respon["error"] != true || respon["pesan"] == nil {
			t.Errorf("Format error 500 harus aman dan konsisten: %v", respon)
		}
	})
}
