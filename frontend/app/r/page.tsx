"use client";

import { useEffect, useState } from "react";
import { dapatkanUrlApi } from "@/pustaka/api";

export default function HalamanPengalihanLangsung() {
  const [status, setStatus] = useState<"memuat" | "berhasil" | "gagal">("memuat");
  const [urlTujuan, setUrlTujuan] = useState<string>("");
  const [pesan, setPesan] = useState<string>("Menganalisis profil pengunjung & aturan routing...");

  useEffect(() => {
    async function prosesPengalihan() {
      try {
        const urlBackend = dapatkanUrlApi();
        const userAgent = navigator.userAgent;
        const referrer = document.referrer;

        // Ambil query params jika ada
        const urlParams = new URLSearchParams(window.location.search);
        const negara = urlParams.get("negara") || "";
        const perangkat = urlParams.get("perangkat") || "";
        const peramban = urlParams.get("peramban") || "";

        const res = await fetch(`${urlBackend}/api/periksa`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            negara,
            perangkat,
            peramban,
            agen_pengguna: userAgent,
            asal_rujukan: referrer,
          }),
        });

        if (res.ok) {
          const json = await res.json();
          let target = json.data?.url_tujuan || "/";
          if (!target.startsWith("http://") && !target.startsWith("https://") && !target.startsWith("/")) {
            target = "https://" + target;
          }

          setUrlTujuan(target);
          setStatus("berhasil");
          setPesan(`Mengalihkan ke ${target}...`);

          // Langsung alihkan peramban ke website tujuan
          setTimeout(() => {
            window.location.replace(target);
          }, 400);
        } else {
          setStatus("gagal");
          setPesan("Gagal mendapatkan keputusan routing dari server.");
        }
      } catch (err) {
        console.error("Kesalahan pengalihan:", err);
        setStatus("gagal");
        setPesan("Terjadi kendala saat menghubungi server routing.");
      }
    }

    void prosesPengalihan();
  }, []);

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 flex flex-col items-center justify-center p-6">
      <div className="max-w-md w-full bg-slate-900/90 border border-slate-800 rounded-3xl p-8 text-center shadow-2xl backdrop-blur relative overflow-hidden">
        <div className="absolute top-0 left-0 right-0 h-1 bg-gradient-to-r from-cyan-400 via-indigo-500 to-emerald-400"></div>

        <div className="w-16 h-16 rounded-2xl bg-indigo-500/10 border border-indigo-500/20 text-indigo-400 flex items-center justify-center mx-auto mb-5 shadow-inner">
          {status === "memuat" ? (
            <svg className="w-8 h-8 animate-spin" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0l3.181 3.183a8.25 8.25 0 0013.803-3.7M4.031 9.865a8.25 8.25 0 0113.803-3.7l3.181 3.182m0-4.991v4.99" />
            </svg>
          ) : status === "berhasil" ? (
            <svg className="w-8 h-8 text-emerald-400" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" d="M4.5 12.75l6 6 9-13.5" />
            </svg>
          ) : (
            <svg className="w-8 h-8 text-rose-400" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z" />
            </svg>
          )}
        </div>

        <h1 className="text-xl font-bold text-white mb-2">
          {status === "memuat" ? "Sivilize Gateway Router" : status === "berhasil" ? "Pengalihan Trafik Aktif" : "Pengalihan Terkendala"}
        </h1>

        <p className="text-slate-300 text-sm mb-6 leading-relaxed">
          {pesan}
        </p>

        {urlTujuan && (
          <div className="space-y-4">
            <div className="p-3.5 rounded-xl bg-slate-950 border border-slate-800 text-xs font-mono text-cyan-300 break-all">
              {urlTujuan}
            </div>
            <a
              href={urlTujuan}
              className="inline-flex items-center justify-center w-full py-3 px-4 rounded-xl bg-gradient-to-r from-indigo-600 to-violet-600 hover:from-indigo-500 hover:to-violet-500 text-white font-semibold text-sm transition-all shadow-lg shadow-indigo-600/30"
            >
              Klik Di Sini Jika Tidak Beralih Otomatis →
            </a>
          </div>
        )}

        {status === "gagal" && (
          <a
            href="/"
            className="inline-block py-2 px-4 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-300 text-xs font-medium transition"
          >
            ← Kembali ke Dasbor
          </a>
        )}
      </div>
    </div>
  );
}
