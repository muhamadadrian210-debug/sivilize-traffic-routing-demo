package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"sivilize-traffic-routing-demo/backend/internal/basisdata"
)

// cmd/migrasi — dijalankan sekali untuk membuat struktur tabel database.
// Gunakan: go run ./cmd/migrasi
func main() {
	// Muat .env jika ada
	_ = godotenv.Load()

	urlDatabase := os.Getenv("DATABASE_URL")
	if urlDatabase == "" {
		urlDatabase = "postgres://postgres:postgres@localhost:5432/sivilize_trafik?sslmode=disable"
	}

	db, err := sql.Open("postgres", urlDatabase)
	if err != nil {
		log.Fatalf("Gagal membuka koneksi database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Database tidak dapat dijangkau. Pastikan PostgreSQL sudah berjalan. Error: %v", err)
	}

	fmt.Println("Koneksi berhasil. Menjalankan migrasi...")

	if err := basisdata.JalankanMigrasi(db); err != nil {
		log.Fatalf("Gagal menjalankan migrasi: %v", err)
	}

	fmt.Println("Migrasi berhasil dijalankan.")
	fmt.Println("Tabel yang tersedia: tujuan, aturan, catatan_trafik, traffic_logs, compliance_checklist, compliance_audits, compliance_findings")
}
