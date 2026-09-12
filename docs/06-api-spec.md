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

### `GET /providers`
Provider yang tersedia + kesiapannya. UI memakai `configured` untuk menandai
provider OAuth yang kredensial platformnya belum diisi di env backend.
```json
[ { "provider":"gdrive","kind":"oauth","configured":false },
  { "provider":"s3","kind":"keys","configured":true,
    "fields":["provider","access_key_id","secret_access_key","region","endpoint","location_constraint","acl","bucket"],
    "required":["access_key_id","secret_access_key","bucket"] } ]
```

### `POST /accounts/connect`
Dua bentuk, tergantung `kind` provider.

**OAuth** (`gdrive`/`dropbox`/`onedrive`) — body: `{ "provider":"gdrive","label":"GDrive utama" }`.
Resp: `{ "auth_url":"https://accounts.google.com/o/oauth2/..." }` → frontend redirect.
Kredensial OAuth platform belum diset → `400 INVALID_ARGUMENT`.

**Berbasis key** (`s3`/`b2`/`r2`) — tanpa redirect; account langsung jadi:
```json
{ "provider":"s3", "label":"Arsip", "fields": {
    "access_key_id":"...", "secret_access_key":"...",
    "endpoint":"https://s3.example.com", "region":"us-east-1",
    "bucket":"nama-bucket" } }
```
Resp `201`: `{ "account_id":"uuid","status":"active","account":{...} }`.
`bucket` wajib: akar remote penyimpanan objek adalah daftar bucket, bukan tempat
menaruh file — nilainya jadi akar penyimpanan account (`accounts.root_path`).

### `POST /accounts/callback`
Tukar code jadi token. Body: `{ "provider":"gdrive","code":"...","state":"..." }`.
Resp: `{ "account_id":"uuid","status":"active" }`. Memicu initial sync.

### `POST /accounts/{id}/sync`
Trigger sync ulang index & kuota account.
Resp: `{ "synced":true,"files_indexed":1234,"total_bytes":0,"used_bytes":0,"free_bytes":0 }`.
Provider tanpa dukungan `about` mengembalikan kuota 0 — bukan kegagalan.

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

### `POST /files/upload?folder_id=&name=&job_id=&size=`
Upload stream. Body boleh multipart **atau** byte mentah (`application/octet-stream`).
Nama file diambil dari `?name=`, header `X-File-Name`, atau nama part multipart.
Tujuan folder dari `?folder_id=` atau `?path=` (kosong = root).

`job_id` opsional: isi dengan id yang sama seperti yang dipakai membuka SSE agar
progres bisa diikuti sejak byte pertama; kalau kosong, backend membuatnya dan
mengembalikannya di header `X-Job-Id`. `size` melengkapi total byte untuk
progress bar saat body multipart (Content-Length tak menggambarkan ukuran file).

Backend memilih account via smart routing. Nama yang bentrok di folder yang sama
di-suffix otomatis (`laporan (1).pdf`).
Resp `201`:
```json
{ "id":"uuid","name":"foto.jpg","account_id":"uuid","account_label":"GDrive-2",
  "routed_by":"most-free","job_id":"...","file":{ ... } }
```
Header respons: `X-Job-Id`, `X-Account-Label`, `X-Routed-By`.
Gagal muat semua account (Model A): `409 { "error":{ "code":"NO_ROOM","message":"..." } }`.

### `GET /files/{id}/download?inline=`
Stream file (stream-through). Resp: binary + `Content-Disposition: attachment`.
`?inline=1` mengubahnya jadi `inline`, sehingga browser merender isinya
(pratinjau PDF/gambar/video) alih-alih memicu unduhan.

### `POST /files/{id}/move`
Pindah **fisik antar account** (transfer data). Body: `{ "dest_account_id":"uuid" }`.
Resp: `{ "moved":true }`. Catatan: ini beda dari pindah folder (di bawah), yang hanya
mengubah organisasi virtual tanpa transfer data.

### `PATCH /files/{id}`
Ubah organisasi virtual file (rename / pindah folder). Hanya update DB, **tak sentuh provider**.
Body: `{ "name":"baru.pdf", "folder_id":"uuid-or-null" }` (field opsional).
Resp: `{ "id":"uuid","name":"baru.pdf","virtual_path":"/Kerjaan/baru.pdf" }`.

### `DELETE /files/{id}`
Hapus file (dan block-nya). Resp: `204`.

---

## Folders (VFS)
Folder virtual — hanya di DB, tak dibuat di provider. Lihat [doc 09](09-virtual-filesystem.md).
Semua operasi di sini transaksi DB murni (instan, tanpa transfer data).

### `GET /folders?parent_id=`
List subfolder di bawah `parent_id` (kosong = root).
```json
[ { "id":"uuid","name":"Kerjaan","path":"/Kerjaan","parent_id":null } ]
```

### `POST /folders`
Buat folder. Body: `{ "name":"Kerjaan","parent_id":null }`.
Resp: `{ "id":"uuid","name":"Kerjaan","path":"/Kerjaan","parent_id":null }`.
Path duplikat → `409 { "error":{ "code":"PATH_EXISTS" } }`.

### `PATCH /folders/{id}`
Rename atau pindah folder. Body: `{ "name":"...", "parent_id":"uuid-or-null" }` (opsional).
Backend recompute `path` folder ini + seluruh turunannya. Resp: folder terbaru.

### `DELETE /folders/{id}?recursive=false`
Hapus folder. `recursive=false` → tolak jika folder tak kosong (`409 FOLDER_NOT_EMPTY`).
`recursive=true` → hapus semua isi (file fisik di provider ikut terhapus) + subfolder.
Resp: `204`.

---

## Settings

### `GET /settings`
Konfigurasi runtime backend (read-only). Dibaca dari environment saat proses start;
mengubahnya berarti menyunting `.env` lalu me-restart layanan.
```json
{ "routing_strategy":"most-free","remote_base_dir":"PolyCloud",
  "storage_model":"A","chunking":false,"sync_recurse":true,"multi_user":false }
```

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
| `NOT_FOUND` | 404 | file/account/folder tak ada |
| `RATE_LIMITED` | 429 | provider rate-limit |
| `PATH_EXISTS` | 409 | folder/file dgn path sama sudah ada |
| `FOLDER_NOT_EMPTY` | 409 | hapus folder berisi tanpa `recursive=true` |
| `INVALID_ARGUMENT` | 400 | body/param tak valid, provider belum dikonfigurasi |
| `UNSUPPORTED` | 501 | operasi belum didukung impl aktif (mis. file chunked) |
| `INTERNAL` | 500 | kesalahan tak terduga |
