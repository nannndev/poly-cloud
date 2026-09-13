# Poly Cloud

**Satu ruang, semua drive.** Poly Cloud menyatukan Google Drive, Dropbox,
OneDrive, S3, R2, dan B2 jadi satu drive virtual. Unggahan diarahkan otomatis ke
akun dengan sisa ruang terbanyak — Anda tak perlu memikirkan tujuannya.

Self-hosted sepenuhnya: berkas tetap tinggal di provider asalnya, dan tak ada
salinan yang mengendap di server Anda.

| Komponen | Teknologi |
|---|---|
| Backend | Go 1.25 |
| Frontend | Nuxt 4 + Nuxt UI |
| Index metadata | PostgreSQL 16 |
| Engine berkas | rclone (daemon RC) |

---

## Daftar isi

- [Prasyarat](#prasyarat)
- [Setup](#setup)
- [Menjalankan dengan Docker](#menjalankan-dengan-docker) ← cara yang disarankan
- [Menjalankan tanpa Docker](#menjalankan-tanpa-docker)
- [Menambah akun penyimpanan](#menambah-akun-penyimpanan)
- [Konfigurasi](#konfigurasi)
- [Pengembangan](#pengembangan)
- [Pemecahan masalah](#pemecahan-masalah)
- [Arsitektur](#arsitektur)

---

## Prasyarat

**Jalur Docker** — hanya butuh satu hal:

- Docker + Docker Compose ([Docker Desktop](https://docs.docker.com/desktop/)
  sudah memuat keduanya)

**Jalur tanpa Docker** — empat hal, dipasang sendiri:

- Go 1.25 atau lebih baru
- Node.js 20 atau lebih baru
- PostgreSQL 16
- rclone

---

## Setup

Langkah ini sama untuk kedua jalur.

```sh
git clone https://github.com/nannndev/poly-cloud
cd poly-cloud
cp .env.example .env
```

Lalu **ganti dua nilai** di `.env` sebelum menjalankan apa pun:

```sh
# Hasilkan nilai acak
openssl rand -hex 32   # untuk TOKEN_ENC_KEY
openssl rand -hex 24   # untuk RCLONE_RC_PASS
```

```ini
TOKEN_ENC_KEY=<hasil perintah pertama>
RCLONE_RC_PASS=<hasil perintah kedua>
```

Keduanya berisi `change-me-...` secara bawaan dan **tidak aman dipakai**:

- `TOKEN_ENC_KEY` mengenkripsi token provider Anda at-rest (AES-256-GCM).
  Minimal 16 karakter — backend menolak start bila kurang.
- `RCLONE_RC_PASS` melindungi daemon rclone, yang memegang seluruh kredensial
  penyimpanan Anda.

> **Mengganti `TOKEN_ENC_KEY` setelah ada akun terhubung** membuat token lama tak
> bisa didekripsi — setiap akun OAuth harus dihubungkan ulang lewat consent.
> Tetapkan sekali di awal.

---

## Menjalankan dengan Docker

Cara yang disarankan. Satu perintah menjalankan keempat service:

```sh
make dev
```

Tunggu sampai build selesai, lalu buka:

| | Alamat |
|---|---|
| **Frontend** | http://localhost:3000 |
| API health | http://localhost:8080/healthz |
| Postgres | localhost:5432 |

`/healthz` menjawab `{"db":"ok","engine":"ok","status":"ok"}` bila semuanya
sehat.

### Perintah lain

```sh
make logs                    # ikuti log seluruh service
make down                    # hentikan (volume tetap ada)
docker compose down -v       # hentikan + hapus seluruh data
make test                    # unit test backend, lewat container Go
make fmt                     # gofmt -w, juga lewat container
```

`make test` dan `make fmt` berjalan di dalam container Go, jadi Anda tak perlu
memasang Go untuk berkontribusi ke backend.

### Menjalankan ulang setelah mengubah kode

```sh
docker compose up -d --build backend    # hanya backend
docker compose up -d --build frontend   # hanya frontend
```

---

## Menjalankan tanpa Docker

Berguna saat Anda ingin iterasi cepat di backend atau frontend tanpa menunggu
image dibangun ulang. Butuh **empat terminal**.

### 1. PostgreSQL

Pasang dan jalankan Postgres, lalu buat user, database, dan skemanya:

```sh
createuser -s polycloud
createdb -O polycloud polycloud
psql -d polycloud -c "ALTER USER polycloud PASSWORD 'polycloud';"
psql -d polycloud -f apps/backend/migrations/001_init.sql
```

Nama dan password di atas mengikuti `.env.example`, jadi `DATABASE_URL` di
langkah 3 bisa dipakai apa adanya. Bila Anda memakai kredensial lain,
sesuaikan string koneksinya.

### 2. rclone

rclone berjalan sebagai daemon RC — backend berbicara dengannya lewat HTTP,
bukan memanggil perintah rclone.

```sh
rclone rcd \
  --rc-addr=127.0.0.1:5572 \
  --rc-user=polycloud \
  --rc-pass="<RCLONE_RC_PASS dari .env Anda>" \
  --rc-serve
```

> `--rc-addr=127.0.0.1` bukan `0.0.0.0`. RC API bersifat **all-or-nothing**:
> siapa pun yang bisa menjangkaunya dapat membaca seluruh kredensial penyimpanan
> Anda (ADR-011). Jangan pernah membukanya ke jaringan.

`--rc-serve` wajib — tanpa itu unduhan dan pratinjau berkas tak berfungsi.

### 3. Backend

Dua nilai di `.env` menunjuk nama host Docker (`postgres`, `rclone`), jadi harus
ditimpa ke `localhost`. Muat seluruh `.env` lalu timpa keduanya:

```sh
cd apps/backend

set -a && source ../../.env && set +a

DATABASE_URL="postgres://polycloud:polycloud@localhost:5432/polycloud?sslmode=disable" \
RCLONE_RC_URL="http://localhost:5572" \
go run ./cmd/api
```

Variabel lain — `TOKEN_ENC_KEY`, `RCLONE_RC_PASS`, kredensial OAuth — ikut
terbaca dari `.env` lewat `source` di atas.

Backend mendengarkan di `:8080` dan membuat user default otomatis saat start.
Tandanya berhasil:

```
{"level":"INFO","msg":"polycloud API listening","port":"8080","routing":"most-free"}
```

### 4. Frontend

```sh
cd apps/frontend
npm install        # sekali saja
npm run dev
```

Buka http://localhost:3000.

### Ringkasan perbedaan konfigurasi

| Variabel | Docker | Tanpa Docker |
|---|---|---|
| `DATABASE_URL` | `...@postgres:5432/...` | `...@localhost:5432/...` |
| `RCLONE_RC_URL` | `http://rclone:5572` | `http://localhost:5572` |

Sisanya sama persis.

---

## Menambah akun penyimpanan

Klik **Connect New Account** di halaman Connected Accounts. Ada dua jenis
provider dengan alur berbeda.

### Provider berbasis kunci — S3, R2, B2

Langsung jalan tanpa setup tambahan. Isi access key, secret, dan **nama bucket**.

Bucket wajib: akar sebuah remote penyimpanan objek adalah *daftar bucket*, bukan
tempat menaruh berkas.

Lewat API:

```sh
curl -X POST localhost:8080/api/v1/accounts/connect \
  -H 'Content-Type: application/json' \
  -d '{"provider":"s3","label":"Arsip","fields":{
        "access_key_id":"...","secret_access_key":"...",
        "endpoint":"https://s3.example.com","region":"us-east-1",
        "bucket":"nama-bucket"}}'
```

### Provider OAuth — Google Drive, Dropbox, OneDrive

Perlu aplikasi OAuth milik Anda sendiri lebih dulu (ADR-010). Daftarkan aplikasi
di masing-masing provider dengan redirect URI:

```
http://localhost:3000/connect/callback
```

Lalu isi kredensialnya di `.env`:

```ini
GOOGLE_CLIENT_ID=...
GOOGLE_CLIENT_SECRET=...
```

Restart backend agar terbaca. Halaman Connect menandai provider mana yang
kredensialnya sudah terisi — yang belum akan menolak otorisasi, jadi tombolnya
dinonaktifkan dengan penjelasan.

Setelah terhubung, jalankan **Sync** untuk mengindeks isi akun.

---

## Konfigurasi

Seluruh pengaturan dibaca dari lingkungan saat backend start. Mengubahnya berarti
menyunting `.env` lalu me-restart — halaman Settings menampilkan nilainya, bukan
menyuntingnya.

| Variabel | Bawaan | Keterangan |
|---|---|---|
| `TOKEN_ENC_KEY` | — | **Wajib.** Kunci enkripsi token, min. 16 karakter |
| `DATABASE_URL` | — | **Wajib.** String koneksi Postgres |
| `RCLONE_RC_PASS` | — | Password daemon rclone |
| `RCLONE_RC_URL` | `http://rclone:5572` | Alamat daemon RC |
| `RCLONE_BASE_DIR` | `PolyCloud` | Folder fisik di tiap akun |
| `ROUTING_STRATEGY` | `most-free` | `most-free` atau `round-robin` |
| `PORT` | `8080` | Port backend |
| `CORS_ORIGINS` | `http://localhost:3000` | Origin yang diizinkan, dipisah koma |
| `OAUTH_REDIRECT_URL` | `http://localhost:3000/connect/callback` | Halaman callback |
| `SYNC_RECURSE` | `true` | Telusuri subfolder saat sync |
| `NUXT_PUBLIC_API_BASE` | `http://localhost:8080/api/v1` | Alamat API bagi frontend |

---

## Pengembangan

```
apps/backend     Go API — handler HTTP, storage service, engine rclone, index
apps/frontend    Nuxt 4 — antarmuka yang dipakai sehari-hari
apps/landing     Halaman publik, statis, di-deploy terpisah ke Vercel
docs/            Arsitektur, data model, spesifikasi API, dan ADR
```

Sebelum perubahan besar, `docs/` layak dibaca — `docs/08-adr.md` mencatat alasan
di balik beberapa keputusan yang tampak ganjil tanpa konteksnya.

### Menjalankan pemeriksaan

```sh
# Backend — lewat container, tak perlu Go lokal
make test

# Backend — bila Go terpasang
cd apps/backend && gofmt -l . && go vet ./... && go test ./...

# Frontend
cd apps/frontend
npx vue-tsc --noEmit -p .nuxt/tsconfig.json
npx nuxt build
```

Panduan kontribusi lengkap ada di [CONTRIBUTING.md](CONTRIBUTING.md).

### Landing page

```sh
cd apps/landing
npm install
npm run dev        # http://localhost:3100
```

Port 3100 dipakai agar tak bentrok dengan aplikasi utama di 3000. Halaman ini
statis dan tak pernah memanggil API backend.

---

## Pemecahan masalah

**Port 3000 sudah dipakai proses lain**
Buka `http://127.0.0.1:3000` — origin itu sudah termasuk di `CORS_ORIGINS`
bawaan.

**`/healthz` melaporkan `engine` degraded**
Backend tak bisa menjangkau daemon rclone. Pada mode Docker, periksa
`docker compose ps` — service `rclone` harus `healthy`. Pada mode tanpa Docker,
pastikan `rclone rcd` berjalan dan `RCLONE_RC_URL` menunjuk `localhost:5572`.

**`TOKEN_ENC_KEY wajib diisi (minimal 16 karakter)`**
`.env` belum disalin, atau nilainya masih placeholder. Lihat [Setup](#setup).

**Mengubah skema database tak berpengaruh**
Migrasi hanya berjalan otomatis saat volume Postgres masih kosong. Setelah
volume terbentuk, jalankan `docker compose down -v` (data hilang) atau terapkan
perubahan secara manual lewat `psql`.

**Berkas hasil sync lama tak bisa diunduh**
Baris index yang dibuat versi terdahulu menyimpan referensi objek dengan format
berbeda. Jalankan **Resync All Accounts** di halaman Settings sekali untuk
memperbaruinya.

**Pratinjau terasa lambat terbuka**
Wajar. Byte dialirkan langsung dari provider saat itu juga — tak ada cache di
server — jadi jeda beberapa detik untuk berkas besar adalah latensi jaringan ke
provider, bukan kemacetan.

---

## Arsitektur

```
Frontend (Nuxt) → Backend (Go) → rclone rcd → provider
                       ↓
                  Postgres (index & token terenkripsi)
```

Beberapa hal yang membentuk desainnya:

**Berkas tak pernah singgah di server.** Unggahan dan unduhan dialirkan langsung
antara peramban dan provider. Database hanya menyimpan metadata: nama, ukuran,
dan di akun mana sebuah berkas berada.

**Folder bersifat virtual.** Struktur folder hidup di database platform, bukan di
provider. Memindahkan berkas antar folder hanya mengubah satu baris — tak ada
data yang berpindah.

**Satu berkas disimpan utuh di satu akun** (Model A). Pemecahan berkas lintas
akun (Model B) belum tersedia, jadi berkas yang lebih besar daripada sisa ruang
akun terlega akan ditolak, bukan dipecah diam-diam. `StorageService` sudah
menjadi titik tukarnya bila nanti diimplementasikan.

**rclone tak pernah terekspos.** Pada mode Docker, port-nya tidak dipublikasikan
ke host dan hanya dijangkau dari jaringan privat compose. RC API bersifat
all-or-nothing — siapa pun yang bisa menjangkaunya dapat membaca seluruh
kredensial tersimpan (ADR-011).

Desain lengkap ada di [`docs/`](docs/): arsitektur, data model, spesifikasi API,
sequence flow, dan ADR.
