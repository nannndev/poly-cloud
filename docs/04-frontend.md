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
- **accounts store:** daftar account, status, kuota per account, total agregat.
- **files store:** entri file di path aktif, hasil search, seleksi, breadcrumb path.
- Data di-fetch dari backend (yang baca DB cache) → cepat, tak nunggu provider.

## 4. Komponen Kunci
- **FileExplorer:** tampilkan file gabungan semua account; tiap file punya badge account asal
  (mis. "GDrive utama"). Toggle view: gabungan vs per-account. Sort & pagination.
- **UploadDropzone:** drop file → kirim ke backend; backend yang pilih account (smart routing).
  UI cukup tampilkan "→ menuju GDrive-2" hasil keputusan router + progress bar (SSE).
- **QuotaDashboard:** bar terpakai/bebas per account + satu angka agregat besar (value inti).
- **AccountCard:** kalau `needs_reconnect`, tampilkan tombol reconnect.

## 5. Alur UI Penting
- **Connect:** klik "Tambah Account" → pilih provider → redirect OAuth → `callback.vue`
  kirim code ke backend → account muncul + initial sync jalan (spinner/index progress).
- **Upload:** drop → POST stream → SSE progress → item baru muncul di explorer dgn badge account.
- **Download:** klik file → GET stream → browser save.
- **Search:** ketik → debounce → backend (query DB) → render hasil lintas account.

## 6. Kontrak Tipe
Impor dari `packages/shared-types` (`FileEntry`, `Account`, `Quota`, response API).
Jaga sinkron dengan Go structs (lihat [02 §4](02-repo-structure.md)).

## 7. UX Prinsip
- Utamakan kesan "satu drive": default view = gabungan, badge account sekadar info.
- Router menyembunyikan kompleksitas: user tak perlu memilih account saat upload (boleh override manual).
- Feedback jelas untuk operasi async (progress, status account).
