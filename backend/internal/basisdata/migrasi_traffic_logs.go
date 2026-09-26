package basisdata

import (
	"database/sql"
	"fmt"
)

// JalankanMigrasiTrafficLogs membuat tabel traffic_logs dan index yang dibutuhkan untuk Batch 5.
func JalankanMigrasiTrafficLogs(db *sql.DB) error {
	sqlMigrasi := `
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
		return fmt.Errorf("gagal menjalankan migrasi traffic_logs: %w", err)
	}

	return nil
}
