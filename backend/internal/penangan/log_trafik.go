package penangan

import (
	"database/sql"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"sivilize-traffic-routing-demo/backend/internal/model"
)

// InfoPaginasi menyimpan metadata paginasi untuk daftar log trafik.
type InfoPaginasi struct {
	Page         int `json:"page"`
	Limit        int `json:"limit"`
	Total        int `json:"total"`
	TotalHalaman int `json:"total_halaman"`
}

// DaftarLogTrafik menangani GET /api/log-trafik.
// Mendukung paginasi dengan query parameter ?page=1&limit=20 (maksimal limit 100).
// Mengurutkan dari yang terbaru (dibuat_pada DESC).
func DaftarLogTrafik(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		pageStr := c.DefaultQuery("page", "1")
		limitStr := c.DefaultQuery("limit", "20")

		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": true,
				"pesan": "Nomor halaman tidak valid (minimal 1)",
			})
			return
		}

		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": true,
				"pesan": "Limit tidak valid (minimal 1)",
			})
			return
		}

		// Batasi limit maksimal menjadi 100 sesuai spesifikasi
		if limit > 100 {
			limit = 100
		}

		// Hitung total data log di tabel traffic_logs
		var total int
		err = db.QueryRow("SELECT COUNT(*) FROM traffic_logs").Scan(&total)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": true,
				"pesan": "Gagal menghitung total data log trafik",
			})
			return
		}

		totalHalaman := 1
		if total > 0 {
			totalHalaman = int(math.Ceil(float64(total) / float64(limit)))
		}

		offset := (page - 1) * limit

		kueriData := `
			SELECT id, negara, perangkat, peramban, agen_pengguna, asal_rujukan,
			       aturan_id, tujuan_id, url_tujuan, hasil, dibuat_pada
			FROM traffic_logs
			ORDER BY dibuat_pada DESC, id DESC
			LIMIT $1 OFFSET $2
		`

		baris, err := db.Query(kueriData, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": true,
				"pesan": "Gagal mengambil daftar log trafik dari database",
			})
			return
		}
		defer baris.Close()

		daftarLog := make([]model.LogTrafikJSON, 0)
		for baris.Next() {
			var l model.LogTrafik
			err := baris.Scan(
				&l.ID, &l.Negara, &l.Perangkat, &l.Peramban, &l.AgenPengguna, &l.AsalRujukan,
				&l.AturanID, &l.TujuanID, &l.URLTujuan, &l.Hasil, &l.DibuatPada,
			)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": true,
					"pesan": "Gagal membaca baris log trafik",
				})
				return
			}
			daftarLog = append(daftarLog, l.KeJSON())
		}

		if err := baris.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": true,
				"pesan": "Terjadi kesalahan saat memproses log trafik",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": daftarLog,
			"pagination": InfoPaginasi{
				Page:         page,
				Limit:        limit,
				Total:        total,
				TotalHalaman: totalHalaman,
			},
		})
	}
}

// SatuLogTrafik menangani GET /api/log-trafik/:id.
// Mengambil detail satu entri log trafik berdasarkan ID.
func SatuLogTrafik(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": true,
				"pesan": "ID log tidak valid",
			})
			return
		}

		kueri := `
			SELECT id, negara, perangkat, peramban, agen_pengguna, asal_rujukan,
			       aturan_id, tujuan_id, url_tujuan, hasil, dibuat_pada
			FROM traffic_logs
			WHERE id = $1
		`

		var l model.LogTrafik
		err = db.QueryRow(kueri, id).Scan(
			&l.ID, &l.Negara, &l.Perangkat, &l.Peramban, &l.AgenPengguna, &l.AsalRujukan,
			&l.AturanID, &l.TujuanID, &l.URLTujuan, &l.Hasil, &l.DibuatPada,
		)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error": true,
				"pesan": "Log trafik tidak ditemukan",
			})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": true,
				"pesan": "Gagal mengambil data log trafik",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": l.KeJSON(),
		})
	}
}
