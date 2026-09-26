// Tipe entitas data untuk Sivilize Traffic Routing Demo

export interface Tujuan {
  id: number;
  nama: string;
  url: string;
  status: boolean;
  dibuat_pada: string;
  diperbarui_pada: string;
}

export interface Aturan {
  id: number;
  nama: string;
  negara: string | null;
  perangkat: string | null;
  peramban: string | null;
  agen_pengguna: string | null;
  asal_rujukan: string | null;
  tujuan_id: number;
  tujuan_nama: string;
  tujuan_url: string;
  prioritas: number;
  status: boolean;
  dibuat_pada: string;
  diperbarui_pada: string;
}

export interface HasilRouting {
  cocok: boolean;
  aturan_id: number | null;
  tujuan_id: number | null;
  url_tujuan: string;
  hasil: "aturan" | "fallback";
  alasan: string;
}

export interface LogTrafik {
  id: number;
  negara: string | null;
  perangkat: string | null;
  peramban: string | null;
  agen_pengguna?: string | null;
  asal_rujukan: string | null;
  aturan_id: number | null;
  tujuan_id: number | null;
  url_tujuan: string;
  hasil: "aturan" | "fallback";
  dibuat_pada: string;
}

export interface InfoPaginasi {
  page: number;
  limit: number;
  total: number;
  total_halaman: number;
}
