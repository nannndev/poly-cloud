# 08 — Architecture Decision Records (ADR)

Format ringkas: Konteks → Keputusan → Konsekuensi.

---

## ADR-001 — Model A (whole-file) dulu, bukan Model B (chunked)
**Konteks:** Masalah inti = kapasitas agregat penuh, bukan file tunggal melebihi 1 akun.
Model B kompleks & berisiko data-loss (1 akun raib = file rusak) dan sering langgar ToS.
**Keputusan:** Mulai Model A + smart routing. Sediakan jalur upgrade ke B via abstraksi.
**Konsekuensi:** (+) cepat, aman, cukup selesaikan masalah nyata. (−) file > kapasitas
akun terlowong belum bisa (ditandai `ErrNoRoom`) sampai B diadopsi.

## ADR-002 — rclone sebagai Engine
**Konteks:** Butuh integrasi banyak provider (OAuth, refresh, file ops). Menulis sendiri
per provider = pekerjaan sangat besar & terus berubah.
**Keputusan:** Pakai rclone (70+ provider) sebagai engine, dibungkus interface `Engine`.
**Konsekuensi:** (+) time-to-MVP cepat, maintenance provider ditanggung rclone.
(−) terikat model operasi rclone; mitigasi: interface `Engine` bisa diganti impl.
Mode awal: CLI subprocess; bisa pindah library mode tanpa ubah pemanggil.

## ADR-003 — Backend Go, Frontend Nuxt 4
**Konteks:** Butuh I/O streaming efisien di backend; frontend selaras stack existing.
**Keputusan:** Go (backend) + Nuxt 4 (frontend).
**Konsekuensi:** (+) Go kuat untuk stream & subprocess rclone (Go-native); Nuxt sesuai
keahlian tim. (−) dua bahasa dalam satu repo → kontrak API dijaga manual: struct Go jadi otoritas, `apps/frontend/app/types/index.ts` salinannya (lihat docs/02 §4).

## ADR-004 — Monorepo
**Konteks:** Solo dev, backend & frontend berbagi kontrak API, target deploy satu unit.
**Keputusan:** Monorepo (`apps/`, `packages/`, `docs/`, satu docker-compose).
**Konsekuensi:** (+) satu clone, perubahan API lintas backend/FE dalam satu PR, tak drift,
deploy tunggal. (−) CI bisa lebih berat seiring tumbuh; trigger pindah multi-repo
didefinisikan (lihat [02 §7](02-repo-structure.md)).

## ADR-005 — Metadata di DB, file tetap di provider (stream-through)
**Konteks:** Menyimpan file di server = biaya storage & risiko; user ingin file tetap di provider.
**Keputusan:** DB hanya simpan metadata/index; operasi file di-stream (rcat/cat), tak persist.
**Konsekuensi:** (+) murah, tak duplikasi storage, browse cepat dari cache. (−) operasi
bergantung ketersediaan & bandwidth provider saat itu.

## ADR-006 — Token dienkripsi & dipisah tabel
**Konteks:** Access/refresh token sangat sensitif.
**Keputusan:** Simpan terenkripsi (AES-GCM) di `account_tokens`, terpisah dari `accounts`.
**Konsekuensi:** (+) isolasi data sensitif, mudah batasi akses. (−) perlu manajemen key
(`TOKEN_ENC_KEY`) yang aman.

## ADR-007 — Multi-tenant sejak skema, aktif nanti
**Konteks:** v1 personal, tapi target self-hosted multi-user.
**Keputusan:** Sertakan `user_id` di semua tabel; aktifkan RLS saat mode multi-user.
**Konsekuensi:** (+) tak perlu migrasi besar saat scale. (−) sedikit overhead kolom di v1.

## ADR-008 — Smart routing default: most-free-space
**Konteks:** Perlu strategi memilih account tujuan upload.
**Keputusan:** Default "most-free" (isi account terlowong dulu). Strategi pluggable.
**Konsekuensi:** (+) distribusi merata, sederhana, deterministik. (−) bisa suboptimal
untuk kasus khusus (mis. ingin kelompokkan tipe file) → tersedia strategi alternatif.

## ADR-009 — Virtual Filesystem murni (bukan mirror provider)
**Konteks:** File user tersebar di banyak akun, tapi user butuh organisasi (folder) yang rapi.
Dua opsi: (1) struktur folder hanya di DB platform (virtual path murni), atau (2) mirror
struktur folder ke provider aslinya.
**Keputusan:** Pakai Model 1 — virtual filesystem murni. Folder hidup di tabel `folders`
di DB; file fisik ditaruh flat di satu folder khusus (mis. `PolyCloud/`) per akun.
Struktur folder tak pernah dibuat di provider.
**Konsekuensi:** (+) organisasi tak dibatasi lokasi fisik — folder logis bisa berisi file
dari banyak akun sekaligus; buat/rename/pindah folder = update DB murni (instan, tak transfer
data); folder kosong bisa ada. (−) bila user membuka provider langsung (di luar platform),
file terlihat flat/tak terstruktur. Model 2 ditolak karena satu folder logis mustahil
di-mirror utuh ketika isinya tersebar di banyak akun. Detail: [doc 09](09-virtual-filesystem.md).

<a name="adr-010"></a>
## ADR-010 — Provisioning akun: backend-managed OAuth (Opsi A)
**Konteks:** User menambah akun lewat frontend, bukan `rclone config` terminal. Perlu
jembatan antara OAuth di browser dan remote rclone. Dua opsi: (A) backend mengelola OAuth
sendiri lalu inject token ke rclone; (B) delegasikan OAuth ke rclone via RC.
**Keputusan:** Opsi A. Backend memegang OAuth app (client_id/secret platform), menukar code,
menyimpan token (encrypted), lalu membuat remote rclone via `config/create` dengan token itu.
**Konsekuensi:** (+) platform mengontrol penuh token & refresh, isolasi multi-user rapi
(remote per-UUID), UX konsisten lintas provider. (+) rclone murni jadi eksekutor. (−) backend
menanggung logika OAuth tiap provider + wajib sinkronkan token ke remote saat refresh
(`config/update`). Detail: [doc 10](10-account-provisioning.md).

<a name="adr-011"></a>
## ADR-011 — rclone dijalankan sebagai daemon (rcd) yang terkunci ketat
**Konteks:** Opsi A butuh memanggil rclone secara terprogram (`config/create`, operasi file).
Mode CLI subprocess (dipakai spike) tak praktis untuk provisioning dinamis. rclone RC API
sangat kuat & all-or-nothing: aksesnya = bisa jalankan perintah OS dan baca semua kredensial.
**Keputusan:** Jalankan `rclone rcd` sebagai daemon internal; hanya backend yang memanggilnya.
Ikat ketat: bind `127.0.0.1`, aktifkan auth, jangan expose port keluar host/pod, matikan
fitur serve yang tak perlu.
**Konsekuensi:** (+) provisioning & operasi terprogram, engine bisa pindah dari CLI ke RC
tanpa ubah interface `Engine`. (−) daemon jadi komponen berhak-tinggi yang wajib diamankan;
kompromi pada daemon = kompromi semua backend terkonfigurasi. Mitigasi = isolasi jaringan
privat + auth. Detail: [doc 10 §6](10-account-provisioning.md).
