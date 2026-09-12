# Poly Cloud - Landing Page

Halaman publik untuk Poly Cloud. Dibangun sepenuhnya statis dan
**tidak memanggil API backend sama sekali**: backend Poly Cloud bersifat
self-hosted sehingga tak dapat dijangkau dari halaman publik.

## Pengembangan

```sh
npm install
npm run dev        # http://localhost:3100
```

Port 3100 dipakai agar tidak bentrok dengan aplikasi di `apps/frontend`
yang memakai port 3000.

## Build

```sh
npm run generate   # keluaran statis di .output/public
npx serve .output/public
```

## Deploy ke Vercel

Buat proyek baru di Vercel dari repo ini, lalu setel:

| Pengaturan | Nilai |
|---|---|
| Root Directory | `apps/landing` |
| Framework Preset | Nuxt.js |
| Build Command | `npm run generate` |
| Output Directory | `.output/public` |

`vercel.json` di folder ini sudah memuat build command, output directory,
dan header keamanan, jadi umumnya cukup menyetel Root Directory saja.

## Tangkapan layar

Berkas gambar ada di `public/shots/`. Ketiganya sudah disensor dari
identitas pribadi sebelum dimasukkan ke repo. Bila menambah gambar baru,
periksa lebih dulu bahwa tidak ada alamat surel, nama akun, atau tab
peramban pribadi yang ikut terekam.

Gambar yang dipakai halaman:

- `explorer.png` - daftar berkas lintas akun
- `accounts.png` - akun terhubung dan kuotanya
- `connect.png` - dialog menambah akun

## Catatan isi

Klaim di halaman ini dijaga agar sesuai dengan kemampuan yang benar-benar
berjalan. Bagian arsitektur menyebut secara eksplisit bahwa pemecahan berkas
lintas akun belum tersedia, agar calon pengguna tidak salah menduga.
