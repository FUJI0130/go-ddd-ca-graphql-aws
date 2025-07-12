-- まずテーブルを削除（依存関係の順番に注意）
DROP TABLE IF EXISTS effort_records;
DROP TABLE IF EXISTS status_history;
DROP TABLE IF EXISTS test_cases;
DROP TABLE IF EXISTS test_groups;
DROP TABLE IF EXISTS test_suites;

-- その後、型を削除
DROP TYPE IF EXISTS priority_enum;
DROP TYPE IF EXISTS suite_status_enum;
DROP TYPE IF EXISTS test_status_enum;