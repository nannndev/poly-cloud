# Poly Cloud

Storage aggregator monorepo berdasar `docs/`: Go API, Nuxt 4 frontend, Postgres
metadata index, dan rclone sebagai engine.

## Quick start

```sh
cp .env.example .env
# Ganti TOKEN_ENC_KEY dan RCLONE_RC_PASS dengan nilai acak sebelum jalan.
make dev
```

- Frontend: http://localhost:3000
- API health: http://localhost:8080/healthz
- Postgres: localhost:5432

## Status implementasi

Model A (whole-file + smart routing) jalan end-to-end, frontend sudah tersambung
ke backend: connect akun, provisioning remote rclone, index & kuota, VFS folder,
upload/download stream-through, migrasi antar akun, dan progres upload lewat SSE.
Model B (chunked) belum — `StorageService` sudah jadi titik tukarnya.

Jika port 3000 di mesin Anda sudah dipakai proses lain pada `::1`, buka
`http://127.0.0.1:3000` — origin itu sudah termasuk di `CORS_ORIGINS` bawaan.

## Menambah akun

**Provider berbasis key (S3 / B2 / R2)** langsung jalan tanpa setup tambahan:

```sh
curl -X POST localhost:8080/api/v1/accounts/connect \
  -H 'Content-Type: application/json' \
  -d '{"provider":"s3","label":"Arsip","fields":{
        "access_key_id":"...","secret_access_key":"...",
        "endpoint":"https://s3.example.com","region":"us-east-1",
        "bucket":"nama-bucket"}}'
```

`bucket` wajib untuk provider penyimpanan objek — akar remote S3 adalah daftar
bucket, bukan tempat menaruh file.

**Provider OAuth (Google Drive / Dropbox / OneDrive)** perlu aplikasi OAuth milik
platform lebih dulu (ADR-010). Daftarkan app di masing-masing provider dengan
redirect URI `http://localhost:3000/connect/callback`, lalu isi
`GOOGLE_CLIENT_ID`/`GOOGLE_CLIENT_SECRET` dkk di `.env`. `GET /api/v1/providers`
melaporkan provider mana yang kredensialnya sudah terisi.

## Arsitektur runtime

```
Frontend (Nuxt) → Backend (Go) → rclone rcd → provider
                        ↓
                   Postgres (index & token terenkripsi)
```

rclone berjalan sebagai daemon RC di jaringan privat compose: port-nya tak pernah
dipublikasikan ke host dan dilindungi basic auth. RC API bersifat all-or-nothing —
siapa pun yang bisa menjangkaunya dapat membaca seluruh kredensial tersimpan
(ADR-011), jadi jangan pernah expose service `rclone` keluar.

## Development

```sh
make dev       # bangun & jalankan seluruh stack
make test      # unit test backend (lewat container Go, tak perlu Go lokal)
make logs      # ikuti log
make down      # hentikan (tambah -v untuk hapus volume)
```

Skema DB dijalankan otomatis oleh image Postgres dari
`apps/backend/migrations/` saat volume data masih kosong. Mengubah skema setelah
volume terbentuk butuh `docker compose down -v` (data hilang) atau migrasi manual.

## Landing page

Halaman publik ada di [`apps/landing`](apps/landing) - Nuxt statis yang
di-deploy ke Vercel, terpisah dari unit self-hosted. Halaman itu tak memanggil
API backend sama sekali.

## Dokumentasi

Desain lengkap ada di [`docs/`](docs/) — arsitektur, data model, API spec,
sequence flow, dan ADR.
