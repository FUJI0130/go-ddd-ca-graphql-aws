// ファイル: internal/domain/repository/test_group_repository.go
package repository

import (
	"context"

	"github.com/FUJI0130/go-ddd-ca/internal/domain/entity"
	"github.com/FUJI0130/go-ddd-ca/internal/domain/valueobject"
)

// TestGroupRepository はテストグループの永続化を担当するリポジトリインターフェース
type TestGroupRepository interface {
	Repository[entity.TestGroup]

	// FindBySuiteID は指定されたスイートIDに属するテストグループ一覧を取得する
	FindBySuiteID(ctx context.Context, suiteID string) ([]*entity.TestGroup, error)

	// UpdateStatus は指定されたテストグループのステータスを更新する
	UpdateStatus(ctx context.Context, id string, status valueobject.SuiteStatus) error

	// UpdateDisplayOrder は指定されたテストグループの表示順序を更新する
	UpdateDisplayOrder(ctx context.Context, id string, displayOrder int) error
}
