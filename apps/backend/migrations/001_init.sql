-- Skema awal Poly Cloud (doc 05). Dijalankan otomatis oleh image Postgres
-- lewat /docker-entrypoint-initdb.d saat volume data masih kosong.
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS users (
  id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  email      text UNIQUE NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS accounts (
  id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id       uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  provider      text NOT NULL,                    -- 'gdrive'|'dropbox'|'onedrive'|'s3'|'b2'|'r2'
  label         text NOT NULL,
  rclone_remote text NOT NULL,                    -- 'acc_<uuid tanpa dash>'
  -- Akar penyimpanan di dalam remote. Provider objek (S3/B2/R2) menaruh bucket
  -- di sini; provider drive (GDrive/Dropbox/OneDrive) mengosongkannya.
  root_path     text NOT NULL DEFAULT '',
  total_bytes   bigint,
  used_bytes    bigint,
  status        text NOT NULL DEFAULT 'active',   -- 'active'|'needs_reconnect'|'error'|'syncing'
  created_at    timestamptz NOT NULL DEFAULT now(),
  last_synced   timestamptz,
  UNIQUE (user_id, rclone_remote)
);

-- Token dipisah dari accounts supaya data sensitif terisolasi (ADR-006).
CREATE TABLE IF NOT EXISTS account_tokens (
  account_id  uuid PRIMARY KEY REFERENCES accounts(id) ON DELETE CASCADE,
  enc_payload bytea NOT NULL,                     -- AES-GCM(access+refresh)
  expires_at  timestamptz
);

-- VFS: folder virtual, hanya hidup di DB — tak pernah dibuat di provider (doc 09).
CREATE TABLE IF NOT EXISTS folders (
  id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id    uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  parent_id  uuid REFERENCES folders(id) ON DELETE CASCADE,  -- null = root
  name       text NOT NULL,
  path       text NOT NULL,                       -- materialized cache dari adjacency
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (user_id, path)
);

CREATE TABLE IF NOT EXISTS files_index (
  id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id      uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  folder_id    uuid REFERENCES folders(id) ON DELETE CASCADE, -- null = root
  name         text NOT NULL,
  virtual_path text NOT NULL,
  mime         text,
  size_bytes   bigint NOT NULL DEFAULT 0,
  modified_at  timestamptz,
  is_chunked   boolean NOT NULL DEFAULT false,
  UNIQUE (user_id, virtual_path)
);

-- Model A: tepat 1 baris per file (seq=0). Model B: N baris — struktur sama.
CREATE TABLE IF NOT EXISTS file_blocks (
  id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  file_id      uuid NOT NULL REFERENCES files_index(id) ON DELETE CASCADE,
  account_id   uuid NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
  provider_ref text NOT NULL,
  seq          int NOT NULL DEFAULT 0,
  size_bytes   bigint NOT NULL DEFAULT 0,
  checksum     text
);

CREATE INDEX IF NOT EXISTS idx_files_user_name  ON files_index (user_id, name);
CREATE INDEX IF NOT EXISTS idx_files_user_path  ON files_index (user_id, virtual_path);
CREATE INDEX IF NOT EXISTS idx_files_folder     ON files_index (folder_id);
CREATE INDEX IF NOT EXISTS idx_folders_user     ON folders (user_id);
CREATE INDEX IF NOT EXISTS idx_folders_parent   ON folders (parent_id);
CREATE INDEX IF NOT EXISTS idx_blocks_file_seq  ON file_blocks (file_id, seq);
CREATE INDEX IF NOT EXISTS idx_blocks_account   ON file_blocks (account_id);
