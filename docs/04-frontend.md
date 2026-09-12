# 04 — Frontend Design (Nuxt 4)

## 1. Peran
UI tunggal untuk seluruh storage: connect account, explorer terpadu, search, upload
(dengan progress), download/preview, dashboard kuota. Tidak pernah akses provider/DB
langsung — semua lewat backend REST.

## 2. Struktur
```
app/
├── pages/
│   ├── index.vue            # explorer utama (unified file view)
│   ├── accounts.vue         # kelola account terhubung
│   └── connect/
│       └── callback.vue     # OAuth redirect landing
├── components/
│   ├── FileExplorer.vue     # grid/list file lintas account
│   ├── FileRow.vue          # 1 baris file (nama, size, badge account)
│   ├── UploadDropzone.vue   # drag-drop + progress
│   ├── QuotaDashboard.vue   # bar per account + agregat
│   ├── AccountCard.vue      # status account (active/needs_reconnect)
│   └── SearchBar.vue        # search + filter (tipe/size/account)
├── layouts/default.vue
composables/
├── useAccounts.ts           # list/connect/remove account
├── useFiles.ts              # list/search/move/delete
├── useUpload.ts             # upload + SSE progress
└── useQuota.ts              # kuota agregat & per account
stores/
├── files.ts                 # Pinia: state file & path aktif
└── accounts.ts              # Pinia: state account & kuota
```

## 3. State (Pinia)
- **accounts store:** daftar account, status, kuota per account, total agregat,
  daftar provider beserta kesiapan kredensialnya (`GET /providers`).
- **files store:** folder VFS, entri file di folder aktif, hasil search, dan
  antrean upload.
- Data di-fetch dari backend (yang baca DB cache) → cepat, tak nunggu provider.
- Tiap halaman memuat datanya lewat `useAsyncData(..., { server: false })`:
  backend hanya dijangkau dari browser, bukan dari proses SSR.

### Lapisan API (`composables/useApi.ts`)
Semua panggilan lewat satu pembungkus `$fetch` yang menerjemahkan body error
`{ error: { code, message } }` jadi `ApiError` dengan `code` dari doc 06,
sehingga UI bisa bereaksi spesifik (mis. `NO_ROOM`, `PATH_EXISTS`) alih-alih
menampilkan pesan mentah.

## 4. Komponen Kunci
- **FileExplorer:** tampilkan file gabungan semua account; tiap file punya badge account asal
  (mis. "GDrive utama"). Toggle view: gabungan vs per-account. Sort & pagination.
- **UploadModal:** drop file → kirim ke backend; backend yang pilih account (smart routing).
  UI menampilkan prediksi tujuan sebelum unggah dan akun final hasil keputusan
  router sesudahnya, dengan progress bar yang digerakkan SSE. Upload berjalan
  bisa dibatalkan (request diputus); tak ada pause/resume karena HTTP upload
  biasa tak bisa dilanjutkan tanpa protokol resumable.
- **QuotaDashboard:** bar terpakai/bebas per account + satu angka agregat besar (value inti).
- **AccountCard:** kalau `needs_reconnect`, tampilkan tombol reconnect.

## 5. Alur UI Penting
- **Connect:** klik "Tambah Account" → pilih provider → redirect OAuth → `callback.vue`
  kirim code ke backend → account muncul + initial sync jalan (spinner/index progress).
- **Upload:** drop → POST stream → SSE progress → item baru muncul di explorer dgn badge account.
- **Download:** klik file → GET stream → browser save.
- **Search:** ketik → debounce → backend (query DB) → render hasil lintas account.

## 6. Kontrak Tipe
`app/types/index.ts` mencerminkan `packages/shared-types/types.ts`
(`FileEntry`, `Account`, `Quota`, response API). Jaga sinkron dengan Go structs
(lihat [02 §4](02-repo-structure.md)).

## 6a. Pratinjau Berkas
Pratinjau memakai `GET /files/{id}/download?inline=1` sehingga browser merender
isinya langsung dari stream provider: gambar, video, audio, dan PDF ditangani
penampil bawaan browser; berkas teks diambil isinya untuk ditampilkan bernomor
baris (dibatasi 512 KB). Format yang tak bisa dirender browser (dokumen Office,
arsip) menampilkan tombol unduh, bukan pratinjau tiruan.

## 7. UX Prinsip
- Utamakan kesan "satu drive": default view = gabungan, badge account sekadar info.
- Router menyembunyikan kompleksitas: user tak perlu memilih account saat upload (boleh override manual).
- Feedback jelas untuk operasi async (progress, status account).
