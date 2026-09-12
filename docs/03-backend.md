# 03 — Backend Design (Go)

## 1. Layering
```
HTTP Handlers  →  StorageService  →  { Router, Index, Engine }
                        ↑
                   Auth/Token (cross-cutting)
```
- **Handlers:** parse request, auth check, panggil service, stream response (SSE untuk progress).
- **StorageService:** interface stabil; orkestrasi operasi. Impl A = `WholeFileStore`.
- **Router:** pilih account tujuan upload.
- **Index:** baca/tulis metadata file di DB.
- **Engine:** adapter rclone (eksekusi operasi file per-provider).
- **Auth/Token:** OAuth flow, enkripsi token, refresh lifecycle.

## 2. Interface Inti
```go
type StorageService interface {
    List(ctx context.Context, userID, path string) ([]FileEntry, error)
    Upload(ctx context.Context, userID, path string, r io.Reader, size int64) (FileEntry, error)
    Download(ctx context.Context, userID, fileID string) (io.ReadCloser, error)
    Move(ctx context.Context, userID, fileID, destAccountID string) error
    Delete(ctx context.Context, userID, fileID string) error
    Search(ctx context.Context, userID string, q SearchQuery) ([]FileEntry, error)
    Quota(ctx context.Context, userID string) ([]AccountQuota, error)
}
```
`WholeFileStore` (Model A) dan `ChunkedStore` (Model B, nanti) sama-sama memenuhi interface ini.
Yang dipakai di-inject saat startup — mengganti model = mengganti implementasi, bukan API.

## 3. Engine Adapter
```go
type Engine interface {
    // Provisioning remote (Opsi A — lihat doc 10)
    CreateRemote(ctx context.Context, name, rType string, params map[string]any) error
    UpdateRemote(ctx context.Context, name string, params map[string]any) error // refresh token
    DeleteRemote(ctx context.Context, name string) error
    // Operasi file
    ListRemotes(ctx context.Context) ([]string, error)
    List(ctx context.Context, remote, path string, recurse bool) ([]RemoteEntry, error)
    Stat(ctx context.Context, remote, path string) (*RemoteEntry, error)
    About(ctx context.Context, remote string) (Quota, error)
    Mkdir(ctx context.Context, remote, path string) error
    UploadStream(ctx context.Context, r io.Reader, remote, destDir, name string) error
    Download(ctx context.Context, remote, src string, w io.Writer) error
    Move(ctx context.Context, srcRemote, src, dstRemote, dst string) error
    Delete(ctx context.Context, remote, path string) error
    Ping(ctx context.Context) error
}
```
Parameter `remote` di operasi file menerima **fs target**, bukan sekadar nama remote:
`acc_x` untuk provider drive, `acc_x:bucket` untuk provider penyimpanan objek
(`Account.FsTarget()`). Operasi `config/*` tetap memakai nama remote mentah.
```go
```
Impl produksi = `RcloneDaemonEngine` yang bicara ke **`rclone rcd`** (RC API) — ini yang
memungkinkan provisioning remote dinamis (`config/create`) dari OAuth backend. Daemon wajib
dikunci ketat (localhost + auth, tak expose keluar) — lihat [ADR-011](08-adr.md#adr-011)
& [doc 10 §6](10-account-provisioning.md).

> Catatan: spike fase 0 memakai `RcloneEngine` (CLI subprocess) hanya untuk PoC. Interface
> `Engine` sama, jadi pindah CLI → daemon tak menyentuh StorageService.

## 4. Smart Routing
```go
type Router struct{ Strategy string } // "most-free" (default) | "round-robin" | "by-mime"

func (r *Router) Pick(ctx, accounts []Account, size int64) (*Account, error)
```
- Filter account `active` & `free >= size`.
- Skor sesuai strategi. Default **most-free** (isi account terlowong → distribusi merata).
- Tak ada yang muat → `ErrNoRoom`. Model A: tolak/minta pilih. Model B: trigger chunk.

## 5. Token Lifecycle
- Simpan `access` + `refresh` token **terenkripsi** (AES-GCM) di `account_tokens`, terpisah dari `accounts`.
- Backend (bukan rclone) yang berwenang refresh — token milik platform (Opsi A).
- Sebelum operasi: cek `expires_at`; jika lewat → backend refresh via endpoint provider →
  simpan ulang ke DB **dan** `Engine.UpdateRemote` agar rclone tak memakai token basi.
- Refresh gagal permanen → set account `status = needs_reconnect`; UI minta re-auth.
- Provisioning saat connect: `Engine.CreateRemote(acc_<uuid>, type, {token, client_id, client_secret})`.
  Lihat [doc 10](10-account-provisioning.md).

## 6. Sinkronisasi Index
- **Initial sync** saat connect: walk isi account → tulis `files_index` + `file_blocks`.
- **Incremental sync** berkala / on-demand: bandingkan `modified_at`/delta bila provider mendukung.
- Quota di-cache di `accounts` (refresh saat sync / TTL), agar routing & dashboard cepat.

## 7. Upload / Download Stream-Through
- Upload: body request (io.Reader) → `Engine.UploadStream` (rcat) → tak ada file di disk server.
- Download: `Engine.Download` (cat) → pipe ke response writer. Buffer terbatas, bukan full-load.
- Progress: SSE channel melaporkan bytes transferred ke frontend.

## 8. Error & Reliability
- Bungkus error engine dengan konteks (remote, op).
- Retry backoff untuk transient (rate-limit/network); jangan retry untuk auth/permission.
- Operasi idempotent bila mungkin (mis. upsert index by provider_ref).

## 9. Konfigurasi (env)
`DATABASE_URL`, `TOKEN_ENC_KEY`, `OAUTH_REDIRECT_URL`, `CORS_ORIGINS`, `PORT`,
`LOG_LEVEL`, `ROUTING_STRATEGY`, `RCLONE_RC_URL`/`RCLONE_RC_USER`/`RCLONE_RC_PASS`,
`RCLONE_BASE_DIR`, dan pasangan client id/secret per provider OAuth.
Lihat `.env.example`.

`TOKEN_ENC_KEY` boleh berapa pun panjangnya (minimal 16 karakter); kunci AES-256
diturunkan darinya lewat SHA-256.

## 10. Struktur Paket
```
internal/domain    tipe inti + sentinel error (tanpa dependensi internal)
internal/httpapi   handlers, middleware, router, SSE hub
internal/storage   Service, WholeFileStore, AccountService, Deps
internal/engine    Engine interface + RcloneDaemon (RC API)
internal/routing   Router
internal/index     repo DB (accounts, folders, files, blocks)
internal/auth      oauth provider spec, token crypto, lifecycle
internal/config    env
```
`internal/domain` dipisah supaya `engine` bisa memetakan error ke sentinel tanpa
bergantung pada `storage` — menghindari import cycle.
