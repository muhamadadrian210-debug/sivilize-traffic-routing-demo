"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { DataKesehatan } from "@/tipe/kesehatan";
import { Tujuan, Aturan, HasilRouting, LogTrafik, InfoPaginasi } from "@/tipe/entitas";
import { dapatkanUrlApi } from "@/pustaka/api";

type TabMenu = "ikhtisar" | "tujuan" | "aturan" | "uji" | "log";

export default function DashboardUtama() {
  const urlBackend = dapatkanUrlApi();
  const [tabAktif, setTabAktif] = useState<TabMenu>("ikhtisar");

  // State Sistem & Kesehatan
  const [kesehatan, setKesehatan] = useState<DataKesehatan | null>(null);
  const [backendTerhubung, setBackendTerhubung] = useState<boolean>(false);
  const [memuatKesehatan, setMemuatKesehatan] = useState<boolean>(true);

  // State Data Entitas
  const [daftarTujuan, setDaftarTujuan] = useState<Tujuan[]>([]);
  const [daftarAturan, setDaftarAturan] = useState<Aturan[]>([]);
  const [daftarLog, setDaftarLog] = useState<LogTrafik[]>([]);
  const [paginasiLog, setPaginasiLog] = useState<InfoPaginasi>({
    page: 1,
    limit: 10,
    total: 0,
    total_halaman: 1,
  });

  // State Form & Aksi
  const [pesanNotifikasi, setPesanNotifikasi] = useState<{ tipe: "sukses" | "error"; pesan: string } | null>(null);
  const [memuatAksi, setMemuatAksi] = useState<boolean>(false);

  // Form Tambah Tujuan
  const [formTujuan, setFormTujuan] = useState({ nama: "", url: "" });

  // Form Tambah Aturan
  const [formAturan, setFormAturan] = useState({
    nama: "",
    negara: "",
    perangkat: "",
    peramban: "",
    tujuan_id: 0,
    prioritas: 1,
  });

  // Form Uji Trafik
  const [formUji, setFormUji] = useState({
    negara: "ID",
    perangkat: "mobile",
    peramban: "Chrome",
    agen_pengguna: "",
    asal_rujukan: "",
  });
  const [hasilRouting, setHasilRouting] = useState<HasilRouting | null>(null);

  // 1. Periksa Kesehatan Backend
  async function periksaKesehatan() {
    try {
      const res = await fetch(`${urlBackend}/health`, {
        method: "GET",
        headers: { "Accept": "application/json" },
      });
      if (res.ok) {
        const data = await res.json();
        setKesehatan(data);
        setBackendTerhubung(true);
      } else {
        setBackendTerhubung(false);
      }
    } catch {
      setBackendTerhubung(false);
      setKesehatan(null);
    } finally {
      setMemuatKesehatan(false);
    }
  }

  // 2. Ambil Data Tujuan
  async function muatTujuan() {
    try {
      const res = await fetch(`${urlBackend}/api/tujuan`);
      if (res.ok) {
        const json = await res.json();
        setDaftarTujuan(json.data || []);
        if (json.data && json.data.length > 0 && formAturan.tujuan_id === 0) {
          setFormAturan((prev) => ({ ...prev, tujuan_id: json.data[0].id }));
        }
      }
    } catch {
      // Ditangani oleh banner status utama
    }
  }

  // 3. Ambil Data Aturan
  async function muatAturan() {
    try {
      const res = await fetch(`${urlBackend}/api/aturan`);
      if (res.ok) {
        const json = await res.json();
        setDaftarAturan(json.data || []);
      }
    } catch {
      // Ditangani oleh banner status utama
    }
  }

  // 4. Ambil Data Log Trafik
  async function muatLog(page = 1) {
    try {
      const res = await fetch(`${urlBackend}/api/log-trafik?page=${page}&limit=10`);
      if (res.ok) {
        const json = await res.json();
        setDaftarLog(json.data || []);
        if (json.pagination) {
          setPaginasiLog(json.pagination);
        }
      }
    } catch {
      // Ditangani oleh banner status utama
    }
  }

  // Muat data awal saat halaman pertama kali dibuka
  useEffect(() => {
    void periksaKesehatan();
    void muatTujuan();
    void muatAturan();
    void muatLog(1);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  function tampilkanNotifikasi(tipe: "sukses" | "error", pesan: string) {
    setPesanNotifikasi({ tipe, pesan });
    setTimeout(() => {
      setPesanNotifikasi(null);
    }, 4000);
  }

  // Aksi: Tambah Tujuan
  async function tanganiTambahTujuan(e: React.FormEvent) {
    e.preventDefault();
    if (!formTujuan.nama || !formTujuan.url) {
      tampilkanNotifikasi("error", "Nama dan URL tujuan wajib diisi");
      return;
    }

    setMemuatAksi(true);
    try {
      const res = await fetch(`${urlBackend}/api/tujuan`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ nama: formTujuan.nama, url: formTujuan.url, status: true }),
      });
      const data = await res.json();
      if (res.ok) {
        tampilkanNotifikasi("sukses", `Tujuan '${formTujuan.nama}' berhasil ditambahkan`);
        setFormTujuan({ nama: "", url: "" });
        muatTujuan();
      } else {
        tampilkanNotifikasi("error", data.pesan || "Gagal menambah tujuan");
      }
    } catch {
      tampilkanNotifikasi("error", "Gagal menghubungi server");
    } finally {
      setMemuatAksi(false);
    }
  }

  // Aksi: Hapus Tujuan
  async function tanganiHapusTujuan(id: number, nama: string) {
    if (!confirm(`Hapus tujuan '${nama}'?`)) return;

    try {
      const res = await fetch(`${urlBackend}/api/tujuan/${id}`, { method: "DELETE" });
      const data = await res.json();
      if (res.ok) {
        tampilkanNotifikasi("sukses", `Tujuan '${nama}' berhasil dihapus`);
        muatTujuan();
      } else {
        tampilkanNotifikasi("error", data.pesan || "Gagal menghapus tujuan");
      }
    } catch {
      tampilkanNotifikasi("error", "Gagal menghubungi server");
    }
  }

  // Aksi: Tambah Aturan
  async function tanganiTambahAturan(e: React.FormEvent) {
    e.preventDefault();
    if (!formAturan.nama) {
      tampilkanNotifikasi("error", "Nama aturan wajib diisi");
      return;
    }
    if (!formAturan.tujuan_id) {
      tampilkanNotifikasi("error", "Pilih tujuan yang valid");
      return;
    }

    setMemuatAksi(true);
    try {
      const payload: Record<string, unknown> = {
        nama: formAturan.nama,
        tujuan_id: Number(formAturan.tujuan_id),
        prioritas: Number(formAturan.prioritas) || 1,
        status: true,
      };

      if (formAturan.negara.trim()) payload.negara = formAturan.negara.trim();
      if (formAturan.perangkat.trim()) payload.perangkat = formAturan.perangkat.trim();
      if (formAturan.peramban.trim()) payload.peramban = formAturan.peramban.trim();

      const res = await fetch(`${urlBackend}/api/aturan`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });
      const data = await res.json();
      if (res.ok) {
        tampilkanNotifikasi("sukses", `Aturan '${formAturan.nama}' berhasil dibuat`);
        setFormAturan({
          nama: "",
          negara: "",
          perangkat: "",
          peramban: "",
          tujuan_id: daftarTujuan[0]?.id || 0,
          prioritas: 1,
        });
        muatAturan();
      } else {
        tampilkanNotifikasi("error", data.pesan || "Gagal menambah aturan");
      }
    } catch {
      tampilkanNotifikasi("error", "Gagal menghubungi server");
    } finally {
      setMemuatAksi(false);
    }
  }

  // Aksi: Toggle Status Aturan
  async function tanganiToggleStatusAturan(id: number, statusSekarang: boolean) {
    try {
      const res = await fetch(`${urlBackend}/api/aturan/${id}/status`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ status: !statusSekarang }),
      });
      const data = await res.json();
      if (res.ok) {
        tampilkanNotifikasi("sukses", `Status aturan ID ${id} diubah menjadi ${!statusSekarang ? "Aktif" : "Nonaktif"}`);
        muatAturan();
      } else {
        tampilkanNotifikasi("error", data.pesan || "Gagal mengubah status aturan");
      }
    } catch {
      tampilkanNotifikasi("error", "Gagal menghubungi server");
    }
  }

  // Aksi: Hapus Aturan
  async function tanganiHapusAturan(id: number, nama: string) {
    if (!confirm(`Hapus aturan '${nama}'?`)) return;

    try {
      const res = await fetch(`${urlBackend}/api/aturan/${id}`, { method: "DELETE" });
      const data = await res.json();
      if (res.ok) {
        tampilkanNotifikasi("sukses", `Aturan '${nama}' berhasil dihapus`);
        muatAturan();
      } else {
        tampilkanNotifikasi("error", data.pesan || "Gagal menghapus aturan");
      }
    } catch {
      tampilkanNotifikasi("error", "Gagal menghubungi server");
    }
  }

  // Aksi: Simulator Periksa Trafik
  async function tanganiUjiTrafik(e: React.FormEvent) {
    e.preventDefault();
    setMemuatAksi(true);
    setHasilRouting(null);

    try {
      const payload: Record<string, string> = {};
      if (formUji.negara) payload.negara = formUji.negara;
      if (formUji.perangkat) payload.perangkat = formUji.perangkat;
      if (formUji.peramban) payload.peramban = formUji.peramban;
      if (formUji.agen_pengguna) payload.agen_pengguna = formUji.agen_pengguna;
      if (formUji.asal_rujukan) payload.asal_rujukan = formUji.asal_rujukan;

      const res = await fetch(`${urlBackend}/api/periksa`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });

      const json = await res.json();
      if (res.ok) {
        setHasilRouting(json.data);
        muatLog(1); // Perbarui log otomatis
      } else {
        tampilkanNotifikasi("error", json.pesan || "Gagal memeriksa trafik");
      }
    } catch {
      tampilkanNotifikasi("error", "Gagal menghubungi server backend");
    } finally {
      setMemuatAksi(false);
    }
  }

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 flex flex-col font-sans">
      {/* 1. Header & Status Koneksi */}
      <header className="border-b border-slate-800 bg-slate-900/60 backdrop-blur sticky top-0 z-30">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 py-4 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-xl bg-gradient-to-tr from-indigo-600 to-violet-500 flex items-center justify-center font-black text-white shadow-lg shadow-indigo-500/20">
              S
            </div>
            <div>
              <h1 className="text-lg font-bold tracking-tight text-white flex items-center gap-2">
                Sivilize Traffic Routing
                <span className="text-xs px-2 py-0.5 rounded-full bg-indigo-500/10 text-indigo-400 border border-indigo-500/20 font-medium">
                  MVP Demo
                </span>
              </h1>
              <p className="text-xs text-slate-400">Dasbor Kontrol Perutean Trafik & Aturan</p>
            </div>
          </div>

          <div className="flex items-center gap-3 w-full sm:w-auto justify-between sm:justify-end">
            <div className="flex items-center gap-2 px-3 py-1.5 rounded-lg bg-slate-900 border border-slate-800 text-xs">
              <span className={`w-2.5 h-2.5 rounded-full ${backendTerhubung ? "bg-emerald-400 animate-pulse" : "bg-rose-500"}`}></span>
              <span className="text-slate-300">
                {memuatKesehatan
                  ? "Memeriksa..."
                  : backendTerhubung
                  ? `Backend: Online (${kesehatan?.database || "terhubung"})`
                  : "Backend: Terputus"}
              </span>
            </div>
            <button
              onClick={() => {
                periksaKesehatan();
                muatTujuan();
                muatAturan();
                muatLog(1);
              }}
              title="Muat Ulang Status"
              className="p-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 hover:text-white transition-colors cursor-pointer"
            >
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"></path>
              </svg>
            </button>
          </div>
        </div>
      </header>

      {/* 2. Banner Notifikasi Pesan */}
      {pesanNotifikasi && (
        <div
          className={`fixed bottom-5 right-5 z-50 px-4 py-3 rounded-xl shadow-2xl border flex items-center gap-3 text-sm transition-all animate-bounce ${
            pesanNotifikasi.tipe === "sukses"
              ? "bg-emerald-950 text-emerald-200 border-emerald-700"
              : "bg-rose-950 text-rose-200 border-rose-700"
          }`}
        >
          <span>{pesanNotifikasi.tipe === "sukses" ? "✓" : "⚠"}</span>
          <span>{pesanNotifikasi.pesan}</span>
        </div>
      )}

      {/* 3. Banner Peringatan Jika Backend Terputus */}
      {!backendTerhubung && !memuatKesehatan && (
        <div className="bg-rose-950/80 border-b border-rose-800/80 px-4 py-3 text-rose-200 text-xs sm:text-sm text-center">
          <strong>Backend tidak dapat dihubungi</strong> pada alamat <code className="bg-rose-900/60 px-1.5 py-0.5 rounded">{urlBackend}</code>. Pastikan server Golang telah aktif.
        </div>
      )}

      {/* 4. Tab Navigasi Utama */}
      <div className="border-b border-slate-800 bg-slate-900/30">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 flex overflow-x-auto scrollbar-none gap-2 py-2">
          {[
            { id: "ikhtisar", label: "Ikhtisar & Sistem", icon: "📊" },
            { id: "tujuan", label: `Tujuan (${daftarTujuan.length})`, icon: "🎯" },
            { id: "aturan", label: `Aturan (${daftarAturan.length})`, icon: "⚡" },
            { id: "uji", label: "Simulator Uji Trafik", icon: "🧪" },
            { id: "log", label: `Log Trafik (${paginasiLog.total})`, icon: "📜" },
          ].map((item) => (
            <button
              key={item.id}
              onClick={() => setTabAktif(item.id as TabMenu)}
              className={`px-4 py-2 rounded-xl text-xs sm:text-sm font-medium whitespace-nowrap transition-all flex items-center gap-2 cursor-pointer ${
                tabAktif === item.id
                  ? "bg-indigo-600 text-white shadow-lg shadow-indigo-600/25"
                  : "text-slate-400 hover:text-slate-200 hover:bg-slate-800/60"
              }`}
            >
              <span>{item.icon}</span>
              <span>{item.label}</span>
            </button>
          ))}
          <Link
            href="/compliance"
            className="px-4 py-2 rounded-xl text-xs sm:text-sm font-medium whitespace-nowrap transition-all flex items-center gap-2 text-emerald-400 hover:text-emerald-300 hover:bg-emerald-950/40 border border-emerald-500/20 ml-auto"
          >
            <span>🛡️</span>
            <span>Kepatuhan (AdSense)</span>
          </Link>
        </div>
      </div>

      {/* 5. Konten Halaman Sesuai Tab Aktif */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 py-8 flex-1 w-full">
        {/* TAB 1: IKHTISAR & SISTEM */}
        {tabAktif === "ikhtisar" && (
          <div className="space-y-6">
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
              <div className="bg-slate-900/70 border border-slate-800 p-5 rounded-2xl">
                <span className="text-slate-400 text-xs font-semibold uppercase">Total Tujuan</span>
                <p className="text-3xl font-extrabold text-white mt-1">{daftarTujuan.length}</p>
                <span className="text-slate-500 text-xs mt-1 block">Halaman landing aktif</span>
              </div>
              <div className="bg-slate-900/70 border border-slate-800 p-5 rounded-2xl">
                <span className="text-slate-400 text-xs font-semibold uppercase">Aturan Aktif</span>
                <p className="text-3xl font-extrabold text-emerald-400 mt-1">
                  {daftarAturan.filter((a) => a.status).length} / {daftarAturan.length}
                </p>
                <span className="text-slate-500 text-xs mt-1 block">Rule Engine siap evaluasi</span>
              </div>
              <div className="bg-slate-900/70 border border-slate-800 p-5 rounded-2xl">
                <span className="text-slate-400 text-xs font-semibold uppercase">Total Riwayat Log</span>
                <p className="text-3xl font-extrabold text-indigo-400 mt-1">{paginasiLog.total}</p>
                <span className="text-slate-500 text-xs mt-1 block">Tercatat di traffic_logs</span>
              </div>
              <div className="bg-slate-900/70 border border-slate-800 p-5 rounded-2xl">
                <span className="text-slate-400 text-xs font-semibold uppercase">Status Database</span>
                <p className={`text-2xl font-bold mt-1 ${backendTerhubung ? "text-emerald-400" : "text-rose-400"}`}>
                  {backendTerhubung ? "PostgreSQL Terhubung" : "Terputus"}
                </p>
                <span className="text-slate-500 text-xs mt-1 block">Port 5432 localhost</span>
              </div>
            </div>

            {/* Alur Sistem Visual */}
            <div className="bg-slate-900/50 border border-slate-800 rounded-2xl p-6 sm:p-8">
              <h2 className="text-lg font-bold text-white mb-2">Alur Keputusan Trafik (Rule-based Routing)</h2>
              <p className="text-slate-400 text-sm mb-6">
                Setiap permintaan trafik yang masuk diproses secara bertingkat dari analisis konteks hingga penyimpanan log:
              </p>
              <div className="grid grid-cols-1 md:grid-cols-5 gap-3 text-center">
                <div className="p-4 rounded-xl bg-slate-950 border border-slate-800">
                  <span className="text-indigo-400 font-bold text-sm block">1. Permintaan Masuk</span>
                  <span className="text-xs text-slate-400 mt-1 block">IP, User-Agent, Referer, Negara</span>
                </div>
                <div className="p-4 rounded-xl bg-slate-950 border border-slate-800">
                  <span className="text-indigo-400 font-bold text-sm block">2. Request Analyzer</span>
                  <span className="text-xs text-slate-400 mt-1 block">Klasifikasi perangkat & peramban</span>
                </div>
                <div className="p-4 rounded-xl bg-slate-950 border border-slate-800">
                  <span className="text-indigo-400 font-bold text-sm block">3. Rule Engine</span>
                  <span className="text-xs text-slate-400 mt-1 block">Pencocokan multi-kondisi & prioritas</span>
                </div>
                <div className="p-4 rounded-xl bg-slate-950 border border-slate-800">
                  <span className="text-indigo-400 font-bold text-sm block">4. Penentuan Tujuan</span>
                  <span className="text-xs text-slate-400 mt-1 block">Halaman Tujuan atau Cadangan</span>
                </div>
                <div className="p-4 rounded-xl bg-slate-950 border border-slate-800">
                  <span className="text-indigo-400 font-bold text-sm block">5. Traffic Logging</span>
                  <span className="text-xs text-slate-400 mt-1 block">Catat hasil ke PostgreSQL</span>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* TAB 2: TUJUAN */}
        {tabAktif === "tujuan" && (
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            <div className="lg:col-span-1 bg-slate-900/70 border border-slate-800 p-6 rounded-2xl h-fit">
              <h2 className="text-base font-bold text-white mb-4">Tambah Halaman Tujuan Baru</h2>
              <form onSubmit={tanganiTambahTujuan} className="space-y-4">
                <div>
                  <label className="text-xs font-semibold text-slate-300 block mb-1">Nama Tujuan</label>
                  <input
                    type="text"
                    required
                    placeholder="Contoh: Landing Promo A"
                    value={formTujuan.nama}
                    onChange={(e) => setFormTujuan({ ...formTujuan, nama: e.target.value })}
                    className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3.5 py-2 text-sm text-white focus:outline-none focus:border-indigo-500"
                  />
                </div>
                <div>
                  <label className="text-xs font-semibold text-slate-300 block mb-1">URL / Path Tujuan</label>
                  <input
                    type="text"
                    required
                    placeholder="Contoh: /demo/promo-a"
                    value={formTujuan.url}
                    onChange={(e) => setFormTujuan({ ...formTujuan, url: e.target.value })}
                    className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3.5 py-2 text-sm text-white focus:outline-none focus:border-indigo-500 font-mono"
                  />
                </div>
                <button
                  type="submit"
                  disabled={memuatAksi}
                  className="w-full py-2.5 px-4 bg-indigo-600 hover:bg-indigo-500 text-white font-medium rounded-xl text-sm transition-colors cursor-pointer disabled:bg-slate-800 disabled:text-slate-500"
                >
                  {memuatAksi ? "Menyimpan..." : "+ Tambah Tujuan"}
                </button>
              </form>
            </div>

            <div className="lg:col-span-2 bg-slate-900/70 border border-slate-800 rounded-2xl overflow-hidden">
              <div className="p-4 sm:p-5 border-b border-slate-800 flex justify-between items-center">
                <h2 className="text-base font-bold text-white">Daftar Halaman Tujuan ({daftarTujuan.length})</h2>
                <button onClick={muatTujuan} className="text-xs text-indigo-400 hover:underline">
                  Segarkan
                </button>
              </div>
              <div className="overflow-x-auto">
                <table className="w-full text-left text-xs sm:text-sm">
                  <thead className="bg-slate-950/80 text-slate-400 uppercase text-[10px] tracking-wider border-b border-slate-800">
                    <tr>
                      <th className="py-3 px-4">ID</th>
                      <th className="py-3 px-4">Nama</th>
                      <th className="py-3 px-4">URL</th>
                      <th className="py-3 px-4">Status</th>
                      <th className="py-3 px-4 text-right">Aksi</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-slate-800/60">
                    {daftarTujuan.length === 0 ? (
                      <tr>
                        <td colSpan={5} className="py-8 text-center text-slate-500">
                          Belum ada halaman tujuan yang tersimpan.
                        </td>
                      </tr>
                    ) : (
                      daftarTujuan.map((t) => (
                        <tr key={t.id} className="hover:bg-slate-800/40">
                          <td className="py-3 px-4 text-slate-400 font-mono">{t.id}</td>
                          <td className="py-3 px-4 font-semibold text-white">{t.nama}</td>
                          <td className="py-3 px-4 font-mono text-indigo-300">{t.url}</td>
                          <td className="py-3 px-4">
                            <span className={`px-2 py-0.5 rounded text-[11px] font-medium ${t.status ? "bg-emerald-500/10 text-emerald-400" : "bg-slate-800 text-slate-400"}`}>
                              {t.status ? "Aktif" : "Nonaktif"}
                            </span>
                          </td>
                          <td className="py-3 px-4 text-right">
                            <button
                              onClick={() => tanganiHapusTujuan(t.id, t.nama)}
                              className="text-rose-400 hover:text-rose-300 text-xs px-2 py-1 rounded bg-rose-500/10 hover:bg-rose-500/20 transition cursor-pointer"
                            >
                              Hapus
                            </button>
                          </td>
                        </tr>
                      ))
                    )}
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        )}

        {/* TAB 3: ATURAN */}
        {tabAktif === "aturan" && (
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            <div className="lg:col-span-1 bg-slate-900/70 border border-slate-800 p-6 rounded-2xl h-fit">
              <h2 className="text-base font-bold text-white mb-4">Buat Aturan Routing Baru</h2>
              <form onSubmit={tanganiTambahAturan} className="space-y-3.5">
                <div>
                  <label className="text-xs font-semibold text-slate-300 block mb-1">Nama Aturan</label>
                  <input
                    type="text"
                    required
                    placeholder="Contoh: Indonesia Mobile"
                    value={formAturan.nama}
                    onChange={(e) => setFormAturan({ ...formAturan, nama: e.target.value })}
                    className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3.5 py-2 text-sm text-white focus:outline-none focus:border-indigo-500"
                  />
                </div>
                <div>
                  <label className="text-xs font-semibold text-slate-300 block mb-1">Tujuan Hasil Routing</label>
                  <select
                    value={formAturan.tujuan_id}
                    onChange={(e) => setFormAturan({ ...formAturan, tujuan_id: Number(e.target.value) })}
                    className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3.5 py-2 text-sm text-white focus:outline-none focus:border-indigo-500"
                  >
                    {daftarTujuan.map((t) => (
                      <option key={t.id} value={t.id}>
                        {t.nama} ({t.url})
                      </option>
                    ))}
                  </select>
                </div>
                <div className="grid grid-cols-2 gap-3">
                  <div>
                    <label className="text-xs font-semibold text-slate-300 block mb-1">Negara (Opsional)</label>
                    <input
                      type="text"
                      placeholder="ID, US, SG"
                      value={formAturan.negara}
                      onChange={(e) => setFormAturan({ ...formAturan, negara: e.target.value })}
                      className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3.5 py-2 text-sm text-white focus:outline-none focus:border-indigo-500"
                    />
                  </div>
                  <div>
                    <label className="text-xs font-semibold text-slate-300 block mb-1">Prioritas</label>
                    <input
                      type="number"
                      min={1}
                      required
                      value={formAturan.prioritas}
                      onChange={(e) => setFormAturan({ ...formAturan, prioritas: Number(e.target.value) })}
                      className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3.5 py-2 text-sm text-white focus:outline-none focus:border-indigo-500"
                    />
                  </div>
                </div>
                <div className="grid grid-cols-2 gap-3">
                  <div>
                    <label className="text-xs font-semibold text-slate-300 block mb-1">Perangkat</label>
                    <select
                      value={formAturan.perangkat}
                      onChange={(e) => setFormAturan({ ...formAturan, perangkat: e.target.value })}
                      className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3.5 py-2 text-sm text-white focus:outline-none focus:border-indigo-500"
                    >
                      <option value="">Semua Perangkat</option>
                      <option value="mobile">mobile</option>
                      <option value="tablet">tablet</option>
                      <option value="desktop">desktop</option>
                    </select>
                  </div>
                  <div>
                    <label className="text-xs font-semibold text-slate-300 block mb-1">Peramban</label>
                    <select
                      value={formAturan.peramban}
                      onChange={(e) => setFormAturan({ ...formAturan, peramban: e.target.value })}
                      className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3.5 py-2 text-sm text-white focus:outline-none focus:border-indigo-500"
                    >
                      <option value="">Semua Peramban</option>
                      <option value="Chrome">Chrome</option>
                      <option value="Firefox">Firefox</option>
                      <option value="Safari">Safari</option>
                      <option value="Edge">Edge</option>
                      <option value="Lainnya">Lainnya</option>
                    </select>
                  </div>
                </div>
                <button
                  type="submit"
                  disabled={memuatAksi}
                  className="w-full py-2.5 px-4 bg-indigo-600 hover:bg-indigo-500 text-white font-medium rounded-xl text-sm transition-colors cursor-pointer disabled:bg-slate-800"
                >
                  {memuatAksi ? "Menyimpan..." : "+ Buat Aturan"}
                </button>
              </form>
            </div>

            <div className="lg:col-span-2 bg-slate-900/70 border border-slate-800 rounded-2xl overflow-hidden">
              <div className="p-4 sm:p-5 border-b border-slate-800 flex justify-between items-center">
                <h2 className="text-base font-bold text-white">Daftar Aturan Evaluasi ({daftarAturan.length})</h2>
                <button onClick={muatAturan} className="text-xs text-indigo-400 hover:underline">
                  Segarkan
                </button>
              </div>
              <div className="overflow-x-auto">
                <table className="w-full text-left text-xs sm:text-sm">
                  <thead className="bg-slate-950/80 text-slate-400 uppercase text-[10px] tracking-wider border-b border-slate-800">
                    <tr>
                      <th className="py-3 px-3">Prioritas</th>
                      <th className="py-3 px-3">Nama Aturan</th>
                      <th className="py-3 px-3">Filter Kondisi</th>
                      <th className="py-3 px-3">Tujuan</th>
                      <th className="py-3 px-3">Status</th>
                      <th className="py-3 px-3 text-right">Aksi</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-slate-800/60">
                    {daftarAturan.length === 0 ? (
                      <tr>
                        <td colSpan={6} className="py-8 text-center text-slate-500">
                          Belum ada aturan aktif yang tersimpan.
                        </td>
                      </tr>
                    ) : (
                      daftarAturan.map((a) => (
                        <tr key={a.id} className="hover:bg-slate-800/40">
                          <td className="py-3 px-3 font-mono font-bold text-amber-400">#{a.prioritas}</td>
                          <td className="py-3 px-3 font-semibold text-white">{a.nama}</td>
                          <td className="py-3 px-3">
                            <div className="flex flex-wrap gap-1">
                              {a.negara && <span className="px-1.5 py-0.5 rounded bg-slate-800 text-[10px] text-indigo-300">Negara: {a.negara}</span>}
                              {a.perangkat && <span className="px-1.5 py-0.5 rounded bg-slate-800 text-[10px] text-cyan-300">{a.perangkat}</span>}
                              {a.peramban && <span className="px-1.5 py-0.5 rounded bg-slate-800 text-[10px] text-teal-300">{a.peramban}</span>}
                              {!a.negara && !a.perangkat && !a.peramban && <span className="text-slate-500 text-[11px]">Semua Trafik</span>}
                            </div>
                          </td>
                          <td className="py-3 px-3">
                            <span className="font-semibold text-white block">{a.tujuan_nama}</span>
                            <span className="text-[11px] font-mono text-indigo-300">{a.tujuan_url}</span>
                          </td>
                          <td className="py-3 px-3">
                            <button
                              onClick={() => tanganiToggleStatusAturan(a.id, a.status)}
                              className={`px-2 py-1 rounded text-xs font-semibold cursor-pointer transition ${
                                a.status ? "bg-emerald-500/20 text-emerald-400 border border-emerald-500/30" : "bg-slate-800 text-slate-400"
                              }`}
                            >
                              {a.status ? "Aktif" : "Nonaktif"}
                            </button>
                          </td>
                          <td className="py-3 px-3 text-right">
                            <button
                              onClick={() => tanganiHapusAturan(a.id, a.nama)}
                              className="text-rose-400 hover:text-rose-300 text-xs px-2 py-1 rounded bg-rose-500/10 hover:bg-rose-500/20 cursor-pointer"
                            >
                              Hapus
                            </button>
                          </td>
                        </tr>
                      ))
                    )}
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        )}

        {/* TAB 4: SIMULATOR UJI TRAFIK */}
        {tabAktif === "uji" && (
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <div className="bg-slate-900/70 border border-slate-800 p-6 rounded-2xl">
              <h2 className="text-base font-bold text-white mb-2">Simulator Permintaan Pengunjung</h2>
              <p className="text-slate-400 text-xs mb-6">
                Kirim data kontekstual simulasi untuk melihat keputusan yang diambil oleh Rule Engine secara live.
              </p>
              <form onSubmit={tanganiUjiTrafik} className="space-y-4">
                <div className="grid grid-cols-2 gap-3">
                  <div>
                    <label className="text-xs font-semibold text-slate-300 block mb-1">Negara Pengunjung</label>
                    <input
                      type="text"
                      placeholder="ID, US, SG, MY"
                      value={formUji.negara}
                      onChange={(e) => setFormUji({ ...formUji, negara: e.target.value })}
                      className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3.5 py-2 text-sm text-white focus:outline-none focus:border-indigo-500"
                    />
                  </div>
                  <div>
                    <label className="text-xs font-semibold text-slate-300 block mb-1">Kategori Perangkat</label>
                    <select
                      value={formUji.perangkat}
                      onChange={(e) => setFormUji({ ...formUji, perangkat: e.target.value })}
                      className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3.5 py-2 text-sm text-white focus:outline-none focus:border-indigo-500"
                    >
                      <option value="mobile">mobile</option>
                      <option value="desktop">desktop</option>
                      <option value="tablet">tablet</option>
                    </select>
                  </div>
                </div>

                <div>
                  <label className="text-xs font-semibold text-slate-300 block mb-1">Peramban Pengunjung</label>
                  <select
                    value={formUji.peramban}
                    onChange={(e) => setFormUji({ ...formUji, peramban: e.target.value })}
                    className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3.5 py-2 text-sm text-white focus:outline-none focus:border-indigo-500"
                  >
                    <option value="Chrome">Chrome</option>
                    <option value="Edge">Edge</option>
                    <option value="Firefox">Firefox</option>
                    <option value="Safari">Safari</option>
                    <option value="Lainnya">Lainnya</option>
                  </select>
                </div>

                <div>
                  <label className="text-xs font-semibold text-slate-300 block mb-1">Asal Rujukan (Referrer)</label>
                  <input
                    type="text"
                    placeholder="https://instagram.com atau https://google.com"
                    value={formUji.asal_rujukan}
                    onChange={(e) => setFormUji({ ...formUji, asal_rujukan: e.target.value })}
                    className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3.5 py-2 text-sm text-white focus:outline-none focus:border-indigo-500"
                  />
                </div>

                <div>
                  <label className="text-xs font-semibold text-slate-300 block mb-1">User-Agent Kustom (Opsional)</label>
                  <textarea
                    rows={2}
                    placeholder="Kosongkan untuk menggunakan default peramban"
                    value={formUji.agen_pengguna}
                    onChange={(e) => setFormUji({ ...formUji, agen_pengguna: e.target.value })}
                    className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3.5 py-2 text-xs text-white focus:outline-none focus:border-indigo-500 font-mono"
                  ></textarea>
                </div>

                <button
                  type="submit"
                  disabled={memuatAksi}
                  className="w-full py-3 px-4 bg-indigo-600 hover:bg-indigo-500 text-white font-semibold rounded-xl text-sm transition-all shadow-lg shadow-indigo-600/30 cursor-pointer disabled:bg-slate-800"
                >
                  {memuatAksi ? "Mengevaluasi Aturan..." : "Periksa Trafik (POST /api/periksa)"}
                </button>
              </form>
            </div>

            {/* Hasil Routing Box */}
            <div className="bg-slate-900/70 border border-slate-800 p-6 rounded-2xl flex flex-col justify-between">
              <div>
                <h2 className="text-base font-bold text-white mb-2">Keputusan Perutean Trafik</h2>
                <p className="text-slate-400 text-xs mb-6">Hasil keputusan tujuan yang diputuskan secara real-time oleh backend Go.</p>

                {hasilRouting ? (
                  <div className="space-y-4">
                    <div
                      className={`p-5 rounded-xl border ${
                        hasilRouting.cocok
                          ? "bg-emerald-950/40 border-emerald-800/80 text-emerald-200"
                          : "bg-amber-950/40 border-amber-800/80 text-amber-200"
                      }`}
                    >
                      <div className="flex items-center justify-between mb-2">
                        <span className="text-xs font-bold uppercase tracking-wider">Hasil Evaluasi:</span>
                        <span
                          className={`px-2.5 py-0.5 rounded-full text-xs font-extrabold uppercase ${
                            hasilRouting.cocok ? "bg-emerald-500/20 text-emerald-400" : "bg-amber-500/20 text-amber-400"
                          }`}
                        >
                          {hasilRouting.hasil}
                        </span>
                      </div>
                      <p className="text-sm font-medium">{hasilRouting.alasan}</p>
                    </div>

                    <div className="bg-slate-950 p-4 rounded-xl border border-slate-800 space-y-2">
                      <div className="flex justify-between items-center text-xs">
                        <span className="text-slate-400">URL Tujuan Terpilih:</span>
                        <code className="text-indigo-400 font-mono font-bold text-sm bg-indigo-950/50 px-2 py-1 rounded">
                          {hasilRouting.url_tujuan}
                        </code>
                      </div>
                      <div className="flex justify-between items-center text-xs">
                        <span className="text-slate-400">ID Aturan yang Cocok:</span>
                        <span className="text-slate-200 font-mono">{hasilRouting.aturan_id ?? "Tidak ada (NULL)"}</span>
                      </div>
                      <div className="flex justify-between items-center text-xs">
                        <span className="text-slate-400">ID Tujuan Database:</span>
                        <span className="text-slate-200 font-mono">{hasilRouting.tujuan_id ?? "Fallback (NULL)"}</span>
                      </div>
                    </div>
                  </div>
                ) : (
                  <div className="py-16 text-center text-slate-500 text-xs">
                    Isi parameter simulasi di sebelah kiri dan klik tombol <strong>Periksa Trafik</strong> untuk melihat hasil.
                  </div>
                )}
              </div>

              <div className="mt-6 pt-4 border-t border-slate-800 text-[11px] text-slate-500">
                Pemeriksaan ini otomatis tersimpan ke dalam tabel <code>traffic_logs</code>.
              </div>
            </div>
          </div>
        )}

        {/* TAB 5: LOG TRAFIK */}
        {tabAktif === "log" && (
          <div className="bg-slate-900/70 border border-slate-800 rounded-2xl overflow-hidden">
            <div className="p-4 sm:p-5 border-b border-slate-800 flex flex-col sm:flex-row justify-between items-start sm:items-center gap-2">
              <div>
                <h2 className="text-base font-bold text-white">Catatan Riwayat Trafik (traffic_logs)</h2>
                <span className="text-xs text-slate-400">Diurutkan berdasarkan waktu terbaru (dibuat_pada DESC)</span>
              </div>
              <button
                onClick={() => muatLog(paginasiLog.page)}
                className="px-3 py-1.5 text-xs bg-slate-800 hover:bg-slate-700 text-slate-200 rounded-lg transition cursor-pointer"
              >
                Muat Ulang
              </button>
            </div>

            <div className="overflow-x-auto">
              <table className="w-full text-left text-xs sm:text-sm">
                <thead className="bg-slate-950/80 text-slate-400 uppercase text-[10px] tracking-wider border-b border-slate-800">
                  <tr>
                    <th className="py-3 px-3">ID</th>
                    <th className="py-3 px-3">Waktu</th>
                    <th className="py-3 px-3">Negara</th>
                    <th className="py-3 px-3">Perangkat</th>
                    <th className="py-3 px-3">Peramban</th>
                    <th className="py-3 px-3">Asal Rujukan</th>
                    <th className="py-3 px-3">Hasil</th>
                    <th className="py-3 px-3">URL Tujuan</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-800/60 font-mono text-xs">
                  {daftarLog.length === 0 ? (
                    <tr>
                      <td colSpan={8} className="py-8 text-center text-slate-500 font-sans">
                        Belum ada riwayat trafik pengunjung yang tercatat.
                      </td>
                    </tr>
                  ) : (
                    daftarLog.map((l) => (
                      <tr key={l.id} className="hover:bg-slate-800/40">
                        <td className="py-3 px-3 text-slate-400">#{l.id}</td>
                        <td className="py-3 px-3 text-slate-400 whitespace-nowrap">
                          {new Date(l.dibuat_pada).toLocaleTimeString("id-ID")}
                        </td>
                        <td className="py-3 px-3 font-sans font-bold text-white">{l.negara || "-"}</td>
                        <td className="py-3 px-3 text-cyan-300 font-sans">{l.perangkat || "-"}</td>
                        <td className="py-3 px-3 text-teal-300 font-sans">{l.peramban || "-"}</td>
                        <td className="py-3 px-3 text-slate-400 max-w-[120px] truncate" title={l.asal_rujukan || ""}>
                          {l.asal_rujukan || "-"}
                        </td>
                        <td className="py-3 px-3">
                          <span
                            className={`px-2 py-0.5 rounded text-[10px] font-sans font-bold uppercase ${
                              l.hasil === "aturan" ? "bg-emerald-500/20 text-emerald-400" : "bg-amber-500/20 text-amber-400"
                            }`}
                          >
                            {l.hasil}
                          </span>
                        </td>
                        <td className="py-3 px-3 text-indigo-300 font-semibold">{l.url_tujuan}</td>
                      </tr>
                    ))
                  )}
                </tbody>
              </table>
            </div>

            {/* Paginasi Kontrol */}
            <div className="p-4 border-t border-slate-800 flex items-center justify-between text-xs text-slate-400">
              <div>
                Menampilkan halaman <strong>{paginasiLog.page}</strong> dari <strong>{paginasiLog.total_halaman}</strong> (Total {paginasiLog.total} data)
              </div>
              <div className="flex gap-2">
                <button
                  disabled={paginasiLog.page <= 1}
                  onClick={() => muatLog(paginasiLog.page - 1)}
                  className="px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 disabled:opacity-40 disabled:cursor-not-allowed text-white cursor-pointer"
                >
                  Sebelumnya
                </button>
                <button
                  disabled={paginasiLog.page >= paginasiLog.total_halaman}
                  onClick={() => muatLog(paginasiLog.page + 1)}
                  className="px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 disabled:opacity-40 disabled:cursor-not-allowed text-white cursor-pointer"
                >
                  Berikutnya
                </button>
              </div>
            </div>
          </div>
        )}
      </main>

      {/* Footer */}
      <footer className="border-t border-slate-800 py-6 text-center text-xs text-slate-500">
        Sivilize Traffic Routing Demo &bull; MVP Proof of Concept &bull; Sivilize Corp
      </footer>
    </div>
  );
}
