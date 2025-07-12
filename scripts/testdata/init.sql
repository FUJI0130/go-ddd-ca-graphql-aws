-- scripts/testdata/init.sql
-- テストデータの定義
INSERT INTO test_suites (
    id, 
    name, 
    description, 
    status, 
    estimated_start_date,
    estimated_end_date,
    require_effort_comment,
    created_at,
    updated_at
) VALUES 
    ('TS001-202412', 'テストスイート1', 'テスト用データ1', '準備中', 
     '2024-12-01', '2024-12-31', true, 
     CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('TS002-202412', 'テストスイート2', 'テスト用データ2', '実行中',
     '2024-12-01', '2024-12-31', false,
     CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);