"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { dapatkanUrlApi } from "@/pustaka/api";
import { HasilComplianceCheck, ItemChecklist, AuditRiwayat } from "@/tipe/compliance";

export default function HalamanKepatuhan() {
  const urlBackend = dapatkanUrlApi();

  const [hasilCheck, setHasilCheck] = useState<HasilComplianceCheck | null>(null);
  const [checklist, setChecklist] = useState<ItemChecklist[]>([]);
  const [riwayatAudit, setRiwayatAudit] = useState<AuditRiwayat[]>([]);
  const [memuatAudit, setMemuatAudit] = useState<boolean>(true);
  const [memuatChecklist, setMemuatChecklist] = useState<boolean>(true);
  const [memuatSimpanItem, setMemuatSimpanItem] = useState<string | null>(null);
  const [notifikasi, setNotifikasi] = useState<{ tipe: "sukses" | "error"; pesan: string } | null>(null);

  function tampilkanNotifikasi(tipe: "sukses" | "error", pesan: string) {
    setNotifikasi({ tipe, pesan });
    setTimeout(() => setNotifikasi(null), 4000);
  }

  // 1. Ambil Hasil Pemeriksaan Kepatuhan
  async function jalankanPemeriksaan() {
    try {
      const res = await fetch(`${urlBackend}/api/compliance/check`);
      if (res.ok) {
        const data: HasilComplianceCheck = await res.json();
        setHasilCheck(data);
        tampilkanNotifikasi("sukses", "Audit kepatuhan internal berhasil dijalankan");
      } else {
        tampilkanNotifikasi("error", "Gagal menjalankan pemeriksaan kepatuhan");
      }
    } catch {
      tampilkanNotifikasi("error", "Tidak dapat terhubung ke server backend");
    } finally {
      setMemuatAudit(false);
    }
  }

  // 2. Ambil Data Checklist Manual
  async function muatChecklist() {
    setMemuatChecklist(true);
    try {
      const res = await fetch(`${urlBackend}/api/compliance/checklist`);
      if (res.ok) {
        const json = await res.json();
        setChecklist(json.data || []);
      }
    } catch {
      // diamkan, notifikasi global
    } finally {
      setMemuatChecklist(false);
    }
  }

  // 3. Ambil Riwayat Audit
  async function muatRiwayatAudit() {
    try {
      const res = await fetch(`${urlBackend}/api/compliance/audits`);
      if (res.ok) {
        const json = await res.json();
        setRiwayatAudit(json.data || []);
      }
    } catch {
      // diamkan
    }
  }

  // 4. Toggle Item Checklist Manual
  async function ubahStatusItem(kunci: string, selesaiSekarang: boolean) {
    setMemuatSimpanItem(kunci);
    try {
      const res = await fetch(`${urlBackend}/api/compliance/checklist`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ kunci, selesai: !selesaiSekarang }),
      });
      if (res.ok) {
        const json = await res.json();
        setChecklist(json.data || []);
        tampilkanNotifikasi("sukses", "Status checklist berhasil disimpan");
        // Segarkan hasil audit kepatuhan
        jalankanPemeriksaan();
        muatRiwayatAudit();
      } else {
        tampilkanNotifikasi("error", "Gagal memperbarui checklist");
      }
    } catch {
      tampilkanNotifikasi("error", "Koneksi ke backend gagal");
    } finally {
      setMemuatSimpanItem(null);
    }
  }

  useEffect(() => {
    void jalankanPemeriksaan();
    void muatChecklist();
    void muatRiwayatAudit();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const totalSelesai = checklist.filter((c) => c.selesai).length;
  const persentaseSelesai = checklist.length > 0 ? Math.round((totalSelesai / checklist.length) * 100) : 0;

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 flex flex-col font-sans selection:bg-indigo-500 selection:text-white pb-16">
      {/* 1. Header Halaman */}
      <header className="border-b border-slate-800 bg-slate-900/60 backdrop-blur-md sticky top-0 z-40">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 py-4 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
          <div className="flex items-center gap-3">
            <Link
              href="/"
              className="p-2 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-300 hover:text-white transition-colors text-xs font-medium flex items-center gap-1.5"
            >
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M10 19l-7-7m0 0l7-7m-7 7h18"></path>
              </svg>
              <span>Dashboard Utama</span>
            </Link>
            <div>
              <div className="flex items-center gap-2">
                <div className="p-1.5 rounded-lg bg-emerald-500/15 text-emerald-400 border border-emerald-500/30">
                  <svg className="w-5 h-5" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" d="M9 12.75L11.25 15 15 9.75m-3-7.036A11.959 11.959 0 013.598 6 11.99 11.99 0 003 9.749c0 5.592 3.824 10.29 9 11.623 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.248-8.25-3.285z" />
                  </svg>
                </div>
                <h1 className="text-lg sm:text-xl font-bold tracking-tight text-white">
                  Audit Kepatuhan Google Publisher & AdSense
                </h1>
                <span className="text-[10px] uppercase font-bold tracking-wider px-2 py-0.5 rounded-full bg-emerald-500/20 text-emerald-400 border border-emerald-500/30">
                  Batch 8
                </span>
              </div>
              <p className="text-xs text-slate-400 mt-0.5">
                Pemeriksaan internal konfigurasi traffic routing terhadap pedoman Google Publisher Policy.
              </p>
            </div>
          </div>

          <div className="flex items-center gap-3 w-full sm:w-auto justify-end">
            <button
              onClick={() => {
                jalankanPemeriksaan();
                muatChecklist();
                muatRiwayatAudit();
              }}
              disabled={memuatAudit}
              className="px-3.5 py-2 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-medium transition-all shadow-lg shadow-indigo-600/25 flex items-center gap-2 disabled:opacity-50 cursor-pointer"
            >
              <svg
                className={`w-4 h-4 ${memuatAudit ? "animate-spin" : ""}`}
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth="2"
                  d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
                ></path>
              </svg>
              <span>{memuatAudit ? "Memeriksa..." : "Jalankan Audit Ulang"}</span>
            </button>
          </div>
        </div>
      </header>

      {/* 2. Banner Notifikasi */}
      {notifikasi && (
        <div
          className={`fixed bottom-5 right-5 z-50 px-4 py-3 rounded-xl shadow-2xl border flex items-center gap-3 text-sm transition-all animate-bounce ${
            notifikasi.tipe === "sukses"
              ? "bg-emerald-950 text-emerald-200 border-emerald-700"
              : "bg-rose-950 text-rose-200 border-rose-700"
          }`}
        >
          <svg className="w-4 h-4" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
            {notifikasi.tipe === "sukses" ? (
              <path strokeLinecap="round" strokeLinejoin="round" d="M4.5 12.75l6 6 9-13.5" />
            ) : (
              <path strokeLinecap="round" strokeLinejoin="round" d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z" />
            )}
          </svg>
          <span>{notifikasi.pesan}</span>
        </div>
      )}

      {/* 3. Konten Utama */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 py-6 w-full flex flex-col gap-6">
        {/* Banner Penjelasan & Disclaimer Resmi */}
        <div className="p-4 rounded-2xl bg-amber-950/40 border border-amber-800/60 text-amber-200 text-xs sm:text-sm flex items-start gap-3">
          <div className="p-1 rounded-lg bg-amber-500/20 text-amber-400 border border-amber-500/30 flex-shrink-0 mt-0.5">
            <svg className="w-4 h-4" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126zM12 15.75h.007v.008H12v-.008z" />
            </svg>
          </div>
          <div>
            <strong className="font-semibold block mb-0.5">Penafian Kepatuhan Resmi (Disclaimer):</strong>
            <p className="text-amber-300/90 leading-relaxed">
              &quot;Status ini merupakan pemeriksaan internal berdasarkan dokumentasi kebijakan Google yang tersedia dan bukan keputusan resmi Google.&quot;
            </p>
            <p className="text-amber-400/80 text-[11px] mt-1.5 leading-relaxed">
              Sistem ini adalah demo traffic routing generik untuk kebutuhan bisnis yang sah. Sistem tidak terafiliasi dengan Google, tidak menyediakan jaminan sertifikasi resmi (&quot;Google Approved&quot; / &quot;100% AdSense Compliant&quot;), dan tidak boleh digunakan untuk manipulasi atau penghindaran peninjauan iklan.
            </p>
          </div>
        </div>

        {/* 4. Kartu Status Pemeriksaan Terakhir & Metrik */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          {/* Status Utama */}
          <div className="md:col-span-2 p-5 rounded-2xl bg-slate-900/70 border border-slate-800 flex flex-col justify-between">
            <div>
              <span className="text-xs uppercase tracking-wider font-semibold text-slate-400 block mb-2">
                Status Pemeriksaan Terakhir
              </span>
              <div className="flex items-center gap-3">
                {hasilCheck?.status === "aman" && (
                  <span className="px-3.5 py-1.5 rounded-xl bg-emerald-500/20 text-emerald-400 border border-emerald-500/30 text-sm font-bold flex items-center gap-2">
                    <span className="w-2.5 h-2.5 rounded-full bg-emerald-400"></span>
                    AMAN / TIDAK TERINDIKASI
                  </span>
                )}
                {hasilCheck?.status === "perlu_ditinjau" && (
                  <span className="px-3.5 py-1.5 rounded-xl bg-amber-500/20 text-amber-400 border border-amber-500/30 text-sm font-bold flex items-center gap-2">
                    <span className="w-2.5 h-2.5 rounded-full bg-amber-400 animate-pulse"></span>
                    PERLU DITINJAU
                  </span>
                )}
                {hasilCheck?.status === "berisiko" && (
                  <span className="px-3.5 py-1.5 rounded-xl bg-rose-500/20 text-rose-400 border border-rose-500/30 text-sm font-bold flex items-center gap-2">
                    <span className="w-2.5 h-2.5 rounded-full bg-rose-400 animate-ping"></span>
                    BERISIKO TINGGI
                  </span>
                )}
                {!hasilCheck && <span className="text-xs text-slate-500">Memuat hasil audit...</span>}
              </div>
            </div>

            <div className="mt-4 pt-4 border-t border-slate-800/80 flex flex-wrap items-center justify-between gap-2 text-xs text-slate-400">
              <div>
                Waktu Audit:{" "}
                <span className="text-slate-300 font-mono">
                  {hasilCheck?.waktu_pemeriksaan
                    ? new Date(hasilCheck.waktu_pemeriksaan).toLocaleString("id-ID")
                    : "-"}
                </span>
              </div>
              <div className="text-[11px] text-slate-400">
                Skor Kepatuhan: <span className="text-slate-400 italic">null (tidak ada skor numerik)</span>
              </div>
            </div>
          </div>

          {/* Temuan Peringatan */}
          <div className="p-5 rounded-2xl bg-slate-900/70 border border-slate-800 flex flex-col justify-between">
            <span className="text-xs uppercase tracking-wider font-semibold text-amber-400">Peringatan</span>
            <div className="text-3xl font-extrabold text-amber-300 mt-2">
              {hasilCheck?.ringkasan.peringatan ?? 0}
            </div>
            <p className="text-[11px] text-slate-400 mt-2">Konfigurasi yang memerlukan peninjauan operasional berkala.</p>
          </div>

          {/* Temuan Risiko Tinggi */}
          <div className="p-5 rounded-2xl bg-slate-900/70 border border-slate-800 flex flex-col justify-between">
            <span className="text-xs uppercase tracking-wider font-semibold text-rose-400">Risiko Tinggi</span>
            <div className="text-3xl font-extrabold text-rose-300 mt-2">
              {hasilCheck?.ringkasan.risiko_tinggi ?? 0}
            </div>
            <p className="text-[11px] text-slate-400 mt-2">Indikasi crawler targeting atau invalid traffic sources.</p>
          </div>
        </div>

        {/* 5. Detail Temuan Kepatuhan */}
        <section className="p-6 rounded-2xl bg-slate-900/60 border border-slate-800">
          <div className="flex items-center justify-between mb-4">
            <div>
              <h2 className="text-base font-bold text-white flex items-center gap-2">
                <span>📋</span>
                <span>Detail Temuan Kepatuhan ({hasilCheck?.ringkasan.total ?? 0})</span>
              </h2>
              <p className="text-xs text-slate-400 mt-0.5">
                Daftar temuan hasil verifikasi terhadap seluruh aturan, tujuan, dan konfigurasi routing.
              </p>
            </div>
          </div>

          {(!hasilCheck?.temuan || hasilCheck.temuan.length === 0) ? (
            <div className="p-8 rounded-xl bg-emerald-950/20 border border-emerald-800/40 text-center">
              <span className="text-3xl block mb-2">✅</span>
              <p className="text-sm font-semibold text-emerald-300">Tidak ada temuan pelanggaran terdeteksi.</p>
              <p className="text-xs text-slate-400 mt-1">
                Seluruh konfigurasi routing saat ini berada dalam batas normal dan tidak menargetkan crawler/reviewer.
              </p>
            </div>
          ) : (
            <div className="grid grid-cols-1 gap-3">
              {hasilCheck.temuan.map((t, idx) => (
                <div
                  key={idx}
                  className={`p-4 rounded-xl border flex flex-col sm:flex-row items-start gap-4 transition-all ${
                    t.tingkat === "risiko tinggi"
                      ? "bg-rose-950/30 border-rose-800/60"
                      : t.tingkat === "peringatan"
                      ? "bg-amber-950/30 border-amber-800/60"
                      : "bg-blue-950/30 border-blue-800/60"
                  }`}
                >
                  <div className="flex-shrink-0 pt-0.5">
                    {t.tingkat === "risiko tinggi" && <span className="text-xl">⛔</span>}
                    {t.tingkat === "peringatan" && <span className="text-xl">⚠️</span>}
                    {t.tingkat === "informasi" && <span className="text-xl">ℹ️</span>}
                  </div>

                  <div className="flex-1 min-w-0">
                    <div className="flex flex-wrap items-center gap-2 mb-1">
                      <span
                        className={`text-[10px] uppercase font-bold tracking-wider px-2 py-0.5 rounded-md ${
                          t.tingkat === "risiko tinggi"
                            ? "bg-rose-500/20 text-rose-300 border border-rose-500/30"
                            : t.tingkat === "peringatan"
                            ? "bg-amber-500/20 text-amber-300 border border-amber-500/30"
                            : "bg-blue-500/20 text-blue-300 border border-blue-500/30"
                        }`}
                      >
                        {t.tingkat}
                      </span>
                      <code className="text-xs font-mono text-slate-400 bg-slate-800/60 px-2 py-0.5 rounded">
                        {t.kode}
                      </code>
                    </div>
                    <h3 className="text-sm font-semibold text-slate-100">{t.judul}</h3>
                    <p className="text-xs text-slate-300 mt-1 leading-relaxed">{t.pesan}</p>
                  </div>
                </div>
              ))}
            </div>
          )}
        </section>

        {/* 6. Checklist Manual Kepatuhan (11 Item) */}
        <section className="p-6 rounded-2xl bg-slate-900/60 border border-slate-800">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 mb-4">
            <div>
              <h2 className="text-base font-bold text-white flex items-center gap-2">
                <span>☑️</span>
                <span>Checklist Kepatuhan Mandiri Administrator</span>
              </h2>
              <p className="text-xs text-slate-400 mt-0.5">
                Konfirmasi operasional manual yang wajib dipenuhi publisher sesuai kebijakan Google AdSense.
              </p>
            </div>
            <div className="text-xs font-medium text-slate-300 flex items-center gap-2 bg-slate-800/60 px-3 py-1.5 rounded-xl border border-slate-700/60">
              <span>Progres:</span>
              <strong className="text-indigo-400 font-bold">{totalSelesai} / {checklist.length}</strong>
              <span>({persentaseSelesai}%)</span>
            </div>
          </div>

          {/* Progress Bar */}
          <div className="w-full bg-slate-800 h-2 rounded-full overflow-hidden mb-5">
            <div
              className="bg-indigo-500 h-full transition-all duration-500"
              style={{ width: `${persentaseSelesai}%` }}
            ></div>
          </div>

          {memuatChecklist ? (
            <p className="text-xs text-slate-500 py-4">Memuat data checklist dari database...</p>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
              {checklist.map((item) => (
                <label
                  key={item.kunci}
                  className={`p-3.5 rounded-xl border flex items-start gap-3 cursor-pointer transition-all ${
                    item.selesai
                      ? "bg-slate-900/80 border-slate-700/70 text-slate-200"
                      : "bg-slate-950/50 border-slate-800 text-slate-400 hover:border-slate-700"
                  }`}
                >
                  <input
                    type="checkbox"
                    checked={item.selesai}
                    disabled={memuatSimpanItem === item.kunci}
                    onChange={() => ubahStatusItem(item.kunci, item.selesai)}
                    className="mt-0.5 w-4 h-4 rounded text-indigo-600 bg-slate-800 border-slate-600 focus:ring-indigo-500 focus:ring-offset-slate-900 cursor-pointer"
                  />
                  <div className="flex-1 min-w-0">
                    <span className={`text-xs font-medium ${item.selesai ? "text-slate-100" : "text-slate-300"}`}>
                      {item.label}
                    </span>
                    <div className="text-[10px] text-slate-400 mt-0.5">
                      Kunci: <code className="font-mono text-slate-400">{item.kunci}</code>
                    </div>
                  </div>
                  {memuatSimpanItem === item.kunci && (
                    <span className="text-[10px] text-indigo-400 animate-pulse">Menyimpan...</span>
                  )}
                </label>
              ))}
            </div>
          )}
        </section>

        {/* 7. Panduan Kebijakan & Rujukan Dokumentasi Resmi */}
        <section className="p-6 rounded-2xl bg-slate-900/60 border border-slate-800 flex flex-col gap-5">
          <div>
            <h2 className="text-base font-bold text-white flex items-center gap-2">
              <span>📖</span>
              <span>Dokumentasi Kebijakan Google Publisher & AdSense</span>
            </h2>
            <p className="text-xs text-slate-400 mt-0.5">
              Rangkuman aspek kepatuhan kritis yang wajib dihormati pada implementasi traffic routing.
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-xs">
            {/* A. Invalid Traffic */}
            <div className="p-4 rounded-xl bg-slate-950/60 border border-slate-800/80">
              <h3 className="font-bold text-indigo-300 text-sm mb-2 flex items-center gap-1.5">
                <span>🛑</span> A. Invalid Traffic (Lalu Lintas Tidak Valid)
              </h3>
              <ul className="list-disc list-inside space-y-1 text-slate-300 leading-relaxed">
                <li>Publisher dilarang mengklik iklannya sendiri.</li>
                <li>Dilarang menghasilkan tayangan atau klik iklan secara artifisial.</li>
                <li>Dilarang menggunakan bot otomatis, autosurf, atau click-exchange.</li>
                <li>Lalu lintas berulang yang tidak wajar dapat memicu pembatasan akun.</li>
                <li>Sistem tidak boleh digunakan untuk memanipulasi atau menyembunyikan lalu lintas tidak sah.</li>
              </ul>
            </div>

            {/* B. Incentivized Traffic */}
            <div className="p-4 rounded-xl bg-slate-950/60 border border-slate-800/80">
              <h3 className="font-bold text-indigo-300 text-sm mb-2 flex items-center gap-1.5">
                <span>🎁</span> B. Incentivized Traffic (Lalu Lintas Berinsentif)
              </h3>
              <ul className="list-disc list-inside space-y-1 text-slate-300 leading-relaxed">
                <li>Dilarang memberikan hadiah, poin, atau uang agar pengguna mengklik iklan.</li>
                <li>Dilarang meminta pengguna mengklik iklan secara langsung maupun tidak langsung.</li>
                <li>Dilarang mengondisikan akses konten dengan kewajiban berinteraksi dengan iklan.</li>
                <li>Mekanisme engagement iklan harus sepenuhnya organik dan alami.</li>
              </ul>
            </div>

            {/* C. Ad Placement */}
            <div className="p-4 rounded-xl bg-slate-950/60 border border-slate-800/80">
              <h3 className="font-bold text-indigo-300 text-sm mb-2 flex items-center gap-1.5">
                <span>📐</span> C. Ad Placement (Penempatan Iklan)
              </h3>
              <ul className="list-disc list-inside space-y-1 text-slate-300 leading-relaxed">
                <li>Iklan dilarang menyerupai tombol navigasi situs atau menu.</li>
                <li>Iklan dilarang disamarkan sebagai tombol &quot;Download&quot; atau &quot;Play&quot;.</li>
                <li>Desain tidak boleh memicu accidental clicks (klik tidak disengaja).</li>
                <li>Dilarang menempatkan iklan pada pop-up, pop-under, atau software yang tidak disetujui.</li>
              </ul>
            </div>

            {/* D. Traffic Sources & Content Policy */}
            <div className="p-4 rounded-xl bg-slate-950/60 border border-slate-800/80">
              <h3 className="font-bold text-indigo-300 text-sm mb-2 flex items-center gap-1.5">
                <span>🌐</span> D. Traffic Sources & Kebijakan Konten
              </h3>
              <ul className="list-disc list-inside space-y-1 text-slate-300 leading-relaxed">
                <li>Kualitas traffic berbayar (paid traffic) wajib diverifikasi kemurniannya.</li>
                <li>Status indikator: AMAN, PERLU DITINJAU, atau BERISIKO.</li>
                <li>Konten wajib bebas dari materi ilegal, hak cipta tanpa izin, dan konten seksual eksplisit.</li>
                <li>Dilarang menggunakan routing untuk menyembunyikan konten dari crawler (anti-cloaking).</li>
              </ul>
            </div>
          </div>

          {/* Tautan Resmi Google */}
          <div className="p-4 rounded-xl bg-indigo-950/20 border border-indigo-800/40 mt-1">
            <h4 className="text-xs font-bold text-indigo-300 uppercase tracking-wider mb-2">
              Tautan Dokumentasi Resmi Google:
            </h4>
            <div className="flex flex-wrap gap-2 text-xs">
              <a
                href="https://support.google.com/adsense/answer/48182"
                target="_blank"
                rel="noreferrer"
                className="px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-indigo-300 hover:text-white transition-colors flex items-center gap-1.5"
              >
                <span>Kebijakan Program AdSense</span>
                <span className="text-[10px]">↗</span>
              </a>
              <a
                href="https://support.google.com/adsense/answer/2660562"
                target="_blank"
                rel="noreferrer"
                className="px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-indigo-300 hover:text-white transition-colors flex items-center gap-1.5"
              >
                <span>Panduan Kualitas Lalu Lintas Iklan</span>
                <span className="text-[10px]">↗</span>
              </a>
              <a
                href="https://support.google.com/adsense/answer/16737"
                target="_blank"
                rel="noreferrer"
                className="px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-indigo-300 hover:text-white transition-colors flex items-center gap-1.5"
              >
                <span>Panduan Penempatan Iklan</span>
                <span className="text-[10px]">↗</span>
              </a>
              <a
                href="https://support.google.com/publisherpolicies/answer/10502938"
                target="_blank"
                rel="noreferrer"
                className="px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-indigo-300 hover:text-white transition-colors flex items-center gap-1.5"
              >
                <span>Kebijakan Google Publisher</span>
                <span className="text-[10px]">↗</span>
              </a>
            </div>
          </div>
        </section>

        {/* 8. Log Riwayat Audit Kepatuhan */}
        <section className="p-6 rounded-2xl bg-slate-900/60 border border-slate-800">
          <h2 className="text-base font-bold text-white flex items-center gap-2 mb-3">
            <span>📜</span>
            <span>Riwayat Audit Kepatuhan</span>
          </h2>
          {riwayatAudit.length === 0 ? (
            <p className="text-xs text-slate-500">Belum ada riwayat audit tersimpan.</p>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-left text-xs text-slate-300">
                <thead className="bg-slate-800/60 text-slate-400 font-semibold border-b border-slate-700/60">
                  <tr>
                    <th className="py-2.5 px-3">Audit ID</th>
                    <th className="py-2.5 px-3">Status</th>
                    <th className="py-2.5 px-3">Jumlah Temuan</th>
                    <th className="py-2.5 px-3">Waktu Audit</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-800">
                  {riwayatAudit.map((a) => (
                    <tr key={a.id} className="hover:bg-slate-800/30">
                      <td className="py-2.5 px-3 font-mono text-slate-400">#{a.id}</td>
                      <td className="py-2.5 px-3">
                        <span
                          className={`text-[10px] font-bold px-2 py-0.5 rounded-full ${
                            a.status === "aman"
                              ? "bg-emerald-500/20 text-emerald-400 border border-emerald-500/30"
                              : a.status === "perlu_ditinjau"
                              ? "bg-amber-500/20 text-amber-400 border border-amber-500/30"
                              : "bg-rose-500/20 text-rose-400 border border-rose-500/30"
                          }`}
                        >
                          {a.status}
                        </span>
                      </td>
                      <td className="py-2.5 px-3">{a.temuan?.length ?? 0} temuan</td>
                      <td className="py-2.5 px-3 font-mono text-slate-400">
                        {new Date(a.dibuat_pada).toLocaleString("id-ID")}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </section>
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
            <span className="text-slate-300 font-medium">Compliance & Audit Module</span>
          </div>

          <div className="flex flex-wrap items-center justify-center gap-4 text-slate-300 font-medium">
            <Link href="/" className="text-indigo-400 hover:text-indigo-300 transition-colors">
              ← Kembali ke Dashboard Utama
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
