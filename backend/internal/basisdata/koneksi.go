package basisdata

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

// BukaKoneksi membuka koneksi ke PostgreSQL dan memverifikasi koneksi dengan ping.
// Program berhenti jika database tidak dapat dijangkau.
func BukaKoneksi(urlDatabase string) *sql.DB {
	db, err := sql.Open("postgres", urlDatabase)
	if err != nil {
		log.Fatalf("Gagal membuka koneksi database: %v", err)
	}

	// Konfigurasi pool yang sederhana dan masuk akal untuk skala MVP
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	if err := db.Ping(); err != nil {
		log.Fatalf("Database tidak dapat dijangkau. Pastikan PostgreSQL berjalan dan DATABASE_URL sudah benar. Error: %v", err)
	}

	fmt.Println("Koneksi ke PostgreSQL berhasil.")
	return db
}
