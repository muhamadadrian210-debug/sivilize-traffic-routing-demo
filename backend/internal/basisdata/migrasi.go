package basisdata

import (
	"database/sql"
	"fmt"
)

// JalankanMigrasi mengeksekusi SQL DDL untuk membuat seluruh struktur tabel.
// Menggunakan IF NOT EXISTS di setiap statement, aman dijalankan berulang kali.
func JalankanMigrasi(db *sql.DB) error {
	sqlMigrasi := `
CREATE TABLE IF NOT EXISTS tujuan (
    id              SERIAL PRIMARY KEY,
    nama            TEXT        NOT NULL,
    url             TEXT        NOT NULL,
    status          BOOLEAN     NOT NULL DEFAULT TRUE,
    dibuat_pada     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    diperbarui_pada TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

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
CREATE INDEX IF NOT EXISTS idx_catatan_trafik_negara      ON catatan_trafik(negara);
CREATE INDEX IF NOT EXISTS idx_catatan_trafik_perangkat   ON catatan_trafik(perangkat);
`

	if _, err := db.Exec(sqlMigrasi); err != nil {
		return fmt.Errorf("gagal menjalankan migrasi DDL: %w", err)
	}

	// Jalankan migrasi tabel traffic_logs (Batch 5)
	if err := JalankanMigrasiTrafficLogs(db); err != nil {
		return fmt.Errorf("gagal menjalankan migrasi traffic_logs: %w", err)
	}

	// Jalankan migrasi tabel kepatuhan compliance (Batch 8)
	if err := JalankanMigrasiCompliance(db); err != nil {
		return fmt.Errorf("gagal menjalankan migrasi compliance: %w", err)
	}

	return nil
}
