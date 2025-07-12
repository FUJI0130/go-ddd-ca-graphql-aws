DROP TRIGGER IF EXISTS update_test_cases_updated_at ON test_cases;
DROP TRIGGER IF EXISTS update_test_groups_updated_at ON test_groups;
DROP TRIGGER IF EXISTS update_test_suites_updated_at ON test_suites;
DROP FUNCTION IF EXISTS update_updated_at();