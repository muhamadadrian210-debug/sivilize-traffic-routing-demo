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
    if (!formTujuan.nama || !formTujuan.url) {
      tampilkanNotifikasi("error", "Nama dan URL tujuan wajib diisi");
      return;
    }

    let urlBersih = formTujuan.url.trim();
    // Jika user menginputkan domain tanpa protokol (seperti google.com), otomatis tambahkan https://
    if (!urlBersih.startsWith("http://") && !urlBersih.startsWith("https://") && !urlBersih.startsWith("/")) {
      urlBersih = "https://" + urlBersih;
    }

    setMemuatAksi(true);
    try {
      const res = await fetch(`${urlBackend}/api/tujuan`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ nama: formTujuan.nama, url: urlBersih, status: true }),
      });
      const data = await res.json();
      if (res.ok) {
        tampilkanNotifikasi("sukses", `Tujuan '${formTujuan.nama}' berhasil ditambahkan (${urlBersih})`);
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
      <div className="border-b border-slate-800/80 bg-slate-950/60 backdrop-blur-md">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 flex overflow-x-auto scrollbar-none gap-2 py-2.5">
          {[
            {
              id: "ikhtisar",
              label: "Ikhtisar & Sistem",
              icon: (
                <svg className="w-4 h-4" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" d="M3 13.125C3 12.504 3.504 12 4.125 12h2.25c.621 0 1.125.504 1.125 1.125v6.75C7.5 20.496 6.996 21 6.375 21h-2.25A1.125 1.125 0 013 19.875v-6.75zM9.75 8.625c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125v11.25c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V8.625zM16.5 4.125c0-.621.504-1.125 1.125-1.125h2.25C20.496 3 21 3.504 21 4.125v15.75c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V4.125z" />
                </svg>
              ),
            },
            {
              id: "tujuan",
              label: `Tujuan (${daftarTujuan.length})`,
              icon: (
                <svg className="w-4 h-4" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
                  <circle cx="12" cy="12" r="9" />
                  <path strokeLinecap="round" strokeLinejoin="round" d="M12 7v5l3 3" />
                </svg>
              ),
            },
            {
              id: "aturan",
              label: `Aturan (${daftarAturan.length})`,
              icon: (
                <svg className="w-4 h-4" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" d="M6 13.5V3.75m0 9.75a1.5 1.5 0 010 3m0-3a1.5 1.5 0 000 3m0 0v3.75m6-13.5V3.75m0 9.75a1.5 1.5 0 010 3m0-3a1.5 1.5 0 000 3m0 0v3.75m6-7.5V3.75m0 5.25a1.5 1.5 0 010 3m0-3a1.5 1.5 0 000 3m0 0v9.75" />
                </svg>
              ),
            },
            {
              id: "uji",
              label: "Simulator Uji Trafik",
              icon: (
                <svg className="w-4 h-4" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" d="M5.25 5.653c0-.856.917-1.398 1.667-.986l11.54 6.348a1.125 1.125 0 010 1.971l-11.54 6.347a1.125 1.125 0 01-1.667-.985V5.653z" />
                </svg>
              ),
            },
            {
              id: "log",
              label: `Log Trafik (${paginasiLog.total})`,
              icon: (
                <svg className="w-4 h-4" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" d="M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m0 12.75h7.5m-7.5 3H12M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 00-9-9z" />
                </svg>
              ),
            },
          ].map((item) => (
            <button
              key={item.id}
              onClick={() => setTabAktif(item.id as TabMenu)}
              className={`px-4 py-2 rounded-xl text-xs sm:text-sm font-semibold whitespace-nowrap transition-all flex items-center gap-2.5 cursor-pointer ${
                tabAktif === item.id
                  ? "bg-gradient-to-r from-indigo-600 to-violet-600 text-white shadow-lg shadow-indigo-600/30 border border-indigo-400/30"
                  : "text-slate-400 hover:text-slate-100 hover:bg-slate-900/80 border border-transparent"
              }`}
            >
              <span>{item.icon}</span>
              <span>{item.label}</span>
            </button>
          ))}

          <Link
            href="/compliance"
            className="px-4 py-2 rounded-xl text-xs sm:text-sm font-semibold whitespace-nowrap transition-all flex items-center gap-2.5 text-emerald-300 hover:text-emerald-100 bg-emerald-950/40 hover:bg-emerald-900/50 border border-emerald-500/30 ml-auto shadow-sm"
          >
            <svg className="w-4 h-4 text-emerald-400" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" d="M9 12.75L11.25 15 15 9.75m-3-7.036A11.959 11.959 0 013.598 6 11.99 11.99 0 003 9.749c0 5.592 3.824 10.29 9 11.623 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.248-8.25-3.285z" />
            </svg>
            <span>Kepatuhan (AdSense)</span>
          </Link>
        </div>
      </div>

      {/* 5. Konten Halaman Sesuai Tab Aktif */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 py-8 flex-1 w-full">
        {/* TAB 1: IKHTISAR & SISTEM */}
        {tabAktif === "ikhtisar" && (
          <div className="space-y-8">
            {/* 1. Stat Telemetry Cards - Vibrant Multi-color Harmonies */}
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
              {/* Card 1: Tujuan (Electric Cyan) */}
              <div className="bg-gradient-to-b from-slate-900/90 to-slate-950/90 border border-slate-800/80 hover:border-cyan-500/60 transition-all rounded-2xl p-5 shadow-lg relative overflow-hidden group">
                <div className="absolute top-0 left-0 right-0 h-1 bg-gradient-to-r from-cyan-400 to-blue-500"></div>
                <div className="flex items-center justify-between">
                  <span className="text-slate-300 text-xs font-bold uppercase tracking-wider">Halaman Tujuan</span>
                  <span className="px-2 py-0.5 rounded-full text-[10px] font-bold bg-cyan-500/15 text-cyan-300 border border-cyan-500/30">
                    TARGETS
                  </span>
                </div>
                <p className="text-3xl font-extrabold text-cyan-300 mt-2 tracking-tight">{daftarTujuan.length}</p>
                <div className="flex items-center gap-2 mt-2">
                  <span className="h-1.5 w-1.5 rounded-full bg-cyan-400"></span>
                  <span className="text-slate-300 text-xs font-medium">Halaman landing aktif terdaftar</span>
                </div>
              </div>

              {/* Card 2: Aturan Aktif (Warm Amber / Sunset Gold) */}
              <div className="bg-gradient-to-b from-slate-900/90 to-slate-950/90 border border-slate-800/80 hover:border-amber-500/60 transition-all rounded-2xl p-5 shadow-lg relative overflow-hidden group">
                <div className="absolute top-0 left-0 right-0 h-1 bg-gradient-to-r from-amber-400 to-orange-500"></div>
                <div className="flex items-center justify-between">
                  <span className="text-slate-300 text-xs font-bold uppercase tracking-wider">Aturan Evaluasi</span>
                  <span className="px-2 py-0.5 rounded-full text-[10px] font-bold bg-amber-500/15 text-amber-300 border border-amber-500/30">
                    ACTIVE RULES
                  </span>
                </div>
                <p className="text-3xl font-extrabold text-amber-400 mt-2 tracking-tight">
                  {daftarAturan.filter((a) => a.status).length} <span className="text-slate-400 text-lg font-normal">/ {daftarAturan.length}</span>
                </p>
                <div className="flex items-center gap-2 mt-2">
                  <span className="h-1.5 w-1.5 rounded-full bg-amber-400 animate-pulse"></span>
                  <span className="text-slate-300 text-xs font-medium">Rule Engine siap evaluasi prioritas</span>
                </div>
              </div>

              {/* Card 3: Total Log (Vibrant Violet / Royal Indigo) */}
              <div className="bg-gradient-to-b from-slate-900/90 to-slate-950/90 border border-slate-800/80 hover:border-violet-500/60 transition-all rounded-2xl p-5 shadow-lg relative overflow-hidden group">
                <div className="absolute top-0 left-0 right-0 h-1 bg-gradient-to-r from-violet-400 to-indigo-500"></div>
                <div className="flex items-center justify-between">
                  <span className="text-slate-300 text-xs font-bold uppercase tracking-wider">Riwayat Log Trafik</span>
                  <span className="px-2 py-0.5 rounded-full text-[10px] font-bold bg-violet-500/15 text-violet-300 border border-violet-500/30">
                    AUDIT SINK
                  </span>
                </div>
                <p className="text-3xl font-extrabold text-violet-300 mt-2 tracking-tight">{paginasiLog.total}</p>
                <div className="flex items-center gap-2 mt-2">
                  <span className="h-1.5 w-1.5 rounded-full bg-violet-400"></span>
                  <span className="text-slate-300 text-xs font-medium">Tercatat di tabel traffic_logs</span>
                </div>
              </div>

              {/* Card 4: Status Database (Emerald Mint) */}
              <div className="bg-gradient-to-b from-slate-900/90 to-slate-950/90 border border-slate-800/80 hover:border-emerald-500/60 transition-all rounded-2xl p-5 shadow-lg relative overflow-hidden group">
                <div className={`absolute top-0 left-0 right-0 h-1 bg-gradient-to-r ${backendTerhubung ? "from-emerald-400 to-teal-400" : "from-rose-500 to-rose-400"}`}></div>
                <div className="flex items-center justify-between">
                  <span className="text-slate-300 text-xs font-bold uppercase tracking-wider">Koneksi Database</span>
                  <span className={`px-2 py-0.5 rounded-full text-[10px] font-bold border ${backendTerhubung ? "bg-emerald-500/15 text-emerald-300 border-emerald-500/30" : "bg-rose-500/15 text-rose-300 border-rose-500/30"}`}>
                    {backendTerhubung ? "CONNECTED" : "OFFLINE"}
                  </span>
                </div>
                <p className={`text-2xl font-extrabold mt-2 tracking-tight ${backendTerhubung ? "text-emerald-300" : "text-rose-400"}`}>
                  {backendTerhubung ? "PostgreSQL 17" : "Terputus"}
                </p>
                <div className="flex items-center gap-2 mt-2">
                  <span className={`h-1.5 w-1.5 rounded-full ${backendTerhubung ? "bg-emerald-400 animate-pulse" : "bg-rose-500"}`}></span>
                  <span className="text-slate-300 text-xs font-medium">
                    {backendTerhubung ? "Port 5432 • ACID Logged" : "Periksa koneksi backend"}
                  </span>
                </div>
              </div>
            </div>

            {/* 2. Arsitektur Pipeline Alur Keputusan Trafik - Multi-color Sequential Spectrum */}
            <div className="bg-gradient-to-b from-slate-900/95 via-slate-900/90 to-slate-950 border border-slate-700/80 rounded-2xl p-6 sm:p-8 shadow-2xl relative">
              {/* Header Pipeline */}
              <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-6 border-b border-slate-800/80">
                <div>
                  <div className="flex items-center gap-2 mb-1.5">
                    <span className="px-2.5 py-0.5 rounded-md text-[11px] font-bold uppercase tracking-wider bg-gradient-to-r from-cyan-500/20 via-indigo-500/20 to-emerald-500/20 text-cyan-300 border border-cyan-500/30">
                      PIPELINE ARCHITECTURE
                    </span>
                    <span className="text-slate-500">•</span>
                    <span className="text-xs text-slate-300 font-medium">Deterministic Rule-based Routing</span>
                  </div>
                  <h2 className="text-xl sm:text-2xl font-bold text-white tracking-tight">
                    Alur Keputusan Trafik (Rule-based Routing)
                  </h2>
                  <p className="text-slate-300 text-sm mt-1 max-w-3xl leading-relaxed">
                    Setiap permintaan trafik yang masuk diproses secara berurutan dan deterministik dari analisis konteks hingga persistensi log:
                  </p>
                </div>
                <div className="flex items-center gap-2.5 px-3.5 py-1.5 rounded-xl bg-slate-950/90 border border-slate-700/80 text-xs text-slate-200 self-start sm:self-auto shadow-inner">
                  <span className="h-2 w-2 rounded-full bg-emerald-400 animate-pulse"></span>
                  <span className="font-semibold text-white">5 Tahap Eksekusi</span>
                  <span className="text-slate-500">|</span>
                  <span className="text-cyan-300 font-mono font-medium">&lt; 2ms latency</span>
                </div>
              </div>

              {/* 5-Step Connected Pipeline Cards with Unique Color Accents and Vector SVG Icons */}
              <div className="grid grid-cols-1 md:grid-cols-5 gap-3.5 pt-6 relative">
                {/* Step 1: Ingress (Electric Cyan) */}
                <div className="bg-slate-950/90 border border-slate-800 hover:border-cyan-400/70 transition-all duration-200 rounded-xl p-4 flex flex-col justify-between shadow-lg relative group">
                  <div className="absolute top-0 left-0 right-0 h-0.5 bg-gradient-to-r from-cyan-500 to-blue-500 rounded-t-xl"></div>
                  <div>
                    <div className="flex items-center justify-between mb-3">
                      <span className="px-2 py-0.5 rounded bg-cyan-500/15 text-cyan-300 text-[11px] font-bold font-mono border border-cyan-500/30">
                        STEP 01
                      </span>
                      <div className="p-1.5 rounded-lg bg-cyan-500/10 text-cyan-400 border border-cyan-500/20">
                        <svg className="w-4 h-4" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
                          <circle cx="12" cy="12" r="10" />
                          <path strokeLinecap="round" strokeLinejoin="round" d="M2 12h20M12 2a15.3 15.3 0 014 10 15.3 15.3 0 01-4 10 15.3 15.3 0 01-4-10 15.3 15.3 0 014-10z" />
                        </svg>
                      </div>
                    </div>
                    <h3 className="text-white font-bold text-sm tracking-wide group-hover:text-cyan-300 transition-colors">
                      Permintaan Masuk
                    </h3>
                    <div className="bg-slate-900/90 p-2.5 rounded-lg border border-cyan-950/60 mt-2.5">
                      <p className="text-slate-200 text-xs font-medium leading-relaxed">
                        Ekstraksi metadata HTTP: IP klien, User-Agent, Referer, dan resolusi GeoIP negara.
                      </p>
                    </div>
                  </div>
                  <div className="mt-3.5 pt-2.5 border-t border-slate-800/80 flex items-center justify-between text-[11px]">
                    <span className="text-cyan-300 font-medium">Inbound Ingest</span>
                    <span className="text-slate-400 font-mono">Layer 7</span>
                  </div>
                </div>

                {/* Step 2: Analyzer (Vibrant Violet) */}
                <div className="bg-slate-950/90 border border-slate-800 hover:border-violet-400/70 transition-all duration-200 rounded-xl p-4 flex flex-col justify-between shadow-lg relative group">
                  <div className="absolute top-0 left-0 right-0 h-0.5 bg-gradient-to-r from-violet-500 to-purple-500 rounded-t-xl"></div>
                  <div>
                    <div className="flex items-center justify-between mb-3">
                      <span className="px-2 py-0.5 rounded bg-violet-500/15 text-violet-300 text-[11px] font-bold font-mono border border-violet-500/30">
                        STEP 02
                      </span>
                      <div className="p-1.5 rounded-lg bg-violet-500/10 text-violet-400 border border-violet-500/20">
                        <svg className="w-4 h-4" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
                          <rect x="4" y="4" width="16" height="16" rx="2" />
                          <rect x="9" y="9" width="6" height="6" />
                          <path strokeLinecap="round" strokeLinejoin="round" d="M9 1v3m6-3v3M9 20v3m6-3v3M20 9h3m-3 6h3M1 9h3m-3 6h3" />
                        </svg>
                      </div>
                    </div>
                    <h3 className="text-white font-bold text-sm tracking-wide group-hover:text-violet-300 transition-colors">
                      Request Analyzer
                    </h3>
                    <div className="bg-slate-900/90 p-2.5 rounded-lg border border-violet-950/60 mt-2.5">
                      <p className="text-slate-200 text-xs font-medium leading-relaxed">
                        Klasifikasi tipe perangkat (mobile, tablet, desktop) serta normalisasi jenis peramban.
                      </p>
                    </div>
                  </div>
                  <div className="mt-3.5 pt-2.5 border-t border-slate-800/80 flex items-center justify-between text-[11px]">
                    <span className="text-violet-300 font-medium">Context Normalizer</span>
                    <span className="text-slate-400 font-mono">Regex Engine</span>
                  </div>
                </div>

                {/* Step 3: Rule Engine (Warm Amber / Gold) */}
                <div className="bg-slate-950/90 border border-slate-800 hover:border-amber-400/70 transition-all duration-200 rounded-xl p-4 flex flex-col justify-between shadow-lg relative group">
                  <div className="absolute top-0 left-0 right-0 h-0.5 bg-gradient-to-r from-amber-500 to-orange-500 rounded-t-xl"></div>
                  <div>
                    <div className="flex items-center justify-between mb-3">
                      <span className="px-2 py-0.5 rounded bg-amber-500/15 text-amber-300 text-[11px] font-bold font-mono border border-amber-500/30">
                        STEP 03
                      </span>
                      <div className="p-1.5 rounded-lg bg-amber-500/10 text-amber-400 border border-amber-500/20">
                        <svg className="w-4 h-4" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" d="M3.75 13.5l10.5-11.25L12 10.5h8.25L9.75 21.75 12 13.5H3.75z" />
                        </svg>
                      </div>
                    </div>
                    <h3 className="text-white font-bold text-sm tracking-wide group-hover:text-amber-300 transition-colors">
                      Rule Engine
                    </h3>
                    <div className="bg-slate-900/90 p-2.5 rounded-lg border border-amber-950/60 mt-2.5">
                      <p className="text-slate-200 text-xs font-medium leading-relaxed">
                        Evaluasi multi-kondisi berurutan dari prioritas tertinggi ke terendah secara deterministik.
                      </p>
                    </div>
                  </div>
                  <div className="mt-3.5 pt-2.5 border-t border-slate-800/80 flex items-center justify-between text-[11px]">
                    <span className="text-amber-300 font-medium">Priority Matching</span>
                    <span className="text-slate-400 font-mono">Skor 1 - 100</span>
                  </div>
                </div>

                {/* Step 4: Dispatcher (Vivid Rose / Coral) */}
                <div className="bg-slate-950/90 border border-slate-800 hover:border-rose-400/70 transition-all duration-200 rounded-xl p-4 flex flex-col justify-between shadow-lg relative group">
                  <div className="absolute top-0 left-0 right-0 h-0.5 bg-gradient-to-r from-rose-500 to-pink-500 rounded-t-xl"></div>
                  <div>
                    <div className="flex items-center justify-between mb-3">
                      <span className="px-2 py-0.5 rounded bg-rose-500/15 text-rose-300 text-[11px] font-bold font-mono border border-rose-500/30">
                        STEP 04
                      </span>
                      <div className="p-1.5 rounded-lg bg-rose-500/10 text-rose-400 border border-rose-500/20">
                        <svg className="w-4 h-4" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" d="M7.5 7.5h-.75A2.25 2.25 0 004.5 9.75v7.5a2.25 2.25 0 002.25 2.25h7.5a2.25 2.25 0 002.25-2.25v-.75m0-3.75l4.5 4.5m0-4.5l-4.5 4.5M12 3v9m0 0l3-3m-3 3L9 9" />
                        </svg>
                      </div>
                    </div>
                    <h3 className="text-white font-bold text-sm tracking-wide group-hover:text-rose-300 transition-colors">
                      Penentuan Tujuan
                    </h3>
                    <div className="bg-slate-900/90 p-2.5 rounded-lg border border-rose-950/60 mt-2.5">
                      <p className="text-slate-200 text-xs font-medium leading-relaxed">
                        Memilih URL halaman tujuan aturan atau mengarahkan ke fallback default jika tidak ada kecocokan.
                      </p>
                    </div>
                  </div>
                  <div className="mt-3.5 pt-2.5 border-t border-slate-800/80 flex items-center justify-between text-[11px]">
                    <span className="text-rose-300 font-medium">Target / Fallback</span>
                    <span className="text-slate-400 font-mono">Routing Matrix</span>
                  </div>
                </div>

                {/* Step 5: Logging (Emerald Mint) */}
                <div className="bg-slate-950/90 border border-slate-800 hover:border-emerald-400/70 transition-all duration-200 rounded-xl p-4 flex flex-col justify-between shadow-lg relative group">
                  <div className="absolute top-0 left-0 right-0 h-0.5 bg-gradient-to-r from-emerald-500 to-teal-500 rounded-t-xl"></div>
                  <div>
                    <div className="flex items-center justify-between mb-3">
                      <span className="px-2 py-0.5 rounded bg-emerald-500/15 text-emerald-300 text-[11px] font-bold font-mono border border-emerald-500/30">
                        STEP 05
                      </span>
                      <div className="p-1.5 rounded-lg bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                        <svg className="w-4 h-4" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
                          <ellipse cx="12" cy="5" rx="9" ry="3" />
                          <path strokeLinecap="round" strokeLinejoin="round" d="M21 12c0 1.66-4 3-9 3s-9-1.34-9-3" />
                          <path strokeLinecap="round" strokeLinejoin="round" d="M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5" />
                        </svg>
                      </div>
                    </div>
                    <h3 className="text-white font-bold text-sm tracking-wide group-hover:text-emerald-300 transition-colors">
                      Traffic Logging
                    </h3>
                    <div className="bg-slate-900/90 p-2.5 rounded-lg border border-emerald-950/60 mt-2.5">
                      <p className="text-slate-200 text-xs font-medium leading-relaxed">
                        Penyimpanan seluruh jejak audit evaluasi secara real-time ke tabel PostgreSQL traffic_logs.
                      </p>
                    </div>
                  </div>
                  <div className="mt-3.5 pt-2.5 border-t border-slate-800/80 flex items-center justify-between text-[11px]">
                    <span className="text-emerald-300 font-medium">Audit Trail</span>
                    <span className="text-slate-400 font-mono">ACID Logged</span>
                  </div>
                </div>
              </div>

              {/* Technical Specifications Ribbon */}
              <div className="mt-6 pt-5 border-t border-slate-800/80 grid grid-cols-2 sm:grid-cols-4 gap-3 text-xs">
                <div className="flex items-center gap-2 text-slate-300">
                  <div className="p-1 rounded bg-cyan-500/10 text-cyan-400 border border-cyan-500/20">
                    <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" d="M13 10V3L4 14h7v7l9-11h-7z" />
                    </svg>
                  </div>
                  <span className="text-cyan-400 font-bold">Latensi:</span>
                  <span className="font-mono text-white">&lt; 2ms Overhead</span>
                </div>
                <div className="flex items-center gap-2 text-slate-300">
                  <div className="p-1 rounded bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                    <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" d="M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                  </div>
                  <span className="text-emerald-400 font-bold">Kepatuhan:</span>
                  <span className="text-white">AdSense Safe Guard</span>
                </div>
                <div className="flex items-center gap-2 text-slate-300">
                  <div className="p-1 rounded bg-amber-500/10 text-amber-400 border border-amber-500/20">
                    <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0l3.181 3.183a8.25 8.25 0 0013.803-3.7M4.031 9.865a8.25 8.25 0 0113.803-3.7l3.181 3.182m0-4.991v4.99" />
                    </svg>
                  </div>
                  <span className="text-amber-400 font-bold">Fallback:</span>
                  <span className="text-white">Graceful Routing</span>
                </div>
                <div className="flex items-center gap-2 text-slate-300">
                  <div className="p-1 rounded bg-violet-500/10 text-violet-400 border border-violet-500/20">
                    <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" d="M20.25 6.375c0 2.278-3.694 4.125-8.25 4.125S3.75 8.653 3.75 6.375m16.5 0c0-2.278-3.694-4.125-8.25-4.125S3.75 4.097 3.75 6.375m16.5 0v11.25c0 2.278-3.694 4.125-8.25 4.125s-8.25-1.847-8.25-4.125V6.375m16.5 5.625c0 2.278-3.694 4.125-8.25 4.125s-8.25-1.847-8.25-4.125" />
                    </svg>
                  </div>
                  <span className="text-violet-400 font-bold">Audit Sink:</span>
                  <span className="font-mono text-white">PostgreSQL 100%</span>
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
                  <label className="text-xs font-semibold text-slate-300 block mb-1">
                    URL / Alamat Website Tujuan
                  </label>
                  <input
                    type="text"
                    required
                    placeholder="Contoh: https://google.com atau https://tokopedia.com"
                    value={formTujuan.url}
                    onChange={(e) => setFormTujuan({ ...formTujuan, url: e.target.value })}
                    className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3.5 py-2 text-sm text-white focus:outline-none focus:border-indigo-500 font-mono"
                  />
                  <span className="text-[11px] text-slate-400 mt-1 block">
                    Bisa berupa URL website eksternal mana pun (otomatis diawali https:// jika belum ada).
                  </span>
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
                      <th className="py-3 px-4">URL Tujuan Nyata</th>
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
                          <td className="py-3 px-4 font-mono text-cyan-300">
                            <a
                              href={t.url.startsWith("http") ? t.url : `https://${t.url}`}
                              target="_blank"
                              rel="noopener noreferrer"
                              className="hover:underline inline-flex items-center gap-1.5 text-cyan-300 hover:text-cyan-200"
                              title="Buka Website Tujuan Langsung"
                            >
                              <span className="truncate max-w-[200px] sm:max-w-[280px]">{t.url}</span>
                              <svg className="w-3.5 h-3.5 text-slate-400 group-hover:text-white flex-shrink-0" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" d="M13.5 6H5.25A2.25 2.25 0 003 8.25v10.5A2.25 2.25 0 005.25 21h10.5A2.25 2.25 0 0018 18.75V10.5m-10.5 6L21 3m0 0h-5.25M21 3v5.25" />
                              </svg>
                            </a>
                          </td>
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
                        <code className="text-cyan-300 font-mono font-bold text-sm bg-cyan-950/40 border border-cyan-800/40 px-2 py-1 rounded">
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

                    {/* Tombol Aksi Nyata (Bukan Cuma Simulasi) */}
                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-2.5 pt-2">
                      <a
                        href={hasilRouting.url_tujuan.startsWith("http") ? hasilRouting.url_tujuan : `https://${hasilRouting.url_tujuan}`}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="py-2.5 px-3.5 rounded-xl bg-gradient-to-r from-emerald-600 to-teal-600 hover:from-emerald-500 hover:to-teal-500 text-white font-bold text-xs flex items-center justify-center gap-2 transition-all shadow-md shadow-emerald-600/25"
                      >
                        <span>Buka Website Nyata ↗</span>
                      </a>
                      <a
                        href={`/r?negara=${encodeURIComponent(formUji.negara)}&perangkat=${encodeURIComponent(formUji.perangkat)}&peramban=${encodeURIComponent(formUji.peramban)}`}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="py-2.5 px-3.5 rounded-xl bg-slate-800 hover:bg-slate-700 text-cyan-300 font-bold text-xs flex items-center justify-center gap-2 border border-cyan-500/20 transition-all"
                      >
                        <span>Uji Pengalihan Nyata (/r) ↗</span>
                      </a>
                    </div>
                  </div>
                ) : (
                  <div className="py-16 text-center text-slate-500 text-xs">
                    Isi parameter simulasi di sebelah kiri dan klik tombol <strong>Periksa Trafik</strong> untuk melihat hasil.
                  </div>
                )}
              </div>

              {/* Tautan Gateway Routing Nyata */}
              <div className="mt-6 pt-5 border-t border-slate-800/80">
                <div className="p-4 rounded-xl bg-slate-950/80 border border-slate-800/80 space-y-2.5">
                  <div className="flex items-center justify-between">
                    <span className="text-xs font-bold text-white uppercase tracking-wider flex items-center gap-2">
                      <span className="h-2 w-2 rounded-full bg-emerald-400 animate-pulse"></span>
                      Tautan Gateway Routing Nyata (Live Router)
                    </span>
                    <span className="text-[10px] font-mono px-2 py-0.5 rounded bg-indigo-500/15 text-indigo-300 border border-indigo-500/30">
                      HTTP 302 REDIRECT
                    </span>
                  </div>
                  <p className="text-slate-300 text-xs leading-relaxed">
                    Bagikan tautan ini ke pengunjung atau gunakan di website Anda. Setiap pengunjung riil yang membuka tautan ini akan langsung dianalisis dan dialihkan ke website tujuan yang cocok secara otomatis:
                  </p>
                  <div className="flex items-center gap-2">
                    <input
                      type="text"
                      readOnly
                      value={typeof window !== "undefined" ? `${window.location.origin}/r` : "/r"}
                      className="w-full bg-slate-900 border border-slate-700/80 rounded-lg px-3 py-1.5 text-xs text-cyan-300 font-mono select-all focus:outline-none"
                    />
                    <button
                      type="button"
                      onClick={() => {
                        const url = typeof window !== "undefined" ? `${window.location.origin}/r` : "/r";
                        navigator.clipboard.writeText(url);
                        tampilkanNotifikasi("sukses", "Tautan gateway live berhasil disalin ke clipboard!");
                      }}
                      className="px-3 py-1.5 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-semibold whitespace-nowrap transition cursor-pointer"
                    >
                      Salin Link
                    </button>
                    <a
                      href="/r"
                      target="_blank"
                      rel="noopener noreferrer"
                      className="px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs font-semibold whitespace-nowrap transition"
                    >
                      Buka Gateway ↗
                    </a>
                  </div>
                </div>
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
      <footer className="border-t border-slate-800 bg-slate-950/80 py-8 mt-12">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 flex flex-col md:flex-row items-center justify-between gap-4 text-xs text-slate-300">
          <div className="flex items-center gap-3">
            <span className="flex items-center gap-2 font-semibold text-white">
              <span className="h-2 w-2 rounded-full bg-emerald-400 animate-pulse"></span>
              Sivilize Traffic Platform
            </span>
            <span className="text-slate-600">|</span>
            <span className="text-slate-300 font-medium">Enterprise Rule-based Routing & Compliance Demo</span>
          </div>

          <div className="flex flex-wrap items-center justify-center gap-4 text-slate-300 font-medium">
            <Link href="/compliance" className="text-emerald-400 hover:text-emerald-300 transition-colors flex items-center gap-1.5">
              <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" d="M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <span>Kebijakan Kepatuhan AdSense</span>
            </Link>
            <span className="text-slate-600">•</span>
            <a
              href="https://github.com/muhamadadrian210-debug/sivilize-traffic-routing-demo"
              target="_blank"
              rel="noopener noreferrer"
              className="text-slate-300 hover:text-white transition-colors"
            >
              GitHub Repository
            </a>
            <span className="text-slate-600">•</span>
            <span className="text-slate-400">&copy; {new Date().getFullYear()} Sivilize Corp. All rights reserved.</span>
          </div>
        </div>
      </footer>
    </div>
  );
}
