export type TingkatTemuan = "informasi" | "peringatan" | "risiko tinggi";
export type StatusKepatuhan = "aman" | "perlu_ditinjau" | "berisiko";

export interface ItemChecklist {
  id: number;
  kunci: string;
  label: string;
  selesai: boolean;
  diperbarui_pada: string;
}

export interface TemuanCompliance {
  id?: number;
  audit_id?: number;
  kode: string;
  tingkat: TingkatTemuan;
  judul: string;
  pesan: string;
  dibuat_pada?: string;
}

export interface RingkasanTemuan {
  total: number;
  informasi: number;
  peringatan: number;
  risiko_tinggi: number;
}

export interface HasilComplianceCheck {
  status: StatusKepatuhan;
  skor: null;
  audit_id: number;
  waktu_pemeriksaan: string;
  ringkasan: RingkasanTemuan;
  temuan: TemuanCompliance[];
  penjelasan: string;
}

export interface AuditRiwayat {
  id: number;
  status: StatusKepatuhan;
  dibuat_pada: string;
  temuan?: TemuanCompliance[];
}
