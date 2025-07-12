CREATE TABLE test_suites (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    estimated_start_date DATE,
    estimated_end_date DATE,
    require_effort_comment BOOLEAN DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status suite_status_enum NOT NULL DEFAULT '準備中'
);

CREATE TABLE test_groups (
    id VARCHAR(50) PRIMARY KEY,
    suite_id VARCHAR(50) NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    display_order INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status suite_status_enum NOT NULL DEFAULT '準備中',
    FOREIGN KEY (suite_id) REFERENCES test_suites(id)
);

CREATE TABLE test_cases (
    id VARCHAR(50) PRIMARY KEY,
    group_id VARCHAR(50) NOT NULL,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    status test_status_enum NOT NULL DEFAULT '作成',
    priority priority_enum NOT NULL DEFAULT 'Medium',
    planned_effort FLOAT,
    actual_effort FLOAT DEFAULT 0,
    is_delayed BOOLEAN DEFAULT false,
    delay_days INTEGER DEFAULT 0,
    current_editor VARCHAR(100),
    is_locked BOOLEAN DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (group_id) REFERENCES test_groups(id)
);

CREATE TABLE effort_records (
    id SERIAL PRIMARY KEY,
    test_case_id VARCHAR(50) NOT NULL,
    record_date DATE NOT NULL,
    effort_amount FLOAT NOT NULL,
    is_additional BOOLEAN DEFAULT false,
    comment TEXT,
    recorded_by VARCHAR(100) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (test_case_id) REFERENCES test_cases(id),
    CONSTRAINT check_positive_effort CHECK (effort_amount > 0)
);

CREATE TABLE status_history (
    id SERIAL PRIMARY KEY,
    test_case_id VARCHAR(50) NOT NULL,
    old_status test_status_enum NOT NULL,
    new_status test_status_enum NOT NULL,
    changed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    changed_by VARCHAR(100) NOT NULL,
    reason TEXT,
    FOREIGN KEY (test_case_id) REFERENCES test_cases(id)
);