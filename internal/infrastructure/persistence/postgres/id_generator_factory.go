// internal/infrastructure/persistence/postgres/id_generator_factory.go

package postgres

import (
	"database/sql"
	"os"

	"github.com/FUJI0130/go-ddd-ca/internal/domain/repository"
)

// IsTestEnvironment は現在の実行環境がテスト環境かどうかを判定します
func IsTestEnvironment() bool {
	// 環境変数で判定
	if env := os.Getenv("TEST_ENV"); env == "true" {
		return true
	}

	// テスト用のビルドタグで判定
	// 注: この方法はビルド時に決定するため、実行時の切り替えには使えない
	return false
}

// NewTestSuiteIDGeneratorWithEnv は環境に適したIDジェネレーターを返します
func NewTestSuiteIDGeneratorWithEnv(db *sql.DB) repository.TestSuiteIDGenerator {
	if IsTestEnvironment() {
		return NewTestEnvironmentSuiteIDGenerator(db)
	}
	return NewTestSuiteIDGenerator(db)
}

// NewTestGroupIDGeneratorWithEnv は環境に適したIDジェネレーターを返します
func NewTestGroupIDGeneratorWithEnv(db *sql.DB) repository.TestGroupIDGenerator {
	if IsTestEnvironment() {
		return NewTestEnvironmentGroupIDGenerator(db)
	}
	return NewTestGroupIDGenerator(db)
}

// NewTestCaseIDGeneratorWithEnv は環境に適したIDジェネレーターを返します
func NewTestCaseIDGeneratorWithEnv(db *sql.DB) repository.TestCaseIDGenerator {
	if IsTestEnvironment() {
		return NewTestEnvironmentCaseIDGenerator(db)
	}
	return NewTestCaseIDGenerator(db)
}

// NewUserIDGeneratorWithEnv は環境に適したIDジェネレーターを返します
func NewUserIDGeneratorWithEnv(db *sql.DB) repository.UserIDGenerator {
	if IsTestEnvironment() {
		return NewTestEnvironmentUserIDGenerator(db)
	}
	return NewUserIDGenerator(db)
}
