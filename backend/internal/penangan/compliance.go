package penangan

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"sivilize-traffic-routing-demo/backend/internal/kepatuhan"
	"sivilize-traffic-routing-demo/backend/internal/model"
)

// PeriksaCompliance menangani GET /api/compliance/check.
// Melakukan audit kepatuhan internal terhadap konfigurasi sistem, menyimpan hasil audit ke database,
// dan mengembalikan daftar temuan tanpa skor kepatuhan numerik.
func PeriksaCompliance(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		hasil, err := kepatuhan.LakukanAuditKepatuhan(db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": true,
				"pesan": "Gagal menjalankan audit kepatuhan internal: " + err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, hasil)
	}
}

// AmbilChecklistCompliance menangani GET /api/compliance/checklist.
// Mengembalikan daftar seluruh item checklist kepatuhan manual.
func AmbilChecklistCompliance(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		items, err := kepatuhan.AmbilChecklist(db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": true,
				"pesan": "Gagal mengambil data checklist kepatuhan: " + err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": items,
		})
	}
}

// PermintaanUbahChecklist mendefinisikan payload untuk PUT /api/compliance/checklist.
type PermintaanUbahChecklist struct {
	Kunci   string                         `json:"kunci,omitempty"`
	Selesai *bool                          `json:"selesai,omitempty"`
	Items   []model.ItemChecklistCompliance `json:"items,omitempty"`
}

// UbahChecklistCompliance menangani PUT /api/compliance/checklist.
// Memperbarui status checklist manual oleh administrator.
func UbahChecklistCompliance(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req PermintaanUbahChecklist
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": true,
				"pesan": "Format data checklist tidak valid",
			})
			return
		}

		// Jika update berupa batch list
		if len(req.Items) > 0 {
			for _, it := range req.Items {
				kunciClean := strings.TrimSpace(it.Kunci)
				if kunciClean != "" {
					_ = kepatuhan.PerbaruiItemChecklist(db, kunciClean, it.Selesai)
				}
			}
		} else if strings.TrimSpace(req.Kunci) != "" && req.Selesai != nil {
			// Update single item
			err := kepatuhan.PerbaruiItemChecklist(db, strings.TrimSpace(req.Kunci), *req.Selesai)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": true,
					"pesan": err.Error(),
				})
				return
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": true,
				"pesan": "Parameter 'kunci' & 'selesai' atau daftar 'items' wajib disediakan",
			})
			return
		}

		items, err := kepatuhan.AmbilChecklist(db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": true,
				"pesan": "Gagal mengambil data checklist terbaru: " + err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"pesan": "Checklist kepatuhan berhasil diperbarui",
			"data":  items,
		})
	}
}

// RiwayatAuditCompliance menangani GET /api/compliance/audits.
// Mengembalikan riwayat audit kepatuhan sebelumnya.
func RiwayatAuditCompliance(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		riwayat, err := kepatuhan.AmbilRiwayatAudit(db, 20)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": true,
				"pesan": "Gagal mengambil riwayat audit: " + err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": riwayat,
		})
	}
}
