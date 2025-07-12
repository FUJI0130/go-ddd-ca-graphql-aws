// test/helper/test_data.go
package helper

import (
	"database/sql"
	"time"

	"github.com/FUJI0130/go-ddd-ca/internal/domain/entity"
	"github.com/FUJI0130/go-ddd-ca/internal/domain/valueobject"
)

// CreateTestSuite はテスト用のTestSuiteエンティティを生成します
func CreateTestSuite(id string) *entity.TestSuite {
	return &entity.TestSuite{
		ID:                   id,
		Name:                 "テストスイート" + id,
		Description:          "テスト用データ" + id,
		Status:               valueobject.SuiteStatusPreparation,
		EstimatedStartDate:   time.Now(),
		EstimatedEndDate:     time.Now().AddDate(0, 1, 0),
		RequireEffortComment: true,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}
}

// VerifyTestSuiteExists はテストスイートの存在を検証します
func VerifyTestSuiteExists(db *sql.DB, id string) (bool, error) {
	var exists bool
	err := db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM test_suites WHERE id = $1)",
		id,
	).Scan(&exists)
	return exists, err
}
