-- 009_create_user_sequence.up.sql
-- ユーザーID生成用シーケンス
CREATE SEQUENCE IF NOT EXISTS user_seq
    START WITH 1
    INCREMENT BY 1
    NO MAXVALUE
    CACHE 1;