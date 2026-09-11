# 05 — Data Model

## 1. ERD
```mermaid
erDiagram
    users ||--o{ accounts : owns
    users ||--o{ files_index : owns
    accounts ||--|| account_tokens : has
    accounts ||--o{ file_blocks : stores
    files_index ||--o{ file_blocks : "composed of"

    users {
        uuid id PK
        text email
        timestamptz created_at
    }
    accounts {
        uuid id PK
        uuid user_id FK
        text provider
        text label
        text rclone_remote
        bigint total_bytes
        bigint used_bytes
        text status
        timestamptz created_at
        timestamptz last_synced
    }
    account_tokens {
        uuid account_id PK_FK
        bytea enc_payload
        timestamptz expires_at
    }
    files_index {
        uuid id PK
        uuid user_id FK
        text name
        text virtual_path
        text mime
        bigint size_bytes
        timestamptz modified_at
        boolean is_chunked
    }
    file_blocks {
        uuid id PK
        uuid file_id FK
        uuid account_id FK
        text provider_ref
        int seq
        bigint size_bytes
        text checksum
    }
```

## 2. Skema SQL
```sql
create table users (
  id          uuid primary key default gen_random_uuid(),
  email       text unique not null,
  created_at  timestamptz not null default now()
);

create table accounts (
  id            uuid primary key default gen_random_uuid(),
  user_id       uuid not null references users(id) on delete cascade,
  provider      text not null,               -- 'gdrive'|'dropbox'|'onedrive'|'s3'|'b2'
  label         text not null,
  rclone_remote text not null,               -- 'gdrive1:'
  total_bytes   bigint,
  used_bytes    bigint,
  status        text not null default 'active', -- 'active'|'needs_reconnect'|'error'
  created_at    timestamptz not null default now(),
  last_synced   timestamptz,
  unique (user_id, rclone_remote)
);

create table account_tokens (
  account_id  uuid primary key references accounts(id) on delete cascade,
  enc_payload bytea not null,                -- AES-GCM(access+refresh)
  expires_at  timestamptz
);

create table files_index (
  id           uuid primary key default gen_random_uuid(),
  user_id      uuid not null references users(id) on delete cascade,
  name         text not null,
  virtual_path text not null,
  mime         text,
  size_bytes   bigint not null default 0,
  modified_at  timestamptz,
  is_chunked   boolean not null default false
);

create table file_blocks (
  id           uuid primary key default gen_random_uuid(),
  file_id      uuid not null references files_index(id) on delete cascade,
  account_id   uuid not null references accounts(id) on delete cascade,
  provider_ref text not null,                -- file-id di provider
  seq          int  not null default 0,      -- A: 0 ; B: 0..N-1
  size_bytes   bigint not null default 0,
  checksum     text
);

-- Index bantu
create index idx_files_user_name  on files_index (user_id, name);
create index idx_files_user_path  on files_index (user_id, virtual_path);
create index idx_blocks_file_seq  on file_blocks (file_id, seq);
create index idx_blocks_account   on file_blocks (account_id);
```

## 3. Catatan Desain
- **`file_blocks` = kunci scalability A→B.** Model A: tepat 1 baris per file (`seq=0`).
  Model B: N baris (`seq=0..N-1`). Struktur tabel identik; hanya jumlah baris berbeda.
- **Token dipisah** dari `accounts` agar data sensitif terisolasi dan mudah dibatasi akses.
- **`user_id` ada di mana-mana** → multi-tenant siap; aktifkan Row-Level Security (RLS)
  saat pindah ke mode multi-user.
- **Quota di-cache** di `accounts` (bukan query provider tiap saat) → dashboard & routing cepat.
- **`virtual_path`** memisahkan struktur folder tampilan dari lokasi fisik provider.

## 4. Integritas
- Hapus user → cascade accounts, files, tokens, blocks.
- Hapus account → cascade tokens & blocks account itu (index file dgn block tersisa perlu ditinjau di Model B).
- `unique (user_id, rclone_remote)` cegah duplikasi remote.
