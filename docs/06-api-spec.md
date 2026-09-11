# 06 — API Specification (REST)

Base path: `/api/v1`. Format: JSON. Auth v1: single-user (session/token sederhana);
multi-user (v2) menambah scoping `user_id` dari sesi.

## Konvensi
- Sukses: `2xx` + body JSON.
- Error: `{ "error": { "code": "...", "message": "..." } }` + status sesuai.
- Waktu: ISO-8601 UTC. Ukuran: bytes (integer).

---

## Accounts

### `GET /accounts`
List account terhubung + kuota.
```json
[
  { "id":"uuid","provider":"gdrive","label":"GDrive utama","status":"active",
    "total_bytes":16106127360,"used_bytes":15246680064,"free_bytes":859447296,
    "last_synced":"2026-01-10T08:00:00Z" }
]
```

### `POST /accounts/connect`
Mulai OAuth. Body: `{ "provider":"gdrive","label":"GDrive utama" }`.
Resp: `{ "auth_url":"https://accounts.google.com/o/oauth2/..." }` → frontend redirect.

### `POST /accounts/callback`
Tukar code jadi token. Body: `{ "provider":"gdrive","code":"...","state":"..." }`.
Resp: `{ "account_id":"uuid","status":"active" }`. Memicu initial sync.

### `POST /accounts/{id}/sync`
Trigger sync ulang index & kuota account. Resp: `{ "synced":true,"files_indexed":1234 }`.

### `DELETE /accounts/{id}`
Cabut account (hapus token & index terkait). Resp: `204`.

---

## Files

### `GET /files?path=/&sort=name&page=1`
List file di path terpadu (dari DB cache).
```json
{ "path":"/","items":[
    { "id":"uuid","name":"foto.jpg","mime":"image/jpeg","size_bytes":2400000,
      "modified_at":"2026-01-09T10:00:00Z","account_id":"uuid","account_label":"GDrive-2",
      "is_chunked":false }
  ], "page":1,"total":42 }
```

### `GET /files/search?q=laporan&type=pdf&min_size=0&account_id=`
Search lintas account. Resp: sama seperti list `items`.

### `POST /files/upload?path=/`
Upload stream. Body: multipart/stream file. Backend pilih account (smart routing).
Header respons boleh menyertakan account terpilih. Progres via SSE (lihat Events).
Resp akhir:
```json
{ "id":"uuid","name":"foto.jpg","account_id":"uuid","account_label":"GDrive-2","routed_by":"most-free" }
```
Gagal muat semua account (Model A): `409 { "error":{ "code":"NO_ROOM","message":"..." } }`.

### `GET /files/{id}/download`
Stream file (stream-through). Resp: binary + `Content-Disposition`.

### `POST /files/{id}/move`
Pindah antar account. Body: `{ "dest_account_id":"uuid" }`. Resp: `{ "moved":true }`.

### `DELETE /files/{id}`
Hapus file (dan block-nya). Resp: `204`.

---

## Quota

### `GET /quota`
Agregat + per account.
```json
{ "aggregate":{ "total_bytes":34359738368,"used_bytes":16777216000,"free_bytes":17582522368 },
  "accounts":[ { "account_id":"uuid","label":"GDrive utama","total_bytes":16106127360,
                 "used_bytes":15246680064,"free_bytes":859447296 } ] }
```

---

## Events (SSE)

### `GET /events/uploads/{jobId}`
Stream progres upload.
```
event: progress
data: { "job":"...", "bytes":1048576, "total":2400000 }

event: done
data: { "job":"...", "file_id":"uuid", "account_label":"GDrive-2" }
```

---

## Error Codes
| code | status | arti |
|------|--------|------|
| `NO_ROOM` | 409 | tak ada account muat (Model A) |
| `ACCOUNT_NEEDS_RECONNECT` | 401 | token invalid, minta re-auth |
| `PROVIDER_ERROR` | 502 | error dari provider/engine |
| `NOT_FOUND` | 404 | file/account tak ada |
| `RATE_LIMITED` | 429 | provider rate-limit |
