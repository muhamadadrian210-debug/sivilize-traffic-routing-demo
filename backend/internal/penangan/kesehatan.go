package penangan

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

// PeriksaKesehatan menangani GET /api/kesehatan.
// Menampilkan status server dan status koneksi database.
func PeriksaKesehatan(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		statusDatabase := "terhubung"

		// Ping database saat pemeriksaan untuk memastikan koneksi masih aktif
		if err := db.Ping(); err != nil {
			statusDatabase = "terputus"
		}

		statusTeks := "sehat"
		if c.Request.URL.Path == "/health" {
			statusTeks = "ok"
		}

		c.JSON(http.StatusOK, gin.H{
			"status":   statusTeks,
			"database": statusDatabase,
		})
	}
}
