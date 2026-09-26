package pengujian

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	"sivilize-traffic-routing-demo/backend/internal/basisdata"
	"sivilize-traffic-routing-demo/backend/internal/konfigurasi"
	"sivilize-traffic-routing-demo/backend/internal/model"
	"sivilize-traffic-routing-demo/backend/internal/penangan"
)

// persiapkanRouter menyiapkan mesin Gin untuk pengujian dengan database aktif.
func persiapkanRouter(t *testing.T) (*gin.Engine, *sql.DB) {
	gin.SetMode(gin.TestMode)

	urlDB := os.Getenv("DATABASE_URL")
	if urlDB == "" {
		konf := konfigurasi.MuatKonfigurasi()
		urlDB = konf.UrlDatabase
	}

	db, err := sql.Open("postgres", urlDB)
	if err != nil {
		t.Fatalf("Gagal membuka koneksi database untuk pengujian: %v", err)
	}

	if err := db.Ping(); err != nil {
		t.Fatalf("Database tidak dapat dihubungi: %v", err)
	}

	// Pastikan migrasi tabel sudah dijalankan
	if err := basisdata.JalankanMigrasi(db); err != nil {
		t.Fatalf("Gagal menjalankan migrasi untuk pengujian: %v", err)
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
	}

	return r, db
}

// UjiKesehatan memastikan endpoint /api/kesehatan tetap berfungsi normal.
func TestEndpointKesehatan(t *testing.T) {
	r, db := persiapkanRouter(t)
	defer db.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/kesehatan", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Diharapkan status 200, didapat %d", w.Code)
	}

	var hasil map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &hasil); err != nil {
		t.Fatalf("Gagal mem-parse respon JSON: %v", err)
	}

	if hasil["status"] != "sehat" || hasil["database"] != "terhubung" {
		t.Errorf("Respon kesehatan tidak sesuai: %v", hasil)
	}
}

// UjiCRUDTujuan menguji skenario CRUD pada endpoint /api/tujuan.
func TestCRUDTujuan(t *testing.T) {
	r, db := persiapkanRouter(t)
	defer db.Close()

	var idTujuanDibuat int

	// Bersihkan data tes jika tertinggal
	t.Cleanup(func() {
		if idTujuanDibuat > 0 {
			db.Exec("DELETE FROM aturan WHERE tujuan_id = $1", idTujuanDibuat)
			db.Exec("DELETE FROM tujuan WHERE id = $1", idTujuanDibuat)
		}
	})

	// 1. Membuat tujuan tanpa nama (harus gagal 400)
	t.Run("BuatTujuanTanpaNama", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"nama": "",
			"url":  "/demo/tanpa-nama",
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/tujuan", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Diharapkan status 400, didapat %d", w.Code)
		}
	})

	// 2. Membuat tujuan tanpa url (harus gagal 400)
	t.Run("BuatTujuanTanpaURL", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"nama": "Halaman Tes",
			"url":  "   ",
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/tujuan", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Diharapkan status 400, didapat %d", w.Code)
		}
	})

	// 3. Membuat tujuan valid (harus sukses 201)
	t.Run("BuatTujuanValid", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"nama":   "Halaman Tes Otomatis",
			"url":    "/demo/halaman-tes-otomatis",
			"status": true,
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/tujuan", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("Diharapkan status 201, didapat %d: %s", w.Code, w.Body.String())
		}

		var respon struct {
			Data model.Tujuan `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &respon); err != nil {
			t.Fatalf("Gagal parse respon: %v", err)
		}

		if respon.Data.ID == 0 || respon.Data.Nama != "Halaman Tes Otomatis" {
			t.Errorf("Data tujuan dibuat tidak sesuai: %+v", respon.Data)
		}
		idTujuanDibuat = respon.Data.ID
	})

	// 4. Membaca daftar tujuan (harus sukses 200)
	t.Run("DaftarTujuan", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/api/tujuan", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Diharapkan status 200, didapat %d", w.Code)
		}

		var respon struct {
			Data []model.Tujuan `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &respon); err != nil {
			t.Fatalf("Gagal parse respon: %v", err)
		}

		if len(respon.Data) == 0 {
			t.Errorf("Daftar tujuan tidak boleh kosong")
		}
	})

	// 5. Membaca satu tujuan berdasarkan ID (harus sukses 200)
	t.Run("SatuTujuanValid", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/tujuan/%d", idTujuanDibuat), nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Diharapkan status 200, didapat %d", w.Code)
		}
	})

	// 6. Membaca tujuan dengan ID yang tidak ditemukan (harus 404)
	t.Run("TujuanTidakDitemukan", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/api/tujuan/99999999", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Diharapkan status 404, didapat %d", w.Code)
		}
	})

	// 6b. Mengubah tujuan tidak ditemukan (harus 404)
	t.Run("UbahTujuanTidakDitemukan", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"nama": "Halaman Tidak Ada",
			"url":  "/demo/tidak-ada",
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPut, "/api/tujuan/99999999", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Diharapkan status 404, didapat %d", w.Code)
		}
	})

	// 6c. Menghapus tujuan tidak ditemukan (harus 404)
	t.Run("HapusTujuanTidakDitemukan", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodDelete, "/api/tujuan/99999999", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Diharapkan status 404, didapat %d", w.Code)
		}
	})

	// 7. Mengubah tujuan (harus sukses 200)
	t.Run("UbahTujuan", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"nama":   "Halaman Tes Otomatis Diperbarui",
			"url":    "/demo/halaman-tes-otomatis-v2",
			"status": false,
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/api/tujuan/%d", idTujuanDibuat), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Diharapkan status 200, didapat %d: %s", w.Code, w.Body.String())
		}

		var respon struct {
			Data model.Tujuan `json:"data"`
		}
		json.Unmarshal(w.Body.Bytes(), &respon)
		if respon.Data.Nama != "Halaman Tes Otomatis Diperbarui" || respon.Data.Status != false {
			t.Errorf("Data tujuan yang diubah tidak sesuai: %+v", respon.Data)
		}
	})

	// 8. Gagal menghapus tujuan yang masih digunakan oleh aturan (harus 409)
	t.Run("GagalHapusTujuanTerkaitAturan", func(t *testing.T) {
		// Buat aturan sementara yang mengarah ke idTujuanDibuat
		var idAturanTemp int
		err := db.QueryRow(`
			INSERT INTO aturan (nama, tujuan_id, prioritas, status)
			VALUES ('Aturan Pengunci Tes', $1, 10, true)
			RETURNING id
		`, idTujuanDibuat).Scan(&idAturanTemp)
		if err != nil {
			t.Fatalf("Gagal membuat aturan pengunci: %v", err)
		}

		// Coba hapus tujuan (harus 409 Conflict)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/api/tujuan/%d", idTujuanDibuat), nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusConflict {
			t.Errorf("Diharapkan status 409 saat tujuan masih digunakan, didapat %d", w.Code)
		}

		// Hapus aturan pengunci
		db.Exec("DELETE FROM aturan WHERE id = $1", idAturanTemp)
	})

	// 9. Menghapus tujuan yang tidak terpakai (harus sukses 200)
	t.Run("HapusTujuanSukses", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/api/tujuan/%d", idTujuanDibuat), nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Diharapkan status 200, didapat %d: %s", w.Code, w.Body.String())
		}

		// Pastikan sudah benar-benar hilang (404)
		wCek := httptest.NewRecorder()
		reqCek, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/tujuan/%d", idTujuanDibuat), nil)
		r.ServeHTTP(wCek, reqCek)
		if wCek.Code != http.StatusNotFound {
			t.Errorf("Tujuan harusnya sudah terhapus (404), tapi didapat %d", wCek.Code)
		}

		idTujuanDibuat = 0 // Tandai sudah terhapus
	})
}

// UjiCRUDAturan menguji skenario CRUD pada endpoint /api/aturan.
func TestCRUDAturan(t *testing.T) {
	r, db := persiapkanRouter(t)
	defer db.Close()

	// Siapkan satu tujuan dummy untuk aturan dengan URL unik
	var idTujuanInduk int
	urlTujuan := fmt.Sprintf("/demo/induk-tes-%d", os.Getpid())
	// Bersihkan jika ada yang tersisa dari run sebelumnya
	db.Exec("DELETE FROM tujuan WHERE url LIKE '/demo/induk-tes-%'")

	err := db.QueryRow(`
		INSERT INTO tujuan (nama, url, status)
		VALUES ('Tujuan Induk Aturan Tes', $1, true)
		RETURNING id
	`, urlTujuan).Scan(&idTujuanInduk)
	if err != nil {
		t.Fatalf("Gagal menyiapkan tujuan induk: %v", err)
	}

	var idAturanDibuat int

	t.Cleanup(func() {
		if idAturanDibuat > 0 {
			db.Exec("DELETE FROM aturan WHERE id = $1", idAturanDibuat)
		}
		if idTujuanInduk > 0 {
			db.Exec("DELETE FROM tujuan WHERE id = $1", idTujuanInduk)
		}
	})

	// 1. Membuat aturan tanpa nama (harus 400)
	t.Run("BuatAturanTanpaNama", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"nama":      "",
			"tujuan_id": idTujuanInduk,
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/aturan", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Diharapkan status 400, didapat %d", w.Code)
		}
	})

	// 2. Membuat aturan dengan tujuan yang tidak ditemukan (harus 400)
	t.Run("BuatAturanTujuanTidakAda", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"nama":      "Aturan Tujuan Fiktif",
			"tujuan_id": 99999999,
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/aturan", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Diharapkan status 400, didapat %d", w.Code)
		}
	})

	// 3. Membuat aturan dengan perangkat tidak valid (misal "kulkas") (harus 400)
	t.Run("BuatAturanPerangkatTidakValid", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"nama":      "Aturan Kulkas",
			"tujuan_id": idTujuanInduk,
			"perangkat": "kulkas",
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/aturan", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Diharapkan status 400, didapat %d", w.Code)
		}
	})

	// 4. Membuat aturan dengan peramban tidak valid (misal "Brave") (harus 400)
	t.Run("BuatAturanPerambanTidakValid", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"nama":      "Aturan Peramban Fiktif",
			"tujuan_id": idTujuanInduk,
			"peramban":  "OperaMini",
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/aturan", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Diharapkan status 400, didapat %d", w.Code)
		}
	})

	// 5. Membuat aturan dengan prioritas tidak valid (< 1) (harus 400)
	t.Run("BuatAturanPrioritasTidakValid", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"nama":      "Aturan Prioritas Nol",
			"tujuan_id": idTujuanInduk,
			"prioritas": 0,
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/aturan", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Diharapkan status 400, didapat %d", w.Code)
		}
	})

	// 5b. Membuat aturan dengan negara string kosong (harus 400)
	t.Run("BuatAturanNegaraKosong", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"nama":      "Aturan Negara Kosong",
			"tujuan_id": idTujuanInduk,
			"negara":    "   ",
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/aturan", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Diharapkan status 400, didapat %d", w.Code)
		}
	})

	// 6. Membuat aturan valid (harus 201)
	t.Run("BuatAturanValid", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"nama":          "Aturan Indonesia Chrome Mobile",
			"negara":        "ID",
			"perangkat":     "mobile",
			"peramban":      "Chrome",
			"agen_pengguna": "Mozilla/5.0 Android",
			"asal_rujukan":  "https://google.com",
			"tujuan_id":     idTujuanInduk,
			"prioritas":     5,
			"status":        true,
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/aturan", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("Diharapkan status 201, didapat %d: %s", w.Code, w.Body.String())
		}

		var respon struct {
			Data model.AturanJSON `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &respon); err != nil {
			t.Fatalf("Gagal parse respon: %v", err)
		}

		if respon.Data.ID == 0 || respon.Data.Nama != "Aturan Indonesia Chrome Mobile" {
			t.Errorf("Data aturan dibuat tidak sesuai: %+v", respon.Data)
		}
		if respon.Data.TujuanNama != "Tujuan Induk Aturan Tes" {
			t.Errorf("JOIN nama tujuan tidak sesuai: %s", respon.Data.TujuanNama)
		}

		idAturanDibuat = respon.Data.ID
	})

	// 7. Membaca daftar aturan (harus 200 dan menyertakan data tujuan)
	t.Run("DaftarAturan", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/api/aturan", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Diharapkan status 200, didapat %d", w.Code)
		}

		var respon struct {
			Data []model.AturanJSON `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &respon); err != nil {
			t.Fatalf("Gagal parse respon: %v", err)
		}

		if len(respon.Data) == 0 {
			t.Errorf("Daftar aturan tidak boleh kosong")
		}
	})

	// 8. Membaca aturan berdasarkan ID (harus 200)
	t.Run("SatuAturanValid", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/aturan/%d", idAturanDibuat), nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Diharapkan status 200, didapat %d", w.Code)
		}
	})

	// 9. ID aturan tidak ditemukan (harus 404)
	t.Run("AturanTidakDitemukan", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/api/aturan/99999999", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Diharapkan status 404, didapat %d", w.Code)
		}
	})

	// 10. Mengubah aturan (harus 200)
	t.Run("UbahAturan", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"nama":          "Aturan Indonesia Desktop Edge",
			"negara":        "ID",
			"perangkat":     "desktop",
			"peramban":      "Edge",
			"agen_pengguna": nil,
			"asal_rujukan":  nil,
			"tujuan_id":     idTujuanInduk,
			"prioritas":     2,
			"status":        true,
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/api/aturan/%d", idAturanDibuat), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Diharapkan status 200, didapat %d: %s", w.Code, w.Body.String())
		}

		var respon struct {
			Data model.AturanJSON `json:"data"`
		}
		json.Unmarshal(w.Body.Bytes(), &respon)
		if respon.Data.Nama != "Aturan Indonesia Desktop Edge" || *respon.Data.Perangkat != "desktop" {
			t.Errorf("Data aturan yang diubah tidak sesuai: %+v", respon.Data)
		}
	})

	// 11. Mengubah status aturan (PATCH /api/aturan/:id/status)
	t.Run("UbahStatusAturan", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"status": false,
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("/api/aturan/%d/status", idAturanDibuat), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Diharapkan status 200, didapat %d: %s", w.Code, w.Body.String())
		}

		var respon struct {
			Data model.AturanJSON `json:"data"`
		}
		json.Unmarshal(w.Body.Bytes(), &respon)
		if respon.Data.Status != false {
			t.Errorf("Status aturan gagal diubah menjadi false")
		}
	})

	// 12. Menghapus aturan (harus 200)
	t.Run("HapusAturan", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/api/aturan/%d", idAturanDibuat), nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Diharapkan status 200, didapat %d: %s", w.Code, w.Body.String())
		}

		// Pastikan sudah hilang (404)
		wCek := httptest.NewRecorder()
		reqCek, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/aturan/%d", idAturanDibuat), nil)
		r.ServeHTTP(wCek, reqCek)
		if wCek.Code != http.StatusNotFound {
			t.Errorf("Aturan harusnya sudah terhapus (404), tapi didapat %d", wCek.Code)
		}

		idAturanDibuat = 0 // Tandai sudah terhapus
	})
}
