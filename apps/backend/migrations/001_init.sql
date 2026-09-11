CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS users (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), email text UNIQUE NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS accounts (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  provider text NOT NULL, label text NOT NULL, rclone_remote text NOT NULL,
  total_bytes bigint, used_bytes bigint, status text NOT NULL DEFAULT 'active',
  created_at timestamptz NOT NULL DEFAULT now(), last_synced timestamptz,
  UNIQUE (user_id, rclone_remote)
);
CREATE TABLE IF NOT EXISTS account_tokens (
  account_id uuid PRIMARY KEY REFERENCES accounts(id) ON DELETE CASCADE,
  enc_payload bytea NOT NULL, expires_at timestamptz
);
CREATE TABLE IF NOT EXISTS files_index (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name text NOT NULL, virtual_path text NOT NULL, mime text, size_bytes bigint NOT NULL DEFAULT 0,
  modified_at timestamptz, is_chunked boolean NOT NULL DEFAULT false
);
CREATE TABLE IF NOT EXISTS file_blocks (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), file_id uuid NOT NULL REFERENCES files_index(id) ON DELETE CASCADE,
  account_id uuid NOT NULL REFERENCES accounts(id) ON DELETE CASCADE, provider_ref text NOT NULL,
  seq int NOT NULL DEFAULT 0, size_bytes bigint NOT NULL DEFAULT 0, checksum text
);
CREATE INDEX IF NOT EXISTS idx_files_user_name ON files_index (user_id, name);
CREATE INDEX IF NOT EXISTS idx_files_user_path ON files_index (user_id, virtual_path);
CREATE INDEX IF NOT EXISTS idx_blocks_file_seq ON file_blocks (file_id, seq);
CREATE INDEX IF NOT EXISTS idx_blocks_account ON file_blocks (account_id);
