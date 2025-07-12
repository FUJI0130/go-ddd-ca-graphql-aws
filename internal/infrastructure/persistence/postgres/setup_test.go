// internal/infrastructure/persistence/postgres/setup_test.go
package postgres_test

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

const (
	dbHost     = "localhost"
	dbPort     = 5433
	dbUser     = "test_user"
	dbPassword = "test_pass"
	dbName     = "test_db"
)

// setupTestDB はテスト用のデータベースをセットアップする
func setupTestDB(t *testing.T) (*sql.DB, func()) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName,
	)

	// DB接続を試みる（リトライロジック追加）
	var db *sql.DB
	var err error
	for i := 0; i < 5; i++ {
		db, err = sql.Open("postgres", dsn)
		if err == nil {
			err = db.Ping()
			if err == nil {
				break
			}
		}
		t.Logf("Retrying database connection (%d/5): %v", i+1, err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	// テストデータベースが実行中か確認
	if err := checkDatabaseRunning(t, db); err != nil {
		t.Fatalf("test database is not running correctly: %v", err)
	}

	// まず既存のオブジェクトを削除
	if err := cleanDatabase(t, db); err != nil {
		t.Fatalf("failed to clean database: %v", err)
	}

	// コアマイグレーション実行
	if err := executeMigrations(t, db); err != nil {
		t.Fatalf("failed to execute migrations: %v", err)
	}

	// クリーンアップ関数
	cleanup := func() {
		if err := cleanDatabase(t, db); err != nil {
			t.Errorf("failed to cleanup database: %v", err)
		}
		db.Close()
	}

	return db, cleanup
}

// checkDatabaseRunning はデータベースが実行中かどうかを確認する
func checkDatabaseRunning(t *testing.T, db *sql.DB) error {
	// 簡単なクエリを実行してデータベースが応答するか確認
	var one int
	err := db.QueryRow("SELECT 1").Scan(&one)
	if err != nil {
		return fmt.Errorf("database query failed: %w", err)
	}

	if one != 1 {
		return fmt.Errorf("unexpected query result")
	}

	return nil
}

// cleanDatabase はデータベースを初期状態にクリーンアップする
func cleanDatabase(t *testing.T, db *sql.DB) error {
	// 基本テーブル
	coreTables := []string{
		"status_history",
		"effort_records",
		"test_cases",
		"test_groups",
		"test_suites",
	}

	// 認証関連テーブル
	authTables := []string{
		"login_history",
		"refresh_tokens",
		"users",
	}

	// すべてのテーブルを結合
	allTables := append(coreTables, authTables...)

	// すべてのテーブルをクリーンアップ
	for _, table := range allTables {
		_, err := db.Exec("DROP TABLE IF EXISTS " + table + " CASCADE")
		if err != nil {
			t.Logf("Warning: failed to drop table %s: %v", table, err)
		}
	}

	// enum型の削除
	enums := []string{
		"priority_enum",
		"suite_status_enum",
		"test_status_enum",
		"user_role_enum",
	}

	for _, enum := range enums {
		_, err := db.Exec("DROP TYPE IF EXISTS " + enum + " CASCADE")
		if err != nil {
			t.Logf("Warning: failed to drop enum %s: %v", enum, err)
		}
	}

	// シーケンスの削除
	sequences := []string{
		"user_seq",
	}

	for _, seq := range sequences {
		_, err := db.Exec("DROP SEQUENCE IF EXISTS " + seq + " CASCADE")
		if err != nil {
			t.Logf("Warning: failed to drop sequence %s: %v", seq, err)
		}
	}

	return nil
}

// executeMigrations はマイグレーションファイルを実行する
func executeMigrations(t *testing.T, db *sql.DB) error {
	// 基本マイグレーション
	migrations := []string{
		"000001_create_enums.up.sql",
		"000002_create_tables.up.sql",
		"000003_create_indexes.up.sql",
		"000004_create_triggers.up.sql",
	}

	// マイグレーションパスを取得
	migrationsPath := getMigrationsPath(t)

	for _, migration := range migrations {
		migrationPath := filepath.Join(migrationsPath, migration)
		content, err := os.ReadFile(migrationPath)
		if err != nil {
			t.Logf("Warning: Migration file %s not found: %v", migration, err)
			continue
		}

		if _, err := db.Exec(string(content)); err != nil {
			t.Logf("Warning: Failed to execute migration %s: %v", migration, err)
		}
	}

	return nil
}

// getMigrationsPath はマイグレーションファイルのパスを取得する
func getMigrationsPath(t *testing.T) string {
	// PROJECT_ROOT環境変数があれば使用
	projectRoot := os.Getenv("PROJECT_ROOT")
	if projectRoot != "" {
		return filepath.Join(projectRoot, "scripts", "migrations")
	}

	// カレントディレクトリから推測
	currentDir, err := os.Getwd()
	if err != nil {
		t.Logf("Failed to get current directory: %v", err)
		// 最後の手段として相対パスを返す
		return "../../../../../scripts/migrations"
	}

	// カレントディレクトリからプロジェクトルートを推測
	// 注: このパスはプロジェクト構造に依存するため、調整が必要な場合がある
	for i := 0; i < 5; i++ {
		potentialRoot := currentDir
		for j := 0; j < i; j++ {
			potentialRoot = filepath.Dir(potentialRoot)
		}

		migrationsPath := filepath.Join(potentialRoot, "scripts", "migrations")
		if _, err := os.Stat(migrationsPath); err == nil {
			return migrationsPath
		}
	}

	// 最終的に見つからない場合は相対パスを返す
	return "../../../../../scripts/migrations"
}
