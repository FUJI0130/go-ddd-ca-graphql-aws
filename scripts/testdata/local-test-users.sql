-- ===================================================================
-- ファイル名: local-test-users.sql
-- 配置場所: scripts/testdata/local-test-users.sql
-- 説明: ローカル環境でのGraphQL認証テスト用ユーザーデータ
-- 
-- 用途:
--  - GraphQL認証機能のローカルテスト用データ投入
--  - test_adminユーザー（Admin権限）の作成
--  - JWT認証フローの完全動作確認用
-- 
-- テストユーザー仕様:
--  - username: test_admin
--  - password: password (bcryptハッシュ化済み)
--  - role: Admin (全権限)
--  - ユーザーID: USER001
-- 
-- 注意:
--  - このファイルはマイグレーション実行後に適用してください
--  - パスワードハッシュはbcrypt cost=10で生成済み
--  - ローカル開発専用（本番環境では使用しないでください）
-- ===================================================================

-- GraphQL認証テスト用管理者ユーザーの作成
INSERT INTO users (id, username, password_hash, role, created_at, updated_at) VALUES 
(
    'USER001',
    'test_admin',
    '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi',
    'Admin',
    NOW(),
    NOW()
)
ON CONFLICT (username) DO UPDATE SET
    password_hash = EXCLUDED.password_hash,
    role = EXCLUDED.role,
    updated_at = NOW();

-- GraphQL認証テスト用マネージャーユーザーの作成（権限テスト用）
INSERT INTO users (id, username, password_hash, role, created_at, updated_at) VALUES 
(
    'USER002',
    'test_manager',
    '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi',
    'Manager',
    NOW(),
    NOW()
)
ON CONFLICT (username) DO UPDATE SET
    password_hash = EXCLUDED.password_hash,
    role = EXCLUDED.role,
    updated_at = NOW();

-- GraphQL認証テスト用テスターユーザーの作成（権限テスト用）
INSERT INTO users (id, username, password_hash, role, created_at, updated_at) VALUES 
(
    'USER003',
    'test_tester',
    '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi',
    'Tester',
    NOW(),
    NOW()
)
ON CONFLICT (username) DO UPDATE SET
    password_hash = EXCLUDED.password_hash,
    role = EXCLUDED.role,
    updated_at = NOW();

-- 投入結果の確認
SELECT 
    id,
    username,
    role,
    created_at,
    updated_at,
    last_login_at
FROM users 
WHERE username IN ('test_admin', 'test_manager', 'test_tester')
ORDER BY created_at;

-- 投入完了メッセージ
DO $$
BEGIN
    RAISE NOTICE '========================================';
    RAISE NOTICE 'ローカルテストユーザーデータの投入が完了しました';
    RAISE NOTICE '========================================';
    RAISE NOTICE 'メインテストユーザー:';
    RAISE NOTICE '  - ユーザー名: test_admin';
    RAISE NOTICE '  - パスワード: password';
    RAISE NOTICE '  - ロール: Admin';
    RAISE NOTICE '  - ユーザーID: USER001';
    RAISE NOTICE '';
    RAISE NOTICE '追加テストユーザー:';
    RAISE NOTICE '  - test_manager (Manager権限)';
    RAISE NOTICE '  - test_tester (Tester権限)';
    RAISE NOTICE '';
    RAISE NOTICE 'GraphQL認証テストの実行方法:';
    RAISE NOTICE '1. make local-graphql-server でサーバー起動';
    RAISE NOTICE '2. http://localhost:8080/ にアクセス';
    RAISE NOTICE '3. login mutationでJWTトークンを取得';
    RAISE NOTICE '4. Authorization HeaderにBearer tokenを設定';
    RAISE NOTICE '5. 認証済みAPIを実行してテスト';
    RAISE NOTICE '========================================';
END
$$;