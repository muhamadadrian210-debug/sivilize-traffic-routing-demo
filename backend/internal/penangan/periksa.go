package penangan

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"sivilize-traffic-routing-demo/backend/internal/aturan"
)

// PermintaanPeriksa mendefinisikan payload masukan untuk simulasi pemeriksaan trafik.
type PermintaanPeriksa struct {
	Negara       string `json:"negara"`
	Perangkat    string `json:"perangkat"`
	Peramban     string `json:"peramban"`
	AgenPengguna string `json:"agen_pengguna"`
	AsalRujukan  string `json:"asal_rujukan"`
}

// validasiPerangkatPeriksa mengecek apakah jenis perangkat masukan diperbolehkan.
func validasiPerangkatPeriksa(p string) bool {
	switch strings.ToLower(strings.TrimSpace(p)) {
	case "mobile", "tablet", "desktop":
		return true
	default:
		return false
	}
}

// validasiPerambanPeriksa mengecek apakah nama peramban masukan diperbolehkan.
func validasiPerambanPeriksa(b string) bool {
	switch strings.ToLower(strings.TrimSpace(b)) {
	case "chrome", "firefox", "safari", "edge", "lainnya", "tidak diketahui":
		return true
	default:
		return false
	}
}

// PeriksaTrafik menangani POST /api/periksa.
// Menerima data simulasi trafik, menganalisis konteks pengunjung, mencocokkan dengan rule engine,
// mencatat keputusan ke traffic_logs, dan mengembalikan hasil evaluasi dalam format JSON.
func PeriksaTrafik(db *sql.DB, urlFallback string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req PermintaanPeriksa
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": true,
				"pesan": "Format data permintaan tidak valid",
			})
			return
		}

		// Validasi batas panjang input agar aman dari payload berlebihan
		if len(req.Negara) > 10 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": true,
				"pesan": "Kode negara terlalu panjang (maksimal 10 karakter)",
			})
			return
		}
		if len(req.Perangkat) > 50 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": true,
				"pesan": "Nama perangkat terlalu panjang (maksimal 50 karakter)",
			})
			return
		}
		if len(req.Peramban) > 50 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": true,
				"pesan": "Nama peramban terlalu panjang (maksimal 50 karakter)",
			})
			return
		}
		if len(req.AgenPengguna) > 1000 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": true,
				"pesan": "User-Agent terlalu panjang (maksimal 1000 karakter)",
			})
			return
		}
		if len(req.AsalRujukan) > 2048 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": true,
				"pesan": "Asal rujukan terlalu panjang (maksimal 2048 karakter)",
			})
			return
		}

		// Validasi nilai perangkat jika diisi
		if strings.TrimSpace(req.Perangkat) != "" {
			if !validasiPerangkatPeriksa(req.Perangkat) {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": true,
					"pesan": "Perangkat tidak valid (hanya mobile, tablet, atau desktop)",
				})
				return
			}
		}

		// Validasi nilai peramban jika diisi
		if strings.TrimSpace(req.Peramban) != "" {
			if !validasiPerambanPeriksa(req.Peramban) {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": true,
					"pesan": "Peramban tidak valid (hanya Chrome, Firefox, Safari, Edge, Lainnya, atau Tidak Diketahui)",
				})
				return
			}
		}

		// Ekstrak IP pengunjung dari permintaan
		ip := c.ClientIP()

		// Bangun KonteksPengunjung menggunakan Request Analyzer Batch 4
		konteks := aturan.AnalisisPermintaan(req.AgenPengguna, req.AsalRujukan, ip, req.Negara)

		// Jika pengguna menentukan perangkat secara eksplisit pada payload, gunakan nilai tersebut
		if strings.TrimSpace(req.Perangkat) != "" {
			konteks.Perangkat = aturan.NormalisasiPerangkat(req.Perangkat)
		}

		// Jika pengguna menentukan peramban secara eksplisit pada payload, gunakan nilai tersebut
		if strings.TrimSpace(req.Peramban) != "" {
			konteks.Peramban = aturan.NormalisasiPeramban(req.Peramban)
		}

		// Panggil Rule Engine Batch 4 untuk mendapatkan keputusan routing
		hasil, err := aturan.EvaluasiDenganBasisData(db, konteks, urlFallback)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": true,
				"pesan": "Gagal mengevaluasi aturan trafik dari database",
			})
			return
		}

		// Siapkan nilai yang akan disimpan ke tabel traffic_logs
		var negaraVal, perangkatVal, perambanVal, agenPenggunaVal, asalRujukanVal sql.NullString
		if konteks.Negara != "" {
			negaraVal = sql.NullString{String: konteks.Negara, Valid: true}
		}
		if konteks.Perangkat != "" {
			perangkatVal = sql.NullString{String: konteks.Perangkat, Valid: true}
		}
		if konteks.Peramban != "" {
			perambanVal = sql.NullString{String: konteks.Peramban, Valid: true}
		}
		if konteks.AgenPengguna != "" {
			agenPenggunaVal = sql.NullString{String: konteks.AgenPengguna, Valid: true}
		}
		if konteks.AsalRujukan != "" {
			asalRujukanVal = sql.NullString{String: konteks.AsalRujukan, Valid: true}
		}

		var aturanIDVal, tujuanIDVal sql.NullInt64
		if hasil.AturanID != nil {
			aturanIDVal = sql.NullInt64{Int64: int64(*hasil.AturanID), Valid: true}
		}
		if hasil.TujuanID != nil {
			tujuanIDVal = sql.NullInt64{Int64: int64(*hasil.TujuanID), Valid: true}
		}

		kueriSimpanLog := `
			INSERT INTO traffic_logs (
				negara, perangkat, peramban, agen_pengguna, asal_rujukan,
				aturan_id, tujuan_id, url_tujuan, hasil, dibuat_pada
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
		`

		_, err = db.Exec(
			kueriSimpanLog,
			negaraVal,
			perangkatVal,
			perambanVal,
			agenPenggunaVal,
			asalRujukanVal,
			aturanIDVal,
			tujuanIDVal,
			hasil.URLTujuan,
			hasil.Hasil,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": true,
				"pesan": "Gagal menyimpan log evaluasi trafik ke database",
			})
			return
		}

		// Kembalikan hasil routing dalam format JSON yang konsisten
		c.JSON(http.StatusOK, gin.H{
			"data": hasil,
		})
	}
}
