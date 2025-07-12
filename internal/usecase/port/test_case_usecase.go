package port

import (
	"context"

	"github.com/FUJI0130/go-ddd-ca/internal/usecase/dto"
)

// TestCaseUseCase はテストケースのユースケースインターフェース
type TestCaseUseCase interface {
	// GetCasesByGroupID は指定されたグループIDに属するケース一覧を取得する
	GetCasesByGroupID(ctx context.Context, groupID string) ([]*dto.TestCaseResponseDTO, error)
}
