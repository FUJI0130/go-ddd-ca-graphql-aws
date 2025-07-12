package postgres

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/FUJI0130/go-ddd-ca/internal/domain/repository"
)

// TestGroupIDGenerator はテストグループIDを生成するジェネレーター
type TestGroupIDGenerator struct {
	db *sql.DB
}

// インターフェース実装の確認
var _ repository.TestGroupIDGenerator = (*TestGroupIDGenerator)(nil)

// NewTestGroupIDGenerator は新しいTestGroupIDGeneratorを作成します
func NewTestGroupIDGenerator(db *sql.DB) *TestGroupIDGenerator {
	return &TestGroupIDGenerator{db: db}
}

// GenerateID はスイートIDに紐づいた新しいグループIDを生成します
func (g *TestGroupIDGenerator) GenerateID(suiteID string) (string, error) {
	var seq int
	err := g.db.QueryRow("SELECT nextval('test_group_seq')").Scan(&seq)
	if err != nil {
		return "", fmt.Errorf("failed to generate sequence number: %w", err)
	}

	currentTime := time.Now()
	// スイートIDからプレフィックスを抽出（"TS001"部分）
	suitePrefix := suiteID[:5]
	id := fmt.Sprintf("%sTG%02d-%d%02d", suitePrefix, seq, currentTime.Year(), currentTime.Month())
	return id, nil
}
