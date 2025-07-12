// internal/infrastructure/persistence/postgres/test_environment_Case_id_generator.go

package postgres

import (
	"database/sql"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/FUJI0130/go-ddd-ca/internal/domain/repository"
)

// TestEnvironmentCaseIDGenerator はテスト環境用のテストスイートIDジェネレーター
type TestEnvironmentCaseIDGenerator struct {
	db *sql.DB
}

// インターフェース実装の確認
var _ repository.TestCaseIDGenerator = (*TestEnvironmentCaseIDGenerator)(nil)

// NewTestEnvironmentCaseIDGenerator は新しいTestEnvironmentCaseIDGeneratorを作成します
func NewTestEnvironmentCaseIDGenerator(db *sql.DB) *TestEnvironmentCaseIDGenerator {
	return &TestEnvironmentCaseIDGenerator{db: db}
}

// GenerateID はテスト環境用の一意なIDを生成します
func (g *TestEnvironmentCaseIDGenerator) GenerateID(groupID string) (string, error) {
	// 親IDの検証
	if groupID == "" {
		return "", fmt.Errorf("グループIDが空です")
	}

	// タイムスタンプ生成（ナノ秒精度）
	now := time.Now()
	timestamp := now.UnixNano()
	yearMonth := fmt.Sprintf("%d%02d", now.Year(), now.Month())

	// 乱数生成（Go 1.20以降の方法）
	r := rand.New(rand.NewSource(timestamp))
	randomPart := r.Intn(1000)

	// グループIDからプレフィックスを抽出
	var groupPrefix string
	if strings.HasPrefix(groupID, "TEST-") {
		// TEST-TS001TG01形式のIDからプレフィックスを抽出
		parts := strings.Split(groupID, "-")
		if len(parts) >= 3 {
			// "TS001TG01"部分を取得
			prefixParts := strings.Split(parts[1], "-")
			groupPrefix = prefixParts[0]
		} else {
			return "", fmt.Errorf("無効なテストグループID形式: %s", groupID)
		}
	} else {
		// 通常のTS001TG01-202503形式からプレフィックス抽出
		parts := strings.Split(groupID, "-")
		if len(parts) >= 1 {
			groupPrefix = parts[0] // "TS001TG01"部分を取得
		} else {
			return "", fmt.Errorf("無効なグループID形式: %s", groupID)
		}
	}

	// ID生成
	id := fmt.Sprintf("TEST-%sTC%03d-%s", groupPrefix, randomPart, yearMonth)
	return id, nil
}
