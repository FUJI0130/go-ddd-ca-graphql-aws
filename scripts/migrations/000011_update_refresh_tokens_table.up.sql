-- 000010_update_refresh_tokens_table.up.sql

-- 追加フィールド
ALTER TABLE refresh_tokens 
  ADD COLUMN issued_at TIMESTAMP,
  ADD COLUMN last_used_at TIMESTAMP,
  ADD COLUMN client_info TEXT,
  ADD COLUMN ip_address TEXT,
  ADD COLUMN updated_at TIMESTAMP;

-- 既存レコードのissued_atを設定（既存データの整合性確保）
UPDATE refresh_tokens SET issued_at = created_at WHERE issued_at IS NULL;
UPDATE refresh_tokens SET updated_at = created_at WHERE updated_at IS NULL;

-- NOT NULL制約を追加
ALTER TABLE refresh_tokens 
  ALTER COLUMN issued_at SET NOT NULL,
  ALTER COLUMN updated_at SET NOT NULL;

-- 自動更新トリガーの作成
-- トリガー関数が存在しない場合は作成
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_proc WHERE proname = 'update_updated_at') THEN
    EXECUTE '
    CREATE OR REPLACE FUNCTION update_updated_at()
    RETURNS TRIGGER AS $BODY$
    BEGIN
      NEW.updated_at = CURRENT_TIMESTAMP;
      RETURN NEW;
    END;
    $BODY$ LANGUAGE plpgsql;
    ';
  END IF;
END
$$;

-- 自動更新トリガーの作成
-- 関数が既に存在することを前提に条件チェックを削除
DROP TRIGGER IF EXISTS update_refresh_tokens_updated_at ON refresh_tokens;
CREATE TRIGGER update_refresh_tokens_updated_at
  BEFORE UPDATE ON refresh_tokens
  FOR EACH ROW
  EXECUTE FUNCTION update_updated_at();

-- インデックス追加
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token ON refresh_tokens(token);