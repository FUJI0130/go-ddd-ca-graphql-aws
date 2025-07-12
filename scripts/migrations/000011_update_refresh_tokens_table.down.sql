-- 000010_update_refresh_tokens_table.down.sql

-- トリガー削除
DROP TRIGGER IF EXISTS update_refresh_tokens_updated_at ON refresh_tokens;

-- インデックス削除
DROP INDEX IF EXISTS idx_refresh_tokens_user_id;
DROP INDEX IF EXISTS idx_refresh_tokens_expires_at;
DROP INDEX IF EXISTS idx_refresh_tokens_token;

-- フィールド削除
ALTER TABLE refresh_tokens 
  DROP COLUMN IF EXISTS issued_at,
  DROP COLUMN IF EXISTS last_used_at,
  DROP COLUMN IF EXISTS client_info,
  DROP COLUMN IF EXISTS ip_address,
  DROP COLUMN IF EXISTS updated_at;