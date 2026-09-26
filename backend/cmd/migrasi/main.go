package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
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

	sqlMigrasi := `
-- Tabel tujuan (dibuat dulu karena aturan bergantung padanya)
CREATE TABLE IF NOT EXISTS tujuan (
    id              SERIAL PRIMARY KEY,
    nama            TEXT        NOT NULL,
    url             TEXT        NOT NULL,
    status          BOOLEAN     NOT NULL DEFAULT TRUE,
    dibuat_pada     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    diperbarui_pada TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Tabel aturan
CREATE TABLE IF NOT EXISTS aturan (
    id              SERIAL PRIMARY KEY,
    nama            TEXT        NOT NULL,
    negara          TEXT,
    perangkat       TEXT,
    peramban        TEXT,
    agen_pengguna   TEXT,
    asal_rujukan    TEXT,
    tujuan_id       INTEGER     NOT NULL REFERENCES tujuan(id),
    prioritas       INTEGER     NOT NULL DEFAULT 100,
    status          BOOLEAN     NOT NULL DEFAULT TRUE,
    dibuat_pada     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    diperbarui_pada TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_aturan_status    ON aturan(status);
CREATE INDEX IF NOT EXISTS idx_aturan_prioritas ON aturan(prioritas);
CREATE INDEX IF NOT EXISTS idx_aturan_negara    ON aturan(negara);
CREATE INDEX IF NOT EXISTS idx_aturan_perangkat ON aturan(perangkat);

-- Tabel catatan_trafik
CREATE TABLE IF NOT EXISTS catatan_trafik (
    id            SERIAL PRIMARY KEY,
    alamat_ip     TEXT,
    negara        TEXT,
    perangkat     TEXT,
    peramban      TEXT,
    agen_pengguna TEXT,
    asal_rujukan  TEXT,
    aturan_id     INTEGER REFERENCES aturan(id),
    tujuan_id     INTEGER REFERENCES tujuan(id),
    hasil         TEXT        NOT NULL CHECK (hasil IN ('cocok', 'cadangan', 'gagal')),
    dibuat_pada   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_catatan_trafik_dibuat_pada ON catatan_trafik(dibuat_pada);
-- Tabel traffic_logs (Batch 5)
CREATE TABLE IF NOT EXISTS traffic_logs (
    id            SERIAL PRIMARY KEY,
    negara        TEXT,
    perangkat     TEXT,
    peramban      TEXT,
    agen_pengguna TEXT,
    asal_rujukan  TEXT,
    aturan_id     INTEGER REFERENCES aturan(id),
    tujuan_id     INTEGER REFERENCES tujuan(id),
    url_tujuan    TEXT        NOT NULL,
    hasil         TEXT        NOT NULL CHECK (hasil IN ('aturan', 'fallback')),
    dibuat_pada   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_traffic_logs_dibuat_pada ON traffic_logs(dibuat_pada DESC);
CREATE INDEX IF NOT EXISTS idx_traffic_logs_aturan_id   ON traffic_logs(aturan_id);
CREATE INDEX IF NOT EXISTS idx_traffic_logs_tujuan_id   ON traffic_logs(tujuan_id);
`

	if _, err := db.Exec(sqlMigrasi); err != nil {
		log.Fatalf("Gagal menjalankan migrasi: %v", err)
	}

	fmt.Println("Migrasi berhasil dijalankan.")
	fmt.Println("Tabel yang tersedia: tujuan, aturan, catatan_trafik, traffic_logs")
}
