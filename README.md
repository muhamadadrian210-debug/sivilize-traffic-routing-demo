# Sivilize Traffic Routing Demo

**Sivilize Traffic Routing Demo** adalah aplikasi MVP (*Minimum Viable Product* / *Proof of Concept*) untuk sistem routing trafik berbasis aturan (*rule-based traffic routing*). Sistem ini dirancang untuk menerima data permintaan trafik pengunjung, menganalisis parameter kontekstual (negara, perangkat, peramban), mencocokkannya dengan aturan berprioritas, menentukan halaman tujuan yang sesuai atau tujuan cadangan (*fallback*), mencatat riwayat lalu lintas, serta menampilkan seluruh proses tersebut melalui antarmuka dasbor.

> **Status Proyek**: Demo saja, bukan untuk kebutuhan produksi komersial langsung.  
> **Pemilik**: Sivilize Corp

---

## 1. Tujuan Proyek

1. Menerima data permintaan trafik web.
2. Menganalisis informasi dasar pengunjung (negara, jenis perangkat, jenis peramban).
3. Membandingkan informasi konteks pengunjung dengan daftar aturan yang aktif.
4. Menentukan tujuan yang paling sesuai berdasarkan urutan prioritas aturan.
5. Menyediakan tujuan cadangan (*fallback*) jika tidak ada aturan yang cocok.
6. Mencatat hasil pemeriksaan trafik ke dalam basis data.
7. Menyediakan antarmuka dasbor untuk pemantauan, pengelolaan aturan/tujuan, dan uji simulasi trafik.

---

## 2. Cara Kerja Sistem

```
Pengunjung / Permintaan Uji
           │
           ▼
Penganalisis Permintaan (IP, Perangkat, Peramban, Negara)
           │
           ▼
Mesin Aturan (Pencocokan berdasarkan Prioritas Terkecil)
      ┌────┴────┐
      ▼         ▼
    Cocok?   Tidak Cocok?
      │         │
      │         ▼
      │     Tujuan Cadangan (/demo/cadangan)
      ▼
Tujuan Terpilih (/demo/halaman-a atau /demo/halaman-b)
           │
           ▼
Pencatatan Log Trafik (Tabel catatan_trafik)
           │
           ▼
Tampilan Dasbor & Hasil Pemeriksaan
```

---

## 3. Teknologi yang Digunakan

- **Backend**: Golang dengan kerangka kerja HTTP [Gin](https://github.com/gin-gonic/gin)
- **Frontend**: Next.js (App Router), TypeScript, Tailwind CSS
- **Basis Data**: PostgreSQL (akan diintegrasikan mulai Tahap 2)
- **Komunikasi**: REST API dengan format data JSON
- **Repositori & Versi**: Git / GitHub

---

## 4. Struktur Folder Proyek

```text
sivilize-traffic-routing-demo/
│
├── README.md               # Dokumentasi utama proyek dalam Bahasa Indonesia
├── .gitignore              # Daftar berkas/folder yang diabaikan Git
├── .env.example            # Contoh konfigurasi variabel lingkungan utama
│
├── backend/                # Kode sumber aplikasi backend (Golang)
│   ├── go.mod              # Definisi modul dan dependensi Go
│   ├── go.sum              # Checksum verifikasi dependensi Go
│   ├── cmd/
│   │   └── server/
│   │       └── main.go     # Titik masuk utama aplikasi backend
│   ├── internal/
│   │   ├── konfigurasi/    # Pemuat konfigurasi dari environment
│   │   ├── basisdata/      # Koneksi dan inisialisasi PostgreSQL (Tahap 2)
│   │   ├── model/          # Definisi entitas data (Tahap 2)
│   │   ├── repositori/     # Akses query database langsung (Tahap 2)
│   │   ├── layanan/        # Logika bisnis perantara (Tahap 3-4)
│   │   ├── penangan/       # HTTP handler Gin untuk endpoint API
│   │   ├── middleware/     # Penanganan CORS dan autentikasi/keamanan
│   │   ├── aturan/         # Mesin aturan dan evaluasi prioritas (Tahap 5)
│   │   └── pencatatan/     # Modul pencatatan log trafik (Tahap 6-7)
│   ├── migrasi/            # Skrip DDL tabel database (Tahap 2)
│   └── pengujian/          # Unit test mesin aturan dan handler (Tahap 5 & 11)
│
└── frontend/               # Kode sumber aplikasi antarmuka (Next.js)
    ├── package.json        # Dependensi dan skrip frontend
    ├── tsconfig.json       # Konfigurasi kompilasi TypeScript
    ├── .env.example        # Contoh environment variabel frontend
    ├── .env.local          # Konfigurasi lokal frontend
    ├── app/                # Halaman Next.js App Router
    ├── komponen/           # Komponen UI modular
    ├── pustaka/            # Utilitas pembantu dan pemanggil API
    ├── tipe/               # Antarmuka dan tipe data TypeScript
    └── kait/               # React Custom Hooks
```

---

## 5. Cara Menjalankan Aplikasi Secara Lokal

### Prasyarat
- **Go**: Versi 1.22 atau lebih baru
- **Node.js**: Versi 18 atau lebih baru (npm disertakan)

### Langkah Menjalankan Backend:
1. Buka terminal dan masuk ke folder `backend`:
   ```bash
   cd backend
   ```
2. Jalankan server:
   ```bash
   go run ./cmd/server
   ```
3. Server backend akan aktif pada port `http://localhost:8080`.

### Langkah Menjalankan Frontend:
1. Buka terminal baru dan masuk ke folder `frontend`:
   ```bash
   cd frontend
   ```
2. Pasang dependensi jika belum (hanya saat pertama kali):
   ```bash
   npm install
   ```
3. Jalankan server pengembangan Next.js:
   ```bash
   npm run dev
   ```
4. Buka peramban pada alamat `http://localhost:3000`.

---

## 6. Konfigurasi Basis Data, Migrasi & Seed (Tahap 2)

*Catatan: Sesuai pedoman pembangunan berfokus pada Tahap 1, modul basis data belum diaktifkan pada tahap ini.*

Pada Tahap 2, basis data PostgreSQL akan dihubungkan menggunakan `DATABASE_URL` dengan tabel:
1. `tujuan` (halaman target trafik)
2. `aturan` (kriteria pencocokan konteks pengunjung)
3. `catatan_trafik` (riwayat evaluasi lalu lintas)

---

## 7. Dokumentasi API

Seluruh response API menggunakan format standar:
- Berhasil: `{"data": ...}`
- Error: `{"error": true, "pesan": "..."}`

---

### A. Endpoint Pemeriksaan Sistem

#### `GET /api/kesehatan`
- **Deskripsi**: Memeriksa status kesehatan server dan konektivitas PostgreSQL.
- **Respons (200 OK)**:
  ```json
  {
    "status": "sehat",
    "database": "terhubung"
  }
  ```

---

### B. Endpoint Tujuan

#### `GET /api/tujuan`
- **Deskripsi**: Mengambil daftar seluruh tujuan yang tersimpan (diurutkan berdasarkan `id ASC`).
- **Respons (200 OK)**:
  ```json
  {
    "data": [
      {
        "id": 1,
        "nama": "Halaman A",
        "url": "/demo/halaman-a",
        "status": true,
        "dibuat_pada": "2026-09-26T07:23:55Z",
        "diperbarui_pada": "2026-09-26T07:23:55Z"
      }
    ]
  }
  ```

#### `GET /api/tujuan/:id`
- **Deskripsi**: Mengambil data satu tujuan berdasarkan ID.
- **Respons (200 OK)**:
  ```json
  {
    "data": {
      "id": 1,
      "nama": "Halaman A",
      "url": "/demo/halaman-a",
      "status": true,
      "dibuat_pada": "2026-09-26T07:23:55Z",
      "diperbarui_pada": "2026-09-26T07:23:55Z"
    }
  }
  ```
- **Respons Error (404 Not Found)**:
  ```json
  {
    "error": true,
    "pesan": "Tujuan tidak ditemukan"
  }
  ```

#### `POST /api/tujuan`
- **Deskripsi**: Menambahkan tujuan baru.
- **Request Body**:
  ```json
  {
    "nama": "Halaman Promo",
    "url": "/demo/promo",
    "status": true
  }
  ```
- **Respons (201 Created)**:
  ```json
  {
    "data": {
      "id": 4,
      "nama": "Halaman Promo",
      "url": "/demo/promo",
      "status": true,
      "dibuat_pada": "2026-09-26T07:30:00Z",
      "diperbarui_pada": "2026-09-26T07:30:00Z"
    }
  }
  ```

#### `PUT /api/tujuan/:id`
- **Deskripsi**: Memperbarui tujuan yang sudah ada.
- **Request Body**:
  ```json
  {
    "nama": "Halaman Promo Baru",
    "url": "/demo/promo-baru",
    "status": true
  }
  ```
- **Respons (200 OK)**:
  ```json
  {
    "data": {
      "id": 4,
      "nama": "Halaman Promo Baru",
      "url": "/demo/promo-baru",
      "status": true,
      "dibuat_pada": "2026-09-26T07:30:00Z",
      "diperbarui_pada": "2026-09-26T07:35:00Z"
    }
  }
  ```

#### `DELETE /api/tujuan/:id`
- **Deskripsi**: Menghapus tujuan berdasarkan ID. Jika tujuan masih digunakan oleh aturan, penghapusan ditolak dengan konflik data.
- **Respons (200 OK)**:
  ```json
  {
    "data": {
      "pesan": "Tujuan berhasil dihapus"
    }
  }
  ```
- **Respons Konflik (409 Conflict)**:
  ```json
  {
    "error": true,
    "pesan": "Tujuan masih digunakan oleh aturan dan tidak dapat dihapus"
  }
  ```

---

### C. Endpoint Aturan

#### `GET /api/aturan`
- **Deskripsi**: Mengambil seluruh daftar aturan lengkap dengan data tujuan melalui JOIN PostgreSQL (diurutkan berdasarkan `prioritas ASC, id ASC`).
- **Respons (200 OK)**:
  ```json
  {
    "data": [
      {
        "id": 1,
        "nama": "Indonesia Mobile",
        "negara": "ID",
        "perangkat": "mobile",
        "peramban": null,
        "agen_pengguna": null,
        "asal_rujukan": null,
        "tujuan_id": 1,
        "tujuan_nama": "Halaman A",
        "tujuan_url": "/demo/halaman-a",
        "prioritas": 1,
        "status": true,
        "dibuat_pada": "2026-09-26T07:23:55Z",
        "diperbarui_pada": "2026-09-26T07:23:55Z"
      }
    ]
  }
  ```

#### `GET /api/aturan/:id`
- **Deskripsi**: Mengambil satu aturan berdasarkan ID lengkap dengan data tujuan.
- **Respons (200 OK)**:
  ```json
  {
    "data": {
      "id": 1,
      "nama": "Indonesia Mobile",
      "negara": "ID",
      "perangkat": "mobile",
      "peramban": null,
      "agen_pengguna": null,
      "asal_rujukan": null,
      "tujuan_id": 1,
      "tujuan_nama": "Halaman A",
      "tujuan_url": "/demo/halaman-a",
      "prioritas": 1,
      "status": true,
      "dibuat_pada": "2026-09-26T07:23:55Z",
      "diperbarui_pada": "2026-09-26T07:23:55Z"
    }
  }
  ```

#### `POST /api/aturan`
- **Deskripsi**: Menambahkan aturan pencocokan baru.
- **Request Body**:
  ```json
  {
    "nama": "Aturan Indonesia Desktop",
    "negara": "ID",
    "perangkat": "desktop",
    "peramban": "Chrome",
    "agen_pengguna": null,
    "asal_rujukan": null,
    "tujuan_id": 2,
    "prioritas": 2,
    "status": true
  }
  ```
- **Validasi**:
  - `nama`: Wajib diisi.
  - `tujuan_id`: Wajib merujuk ke ID tujuan yang ada di database.
  - `prioritas`: Wajib integer >= 1.
  - `perangkat`: Opsional (`null`), jika diisi hanya `mobile`, `tablet`, atau `desktop`.
  - `peramban`: Opsional (`null`), jika diisi hanya `Chrome`, `Firefox`, `Safari`, `Edge`, atau `Lainnya`.
  - `negara`: Opsional (`null`), tidak boleh string kosong jika dikirim.
- **Respons (201 Created)**:
  ```json
  {
    "data": {
      "id": 3,
      "nama": "Aturan Indonesia Desktop",
      "negara": "ID",
      "perangkat": "desktop",
      "peramban": "Chrome",
      "agen_pengguna": null,
      "asal_rujukan": null,
      "tujuan_id": 2,
      "tujuan_nama": "Halaman B",
      "tujuan_url": "/demo/halaman-b",
      "prioritas": 2,
      "status": true,
      "dibuat_pada": "2026-09-26T07:40:00Z",
      "diperbarui_pada": "2026-09-26T07:40:00Z"
    }
  }
  ```

#### `PUT /api/aturan/:id`
- **Deskripsi**: Mengubah seluruh konfigurasi aturan.
- **Request Body**: Sama dengan format `POST /api/aturan`.
- **Respons (200 OK)**: Mengembalikan data aturan terkini.

#### `PATCH /api/aturan/:id/status`
- **Deskripsi**: Mengubah status aktif/nonaktif aturan secara spesifik tanpa mengubah data lainnya.
- **Request Body**:
  ```json
  {
    "status": false
  }
  ```
- **Respons (200 OK)**: Mengembalikan data aturan dengan status terkini.

#### `DELETE /api/aturan/:id`
- **Deskripsi**: Menghapus aturan berdasarkan ID.
- **Respons (200 OK)**:
  ```json
  {
    "data": {
      "pesan": "Aturan berhasil dihapus"
    }
  }
  ```

---

### D. Request Analyzer & Mesin Aturan (Batch 4)

#### 1. Konteks Pengunjung (`KonteksPengunjung`)
Menganalisis permintaan pengunjung berdasarkan header HTTP atau masukan simulasi:
- **Negara**: Dinormalisasi menjadi huruf kapital (misal `"id"` -> `"ID"`).
- **Perangkat**: Diklasifikasikan menjadi `"mobile"`, `"tablet"`, `"desktop"`, atau `"unknown"`.
- **Peramban**: Diklasifikasikan menjadi `"Chrome"`, `"Firefox"`, `"Safari"`, `"Edge"`, `"Lainnya"`, atau `"Tidak Diketahui"`.
  - *Catatan*: Edge diprioritaskan sebelum Chrome karena User-Agent Edge memuat token Chrome. Safari diprioritaskan setelah Chrome karena User-Agent Chrome memuat token Safari/WebKit.
- **Agen Pengguna**: String asli User-Agent tetap disimpan utuh tanpa modifikasi.
- **Asal Rujukan (Referrer)**: Referrer asli permintaan.
- **Alamat IP**: Diambil dari `X-Forwarded-For`, `X-Real-IP`, atau `RemoteAddr`.

#### 2. Logika Pencocokan Aturan (`CocokkanAturan`)
- Hanya aturan dengan `status = true` (aktif) yang diikutsertakan.
- Aturan cocok jika **SEMUA** kriteria yang terisi pada aturan cocok dengan konteks pengunjung.
- Kolom bernilai `NULL` dianggap sebagai *wildcard* (tidak membatasi kriteria).
- Filter `agen_pengguna` dan `asal_rujukan` menggunakan pencocokan *case-insensitive substring*.

#### 3. Logika Prioritas & Deterministik
- Jika terdapat lebih dari satu aturan yang cocok, sistem memilih aturan dengan nilai `prioritas` terkecil (angka lebih kecil = prioritas lebih tinggi, contoh prioritas 1 menang atas prioritas 5).
- Jika terdapat aturan dengan nilai prioritas yang sama persis, sistem menggunakan ID aturan terkecil secara deterministik.

#### 4. Penanganan Fallback (Tujuan Cadangan)
- Jika tidak ada aturan yang cocok, atau tujuan pada aturan yang menang sudah tidak tersedia/tidak aktif, sistem mengarahkan trafik ke URL fallback terpusat (`UrlFallback`, default: `/demo/cadangan`).

#### 5. Format Hasil Evaluasi Routing (`HasilRouting`)
Contoh hasil jika aturan cocok:
```json
{
  "cocok": true,
  "aturan_id": 1,
  "tujuan_id": 1,
  "url_tujuan": "/demo/halaman-a",
  "hasil": "aturan",
  "alasan": "Aturan 'Indonesia Mobile' (ID 1) cocok dengan permintaan pengunjung"
}
```

Contoh hasil fallback (tidak ada aturan yang cocok):
```json
{
  "cocok": false,
  "aturan_id": null,
  "tujuan_id": null,
  "url_tujuan": "/demo/cadangan",
  "hasil": "fallback",
  "alasan": "Tidak ada aturan aktif yang cocok"
}
```

---

### E. Endpoint Pemeriksaan Trafik & Riwayat Log (Batch 5)

#### 1. `POST /api/periksa`
- **Deskripsi**: Menerima data simulasi trafik pengunjung, menjalankan Request Analyzer dan Rule Engine, mencatat riwayat ke tabel `traffic_logs`, dan mengembalikan hasil keputusan tujuan dalam format JSON (tanpa redirect HTTP browser).
- **Request Body**:
  ```json
  {
    "negara": "ID",
    "perangkat": "mobile",
    "peramban": "Chrome",
    "agen_pengguna": "Mozilla/5.0 (Linux; Android 13; SM-S908B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Mobile Safari/537.36",
    "asal_rujukan": "https://google.com"
  }
  ```
- **Respons Aturan Cocok (200 OK)**:
  ```json
  {
    "data": {
      "cocok": true,
      "aturan_id": 1,
      "tujuan_id": 1,
      "url_tujuan": "/demo/halaman-a",
      "hasil": "aturan",
      "alasan": "Aturan 'Indonesia Mobile' (ID 1) cocok dengan permintaan pengunjung"
    }
  }
  ```
- **Respons Fallback (200 OK)**:
  ```json
  {
    "data": {
      "cocok": false,
      "aturan_id": null,
      "tujuan_id": null,
      "url_tujuan": "/demo/cadangan",
      "hasil": "fallback",
      "alasan": "Tidak ada aturan aktif yang cocok"
    }
  }
  ```
- **Respons Validasi Error (400 Bad Request)**:
  ```json
  {
    "error": true,
    "pesan": "Perangkat tidak valid (hanya mobile, tablet, atau desktop)"
  }
  ```
- **Contoh curl**:
  ```bash
  curl -X POST http://localhost:8080/api/periksa \
    -H "Content-Type: application/json" \
    -d '{"negara":"ID","perangkat":"mobile","peramban":"Chrome"}'
  ```

#### 2. `GET /api/log-trafik`
- **Deskripsi**: Mengambil riwayat catatan evaluasi trafik dengan dukungan paginasi terurut dari yang paling baru (`dibuat_pada DESC`).
- **Query Parameter**:
  - `page`: Nomor halaman (default: `1`, minimal: `1`).
  - `limit`: Jumlah data per halaman (default: `20`, maksimal: `100`).
- **Respons Berhasil (200 OK)**:
  ```json
  {
    "data": [
      {
        "id": 10,
        "negara": "ID",
        "perangkat": "mobile",
        "peramban": "Chrome",
        "agen_pengguna": "Mozilla/5.0...",
        "asal_rujukan": "https://google.com",
        "aturan_id": 1,
        "tujuan_id": 1,
        "url_tujuan": "/demo/halaman-a",
        "hasil": "aturan",
        "dibuat_pada": "2026-09-26T10:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 45,
      "total_halaman": 3
    }
  }
  ```
- **Contoh curl**:
  ```bash
  curl "http://localhost:8080/api/log-trafik?page=1&limit=10"
  ```

#### 3. `GET /api/log-trafik/:id`
- **Deskripsi**: Mengambil detail satu entri log trafik berdasarkan ID uniknya.
- **Respons (200 OK)**:
  ```json
  {
    "data": {
      "id": 10,
      "negara": "ID",
      "perangkat": "mobile",
      "peramban": "Chrome",
      "agen_pengguna": "Mozilla/5.0...",
      "asal_rujukan": "https://google.com",
      "aturan_id": 1,
      "tujuan_id": 1,
      "url_tujuan": "/demo/halaman-a",
      "hasil": "aturan",
      "dibuat_pada": "2026-09-26T10:00:00Z"
    }
  }
  ```
- **Respons Tidak Ditemukan (404 Not Found)**:
  ```json
  {
    "error": true,
    "pesan": "Log trafik tidak ditemukan"
  }
  ```
- **Contoh curl**:
  ```bash
  curl "http://localhost:8080/api/log-trafik/10"
  ```

---

## 8. Pengujian

- **Backend Build Test**:
  ```bash
  cd backend
  go build -v ./cmd/server
  ```
- **Frontend Build & Typecheck**:
  ```bash
  cd frontend
  npm run build
  ```

---

## 9. Keterbatasan MVP

1. **Deteksi Geolokasi**: Pada versi MVP, penentuan negara pengunjung dimasukkan secara simulasi melalui antarmuka atau parameter masukan, bukan melalui basis data IP geolokasi biner yang rumit.
2. **Kategori Perangkat & Peramban**: Dikelompokkan ke dalam kategori standar (mobile, tablet, desktop) dan peramban populer (Chrome, Firefox, Safari, Edge, Lainnya) tanpa fingerprinting perangkat tingkat rendah.
3. **Keamanan & Ruang Lingkup**: Sistem dirancang strictly sebagai mesin perutean generik dan **tidak** memuat teknik manipulasi bot/crawler, manipulasi platform periklanan, atau cloaking.
