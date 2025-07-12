CREATE INDEX idx_effort_records_date ON effort_records(record_date);
CREATE INDEX idx_test_cases_priority ON test_cases(priority);
CREATE INDEX idx_test_cases_status ON test_cases(status);
CREATE INDEX idx_test_groups_order ON test_groups(display_order, suite_id);
