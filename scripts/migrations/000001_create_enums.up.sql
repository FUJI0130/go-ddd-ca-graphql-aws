CREATE TYPE priority_enum AS ENUM (
    'Critical',
    'High',
    'Medium',
    'Low'
);

CREATE TYPE suite_status_enum AS ENUM (
    '準備中',
    '実行中',
    '完了',
    '中断'
);

CREATE TYPE test_status_enum AS ENUM (
    '作成',
    'テスト',
    '修正',
    'レビュー待ち',
    'レビュー中',
    '完了',
    '再テスト'
);