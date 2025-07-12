-- インデックスの削除
DROP INDEX IF EXISTS idx_users_username;
DROP INDEX IF EXISTS idx_users_role;

-- トリガーの削除
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- テーブルの削除
DROP TABLE IF EXISTS users;