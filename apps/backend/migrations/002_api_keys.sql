-- 002_api_keys.sql: Personal Access Tokens / Developer API Keys untuk integrasi eksternal.
CREATE TABLE IF NOT EXISTS api_keys (
  id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id      uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name         text NOT NULL,
  key_prefix   text NOT NULL,                     -- misal "plc_live_8f3a..."
  key_hash     text UNIQUE NOT NULL,             -- sha256 hex string
  scopes       text[] NOT NULL DEFAULT '{"read", "write"}',
  last_used_at timestamptz,
  expires_at   timestamptz,
  created_at   timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_api_keys_hash ON api_keys (key_hash);
CREATE INDEX IF NOT EXISTS idx_api_keys_user ON api_keys (user_id);
