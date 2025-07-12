-- テストスイート用シーケンス
CREATE SEQUENCE IF NOT EXISTS test_suite_seq
    START WITH 1
    INCREMENT BY 1
    NO MAXVALUE
    CACHE 1;

-- テストグループ用シーケンス
CREATE SEQUENCE IF NOT EXISTS test_group_seq
    START WITH 1
    INCREMENT BY 1
    NO MAXVALUE
    CACHE 1;

-- テストケース用シーケンス
CREATE SEQUENCE IF NOT EXISTS test_case_seq
    START WITH 1
    INCREMENT BY 1
    NO MAXVALUE
    CACHE 1;