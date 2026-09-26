package penangan

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"sivilize-traffic-routing-demo/backend/internal/kepatuhan"
	"sivilize-traffic-routing-demo/backend/internal/model"
)

// PermintaanSimpanAturan mendefinisikan payload untuk membuat atau mengubah data aturan.
type PermintaanSimpanAturan struct {
	Nama         string  `json:"nama"`
	Negara       *string `json:"negara"`
	Perangkat    *string `json:"perangkat"`
	Peramban     *string `json:"peramban"`
	AgenPengguna *string `json:"agen_pengguna"`
	AsalRujukan  *string `json:"asal_rujukan"`
	TujuanID     int     `json:"tujuan_id"`
	Prioritas    *int    `json:"prioritas"`
	Status       *bool   `json:"status"`
}

// PermintaanStatusAturan mendefinisikan payload untuk mengubah status aktif/nonaktif aturan.
type PermintaanStatusAturan struct {
	Status *bool `json:"status"`
}

// validasiPerangkat mengecek apakah nilai perangkat diperbolehkan.
func validasiPerangkat(p string) bool {
	switch p {
	case "mobile", "tablet", "desktop":
		return true
	default:
		return false
	}
}

// validasiPeramban mengecek apakah nilai peramban diperbolehkan.
func validasiPeramban(p string) bool {
	switch p {
	case "Chrome", "Firefox", "Safari", "Edge", "Lainnya":
		return true
	default:
		return false
	}
}

// ambilAturanLengkap mengambil data satu aturan beserta informasi tujuan dengan JOIN.
func ambilAturanLengkap(db *sql.DB, id int) (*model.AturanJSON, error) {
	kueri := `
		SELECT 
			a.id, a.nama, a.negara, a.perangkat, a.peramban, a.agen_pengguna, a.asal_rujukan,
			a.tujuan_id, t.nama AS tujuan_nama, t.url AS tujuan_url,
			a.prioritas, a.status, a.dibuat_pada, a.diperbarui_pada
		FROM aturan a
		JOIN tujuan t ON a.tujuan_id = t.id
		WHERE a.id = $1
	`

	var a model.Aturan
	err := db.QueryRow(kueri, id).Scan(
		&a.ID, &a.Nama, &a.Negara, &a.Perangkat, &a.Peramban, &a.AgenPengguna, &a.AsalRujukan,
		&a.TujuanID, &a.TujuanNama, &a.TujuanURL,
		&a.Prioritas, &a.Status, &a.DibuatPada, &a.DiperbaruiPada,
	)
	if err != nil {
		return nil, err
	}

	jsonHasil := a.KeJSON()
	return &jsonHasil, nil
}

// DaftarAturan menangani GET /api/aturan.
// Mengembalikan seluruh aturan beserta data tujuan dengan JOIN PostgreSQL.
func DaftarAturan(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		kueri := `
			SELECT 
				a.id, a.nama, a.negara, a.perangkat, a.peramban, a.agen_pengguna, a.asal_rujukan,
				a.tujuan_id, t.nama AS tujuan_nama, t.url AS tujuan_url,
				a.prioritas, a.status, a.dibuat_pada, a.diperbarui_pada
			FROM aturan a
			JOIN tujuan t ON a.tujuan_id = t.id
			ORDER BY a.prioritas ASC, a.id ASC
		`

		baris, err := db.Query(kueri)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": true,
				"pesan": "Gagal mengambil daftar aturan dari database",
			})
			return
		}
		defer baris.Close()

		daftar := make([]model.AturanJSON, 0)
		for baris.Next() {
			var a model.Aturan
			err := baris.Scan(
				&a.ID, &a.Nama, &a.Negara, &a.Perangkat, &a.Peramban, &a.AgenPengguna, &a.AsalRujukan,
				&a.TujuanID, &a.TujuanNama, &a.TujuanURL,
				&a.Prioritas, &a.Status, &a.DibuatPada, &a.DiperbaruiPada,
			)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": true,
					"pesan": "Gagal membaca data aturan",
				})
				return
			}
			daftar = append(daftar, a.KeJSON())
		}

		if err := baris.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": true,
				"pesan": "Terjadi kesalahan saat memproses data aturan",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": daftar,
		})
	}
}

// SatuAturan menangani GET /api/aturan/:id.
// Mengembalikan satu aturan lengkap beserta informasi tujuan.
func SatuAturan(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idTeks := c.Param("id")
		id, err := strconv.Atoi(idTeks)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": true,
				"pesan": "ID aturan tidak valid",
			})
			return
		}

		aturan, err := ambilAturanLengkap(db, id)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error": true,
				"pesan": "Aturan tidak ditemukan",
			})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": true,
				"pesan": "Gagal mengambil data aturan",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": aturan,
		})
	}
}

// BuatAturan menangani POST /api/aturan.
// Membuat aturan baru dan menyimpannya di PostgreSQL.
func BuatAturan(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req PermintaanSimpanAturan
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
				"pesan": "Nama aturan wajib diisi",
			})
			return
		}

		if req.TujuanID <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": true,
				"pesan": "Tujuan wajib dipilih",
			})
			return
		}

		// Validasi keberadaan tujuan_id di database
		var cekTujuanID int
		err := db.QueryRow("SELECT id FROM tujuan WHERE id = $1", req.TujuanID).Scan(&cekTujuanID)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": true,
				"pesan": "Tujuan yang dipilih tidak ditemukan",
			})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": true,
				"pesan": "Gagal memverifikasi tujuan",
			})
			return
		}

		prioritas := 1
		if req.Prioritas != nil {
			if *req.Prioritas < 1 {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": true,
					"pesan": "Prioritas harus lebih besar dari atau sama dengan 1",
				})
				return
			}
			prioritas = *req.Prioritas
		}

		statusAktif := true
		if req.Status != nil {
			statusAktif = *req.Status
		}

		// Validasi negara jika diisi
		var negaraVal sql.NullString
		if req.Negara != nil {
			negaraBersih := strings.TrimSpace(*req.Negara)
			if negaraBersih == "" {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": true,
					"pesan": "Negara jika diisi tidak boleh berupa string kosong",
				})
				return
			}
			negaraVal = sql.NullString{String: negaraBersih, Valid: true}
		}

		// Validasi perangkat jika diisi
		var perangkatVal sql.NullString
		if req.Perangkat != nil {
			perangkatBersih := strings.TrimSpace(*req.Perangkat)
			if !validasiPerangkat(perangkatBersih) {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": true,
					"pesan": "Perangkat tidak valid (hanya mobile, tablet, atau desktop)",
				})
				return
			}
			perangkatVal = sql.NullString{String: perangkatBersih, Valid: true}
		}

		// Validasi peramban jika diisi
		var perambanVal sql.NullString
		if req.Peramban != nil {
			perambanBersih := strings.TrimSpace(*req.Peramban)
			if !validasiPeramban(perambanBersih) {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": true,
					"pesan": "Peramban tidak valid (hanya Chrome, Firefox, Safari, Edge, atau Lainnya)",
				})
				return
			}
			perambanVal = sql.NullString{String: perambanBersih, Valid: true}
		}

		// Agen pengguna
		var agenPenggunaVal sql.NullString
		if req.AgenPengguna != nil {
			agenBersih := strings.TrimSpace(*req.AgenPengguna)
			if agenBersih != "" {
				agenPenggunaVal = sql.NullString{String: agenBersih, Valid: true}
			}
		}

		// Asal rujukan
		var asalRujukanVal sql.NullString
		if req.AsalRujukan != nil {
			rujukanBersih := strings.TrimSpace(*req.AsalRujukan)
			if rujukanBersih != "" {
				asalRujukanVal = sql.NullString{String: rujukanBersih, Valid: true}
			}
		}

		kueri := `
			INSERT INTO aturan (nama, negara, perangkat, peramban, agen_pengguna, asal_rujukan, tujuan_id, prioritas, status, dibuat_pada, diperbarui_pada)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
			RETURNING id
		`

		var idBaru int
		err = db.QueryRow(
			kueri,
			namaBersih,
			negaraVal,
			perangkatVal,
			perambanVal,
			agenPenggunaVal,
			asalRujukanVal,
			req.TujuanID,
			prioritas,
			statusAktif,
		).Scan(&idBaru)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": true,
				"pesan": "Gagal menyimpan aturan baru ke database",
			})
			return
		}

		hasil, err := ambilAturanLengkap(db, idBaru)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": true,
				"pesan": "Aturan berhasil dibuat namun gagal memuat data lengkap",
			})
			return
		}

		resp := gin.H{
			"data": hasil,
		}
		if kepatuhan.MenargetkanCrawlerReviewer(namaBersih, agenPenggunaVal.String, asalRujukanVal.String) {
			resp["peringatan"] = "Aturan ini perlu ditinjau karena membedakan pengunjung berdasarkan identitas crawler/reviewer dapat digunakan untuk menghindari pemeriksaan atau kebijakan platform."
		}

		c.JSON(http.StatusCreated, resp)
	}
}

// UbahAturan menangani PUT /api/aturan/:id.
// Mengubah konfigurasi aturan secara menyeluruh.
func UbahAturan(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idTeks := c.Param("id")
		id, err := strconv.Atoi(idTeks)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": true,
				"pesan": "ID aturan tidak valid",
			})
			return
		}

		// Periksa keberadaan aturan
		var cekAturanID int
		err = db.QueryRow("SELECT id FROM aturan WHERE id = $1", id).Scan(&cekAturanID)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error": true,
				"pesan": "Aturan tidak ditemukan",
			})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": true,
				"pesan": "Gagal memeriksa aturan",
			})
			return
		}

		var req PermintaanSimpanAturan
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
				"pesan": "Nama aturan wajib diisi",
			})
			return
		}

		if req.TujuanID <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": true,
				"pesan": "Tujuan wajib dipilih",
			})
			return
		}

		// Validasi keberadaan tujuan_id di database
		var cekTujuanID int
		err = db.QueryRow("SELECT id FROM tujuan WHERE id = $1", req.TujuanID).Scan(&cekTujuanID)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": true,
				"pesan": "Tujuan yang dipilih tidak ditemukan",
			})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": true,
				"pesan": "Gagal memverifikasi tujuan",
			})
			return
		}

		prioritas := 1
		if req.Prioritas != nil {
			if *req.Prioritas < 1 {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": true,
					"pesan": "Prioritas harus lebih besar dari atau sama dengan 1",
				})
				return
			}
			prioritas = *req.Prioritas
		}

		statusAktif := true
		if req.Status != nil {
			statusAktif = *req.Status
		}

		// Validasi negara jika diisi
		var negaraVal sql.NullString
		if req.Negara != nil {
			negaraBersih := strings.TrimSpace(*req.Negara)
			if negaraBersih == "" {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": true,
					"pesan": "Negara jika diisi tidak boleh berupa string kosong",
				})
				return
			}
			negaraVal = sql.NullString{String: negaraBersih, Valid: true}
		}

		// Validasi perangkat jika diisi
		var perangkatVal sql.NullString
		if req.Perangkat != nil {
			perangkatBersih := strings.TrimSpace(*req.Perangkat)
			if !validasiPerangkat(perangkatBersih) {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": true,
					"pesan": "Perangkat tidak valid (hanya mobile, tablet, atau desktop)",
				})
				return
			}
			perangkatVal = sql.NullString{String: perangkatBersih, Valid: true}
		}

		// Validasi peramban jika diisi
		var perambanVal sql.NullString
		if req.Peramban != nil {
			perambanBersih := strings.TrimSpace(*req.Peramban)
			if !validasiPeramban(perambanBersih) {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": true,
					"pesan": "Peramban tidak valid (hanya Chrome, Firefox, Safari, Edge, atau Lainnya)",
				})
				return
			}
			perambanVal = sql.NullString{String: perambanBersih, Valid: true}
		}

		// Agen pengguna
		var agenPenggunaVal sql.NullString
		if req.AgenPengguna != nil {
			agenBersih := strings.TrimSpace(*req.AgenPengguna)
			if agenBersih != "" {
				agenPenggunaVal = sql.NullString{String: agenBersih, Valid: true}
			}
		}

		// Asal rujukan
		var asalRujukanVal sql.NullString
		if req.AsalRujukan != nil {
			rujukanBersih := strings.TrimSpace(*req.AsalRujukan)
			if rujukanBersih != "" {
				asalRujukanVal = sql.NullString{String: rujukanBersih, Valid: true}
			}
		}

		kueri := `
			UPDATE aturan
			SET nama = $1, negara = $2, perangkat = $3, peramban = $4, agen_pengguna = $5, asal_rujukan = $6,
			    tujuan_id = $7, prioritas = $8, status = $9, diperbarui_pada = NOW()
			WHERE id = $10
		`

		_, err = db.Exec(
			kueri,
			namaBersih,
			negaraVal,
			perangkatVal,
			perambanVal,
			agenPenggunaVal,
			asalRujukanVal,
			req.TujuanID,
			prioritas,
			statusAktif,
			id,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": true,
				"pesan": "Gagal memperbarui aturan",
			})
			return
		}

		hasil, err := ambilAturanLengkap(db, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": true,
				"pesan": "Gagal memuat aturan setelah diperbarui",
			})
			return
		}

		resp := gin.H{
			"data": hasil,
		}
		if kepatuhan.MenargetkanCrawlerReviewer(namaBersih, agenPenggunaVal.String, asalRujukanVal.String) {
			resp["peringatan"] = "Aturan ini perlu ditinjau karena membedakan pengunjung berdasarkan identitas crawler/reviewer dapat digunakan untuk menghindari pemeriksaan atau kebijakan platform."
		}

		c.JSON(http.StatusOK, resp)
	}
}

// UbahStatusAturan menangani PATCH /api/aturan/:id/status.
// Mengaktifkan (true) atau menonaktifkan (false) aturan.
func UbahStatusAturan(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idTeks := c.Param("id")
		id, err := strconv.Atoi(idTeks)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": true,
				"pesan": "ID aturan tidak valid",
			})
			return
		}

		var req PermintaanStatusAturan
		if err := c.ShouldBindJSON(&req); err != nil || req.Status == nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": true,
				"pesan": "Status aturan wajib diisi (true atau false)",
			})
			return
		}

		kueri := `
			UPDATE aturan
			SET status = $1, diperbarui_pada = NOW()
			WHERE id = $2
		`

		hasilExec, err := db.Exec(kueri, *req.Status, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": true,
				"pesan": "Gagal mengubah status aturan",
			})
			return
		}

		barisTerdampak, err := hasilExec.RowsAffected()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": true,
				"pesan": "Gagal memverifikasi perubahan status",
			})
			return
		}

		if barisTerdampak == 0 {
			c.JSON(http.StatusNotFound, gin.H{
				"error": true,
				"pesan": "Aturan tidak ditemukan",
			})
			return
		}

		hasil, err := ambilAturanLengkap(db, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": true,
				"pesan": "Gagal memuat aturan setelah status diubah",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": hasil,
		})
	}
}

// HapusAturan menangani DELETE /api/aturan/:id.
// Menghapus aturan dari PostgreSQL berdasarkan ID.
func HapusAturan(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idTeks := c.Param("id")
		id, err := strconv.Atoi(idTeks)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": true,
				"pesan": "ID aturan tidak valid",
			})
			return
		}

		kueri := `DELETE FROM aturan WHERE id = $1`
		hasilExec, err := db.Exec(kueri, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": true,
				"pesan": "Gagal menghapus aturan",
			})
			return
		}

		barisTerdampak, err := hasilExec.RowsAffected()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": true,
				"pesan": "Gagal memverifikasi penghapusan",
			})
			return
		}

		if barisTerdampak == 0 {
			c.JSON(http.StatusNotFound, gin.H{
				"error": true,
				"pesan": "Aturan tidak ditemukan",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": gin.H{
				"pesan": "Aturan berhasil dihapus",
			},
		})
	}
}
