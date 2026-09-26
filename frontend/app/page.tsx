"use client";

import { useEffect, useState } from "react";
import { DataKesehatan } from "@/tipe/kesehatan";
import { dapatkanUrlApi } from "@/pustaka/api";

export default function HalamanBeranda() {
  const [dataKesehatan, setDataKesehatan] = useState<DataKesehatan | null>(null);
  const [sedangMemuat, setSedangMemuat] = useState<boolean>(true);
  const [pesanKesalahan, setPesanKesalahan] = useState<string | null>(null);
  const [waktuPemeriksaan, setWaktuPemeriksaan] = useState<string | null>(null);

  const urlBackend = dapatkanUrlApi();

  async function periksaKoneksiBackend() {
    setSedangMemuat(true);
    setPesanKesalahan(null);

    try {
      const respons = await fetch(`${urlBackend}/api/kesehatan`, {
        method: "GET",
        headers: {
          "Accept": "application/json",
        },
      });

      if (!respons.ok) {
        throw new Error(`Server merespons dengan kode status: ${respons.status}`);
      }

      const hasil: DataKesehatan = await respons.json();
      setDataKesehatan(hasil);
      setWaktuPemeriksaan(new Date().toLocaleTimeString("id-ID"));
    } catch (kesalahan) {
      setDataKesehatan(null);
      if (kesalahan instanceof Error) {
        setPesanKesalahan(kesalahan.message);
      } else {
        setPesanKesalahan("Gagal terhubung ke server backend");
      }
    } finally {
      setSedangMemuat(false);
    }
  }

  useEffect(() => {
    periksaKoneksiBackend();
  }, []);

  return (
    <main className="min-h-screen px-4 py-12 max-w-4xl mx-auto flex flex-col justify-center">
      {/* Kartu Header & Identitas Proyek */}
      <div className="text-center mb-10">
        <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full text-xs font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 mb-4">
          <span className="w-2 h-2 rounded-full bg-emerald-400 animate-pulse"></span>
          Tahap 1: Verifikasi Fondasi Frontend & Backend
        </div>
        <h1 className="text-3xl sm:text-4xl font-extrabold tracking-tight text-white mb-2">
          Sivilize Traffic Routing Demo
        </h1>
        <p className="text-slate-400 text-sm sm:text-base max-w-xl mx-auto">
          Sistem routing trafik berbasis aturan. Memastikan komunikasi REST API antara Next.js dan backend Golang berfungsi dengan stabil.
        </p>
      </div>

      {/* Kartu Status Pemeriksaan Kesehatan API */}
      <div className="bg-slate-900/80 border border-slate-800 rounded-2xl p-6 sm:p-8 backdrop-blur-sm shadow-xl">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between pb-6 border-b border-slate-800 gap-4">
          <div>
            <span className="text-xs uppercase tracking-wider text-slate-400 font-semibold block mb-1">
              Endpoint Pengujian
            </span>
            <code className="text-sm font-mono bg-slate-950 text-indigo-300 px-3 py-1.5 rounded-lg border border-slate-800 inline-block">
              GET {urlBackend}/api/kesehatan
            </code>
          </div>

          <button
            onClick={periksaKoneksiBackend}
            disabled={sedangMemuat}
            className="inline-flex items-center justify-center px-4 py-2 text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-500 disabled:bg-slate-800 disabled:text-slate-500 rounded-lg transition-colors shadow-sm cursor-pointer disabled:cursor-not-allowed"
          >
            {sedangMemuat ? (
              <span className="flex items-center gap-2">
                <svg className="animate-spin h-4 w-4 text-white" viewBox="0 0 24 24">
                  <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none"></circle>
                  <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                Memeriksa...
              </span>
            ) : (
              "Periksa Ulang Koneksi"
            )}
          </button>
        </div>

        {/* Indikator Hasil */}
        <div className="mt-6">
          {sedangMemuat && (
            <div className="py-8 text-center text-slate-400">
              <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-indigo-500 mb-3"></div>
              <p className="text-sm">Sedang menghubungi server backend Golang...</p>
            </div>
          )}

          {!sedangMemuat && dataKesehatan && (
            <div className="space-y-4">
              <div className="flex items-start gap-4 p-4 rounded-xl bg-emerald-950/30 border border-emerald-800/50">
                <div className="w-9 h-9 rounded-lg bg-emerald-500/20 text-emerald-400 flex items-center justify-center shrink-0">
                  <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M5 13l4 4L19 7"></path>
                  </svg>
                </div>
                <div>
                  <h2 className="text-base font-semibold text-emerald-300">
                    Koneksi Berhasil Terhubung
                  </h2>
                  <p className="text-sm text-emerald-400/80 mt-0.5">
                    {dataKesehatan.pesan || `Server: ${dataKesehatan.status} • PostgreSQL: ${dataKesehatan.database || "terhubung"}`}
                  </p>
                  {waktuPemeriksaan && (
                    <span className="text-xs text-slate-400 block mt-2">
                      Terakhir diperiksa pukul {waktuPemeriksaan} WIB
                    </span>
                  )}
                </div>
              </div>

              {/* Data Respons JSON */}
              <div>
                <span className="text-xs font-semibold text-slate-400 uppercase tracking-wider block mb-2">
                  Respons JSON dari Backend:
                </span>
                <pre className="bg-slate-950 p-4 rounded-xl text-xs sm:text-sm font-mono text-emerald-400 border border-slate-800 overflow-x-auto">
                  {JSON.stringify(dataKesehatan, null, 2)}
                </pre>
              </div>
            </div>
          )}

          {!sedangMemuat && pesanKesalahan && (
            <div className="space-y-4">
              <div className="flex items-start gap-4 p-4 rounded-xl bg-rose-950/30 border border-rose-800/50">
                <div className="w-9 h-9 rounded-lg bg-rose-500/20 text-rose-400 flex items-center justify-center shrink-0">
                  <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M6 18L18 6M6 6l12 12"></path>
                  </svg>
                </div>
                <div>
                  <h2 className="text-base font-semibold text-rose-300">
                    Koneksi Terputus atau Backend Belum Aktif
                  </h2>
                  <p className="text-sm text-rose-400/80 mt-0.5">
                    {pesanKesalahan}
                  </p>
                </div>
              </div>

              <div className="p-4 rounded-xl bg-slate-950 border border-slate-800 text-xs sm:text-sm text-slate-300 space-y-2">
                <p className="font-semibold text-slate-200">Panduan Menjalankan Backend:</p>
                <ol className="list-decimal list-inside space-y-1 text-slate-400 font-mono">
                  <li>Buka terminal baru pada direktori backend: <span className="text-indigo-300">cd backend</span></li>
                  <li>Jalankan server: <span className="text-indigo-300">go run ./cmd/server</span></li>
                  <li>Pastikan server berjalan pada port <span className="text-indigo-300">:8080</span></li>
                  <li>Klik tombol &quot;Periksa Ulang Koneksi&quot; di atas.</li>
                </ol>
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Rencana Tahapan Berikutnya */}
      <div className="mt-8 text-center text-xs text-slate-500">
        Status Saat Ini: <span className="text-slate-300">Tahap 1 Selesai</span> &bull; Menunggu instruksi untuk <span className="text-slate-300">Tahap 2 (PostgreSQL: Migrasi, Model & Seed)</span>
      </div>
    </main>
  );
}
