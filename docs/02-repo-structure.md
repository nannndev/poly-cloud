# 02 — Repository Structure

## 1. Keputusan: Monorepo
Backend (Go), frontend (Nuxt 4), shared types, docs, dan deploy config berada dalam satu repo.
Alasan lengkap ada di [ADR-004](08-adr.md#adr-004--monorepo). Ringkas:
- Solo dev → hindari overhead sinkron antar-repo.
- Backend & frontend berbagi kontrak API → satu perubahan, satu PR, tak drift.
- Self-hosted = satu unit deploy (`docker compose up`).
- Migrasi mono→multi jauh lebih mudah daripada sebaliknya (opsi tetap terbuka).

## 2. Struktur Folder
```
storage-platform/
├── apps/
│   ├── backend/                  # Go API + rclone engine adapter
│   │   ├── cmd/api/main.go       # entrypoint
│   │   ├── internal/
│   │   │   ├── http/             # handlers, router, middleware, SSE
│   │   │   ├── storage/          # StorageService + WholeFileStore (+ ChunkedStore nanti)
│   │   │   ├── engine/           # rclone adapter (CLI/lib)
│   │   │   ├── routing/          # smart routing
│   │   │   ├── index/            # repo metadata (DB)
│   │   │   ├── auth/             # OAuth flow, token crypto, lifecycle
│   │   │   └── config/           # env loading
│   │   ├── migrations/           # SQL migrations
│   │   ├── go.mod
│   │   └── Dockerfile
│   └── frontend/                 # Nuxt 4
│       ├── app/                  # pages, layouts, components
│       ├── composables/          # useAccounts, useFiles, useUpload
│       ├── stores/               # Pinia state
│       ├── nuxt.config.ts
│       ├── package.json
│       └── Dockerfile
├── packages/
│   └── shared-types/             # kontrak tipe (sumber kebenaran API)
│       ├── types.ts              # TS untuk frontend
│       └── README.md             # cara sinkron dgn Go structs
├── docs/                         # dokumen ini
├── docker-compose.yml            # backend + frontend + postgres
├── .env.example
├── Makefile                      # task umum (dev, build, migrate)
└── README.md
```

## 3. Batas Tanggung Jawab
| Path | Tanggung jawab | Tidak boleh |
|------|----------------|-------------|
| `apps/backend` | logika bisnis, DB, engine, API | logika presentasi |
| `apps/frontend` | UI, state, panggil API | akses DB / provider langsung |
| `packages/shared-types` | definisi kontrak | logika runtime |
| `docs` | desain & keputusan | kode |

## 4. Shared Contract (anti-drift)
Kontrak API adalah sumber kebenaran tunggal. Alur menjaga sinkron:
1. Go structs (backend) = otoritas bentuk data.
2. `packages/shared-types/types.ts` = cerminan TS-nya untuk frontend.
3. Saat API berubah: update Go struct → update `types.ts` → di satu PR yang sama.
   (Opsional lanjut: generate `types.ts` otomatis dari OpenAPI/Go — lihat ADR jika diadopsi.)

## 5. Tooling
- **Task runner:** `Makefile` di root (`make dev`, `make migrate`, `make build`).
- **Backend:** Go modules; migrasi SQL via tool ringan (mis. `golang-migrate`).
- **Frontend:** pnpm/npm; Nuxt 4.
- **Lint/format:** `gofmt`+`golangci-lint` (Go), ESLint/Prettier (Nuxt).
- **Dev DB:** Postgres via Docker (atau Supabase remote).

## 6. Konvensi Cabang & Commit
- Trunk-based sederhana: `main` + short-lived feature branches.
- Commit menyentuh backend+frontend untuk 1 perubahan API sebaiknya 1 PR (jaga kontrak sinkron).

## 7. Kapan Pecah ke Multi-repo (trigger, bukan sekarang)
- Ada tim terpisah dengan siklus rilis backend vs frontend berbeda.
- Frontend mau di-open-source sementara backend proprietary.
- CI monorepo jadi bottleneck signifikan.
Sampai salah satu terjadi, tetap monorepo.
