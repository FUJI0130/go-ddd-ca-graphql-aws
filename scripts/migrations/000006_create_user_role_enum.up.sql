CREATE TYPE user_role_enum AS ENUM (
    'Admin',    -- 管理者権限
    'Manager',  -- マネージャー権限
    'Tester'    -- テスター権限
);