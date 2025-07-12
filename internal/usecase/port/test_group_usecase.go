package port

import (
	"context"

	"github.com/FUJI0130/go-ddd-ca/internal/usecase/dto"
)

// TestGroupUseCase はテストグループのユースケースインターフェース
type TestGroupUseCase interface {
	// GetGroupsBySuiteID は指定されたスイートIDに属するグループ一覧を取得する
	GetGroupsBySuiteID(ctx context.Context, suiteID string) ([]*dto.TestGroupResponseDTO, error)
	// 追加するメソッド
	CreateTestGroup(ctx context.Context, dto *dto.TestGroupCreateDTO) (*dto.TestGroupResponseDTO, error)
}
