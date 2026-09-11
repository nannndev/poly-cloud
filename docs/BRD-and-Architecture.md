# Storage Aggregator Platform — BRD & Architecture

**Codename:** PolyCloud
**Owner:** Nan (ekaprasetya2244@gmail.com)
**Status:** Draft v1
**Scope:** Personal-first, designed to scale to self-hosted multi-user product.
**Architecture locked:** Model A (index, whole-file) + Smart Routing, with a forward path to Model B (chunked).

---

## PART 1 — BUSINESS REQUIREMENTS DOCUMENT (BRD)

### 1.1 Problem Statement
Free-tier cloud storage (Google Drive 15GB, dll) cepat penuh. Punya banyak akun gratis
di banyak provider = kapasitas total besar, tapi terfragmentasi: harus login-logout,
tiap akun kepisah, gak ada tempat/search terpusat. Tidak ada view tunggal atas semua
storage yang dimiliki.

### 1.2 Goal
Satu platform yang menghubungkan banyak akun cloud (dari provider apa pun) dan
menampilkannya sebagai satu ruang penyimpanan terpadu: browse, search, upload,
download, dan pindah file — tanpa peduli file itu fisik ada di akun/provider mana.

### 1.3 Non-Goals (v1)
- Bukan tempat sinkronisasi realtime dua arah seperti Dropbox client (v1 = manage, bukan sync-daemon).
- Bukan file editor / collaboration suite.
- Belum chunking/splitting file (itu Model B, fase lanjut).
- Belum sharing/permission antar user eksternal.

### 1.4 Target Users
- **v1 (Personal):** 1 user (Nande), banyak akun.
- **v2 (Self-hosted / Product):** multi-user, tiap user punya set akun sendiri, isolated.

### 1.5 Core User Stories
| ID | Sebagai | Ingin | Supaya |
|----|---------|-------|--------|
| US-1 | User | menghubungkan akun cloud via OAuth | akunku terbaca platform |
| US-2 | User | melihat semua file dari semua akun dalam 1 explorer | tak perlu pindah-pindah login |
| US-3 | User | search file lintas akun | cepat menemukan file |
| US-4 | User | upload file, sistem otomatis memilih akun tujuan | tak perlu mikir akun mana yang muat |
| US-5 | User | download / preview file | mengambil file |
| US-6 | User | melihat kapasitas terpakai & sisa per akun + total | tahu kondisi storage |
| US-7 | User | memindah file antar akun | menyeimbangkan / merapikan |
| US-8 | User | menghapus akun terhubung | mencabut akses |

### 1.6 Functional Requirements
- **FR-1 Connect:** OAuth ke provider; simpan token (encrypted at-rest) + metadata akun.
- **FR-2 Index:** Sinkron metadata file (nama, path, size, mime, modified, akun-asal) ke DB.
- **FR-3 Unified Explorer:** Tampilan folder gabungan / per-akun; sortable; paginated.
- **FR-4 Search:** Full-text nama file lintas akun; filter tipe/ukuran/akun.
- **FR-5 Upload + Smart Routing:** Pilih akun tujuan otomatis (default: most-free-space
  yang muat) atau manual override.
- **FR-6 Download / Stream:** Ambil file lewat platform (proxy stream, bukan simpan permanen).
- **FR-7 Quota Dashboard:** Sisa & terpakai per akun + agregat.
- **FR-8 Move / Delete:** Operasi file antar / dalam akun.
- **FR-9 Token Lifecycle:** Auto-refresh; tandai akun "needs reconnect" bila expired.

### 1.7 Non-Functional Requirements
- **Security:** Token & refresh-token dienkripsi (AES-GCM) pakai key dari KMS/env; tidak
  pernah plaintext di DB. File data tidak disimpan permanen di server (stream-through).
- **Portability:** Jalan sebagai self-hosted (Docker Compose) — sesuai infra self-hosting kamu.
- **Extensibility:** Tambah provider baru = tambah 1 adapter, tanpa sentuh UI/DB.
- **Scalability:** Model A → Model B tanpa rewrite (lihat Part 2.6).
- **Reliability:** Operasi idempotent; retry pada kegagalan transient.

### 1.8 Success Metrics
- Bisa connect ≥3 provider berbeda.
- 1 explorer menampilkan file dari semua akun < 2 dtk (cached index).
- Upload auto-route ke akun benar tanpa manual pilih.
- Migrasi ke self-hosted = 1 perintah `docker compose up`.

### 1.9 Risks & Mitigasi
| Risiko | Dampak | Mitigasi |
|--------|--------|----------|
| Provider API rate-limit saat index | Sync lambat | Incremental sync + backoff; index bertahap |
| Token bocor | Akun user tereskpos | Enkripsi at-rest, scope minimal, rotasi |
| ToS provider soal automated access | Akun kena limit | Pakai OAuth resmi + scope wajar; hindari abuse pattern |
| File besar via server (proxy) | Bandwidth/memori server | Stream (io.Copy), bukan buffer penuh; opsi direct-link |
| Ketergantungan rclone | Terkunci behavior rclone | Bungkus di interface `Provider`; bisa ganti impl |

---

## PART 2 — ARCHITECTURE

### 2.1 High-Level Diagram
```
                    ┌─────────────────────────────┐
                    │   Nuxt 4 Frontend (SPA)     │
                    │  Explorer · Search · Upload  │
                    └──────────────┬──────────────┘
                                   │ REST/JSON (+ SSE for progress)
                    ┌──────────────▼──────────────┐
                    │      Go Backend (API)        │
                    │  ┌────────────────────────┐  │
                    │  │  StorageService (iface) │  │  ◄── titik abstraksi kunci
                    │  └───────────┬────────────┘  │
                    │   Routing · Index · Auth     │
                    └──────┬───────────────┬───────┘
                           │               │
                  ┌────────▼──────┐  ┌──────▼────────┐
                  │  Postgres     │  │   rclone       │
                  │  (Supabase)   │  │   engine       │
                  │ accounts,     │  │ (CLI/lib)      │
                  │ files_index,  │  └──────┬────────┘
                  │ tokens(enc)   │         │ per-provider ops
                  └───────────────┘  ┌──────┴───────────────────────┐
                                     ▼        ▼        ▼        ▼
                                  GDrive  Dropbox  OneDrive  S3/B2 ...
```

### 2.2 Components
- **Frontend (Nuxt 4):** File explorer, connect-account wizard, upload dropzone, quota
  dashboard, search bar. Bicara ke backend via REST; progress upload/download via SSE/WebSocket.
- **Backend (Go):** API server + `StorageService`. Menyimpan business logic:
  smart routing, indexing, token lifecycle. Tidak menyimpan file permanen.
- **rclone engine:** Eksekutor operasi file per-provider (list/upload/download/move/delete,
  OAuth, refresh). Dipanggil via library atau CLI subprocess.
- **Postgres (Supabase):** Sumber kebenaran untuk metadata & routing. Token dienkripsi.

### 2.3 The Key Abstraction (inti scalability)
```go
// Interface stabil — UI & DB tak peduli implementasi di bawahnya.
type StorageService interface {
    List(ctx, path) ([]FileEntry, error)
    Upload(ctx, path, io.Reader, size) (FileEntry, error)   // routing terjadi di sini
    Download(ctx, fileID) (io.ReadCloser, error)
    Move(ctx, fileID, destAccount) error
    Delete(ctx, fileID) error
    Search(ctx, query, filters) ([]FileEntry, error)
    Quota(ctx) ([]AccountQuota, error)
}

// Model A: satu file = satu objek di satu akun.
type WholeFileStore struct { router Router; index Index; engine RcloneEngine }

// Model B (nanti): satu file = N chunk tersebar. Interface SAMA.
type ChunkedStore struct   { router Router; index Index; engine RcloneEngine; splitter Splitter }
```
UI memanggil `StorageService`. Ganti dari A ke B = ganti implementasi yang di-inject,
bukan mengubah API/UI.

### 2.4 Smart Routing (Model A)
Saat `Upload`, pilih akun tujuan:
1. Filter akun yang `free_space >= file_size` dan status `active`.
2. Skor kandidat. Strategi default **most-free-space** (isi akun paling lowong dulu).
   Alternatif pluggable: round-robin, by-mime (foto→akun A), by-provider-preference.
3. Kalau tak ada satu akun pun yang muat → di Model A: tolak / minta user pilih.
   (Di Model B nanti: inilah trigger untuk chunk & sebar.)
4. Eksekusi upload via rclone ke akun terpilih; catat di `files_index`.

```
UploadRequest(size) → Router.Pick(accounts, size, strategy) → account
                    → engine.Upload(account, ...) → index.Insert(...)
```

### 2.5 Data Model (Postgres)
```sql
-- Akun cloud yang terhubung
accounts (
  id            uuid pk,
  user_id       uuid,                 -- multi-tenant ready sejak awal
  provider      text,                 -- 'gdrive' | 'dropbox' | 'onedrive' | 's3' | 'b2'
  label         text,                 -- 'GDrive utama', 'Dropbox kerja'
  rclone_remote text,                 -- nama remote di rclone
  total_bytes   bigint,
  used_bytes    bigint,
  status        text,                 -- 'active' | 'needs_reconnect' | 'error'
  created_at    timestamptz,
  last_synced   timestamptz
)

-- Kredensial (TERPISAH, terenkripsi)
account_tokens (
  account_id    uuid pk fk,
  enc_payload   bytea,                -- AES-GCM(access+refresh token)
  expires_at    timestamptz
)

-- Index metadata file (bukan file-nya)
files_index (
  id            uuid pk,
  user_id       uuid,
  name          text,
  virtual_path  text,                 -- path di explorer terpadu
  mime          text,
  size_bytes    bigint,
  modified_at   timestamptz,
  -- Kunci scalability: file punya SATU-atau-BANYAK block.
  -- Model A: selalu 1 baris di file_blocks. Model B: N baris.
  is_chunked    boolean default false
)

-- Lokasi fisik. Di A: 1 baris/ file. Di B: N baris/ file.
file_blocks (
  id            uuid pk,
  file_id       uuid fk,
  account_id    uuid fk,              -- akun tempat block ini berada
  provider_ref  text,                 -- file-id di provider (mis. GDrive fileId)
  seq           int default 0,        -- urutan chunk (A: 0). B: 0..N-1
  size_bytes    bigint,
  checksum      text
)

-- index bantu
create index on files_index (user_id, name);
create index on file_blocks (file_id, seq);
```
Skema ini **sudah Model-B-ready**: struktur `file_blocks` sama, Model A hanya
selalu mengisi 1 baris per file.

### 2.6 Migration Path A → B (tanpa rewrite)
Yang berubah hanya: (a) implementasi `StorageService` yang di-inject,
(b) pengisian `file_blocks` (1 baris → N baris), (c) tambah modul `Splitter`.
Yang **tidak** berubah: UI, REST API, skema DB, smart-router (tinggal extend), rclone engine.

Prosedur upgrade sebuah file A→B:
1. Baca file utuh via `WholeFileStore`.
2. `Splitter` pecah + (opsional) encrypt jadi N chunk.
3. Router pilih N akun tujuan; engine upload tiap chunk.
4. Tulis N baris `file_blocks`, set `files_index.is_chunked = true`.
5. Hapus objek lama. Reversible.

### 2.7 Tech Stack
| Layer | Pilihan | Alasan |
|-------|---------|--------|
| Frontend | Nuxt 4 | stack existing kamu |
| Backend | Go | performa I/O, cocok wrap rclone (Go-native) |
| Engine | rclone (lib/CLI) | 70+ provider gratis, OAuth & refresh included |
| DB | Postgres / Supabase | stack existing, RLS untuk multi-tenant |
| Deploy | Docker Compose | selaras infra self-hosting kamu |
| Enkripsi token | AES-GCM (key dari env/KMS) | standar, sederhana |

### 2.8 Roadmap
- **Fase 0 — Spike:** rclone bisa list+upload 1 GDrive dari Go. Bukti engine jalan.
- **Fase 1 — MVP Personal (Model A):** connect multi-akun, index, explorer, upload+smart-routing, download, quota. (Target inti.)
- **Fase 2 — Hardening:** enkripsi token, refresh lifecycle, incremental sync, search.
- **Fase 3 — Self-hosted multi-user:** user_id isolation, RLS, Docker Compose rilis.
- **Fase 4 — Model B (opsional):** chunking, redundancy/parity, upgrade file A→B.
