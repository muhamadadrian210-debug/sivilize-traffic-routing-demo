package penangan

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"sivilize-traffic-routing-demo/backend/internal/model"
)

// PermintaanSimpanTujuan mendefinisikan payload untuk membuat atau mengubah data tujuan.
type PermintaanSimpanTujuan struct {
	Nama   string `json:"nama"`
	URL    string `json:"url"`
	Status *bool  `json:"status"`
}

// DaftarTujuan menangani GET /api/tujuan.
// Mengambil semua data tujuan dari PostgreSQL dengan urutan id ASC.
func DaftarTujuan(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		kueri := `
			SELECT id, nama, url, status, dibuat_pada, diperbarui_pada
			FROM tujuan
			ORDER BY id ASC
		`

		baris, err := db.Query(kueri)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": true,
				"pesan": "Gagal mengambil daftar tujuan dari database",
			})
			return
		}
		defer baris.Close()

		daftar := make([]model.Tujuan, 0)
		for baris.Next() {
			var t model.Tujuan
			if err := baris.Scan(&t.ID, &t.Nama, &t.URL, &t.Status, &t.DibuatPada, &t.DiperbaruiPada); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": true,
					"pesan": "Gagal membaca data tujuan",
				})
				return
			}
			daftar = append(daftar, t)
		}

		if err := baris.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": true,
				"pesan": "Terjadi kesalahan saat memproses data tujuan",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": daftar,
		})
	}
}

// SatuTujuan menangani GET /api/tujuan/:id.
// Mengambil satu data tujuan berdasarkan ID.
func SatuTujuan(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idTeks := c.Param("id")
		id, err := strconv.Atoi(idTeks)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": true,
				"pesan": "ID tujuan tidak valid",
			})
			return
		}

		kueri := `
			SELECT id, nama, url, status, dibuat_pada, diperbarui_pada
			FROM tujuan
			WHERE id = $1
		`

		var t model.Tujuan
		err = db.QueryRow(kueri, id).Scan(&t.ID, &t.Nama, &t.URL, &t.Status, &t.DibuatPada, &t.DiperbaruiPada)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error": true,
				"pesan": "Tujuan tidak ditemukan",
			})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": true,
				"pesan": "Gagal mengambil data tujuan",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": t,
		})
	}
}

// BuatTujuan menangani POST /api/tujuan.
// Membuat entri tujuan baru ke PostgreSQL.
func BuatTujuan(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req PermintaanSimpanTujuan
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": true,
				"pesan": "Format data tidak valid",
			})
			return
		}

		namaBersih := strings.TrimSpace(req.Nama)
		if namaBersih == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": true,
				"pesan": "Nama tujuan wajib diisi",
			})
			return
		}

		urlBersih := strings.TrimSpace(req.URL)
		if urlBersih == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": true,
				"pesan": "URL tujuan wajib diisi",
			})
			return
		}

		statusAktif := true
		if req.Status != nil {
			statusAktif = *req.Status
		}

		kueri := `
			INSERT INTO tujuan (nama, url, status, dibuat_pada, diperbarui_pada)
			VALUES ($1, $2, $3, NOW(), NOW())
			RETURNING id, nama, url, status, dibuat_pada, diperbarui_pada
		`

		var t model.Tujuan
		err := db.QueryRow(kueri, namaBersih, urlBersih, statusAktif).Scan(
			&t.ID, &t.Nama, &t.URL, &t.Status, &t.DibuatPada, &t.DiperbaruiPada,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": true,
				"pesan": "Gagal menyimpan tujuan baru",
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"data": t,
		})
	}
}

// UbahTujuan menangani PUT /api/tujuan/:id.
// Mengubah seluruh konfigurasi tujuan yang ada.
func UbahTujuan(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idTeks := c.Param("id")
		id, err := strconv.Atoi(idTeks)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": true,
				"pesan": "ID tujuan tidak valid",
			})
			return
		}

		var req PermintaanSimpanTujuan
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": true,
				"pesan": "Format data tidak valid",
			})
			return
		}

		namaBersih := strings.TrimSpace(req.Nama)
		if namaBersih == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": true,
				"pesan": "Nama tujuan wajib diisi",
			})
			return
		}

		urlBersih := strings.TrimSpace(req.URL)
		if urlBersih == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": true,
				"pesan": "URL tujuan wajib diisi",
			})
			return
		}

		statusAktif := true
		if req.Status != nil {
			statusAktif = *req.Status
		}

		kueri := `
			UPDATE tujuan
			SET nama = $1, url = $2, status = $3, diperbarui_pada = NOW()
			WHERE id = $4
			RETURNING id, nama, url, status, dibuat_pada, diperbarui_pada
		`

		var t model.Tujuan
		err = db.QueryRow(kueri, namaBersih, urlBersih, statusAktif, id).Scan(
			&t.ID, &t.Nama, &t.URL, &t.Status, &t.DibuatPada, &t.DiperbaruiPada,
		)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error": true,
				"pesan": "Tujuan tidak ditemukan",
			})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": true,
				"pesan": "Gagal memperbarui tujuan",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": t,
		})
	}
}

// HapusTujuan menangani DELETE /api/tujuan/:id.
// Memeriksa relasi dengan aturan sebelum menghapus. Jika masih digunakan, mengembalikan HTTP 409.
func HapusTujuan(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idTeks := c.Param("id")
		id, err := strconv.Atoi(idTeks)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": true,
				"pesan": "ID tujuan tidak valid",
			})
			return
		}

		// Periksa apakah tujuan ada di database
		var existsID int
		kueriCek := `SELECT id FROM tujuan WHERE id = $1`
		err = db.QueryRow(kueriCek, id).Scan(&existsID)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error": true,
				"pesan": "Tujuan tidak ditemukan",
			})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": true,
				"pesan": "Gagal memeriksa data tujuan",
			})
			return
		}

		// Periksa apakah masih ada aturan yang menggunakan tujuan ini
		var jumlahPenggunaan int
		kueriRelasi := `SELECT COUNT(*) FROM aturan WHERE tujuan_id = $1`
		err = db.QueryRow(kueriRelasi, id).Scan(&jumlahPenggunaan)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": true,
				"pesan": "Gagal memeriksa keterkaitan aturan",
			})
			return
		}

		if jumlahPenggunaan > 0 {
			c.JSON(http.StatusConflict, gin.H{
				"error": true,
				"pesan": "Tujuan masih digunakan oleh aturan dan tidak dapat dihapus",
			})
			return
		}

		// Hapus tujuan
		kueriHapus := `DELETE FROM tujuan WHERE id = $1`
		_, err = db.Exec(kueriHapus, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": true,
				"pesan": "Gagal menghapus tujuan",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": gin.H{
				"pesan": "Tujuan berhasil dihapus",
			},
		})
	}
}
