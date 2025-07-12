package repository

import (
	"context"

	"github.com/FUJI0130/go-ddd-ca/internal/domain/entity"
	"github.com/FUJI0130/go-ddd-ca/internal/domain/valueobject"
	"github.com/FUJI0130/go-ddd-ca/internal/usecase/dto"
)

// TestSuiteRepository はテストスイートの永続化を担当するリポジトリインターフェース
type TestSuiteRepository interface {
	Repository[entity.TestSuite]

	// UpdateStatus は指定されたテストスイートのステータスを更新する
	// ctx: コンテキスト
	// id: テストスイートID
	// status: 更新後のステータス
	UpdateStatus(ctx context.Context, id string, status valueobject.SuiteStatus) error

	// FindByStatus は指定されたステータスのテストスイート一覧を取得する
	// ctx: コンテキスト
	// status: 検索対象のステータス
	FindByStatus(ctx context.Context, status valueobject.SuiteStatus) ([]*entity.TestSuite, error)

	// TestSuiteRepository interfaceに追加
	FindWithFilters(ctx context.Context, params *dto.TestSuiteQueryParamDTO) ([]*entity.TestSuite, int, error)
}
