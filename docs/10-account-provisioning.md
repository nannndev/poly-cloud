# 10 — Account Provisioning (OAuth → rclone)

Bagaimana user menambah akun (Google Drive, Dropbox, dll) **lewat frontend**, tanpa pernah
menyentuh `rclone config` di terminal. Ini menutup gap antara "OAuth di browser" dan
"rclone butuh remote terkonfigurasi".

## 1. Keputusan: Opsi A — Backend-managed OAuth
Backend platform yang mengelola seluruh OAuth flow (bukan rclone). Setelah backend
memegang token, backend **menyuntikkan** token itu ke rclone sebagai remote baru secara
terprogram. Alasan & alternatif yang ditolak: [ADR-010](08-adr.md#adr-010).

Ringkas: satu OAuth app milik platform (per provider), token disimpan platform (encrypted),
rclone hanya jadi eksekutor operasi file.

## 2. Komponen
- **OAuth Client (backend):** registrasi app di tiap provider (Google, Dropbox, …) →
  dapat `client_id` + `client_secret` milik platform. Simpan di env/secret.
- **rclone daemon (`rclone rcd`):** rclone dijalankan sebagai daemon dengan HTTP RC API.
  Backend memanggilnya untuk membuat remote & operasi file. **Wajib** diikat ketat (§6).
- **Remote registry:** tiap akun user = satu remote rclone unik (mis. `acc_<uuid>:`),
  dipetakan di kolom `accounts.rclone_remote`.

## 3. Alur Tambah Akun (end-to-end)
1. User klik "Tambah Akun → Google Drive" di frontend.
2. FE → `POST /accounts/connect { provider:"gdrive", label:"GDrive kerja" }`.
3. Backend bangun `auth_url` (pakai `client_id` platform, scope minimal, `state` acak
   terikat user+session) → balas `{ auth_url }`.
4. FE redirect user ke consent Google. User login + izinkan.
5. Google redirect balik ke `POST /accounts/callback { code, state }`.
6. Backend validasi `state`, tukar `code` → `access_token` + `refresh_token`
   (pakai `client_secret` platform).
7. Backend simpan `accounts` (baris baru) + `account_tokens` (token, AES-GCM).
8. Backend **provisioning remote rclone**: panggil RC `config/create` dengan:
   - `name = acc_<account_uuid>`
   - `type = drive` (mapping provider→tipe rclone)
   - parameter token (`token = {access,refresh,expiry}` JSON) + `client_id`/`client_secret` platform.
9. Backend jalankan initial sync (list → tulis `files_index`).
10. FE tampilkan akun baru + progress index.

Setelah ini, semua operasi (list/upload/download) memakai remote `acc_<uuid>:` via daemon.

## 4. Provider → tipe rclone (mapping)
| Provider UI | rclone `type` | OAuth? |
|-------------|---------------|--------|
| Google Drive | `drive` | ya |
| Dropbox | `dropbox` | ya |
| OneDrive | `onedrive` | ya |
| S3 (AWS/kompatibel) | `s3` | tidak (access key/secret) |
| Backblaze B2 | `b2` | tidak (key id/app key) |

Provider non-OAuth (S3/B2) lewat form input key (bukan redirect), lalu `config/create`
dengan kredensial tsb. Alur langkah 3–6 diganti form; langkah 7–9 sama.

## 5. Token Refresh (siapa yang pegang)
Karena token milik platform (Opsi A), **backend** yang berwenang refresh:
- Backend refresh via endpoint token provider memakai `refresh_token` + `client_secret`.
- Setelah refresh, backend update `account_tokens` **dan** perbarui token di remote rclone
  (`config/update` remote terkait) agar rclone tak memakai token basi.
- Refresh gagal permanen → `accounts.status = needs_reconnect`; UI minta re-auth.

## 6. Keamanan rclone RC daemon (WAJIB)
rclone RC API bersifat **all-or-nothing dan sangat kuat**: pemanggil yang bisa menjangkau
API dapat menjalankan perintah OS, membaca balik semua kredensial tersimpan (`config/dump`),
dan mengubah/menghapus remote. Karena itu:
- **Bind localhost saja** (`--rc-addr 127.0.0.1:5572`); jangan pernah expose ke publik/jaringan.
- **Aktifkan auth** (`--rc-user`/`--rc-pass`) walau localhost; simpan kredensial di env backend.
- Daemon & backend **satu jaringan privat** (same host / same pod). Di Docker Compose:
  daemon tak `expose`/`publish` port ke host.
- Jangan aktifkan `--rc-serve`/`--rc-files` kecuali diperlukan (mengurangi permukaan serang).
- Perlakukan daemon seperti komponen berhak-tinggi: hanya backend yang boleh memanggilnya.
Detail keputusan ini: [ADR-011](08-adr.md#adr-011).

## 7. Multi-akun (termasuk banyak akun Google)
- Tiap akun = remote unik `acc_<uuid>`, jadi banyak akun Google hidup berdampingan
  (`acc_a1b2…`, `acc_c3d4…`) tanpa tabrakan. Diperkuat `unique (user_id, rclone_remote)`.
- Satu OAuth app platform melayani semua akun Google user; tiap akun punya token sendiri.
- Multi-user: nama remote pakai UUID akun (bukan email) → tak bocor identitas & tak bentrok
  antar user.

## 8. Yang berubah di dokumen lain
- [03 Backend](03-backend.md): Engine kini menargetkan rclone **daemon (rcd)**, bukan CLI
  subprocess; tambah tanggung jawab provisioning (`config/create`/`update`/`delete`).
- [07 Sequence](07-sequence-flows.md): flow connect diperbarui memakai backend-managed OAuth
  + `config/create`.
- [06 API](06-api-spec.md): `/accounts/connect` & `/accounts/callback` sudah selaras;
  provider non-OAuth memakai form (bukan redirect).

## 9. Catatan Dev vs Prod
- **Spike (fase 0)** memakai `rclone config` manual + CLI — itu hanya untuk proof-of-concept,
  **bukan** jalur produk. Provisioning terprogram (dokumen ini) menggantikannya untuk MVP.
- Migrasi dari spike: ganti `RcloneEngine` (CLI) → `RcloneDaemonEngine` (RC API). Interface
  `Engine` tetap sama, jadi pemanggil (StorageService) tak berubah.
