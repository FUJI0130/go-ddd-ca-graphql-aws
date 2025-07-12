package postgres

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/FUJI0130/go-ddd-ca/internal/domain/repository"
)

// インターフェース実装の確認
var _ repository.TestSuiteIDGenerator = (*TestSuiteIDGenerator)(nil)

type TestSuiteIDGenerator struct {
	db *sql.DB
}

func NewTestSuiteIDGenerator(db *sql.DB) *TestSuiteIDGenerator {
	return &TestSuiteIDGenerator{db: db}
}

func (g *TestSuiteIDGenerator) GenerateID() (string, error) {
	var seq int
	err := g.db.QueryRow("SELECT nextval('test_suite_seq')").Scan(&seq)
	if err != nil {
		return "", fmt.Errorf("failed to generate sequence number: %w", err)
	}

	currentTime := time.Now()
	id := fmt.Sprintf("TS%03d-%d%02d", seq, currentTime.Year(), currentTime.Month())
	return id, nil
}
