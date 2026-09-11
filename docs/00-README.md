# Storage Aggregator Platform — Documentation

Satu platform yang menghubungkan banyak akun cloud (Google Drive, Dropbox, OneDrive,
S3, Backblaze B2, dll) dan menampilkannya sebagai satu ruang penyimpanan terpadu.

- **Arsitektur terkunci:** Model A (index, whole-file) + Smart Routing, dengan jalur upgrade ke Model B (chunked) tanpa rewrite.
- **Repo:** Monorepo.
- **Scope:** Personal-first → self-hosted multi-user product.
- **Engine:** rclone (70+ provider, OAuth & refresh included).

## Peta Dokumen

| # | Dokumen | Isi |
|---|---------|-----|
| 01 | [Architecture](01-architecture.md) | SAD: konteks, C4, komponen, aliran data, model A→B |
| 02 | [Repo Structure](02-repo-structure.md) | Keputusan monorepo, struktur folder, tooling |
| 03 | [Backend Design](03-backend.md) | Go: layering, StorageService, engine, routing, lifecycle token |
| 04 | [Frontend Design](04-frontend.md) | Nuxt 4: struktur, state, komponen explorer, upload/download |
| 05 | [Data Model](05-data-model.md) | ERD + skema Postgres final |
| 06 | [API Spec](06-api-spec.md) | Kontrak REST endpoint |
| 07 | [Sequence Flows](07-sequence-flows.md) | Diagram: OAuth connect, upload+routing, download, refresh |
| 08 | [ADR](08-adr.md) | Architecture Decision Records |
| 09 | [Virtual Filesystem](09-virtual-filesystem.md) | Folder virtual: organisasi terpisah dari lokasi fisik |

## Glosarium Singkat
- **Provider** — layanan cloud (GDrive, Dropbox, …).
- **Account** — 1 akun user pada 1 provider (bisa banyak akun per provider).
- **Remote** — nama konfigurasi rclone untuk 1 account (mis. `gdrive1:`).
- **Engine** — lapisan yang mengeksekusi operasi file (rclone).
- **Smart Routing** — logika memilih account tujuan saat upload.
- **Model A** — 1 file = 1 objek di 1 account (whole-file).
- **Model B** — 1 file = N chunk tersebar di N account (chunked, fase lanjut).
- **Index** — metadata file yang di-cache di DB (bukan file-nya).
- **VFS (Virtual Filesystem)** — struktur folder yang user lihat; hidup di DB platform, terpisah dari lokasi fisik di provider.
- **Virtual path** — path folder logis (mis. `/Kerjaan/laporan.pdf`) yang tak tercermin di provider.

## Status
Draft v1 — dokumen desain, mendahului implementasi.
