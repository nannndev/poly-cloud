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
keahlian tim. (−) dua bahasa dalam satu repo → dijembatani `shared-types`.

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
