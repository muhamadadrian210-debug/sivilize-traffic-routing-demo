package konfigurasi

import (
	"os"

	"github.com/joho/godotenv"
)

// Konfigurasi menyimpan pengaturan server backend termasuk database.
type Konfigurasi struct {
	Port        string
	UrlFrontend string
	UrlDatabase string
	UrlFallback string
}

// MuatKonfigurasi membaca konfigurasi dari file .env (jika ada) dan environment variable.
func MuatKonfigurasi() Konfigurasi {
	// Coba muat .env — tidak masalah jika tidak ada (pakai env yang sudah ada di sistem)
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	urlFrontend := os.Getenv("FRONTEND_URL")
	if urlFrontend == "" {
		urlFrontend = "http://localhost:3000"
	}

	urlDatabase := os.Getenv("DATABASE_URL")
	if urlDatabase == "" {
		// Nilai bawaan hanya untuk kemudahan pengembangan lokal
		urlDatabase = "postgres://postgres:postgres@localhost:5432/sivilize_trafik?sslmode=disable"
	}

	urlFallback := os.Getenv("FALLBACK_URL")
	if urlFallback == "" {
		urlFallback = "/demo/cadangan"
	}

	return Konfigurasi{
		Port:        port,
		UrlFrontend: urlFrontend,
		UrlDatabase: urlDatabase,
		UrlFallback: urlFallback,
	}
}

// MaskedUrlDatabase mengembalikan URL database tanpa menampilkan password.
// Digunakan untuk logging agar credential tidak bocor ke log.
func (k Konfigurasi) MaskedUrlDatabase() string {
	return "postgres://***@[host]/[database]"
}
