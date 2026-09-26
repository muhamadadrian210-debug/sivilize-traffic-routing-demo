package basisdata

import (
	"database/sql"
	"fmt"
)

// JalankanMigrasiCompliance membuat tabel-tabel kepatuhan dan audit Google Publisher / AdSense (Batch 8).
func JalankanMigrasiCompliance(db *sql.DB) error {
	sqlMigrasi := `
CREATE TABLE IF NOT EXISTS compliance_checklist (
    id              SERIAL PRIMARY KEY,
    kunci           TEXT        NOT NULL UNIQUE,
    label           TEXT        NOT NULL,
    selesai         BOOLEAN     NOT NULL DEFAULT FALSE,
    diperbarui_pada TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS compliance_audits (
    id          SERIAL PRIMARY KEY,
    status      TEXT        NOT NULL CHECK (status IN ('aman', 'perlu_ditinjau', 'berisiko')),
    dibuat_pada TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS compliance_findings (
    id          SERIAL PRIMARY KEY,
    audit_id    INTEGER     NOT NULL REFERENCES compliance_audits(id) ON DELETE CASCADE,
    kode        TEXT        NOT NULL,
    tingkat     TEXT        NOT NULL CHECK (tingkat IN ('informasi', 'peringatan', 'risiko tinggi')),
    judul       TEXT        NOT NULL,
    pesan       TEXT        NOT NULL,
    dibuat_pada TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_compliance_audits_dibuat_pada   ON compliance_audits(dibuat_pada DESC);
CREATE INDEX IF NOT EXISTS idx_compliance_findings_audit_id    ON compliance_findings(audit_id);
CREATE INDEX IF NOT EXISTS idx_compliance_findings_tingkat     ON compliance_findings(tingkat);

-- Masukkan 11 item checklist kebijakan kepatuhan standar jika belum ada
INSERT INTO compliance_checklist (kunci, label, selesai)
VALUES 
    ('no_self_clicks', 'Tidak melakukan klik pada iklan sendiri', false),
    ('no_auto_traffic', 'Tidak menggunakan traffic otomatis untuk menghasilkan klik/tayangan', false),
    ('no_traffic_exchange', 'Tidak menggunakan paid-to-click / paid-to-surf / traffic exchange', false),
    ('no_incentivized_clicks', 'Tidak memberikan insentif untuk klik iklan', false),
    ('no_asking_for_clicks', 'Tidak meminta pengguna mengklik iklan', false),
    ('accurate_ad_placement', 'Penempatan iklan tidak menyesatkan', false),
    ('content_policy_checked', 'Konten website telah diperiksa', false),
    ('traffic_sources_reviewed', 'Sumber traffic telah diperiksa', false),
    ('no_crawler_evasion', 'Tidak menggunakan routing untuk menghindari pemeriksaan Google', false),
    ('no_cloaking', 'Tidak menggunakan cloaking untuk menghindari kebijakan iklan', false),
    ('no_targeting_manipulation', 'Tidak melakukan manipulasi targeting iklan untuk meningkatkan pendapatan', false)
ON CONFLICT (kunci) DO NOTHING;
`

	if _, err := db.Exec(sqlMigrasi); err != nil {
		return fmt.Errorf("gagal menjalankan migrasi compliance: %w", err)
	}

	return nil
}
