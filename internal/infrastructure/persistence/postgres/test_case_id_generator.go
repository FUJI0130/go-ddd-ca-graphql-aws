package postgres

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/FUJI0130/go-ddd-ca/internal/domain/repository"
)

// TestCaseIDGenerator はテストケースIDを生成するジェネレーター
type TestCaseIDGenerator struct {
	db *sql.DB
}

// インターフェース実装の確認
var _ repository.TestCaseIDGenerator = (*TestCaseIDGenerator)(nil)

// NewTestCaseIDGenerator は新しいTestCaseIDGeneratorを作成します
func NewTestCaseIDGenerator(db *sql.DB) *TestCaseIDGenerator {
	return &TestCaseIDGenerator{db: db}
}

// GenerateID はグループIDに紐づいた新しいケースIDを生成します
func (g *TestCaseIDGenerator) GenerateID(groupID string) (string, error) {
	var seq int
	err := g.db.QueryRow("SELECT nextval('test_case_seq')").Scan(&seq)
	if err != nil {
		return "", fmt.Errorf("failed to generate sequence number: %w", err)
	}

	currentTime := time.Now()
	// グループIDからプレフィックスを抽出（"TS001TG01"部分）
	groupPrefix := groupID[:8]
	id := fmt.Sprintf("%sTC%03d-%d%02d", groupPrefix, seq, currentTime.Year(), currentTime.Month())
	return id, nil
}
