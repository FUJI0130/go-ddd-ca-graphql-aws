// ファイル: internal/domain/repository/test_case_repository.go
package repository

import (
	"context"

	"github.com/FUJI0130/go-ddd-ca/internal/domain/entity"
)

// TestCaseRepository はテストケースの永続化を担当するリポジトリインターフェース
type TestCaseRepository interface {
	Repository[entity.TestCase]

	// FindByGroupID は指定されたグループIDに属するテストケース一覧を取得する
	FindByGroupID(ctx context.Context, groupID string) ([]*entity.TestCase, error)

	// UpdateStatus は指定されたテストケースのステータスを更新する
	UpdateStatus(ctx context.Context, id string, status entity.TestStatus) error

	// AddEffort は指定されたテストケースに工数を追加する
	AddEffort(ctx context.Context, id string, effort float64) error

	// FindByStatus は指定されたステータスのテストケース一覧を取得する
	FindByStatus(ctx context.Context, status entity.TestStatus) ([]*entity.TestCase, error)
}
