// internal/infrastructure/persistence/postgres/test_environment_Group_id_generator.go

package postgres

import (
	"database/sql"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/FUJI0130/go-ddd-ca/internal/domain/repository"
)

// TestEnvironmentGroupIDGenerator はテスト環境用のテストスイートIDジェネレーター
type TestEnvironmentGroupIDGenerator struct {
	db *sql.DB
}

// インターフェース実装の確認
var _ repository.TestGroupIDGenerator = (*TestEnvironmentGroupIDGenerator)(nil)

// NewTestEnvironmentGroupIDGenerator は新しいTestEnvironmentGroupIDGeneratorを作成します
func NewTestEnvironmentGroupIDGenerator(db *sql.DB) *TestEnvironmentGroupIDGenerator {
	return &TestEnvironmentGroupIDGenerator{db: db}
}

// GenerateID はテスト環境用の一意なIDを生成します
func (g *TestEnvironmentGroupIDGenerator) GenerateID(suiteID string) (string, error) {
	// 親IDの検証
	if suiteID == "" {
		return "", fmt.Errorf("スイートIDが空です")
	}

	// タイムスタンプ生成（ナノ秒精度）
	now := time.Now()
	timestamp := now.UnixNano()
	yearMonth := fmt.Sprintf("%d%02d", now.Year(), now.Month())

	// 乱数生成（Go 1.20以降の方法）
	r := rand.New(rand.NewSource(timestamp))
	randomPart := r.Intn(100)

	// スイートIDからプレフィックスを抽出
	var suitePrefix string
	if strings.HasPrefix(suiteID, "TEST-") {
		// TEST-TS形式のIDからプレフィックスを抽出
		parts := strings.Split(suiteID, "-")
		if len(parts) >= 3 {
			suitePrefix = parts[1] // "TS001"部分を取得
		} else {
			return "", fmt.Errorf("無効なテストスイートID形式: %s", suiteID)
		}
	} else {
		// 通常のTS001-202503形式からプレフィックス抽出
		parts := strings.Split(suiteID, "-")
		if len(parts) >= 1 {
			suitePrefix = parts[0] // "TS001"部分を取得
		} else {
			return "", fmt.Errorf("無効なスイートID形式: %s", suiteID)
		}
	}

	// ID生成
	id := fmt.Sprintf("TEST-%sTG%02d-%s", suitePrefix, randomPart, yearMonth)
	return id, nil
}
