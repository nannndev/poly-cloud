# Fase 0 Spike — rclone + Go

Tujuan: buktikan rclone bisa di-drive dari Go. Connect banyak akun, baca quota,
list file, simulasi smart routing, dan upload beneran. **Zero dependency** (stdlib saja) —
rclone dipanggil sebagai subprocess.

## 1. Prasyarat
- Go 1.22+  → https://go.dev/dl
- rclone    → https://rclone.org/downloads (atau `sudo apt install rclone` / `brew install rclone`)

Cek:
```bash
go version
rclone version
```

## 2. Connect akun (sekali per akun)
rclone yang urus OAuth. Ulangi untuk tiap akun yang mau digabung.

```bash
rclone config
# n) New remote
# name> gdrive1        (nama bebas, unik per akun — mis. gdrive1, gdrive2, dropbox1)
# Storage> drive       (pilih Google Drive; untuk lain: dropbox, onedrive, s3, b2)
# ...ikuti wizard, browser kebuka untuk login/consent...
# y) Yes this is OK
```

Tambah akun kedua GDrive? Ulangi dengan `name> gdrive2` pakai email berbeda.

Cek remote yang terdaftar:
```bash
rclone listremotes
# gdrive1:
# gdrive2:
# dropbox1:
```

## 3. Jalankan spike
```bash
cd spike
go run .                         # inspeksi: quota + list + simulasi routing tiap akun
go run . ~/Downloads/foto.jpg    # + upload beneran ke akun yang dipilih router
```

## Contoh output
```
Ditemukan 3 remote: [gdrive1: gdrive2: dropbox1:]

[gdrive1:] total=15.0 GB used=14.2 GB free=820.0 MB
    12 item di root. Contoh:
      - [dir] Documents (0 B)
      - [file] cv.pdf (240.1 KB)
[gdrive2:] total=15.0 GB used=1.1 GB free=13.9 GB
[dropbox1:] total=2.0 GB used=0.3 GB free=1.7 GB

=== AGREGAT SEMUA AKUN ===
Kapasitas total: 32.0 GB | Terpakai: 15.6 GB | Bebas: 16.4 GB

=== SIMULASI SMART ROUTING (most-free) ===
Upload 50.0 MB -> pilih gdrive2: (free 13.9 GB)
Upload 2.0 GB  -> pilih gdrive2: (free 13.9 GB)
Upload 20.0 GB -> no single account has enough free space (Model B akan pecah & sebar)
```

## Yang dibuktikan spike ini
- ✅ Multi-akun/multi-provider kebaca dari satu program (US-1, US-2)
- ✅ Quota agregat lintas akun (US-6) — ini value inti platform
- ✅ Smart routing memilih akun tujuan otomatis (US-4, FR-5)
- ✅ `ErrNoRoom` = titik masuk Model B nanti (file > kapasitas 1 akun)
- ✅ Upload/download stream-through tanpa simpan permanen di server (UploadStream/Download)

## File
- `engine.go`  — wrapper rclone (List, About, Upload, UploadStream, Download). Ini "Engine" di arsitektur.
- `router.go`  — smart routing Model A (most-free-space) + `ErrNoRoom`.
- `main.go`    — demo end-to-end.

## Langkah setelah spike hijau
1. Bungkus `engine` + `router` di balik interface `StorageService` (lihat BRD §2.3).
2. Ganti "baca quota via About tiap kali" → cache ke Postgres (`accounts`, `files_index`).
3. Bikin REST API (Go) + explorer Nuxt di atasnya.
