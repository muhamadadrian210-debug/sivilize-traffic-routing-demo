# Sivilize Traffic Routing Demo

## Deskripsi

**Sivilize Traffic Routing Demo** adalah aplikasi MVP (*Minimum Viable Product* / *Proof of Concept*) untuk sistem routing trafik berbasis aturan (*rule-based traffic routing*). Sistem ini menerima data konteks pengunjung (negara, jenis perangkat, peramban, User-Agent, referer), mencocokkannya dengan aturan berprioritas, menentukan halaman tujuan yang paling relevan atau tujuan cadangan (*fallback*), serta mencatat riwayat evaluasi ke PostgreSQL secara real-time.

> **Status Proyek**: DEMO / MVP (Proof of Concept), bukan sistem enterprise production langsung.  
> **Pemilik**: Sivilize Corp

---

## Arsitektur Sistem

```text
Pengunjung / Uji Simulasi
           │
           ▼
    Frontend Dashboard (Next.js)
           │
           ▼ (REST API / JSON)
     Backend Server (Golang Gin)
           │
           ▼
    Request Analyzer (Ekstraksi & Normalisasi Konteks)
           │
           ▼
    Rule Engine (Evaluasi Aturan Aktif & Prioritas)
      ┌────┴────┐
      ▼         ▼
    Cocok    Fallback (Default /demo/cadangan)
      │         │
      └────┬────┘
           ▼
    Traffic Logging (PostgreSQL: traffic_logs)
           │
           ▼
    Hasil Routing JSON
```

---

## Tech Stack

- **Frontend**: Next.js 16 (App Router), React 19, TypeScript, Tailwind CSS
- **Backend**: Golang 1.23, Gin Web Framework
- **Basis Data**: PostgreSQL 17
- **Komunikasi**: REST API (JSON)
- **Deployment**: Vercel (Frontend), Docker / Render / Railway (Backend), Managed PostgreSQL

---

## Fitur Utama

1. **Pemeriksaan Kesehatan (`/health` & `/api/kesehatan`)**: Memantau ketersediaan server dan koneksi PostgreSQL aktif.
2. **Manajemen Tujuan (`/api/tujuan`)**: CRUD halaman target trafik, lengkap dengan pencegahan penghapusan jika masih dijadikan referensi oleh aturan (`409 Conflict`).
3. **Manajemen Aturan (`/api/aturan`)**: CRUD aturan multi-kriteria (negara, perangkat, peramban, User-Agent, referer) dengan prioritas dan toggle aktif/nonaktif (`PATCH /api/aturan/:id/status`).
4. **Request Analyzer**: Mengklasifikasikan perangkat (`mobile`, `tablet`, `desktop`, `unknown`) dan peramban (`Edge`, `Chrome`, `Firefox`, `Safari`, `Lainnya`), dengan normalisasi string tanpa merusak nilai asli User-Agent.
5. **Rule Engine & Prioritas**: Evaluasi deterministik di mana angka prioritas lebih kecil menang (`1` > `5`), dengan penanganan fallback otomatis.
6. **Pencatatan Trafik (`traffic_logs`)**: Riwayat evaluasi trafik real-time tersimpan ke PostgreSQL dengan indeks performa.
7. **Paginasi Log (`/api/log-trafik`)**: Mengambil log terurut dari yang terbaru dengan paginasi (`?page=1&limit=20`, maks limit 100).
8. **Dasbor UI Interaktif**: Antarmuka visual untuk memantau status sistem, mengelola tujuan & aturan, mencoba simulasi perutean, dan memantau log.

---

## Environment Variables

### Backend (`backend/.env`)
```env
PORT=8080
FRONTEND_URL=http://localhost:3000
DATABASE_URL=postgres://pengguna:kata_sandi@localhost:5432/sivilize_trafik?sslmode=disable
FALLBACK_URL=/demo/cadangan
```

### Frontend (`frontend/.env.local`)
```env
NEXT_PUBLIC_API_URL=http://localhost:8080
```

> File `.env` asli tidak boleh di-commit ke Git. Gunakan `.env.example` sebagai acuan.

---

## Local Development

### 1. Menjalankan PostgreSQL
Jalankan skrip pembantu atau jalankan `pg_ctl`:
```powershell
.\jalankan-postgres.ps1
```

### 2. Menjalankan Migrasi & Data Seed
```powershell
cd backend
go run cmd/migrasi/main.go
go run cmd/seed/main.go
```

### 3. Menjalankan Backend
```powershell
cd backend
go run cmd/server/main.go
```
*Backend aktif pada `http://localhost:8080`.*

### 4. Menjalankan Frontend
```powershell
cd frontend
npm install
npm run dev
```
*Buka browser pada `http://localhost:3000`.*

---

## Dokumentasi API

Format Respons Standar:
- **Sukses**: `{"data": ...}`
- **Gagal**: `{"error": true, "pesan": "..."}`

### A. Health Check
- `GET /health` → `{"status": "ok", "database": "terhubung"}`
- `GET /api/kesehatan` → `{"status": "sehat", "database": "terhubung"}`

### B. Tujuan (`/api/tujuan`)
- `GET /api/tujuan`: Mengambil seluruh daftar tujuan.
- `GET /api/tujuan/:id`: Mengambil satu tujuan berdasarkan ID.
- `POST /api/tujuan`: Menambahkan tujuan baru (`nama`, `url`, `status`).
- `PUT /api/tujuan/:id`: Mengubah tujuan yang sudah ada.
- `DELETE /api/tujuan/:id`: Menghapus tujuan (menolak dengan `409` jika masih dipakai aturan).

### C. Aturan (`/api/aturan`)
- `GET /api/aturan`: Mengambil daftar aturan dengan JOIN nama & URL tujuan.
- `GET /api/aturan/:id`: Mengambil satu aturan berdasarkan ID.
- `POST /api/aturan`: Menambahkan aturan (`nama`, `tujuan_id`, `prioritas`, `negara`, `perangkat`, `peramban`).
- `PUT /api/aturan/:id`: Memperbarui seluruh konfigurasi aturan.
- `PATCH /api/aturan/:id/status`: Mengaktifkan (`true`) atau menonaktifkan (`false`) aturan.
- `DELETE /api/aturan/:id`: Menghapus aturan.

### D. Pemeriksaan & Log Trafik
- `POST /api/periksa`: Menerima simulasi permintaan dan menghasilkan keputusan routing dalam JSON.
  ```json
  {
    "negara": "ID",
    "perangkat": "mobile",
    "peramban": "Chrome",
    "asal_rujukan": "https://instagram.com"
  }
  ```
- `GET /api/log-trafik?page=1&limit=20`: Mengambil riwayat log trafik dengan paginasi.
- `GET /api/log-trafik/:id`: Mengambil rincian satu log trafik.

---

## Demo Flow

1. **Buka Dasbor**: Kunjungi `http://localhost:3000` dan pastikan status backend *Online*.
2. **Kelola Tujuan**: Buka tab **Tujuan** untuk melihat Landing A, Landing B, dan Cadangan atau menambahkan tujuan baru.
3. **Konfigurasi Aturan**: Buka tab **Aturan** untuk melihat aturan aktif dengan urutan prioritas (`ID + Mobile → Landing A`, `ID + Desktop → Landing B`).
4. **Uji Simulasi Trafik**: Buka tab **Simulator Uji Trafik**, isi negara `ID`, perangkat `mobile`, klik **Periksa Trafik**. Sistem akan mencocokkan ke Aturan #1 dan mengarahkan ke Landing A.
5. **Lihat Log**: Buka tab **Log Trafik**, periksa entri baru yang langsung tercatat di tabel `traffic_logs`.
6. **Uji Kasus Fallback**: Nonaktifkan aturan atau masukkan parameter di luar jangkauan aturan (misal negara `US`), periksa trafik kembali. Sistem akan mengarahkan ke URL Fallback `/demo/cadangan`.

---

## Panduan Deployment

### 1. Database (Managed PostgreSQL)
Gunakan penyedia PostgreSQL gratis seperti **Neon.tech**, **Supabase**, atau **Render PostgreSQL**. Ambil connection string yang diberikan untuk variabel `DATABASE_URL`.

### 2. Backend (Golang)
Dapat di-deploy ke **Render**, **Railway**, atau **Fly.io** menggunakan `backend/Dockerfile` yang sudah disediakan:
1. Hubungkan repositori GitHub ke Render/Railway.
2. Atur Root Directory ke `backend` (atau gunakan Dockerfile).
3. Tambahkan Environment Variable:
   - `DATABASE_URL`: Connection string PostgreSQL production.
   - `FRONTEND_URL`: URL domain frontend production (misal `https://sivilize-demo.vercel.app`).
   - `FALLBACK_URL`: `/demo/cadangan`.
4. Jalankan migrasi sekali dengan menjalankan `go run cmd/migrasi/main.go` atau mengeksekusi DDL.

### 3. Frontend (Next.js)
Dapat di-deploy ke **Vercel**:
1. Impor repositori GitHub di dashboard Vercel.
2. Set Root Directory ke `frontend`.
3. Tambahkan Environment Variable:
   - `NEXT_PUBLIC_API_URL`: URL domain backend production (misal `https://sivilize-backend.up.railway.app`).
4. Klik **Deploy**.

---

## Google AdSense & Publisher Compliance

Sistem **Sivilize Traffic Routing Demo** dilengkapi dengan modul audit dan kepatuhan internal terhadap kebijakan Google Publisher & Google AdSense.

### Pernyataan Kepatuhan & Tanggung Jawab

- **Demo Routing Generik**: Sistem ini merupakan generic traffic routing demo (*Proof of Concept*) untuk keperluan optimasi pengalaman pengguna dan kebutuhan bisnis yang sah (seperti A/B testing, segmentasi bahasa/negara, dan adaptasi tata letak perangkat).
- **Tanggung Jawab Administrator**: Administrator dan pemilik situs bertanggung jawab penuh terhadap segala konfigurasi aturan, konten tujuan, serta kualitas sumber trafik yang diarahkan.
- **Bukan Sistem Cloaking atau Evasion**: Sistem ini **tidak dirancang** dan **tidak boleh digunakan** untuk:
  - Mengelabui crawler atau sistem review iklan Google;
  - Menampilkan halaman berbeda kepada crawler/reviewer dibanding pengguna untuk menghindari peninjauan;
  - Melakukan bot evasion atau fingerprinting agresif;
  - Menyembunyikan trafik tidak valid (*invalid traffic*);
  - Memanipulasi penargetan iklan demi meningkatkan RPM/CPC secara artifisial.
- **Aturan Perlindungan Otomatis**: Jika ada aturan routing yang secara eksplisit menargetkan crawler/reviewer Google (seperti Googlebot, AdsBot, Mediapartners-Google), sistem otomatis menandainya sebagai risiko tinggi dan **tidak akan menjalankan routing khusus tersebut** demi mencegah pelanggaran cloaking.
- **Tidak Ada Jaminan Kepatuhan Resmi**: Sistem **tidak memberikan jaminan** kepatuhan resmi ("Google Approved", "100% AdSense Compliant", atau "Guaranteed Safe"). Status kepatuhan internal yang dihasilkan (`aman`, `perlu_ditinjau`, `berisiko`) dengan skor `null` merupakan estimasi diagnostik internal, bukan keputusan resmi dari Google LLC.
- **Kebijakan Bersifat Dinamis**: Kebijakan Google dapat berubah sewaktu-waktu. Pengguna dan pengelola situs wajib selalu membaca serta mematuhi dokumentasi kebijakan resmi Google sebelum mengaktifkan periklanan AdSense.

### Referensi Resmi Kebijakan Google

1. [Kebijakan Program Google AdSense](https://support.google.com/adsense/answer/48182)
2. [Panduan Kualitas Lalu Lintas Iklan (Traffic Quality Guidelines)](https://support.google.com/adsense/answer/2660562)
3. [Panduan Penempatan Iklan (Ad Placement Policies)](https://support.google.com/adsense/answer/16737)
4. [Kebijakan Google Publisher (Publisher Policies)](https://support.google.com/publisherpolicies/answer/10502938)

### Endpoint Kepatuhan (Batch 8 API)

- `GET /api/compliance/check`: Menjalankan 7 lapisan audit kepatuhan internal tanpa menghubungi server Google luar, menghasilkan status (`aman` | `perlu_ditinjau` | `berisiko`), skor `null`, ringkasan metrik, serta menyimpan riwayat ke tabel `compliance_audits` & `compliance_findings`.
- `GET /api/compliance/checklist`: Mengambil daftar 11 item checklist kepatuhan manual dari basis data.
- `PUT /api/compliance/checklist`: Memperbarui status checklist kepatuhan mandiri oleh administrator.
- `GET /api/compliance/audits`: Mengambil riwayat catatan audit kepatuhan sebelumnya.

### Dasbor Kepatuhan (`/compliance`)

Antarmuka web interaktif yang menyajikan:
- Status audit terkini dengan badge warna informatif;
- Ringkasan temuan berdasarkan tingkat keparahan (*informasi*, *peringatan*, *risiko tinggi*);
- Rincian kartu temuan lengkap dengan kode identifikasi kebijakan;
- 11 item checklist mandiri administrator yang tersimpan di basis data secara real-time;
- Ringkasan dokumentasi kebijakan AdSense beserta tautan rujukan resmi Google.

---

## Limitasi MVP

1. **Simulasi Geolokasi**: Pada versi demo/MVP, penentuan negara pengunjung dimasukkan secara simulasi melalui antarmuka atau parameter masukan, bukan melalui basis data IP geolokasi biner offline yang berat.
2. **Kategori Standar**: Perangkat dikelompokkan ke dalam kategori umum (`mobile`, `tablet`, `desktop`) dan peramban utama tanpa teknik *device fingerprinting* tingkat rendah.
3. **Bukan Sistem Evasion**: Sistem ini strictly dirancang sebagai mesin demonstrasi perutean trafik berbasis aturan bisnis dan **tidak** memuat teknik manipulasi bot/crawler, manipulasi platform periklanan, atau cloaking.
