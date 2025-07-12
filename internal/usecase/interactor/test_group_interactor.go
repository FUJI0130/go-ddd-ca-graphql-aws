package interactor

import (
	"context"
	"time"

	"github.com/FUJI0130/go-ddd-ca/internal/domain/entity"
	"github.com/FUJI0130/go-ddd-ca/internal/domain/repository"
	"github.com/FUJI0130/go-ddd-ca/internal/domain/valueobject"
	"github.com/FUJI0130/go-ddd-ca/internal/usecase/dto"
	"github.com/FUJI0130/go-ddd-ca/pkg/errors"
)

// TestGroupInteractor はテストグループのユースケース実装
type TestGroupInteractor struct {
	testGroupRepo repository.TestGroupRepository
	idGenerator   repository.TestGroupIDGenerator
}

// NewTestGroupInteractor は新しいTestGroupInteractorを作成します
func NewTestGroupInteractor(testGroupRepo repository.TestGroupRepository, idGenerator repository.TestGroupIDGenerator) *TestGroupInteractor {
	return &TestGroupInteractor{
		testGroupRepo: testGroupRepo,
		idGenerator:   idGenerator,
	}
}

// GetGroupsBySuiteID は指定されたスイートIDに属するグループ一覧を取得する
func (i *TestGroupInteractor) GetGroupsBySuiteID(ctx context.Context, suiteID string) ([]*dto.TestGroupResponseDTO, error) {
	// リポジトリからデータを取得
	groups, err := i.testGroupRepo.FindBySuiteID(ctx, suiteID)
	if err != nil {
		return nil, err
	}

	// エンティティからDTOに変換
	result := make([]*dto.TestGroupResponseDTO, len(groups))
	for j, group := range groups {
		result[j] = &dto.TestGroupResponseDTO{
			ID:           group.ID,
			SuiteID:      group.SuiteID,
			Name:         group.Name,
			Description:  group.Description,
			DisplayOrder: group.DisplayOrder,
			Status:       group.Status.String(),
			CreatedAt:    group.CreatedAt,
			UpdatedAt:    group.UpdatedAt,
		}
	}

	return result, nil
}

// CreateTestGroup は新しいテストグループを作成します
func (i *TestGroupInteractor) CreateTestGroup(ctx context.Context, createDTO *dto.TestGroupCreateDTO) (*dto.TestGroupResponseDTO, error) {
	// 入力検証
	if createDTO.SuiteID == "" {
		return nil, errors.NewDomainValidationError("スイートIDは必須です", nil)
	}
	if createDTO.Name == "" {
		return nil, errors.NewDomainValidationError("グループ名は必須です", nil)
	}

	// IDの生成
	id, err := i.idGenerator.GenerateID(createDTO.SuiteID)
	if err != nil {
		return nil, err
	}

	// ステータスの設定
	status, err := valueobject.NewSuiteStatus("準備中")
	if err != nil {
		return nil, err
	}

	// エンティティの作成
	now := time.Now()
	group := &entity.TestGroup{
		ID:           id,
		SuiteID:      createDTO.SuiteID,
		Name:         createDTO.Name,
		Description:  createDTO.Description,
		DisplayOrder: createDTO.DisplayOrder,
		Status:       status,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// リポジトリに保存
	err = i.testGroupRepo.Create(ctx, group)
	if err != nil {
		if errors.IsDomainError(err) {
			return nil, err
		}
		return nil, errors.NewSystemError("テストグループの作成に失敗しました", err)
	}

	// レスポンスDTOの作成
	return &dto.TestGroupResponseDTO{
		ID:           group.ID,
		SuiteID:      group.SuiteID,
		Name:         group.Name,
		Description:  group.Description,
		DisplayOrder: group.DisplayOrder,
		Status:       group.Status.String(),
		CreatedAt:    group.CreatedAt,
		UpdatedAt:    group.UpdatedAt,
	}, nil
}
