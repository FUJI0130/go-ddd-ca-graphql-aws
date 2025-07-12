// internal/infrastructure/persistence/postgres/test_environment_suite_id_generator.go

package postgres

import (
	"database/sql"
	"fmt"
	"math/rand"
	"time"

	"github.com/FUJI0130/go-ddd-ca/internal/domain/repository"
)

// TestEnvironmentSuiteIDGenerator はテスト環境用のテストスイートIDジェネレーター
type TestEnvironmentSuiteIDGenerator struct {
	db *sql.DB
}

// インターフェース実装の確認
var _ repository.TestSuiteIDGenerator = (*TestEnvironmentSuiteIDGenerator)(nil)

// NewTestEnvironmentSuiteIDGenerator は新しいTestEnvironmentSuiteIDGeneratorを作成します
func NewTestEnvironmentSuiteIDGenerator(db *sql.DB) *TestEnvironmentSuiteIDGenerator {
	return &TestEnvironmentSuiteIDGenerator{db: db}
}

// GenerateID はテスト環境用の一意なIDを生成します
func (g *TestEnvironmentSuiteIDGenerator) GenerateID() (string, error) {
	// タイムスタンプ生成（ナノ秒精度）
	now := time.Now()
	timestamp := now.UnixNano()
	yearMonth := fmt.Sprintf("%d%02d", now.Year(), now.Month())

	// 乱数生成（Go 1.20以降の方法）
	r := rand.New(rand.NewSource(timestamp))
	randomPart := r.Intn(1000)

	// TEST-TS{ランダム部分}-{年月} 形式のID生成
	id := fmt.Sprintf("TEST-TS%03d-%s", randomPart, yearMonth)

	return id, nil
}
