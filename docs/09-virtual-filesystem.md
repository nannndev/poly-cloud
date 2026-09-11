# 09 — Virtual Filesystem (VFS) Spec

## 1. Konsep Inti
Organisasi file yang dilihat user **terpisah** dari lokasi fisik file di provider.
Folder, path, dan struktur yang user buat hanya hidup di database platform
(`files_index.virtual_path` + tabel `folders`). Provider tidak tahu-menahu soal
struktur ini.

> **Prinsip:** *Organisasi (apa yang user lihat) ≠ Lokasi fisik (di mana file disimpan).*
> Smart routing menentukan **di akun mana** file ditaruh (berdasar sisa ruang).
> VFS menentukan **di folder mana** file itu muncul (berdasar keinginan user).
> Dua hal yang independen.

## 2. Model Terpilih: Virtual Path Murni (Model 1)
File secara fisik ditaruh di satu folder khusus di tiap akun (mis. `PolyCloud/`),
sedangkan struktur folder yang terlihat user 100% dari DB platform.

Alternatif yang **ditolak** (Model 2 — mirror struktur ke provider) beserta alasannya
ada di [ADR-009](08-adr.md#adr-009--virtual-filesystem-murni-bukan-mirror-provider).
Alasan ringkas penolakan: ketika file tersebar di banyak akun, satu folder logis
(mis. `/Kerjaan`) isinya mustahil di-mirror utuh karena tak ada satu akun pun yang
memuat seluruh isinya.

### Ilustrasi
Yang user lihat di platform:
```
/Kerjaan/
   ├── laporan.pdf      (fisik: GDrive-1, di folder PolyCloud/)
   └── budget.xlsx      (fisik: Dropbox, di folder PolyCloud/)
/Foto/
   └── liburan.jpg      (fisik: GDrive-2, di folder PolyCloud/)
```
Yang ada di provider (contoh GDrive-1): hanya `PolyCloud/laporan.pdf`.
Folder `Kerjaan` / `Foto` **tidak ada** di provider mana pun — hanya di DB platform.

## 3. Perilaku Operasi
| Operasi user | Yang terjadi | Sentuh provider? |
|--------------|--------------|------------------|
| Buat folder `/abc` | Insert baris di `folders` | ❌ (DB only) |
| Rename folder | Update `name`/`path` di `folders` (+ path anak) | ❌ (DB only) |
| Pindah file antar folder | Update `virtual_path`/`folder_id` di `files_index` | ❌ (DB only, instan) |
| Pindah folder (+isinya) | Update path folder & turunannya di DB | ❌ (DB only) |
| Hapus folder kosong | Hapus baris `folders` | ❌ |
| Hapus folder berisi | Hapus file fisik tiap isi + baris DB | ✅ (hapus objek) |
| Upload ke `/abc` | Router pilih akun → upload fisik → set `folder_id`=abc | ✅ (upload) |
| Download file | Ambil dari akun fisiknya (via `file_blocks`) | ✅ (stream) |

**Kunci:** operasi *organisasi* (buat/rename/pindah folder) itu **murah & instan**
karena cuma update DB. Hanya operasi *data* (upload/download/delete-isi) yang menyentuh provider.

## 4. Aturan & Batasan
- **Path unik per user:** tidak boleh dua item dengan `virtual_path` sama dalam satu user.
- **Nama file duplikat di folder sama:** tolak atau auto-suffix (`file (1).pdf`) — pilih di API.
- **Folder kosong boleh ada:** karena folder adalah entitas DB (tabel `folders`), bukan
  turunan dari keberadaan file. (Beda dgn provider yang kadang tak punya konsep folder kosong.)
- **Move lintas folder tak memindah data fisik:** file tetap di akun asalnya; hanya
  `folder_id`/`virtual_path` berubah. (Beda dgn "Move antar akun" di API `/files/{id}/move`
  yang benar-benar transfer data — dua hal berbeda.)
- **Konsistensi:** `files_index.folder_id` harus menunjuk folder milik user yang sama.

## 5. Path vs Folder ID (desain penyimpanan)
Dua pendekatan menyimpan hierarki; kita pakai **kombinasi**:
- **`folder_id` (adjacency):** tiap folder punya `parent_id`. Sumber kebenaran hierarki.
  Rename/move cukup ubah satu baris.
- **`path` (materialized, cache):** string `/Kerjaan/Sub` disimpan juga untuk query cepat
  & tampilan. Di-recompute saat folder dipindah/rename (update path folder + semua turunannya).

Alasan: adjacency mudah dimutasi; materialized path cepat dibaca. Materialized path
adalah turunan (derivable) dari adjacency, jadi bila keduanya tak sinkron, adjacency menang.

## 6. Dampak ke Model B (chunked) — tetap kompatibel
VFS beroperasi di level `files_index` (logis). Model B mengubah *bagaimana file fisik
disimpan* (`file_blocks` jadi N baris), **bukan** bagaimana file diorganisasi. Jadi
folder virtual bekerja identik untuk file whole (A) maupun chunked (B). Tak ada perubahan
di lapisan VFS saat upgrade.

## 7. Ringkas untuk Implementasi
- Tambah tabel `folders` (lihat [05 §Data Model](05-data-model.md#vfs-folders)).
- Tambah `folder_id` di `files_index`.
- Endpoint folder baru (lihat [06 §Folders](06-api-spec.md#folders)).
- Operasi organisasi = transaksi DB murni; tak panggil engine.
