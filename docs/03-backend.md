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
    ListRemotes(ctx context.Context) ([]string, error)
    List(ctx context.Context, remote, path string) ([]FileEntry, error)
    About(ctx context.Context, remote string) (Quota, error)
    UploadStream(ctx context.Context, r io.Reader, remote, dest string) error
    Download(ctx context.Context, remote, src string, w io.Writer) error
    Move(ctx context.Context, srcRemote, src, dstRemote, dst string) error
    Delete(ctx context.Context, remote, path string) error
}
```
Impl v1 = `RcloneEngine` (subprocess). Membungkus rclone di balik interface memungkinkan
ganti ke library mode atau engine lain tanpa menyentuh service.

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
- Sebelum operasi: cek `expires_at`; jika lewat → refresh via engine → simpan ulang.
- Refresh gagal permanen → set account `status = needs_reconnect`; UI minta re-auth.

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
`DATABASE_URL`, `RCLONE_CONFIG_PATH`, `TOKEN_ENC_KEY` (32-byte), `OAUTH_REDIRECT_URL`,
`PORT`, `LOG_LEVEL`. Lihat `.env.example`.

## 10. Struktur Paket
```
internal/http      handlers, middleware, sse
internal/storage   StorageService, WholeFileStore
internal/engine    RcloneEngine
internal/routing   Router
internal/index     repo DB (accounts, files, blocks)
internal/auth      oauth, token crypto, lifecycle
internal/config    env
```
