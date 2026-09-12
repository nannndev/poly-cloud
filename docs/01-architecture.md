# 01 — Software Architecture Document (SAD)

## 1. Tujuan & Konteks
Menghilangkan fragmentasi storage: user punya banyak akun cloud gratis dengan kapasitas
total besar namun terpisah-pisah. Platform ini menyatukannya jadi satu view: browse,
search, upload (auto-route), download — tanpa peduli file fisik ada di akun/provider mana.

## 2. Prinsip Arsitektur
1. **Engine-agnostic.** Operasi file disembunyikan di balik interface; rclone hanya salah satu implementasi.
2. **Storage-model-agnostic.** UI & API tak tahu file itu whole (A) atau chunked (B).
3. **Metadata di DB, file di provider.** Server tak menyimpan file permanen (stream-through).
4. **Multi-tenant sejak awal.** `user_id` ada di skema walau v1 single-user.
5. **Self-hostable.** Seluruh sistem jalan via Docker Compose.

## 3. C4 — Level 1: System Context
```mermaid
graph TB
    user([User])
    subgraph platform[Storage Aggregator Platform]
        sys[Aggregator System]
    end
    gdrive[(Google Drive)]
    dropbox[(Dropbox)]
    onedrive[(OneDrive)]
    s3[(S3 / Backblaze B2)]

    user -->|browse, search, upload, download| sys
    sys -->|OAuth + file ops| gdrive
    sys -->|OAuth + file ops| dropbox
    sys -->|OAuth + file ops| onedrive
    sys -->|keys + file ops| s3
```

## 4. C4 — Level 2: Container
```mermaid
graph TB
    user([User])
    subgraph mono[Monorepo Deploy Unit]
        fe[Frontend<br/>Nuxt 4 SPA]
        be[Backend API<br/>Go]
        db[(Postgres<br/>Supabase)]
        engine[rclone Engine<br/>rcd daemon, RC API]
    end
    providers[(Cloud Providers<br/>GDrive/Dropbox/OneDrive/S3/B2)]

    user -->|HTTPS| fe
    fe -->|REST + SSE| be
    be -->|SQL| db
    be -->|exec ops| engine
    engine -->|API calls| providers
```

## 5. C4 — Level 3: Component (Backend)
```mermaid
graph TB
    subgraph be[Backend Go]
        http[HTTP Handlers<br/>REST + SSE]
        svc[StorageService<br/>interface]
        whole[WholeFileStore<br/>Model A impl]
        chunk[ChunkedStore<br/>Model B impl - future]
        router[Router<br/>smart routing]
        index[Index Repo<br/>DB metadata]
        auth[Auth / Token<br/>lifecycle + crypto]
        eng[Engine adapter<br/>rclone]
    end

    http --> svc
    svc --> whole
    svc -. future .-> chunk
    whole --> router
    whole --> index
    whole --> eng
    chunk -. future .-> router
    auth --> eng
    auth --> index
```

## 6. Aliran Data Inti
- **Connect:** User → FE → BE (`/accounts/connect`) → OAuth via engine → token disimpan (encrypted) di DB.
- **Index sync:** BE `List` per account via engine → tulis metadata ke `files_index`/`file_blocks`.
- **Browse/Search:** FE → BE → baca dari DB (cepat, tak hit provider) → render explorer.
- **Upload:** FE stream → BE → Router pilih account → engine `UploadStream` → catat index.
- **Download:** FE minta → BE engine `Download` (stream-through) → pipe ke FE.

## 7. Model A → Model B (Scalability)
Titik abstraksi: `StorageService`. Skema DB `file_blocks` sudah menampung 1..N block per file.

| Elemen | Model A | Model B | Berubah? |
|--------|---------|---------|----------|
| UI / REST API | sama | sama | ❌ |
| Skema DB | 1 block/file | N block/file | ❌ (struktur), ✅ (isi) |
| StorageService impl | WholeFileStore | ChunkedStore | ✅ (tambah impl) |
| Router | pick 1 account | pick N account | ➕ (extend) |
| Modul baru | — | Splitter + reassembler | ➕ |
| Engine (rclone) | sama | sama | ❌ |

Upgrade file A→B = baca utuh → pecah → sebar → tulis N `file_blocks` → set `is_chunked=true`. Reversible.

## 8. Kualitas Atribut
- **Security:** token AES-GCM at-rest; scope OAuth minimal; file tak persist di server.
- **Performance:** browse/search dari DB cache < 2 dtk; upload/download stream (bukan buffer penuh).
- **Extensibility:** provider baru = 1 remote rclone, tanpa sentuh UI/DB.
- **Portability:** Docker Compose, env-driven config.
- **Reliability:** operasi idempotent; retry backoff pada error transient; incremental sync.

## 9. Batasan & Asumsi
- rclone berjalan sebagai daemon RC terpisah di jaringan privat, hanya dijangkau backend (ADR-011).
- Provider yang dipakai mendukung `about` (quota) — jika tidak, kapasitas ditandai unknown.
- v1: operasi manajemen (bukan sync-daemon dua arah realtime).
