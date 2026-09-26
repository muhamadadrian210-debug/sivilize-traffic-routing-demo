package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// cmd/seed — mengisi data awal demo ke database.
// Aman dijalankan berulang kali (idempoten menggunakan ON CONFLICT DO NOTHING).
// Gunakan: go run ./cmd/seed
func main() {
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
		log.Fatalf("Database tidak dapat dijangkau. Pastikan migrasi sudah dijalankan terlebih dahulu. Error: %v", err)
	}

	fmt.Println("Koneksi berhasil. Memasukkan data awal...")

	if err := isiDataAwal(db); err != nil {
		log.Fatalf("Gagal memasukkan data awal: %v", err)
	}

	fmt.Println("Data awal berhasil dimasukkan.")
}

func isiDataAwal(db *sql.DB) error {
	// Masukkan tiga tujuan demo.
	// ON CONFLICT DO NOTHING → aman dijalankan berulang, tidak membuat duplikat.
	// Kolom url dijadikan acuan konflik karena URL bersifat unik per tujuan.
	_, err := db.Exec(`
		ALTER TABLE tujuan DROP CONSTRAINT IF EXISTS tujuan_url_key;
		ALTER TABLE tujuan ADD CONSTRAINT tujuan_url_key UNIQUE (url);
	`)
	if err != nil {
		return fmt.Errorf("gagal menambahkan constraint unique pada tujuan.url: %w", err)
	}

	tujuanAwal := []struct {
		nama string
		url  string
	}{
		{"Halaman A", "/demo/halaman-a"},
		{"Halaman B", "/demo/halaman-b"},
		{"Cadangan", "/demo/cadangan"},
	}

	for _, t := range tujuanAwal {
		_, err := db.Exec(`
			INSERT INTO tujuan (nama, url, status)
			VALUES ($1, $2, TRUE)
			ON CONFLICT (url) DO NOTHING
		`, t.nama, t.url)
		if err != nil {
			return fmt.Errorf("gagal menyisipkan tujuan '%s': %w", t.nama, err)
		}
	}

	// Ambil ID tujuan Halaman A dan Halaman B untuk relasi aturan
	var idHalamanA, idHalamanB int
	if err := db.QueryRow(`SELECT id FROM tujuan WHERE url = '/demo/halaman-a'`).Scan(&idHalamanA); err != nil {
		return fmt.Errorf("gagal mengambil ID Halaman A: %w", err)
	}
	if err := db.QueryRow(`SELECT id FROM tujuan WHERE url = '/demo/halaman-b'`).Scan(&idHalamanB); err != nil {
		return fmt.Errorf("gagal mengambil ID Halaman B: %w", err)
	}

	// Tambahkan unique constraint pada nama aturan jika belum ada
	_, err = db.Exec(`
		ALTER TABLE aturan DROP CONSTRAINT IF EXISTS aturan_nama_key;
		ALTER TABLE aturan ADD CONSTRAINT aturan_nama_key UNIQUE (nama);
	`)
	if err != nil {
		return fmt.Errorf("gagal menambahkan constraint unique pada aturan.nama: %w", err)
	}

	// Masukkan dua aturan demo
	aturanAwal := []struct {
		nama      string
		negara    string
		perangkat string
		tujuanID  int
		prioritas int
	}{
		{"Indonesia Mobile", "ID", "mobile", idHalamanA, 1},
		{"Indonesia Desktop", "ID", "desktop", idHalamanB, 2},
	}

	for _, a := range aturanAwal {
		_, err := db.Exec(`
			INSERT INTO aturan (nama, negara, perangkat, tujuan_id, prioritas, status)
			VALUES ($1, $2, $3, $4, $5, TRUE)
			ON CONFLICT (nama) DO NOTHING
		`, a.nama, a.negara, a.perangkat, a.tujuanID, a.prioritas)
		if err != nil {
			return fmt.Errorf("gagal menyisipkan aturan '%s': %w", a.nama, err)
		}
	}

	fmt.Printf("  ✓ Tujuan: Halaman A, Halaman B, Cadangan\n")
	fmt.Printf("  ✓ Aturan: Indonesia Mobile (prioritas 1), Indonesia Desktop (prioritas 2)\n")

	return nil
}
