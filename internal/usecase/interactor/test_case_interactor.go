package interactor

import (
	"context"
	"time"

	"github.com/FUJI0130/go-ddd-ca/internal/domain/entity"
	"github.com/FUJI0130/go-ddd-ca/internal/domain/repository"
	"github.com/FUJI0130/go-ddd-ca/internal/usecase/dto"
	"github.com/FUJI0130/go-ddd-ca/pkg/errors"
)

// TestCaseInteractor はテストケースのユースケース実装
type TestCaseInteractor struct {
	testCaseRepo repository.TestCaseRepository
	idGenerator  repository.TestCaseIDGenerator
}

// NewTestCaseInteractor は新しいTestCaseInteractorを作成します
func NewTestCaseInteractor(testCaseRepo repository.TestCaseRepository, idGenerator repository.TestCaseIDGenerator) *TestCaseInteractor {
	return &TestCaseInteractor{
		testCaseRepo: testCaseRepo,
		idGenerator:  idGenerator,
	}
}

// GetCasesByGroupID は指定されたグループIDに属するケース一覧を取得する
func (i *TestCaseInteractor) GetCasesByGroupID(ctx context.Context, groupID string) ([]*dto.TestCaseResponseDTO, error) {
	// リポジトリからデータを取得
	cases, err := i.testCaseRepo.FindByGroupID(ctx, groupID)
	if err != nil {
		return nil, err
	}

	// エンティティからDTOに変換
	result := make([]*dto.TestCaseResponseDTO, len(cases))
	for j, tc := range cases {
		result[j] = &dto.TestCaseResponseDTO{
			ID:            tc.ID,
			GroupID:       tc.GroupID,
			Title:         tc.Title,
			Description:   tc.Description,
			Status:        string(tc.Status),
			Priority:      string(tc.Priority),
			PlannedEffort: tc.PlannedEffort,
			ActualEffort:  tc.ActualEffort,
			IsDelayed:     tc.IsDelayed,
			DelayDays:     tc.DelayDays,
			CreatedAt:     tc.CreatedAt,
			UpdatedAt:     tc.UpdatedAt,
		}
	}

	return result, nil
}

// CreateTestCase は新しいテストケースを作成します
func (i *TestCaseInteractor) CreateTestCase(ctx context.Context, createDTO *dto.TestCaseCreateDTO) (*dto.TestCaseResponseDTO, error) {
	// 入力検証
	if createDTO.GroupID == "" {
		return nil, errors.NewDomainValidationError("グループIDは必須です", nil)
	}
	if createDTO.Title == "" {
		return nil, errors.NewDomainValidationError("タイトルは必須です", nil)
	}

	// IDの生成
	id, err := i.idGenerator.GenerateID(createDTO.GroupID)
	if err != nil {
		return nil, err
	}

	// ステータスの設定（デフォルトは「作成」）
	status := entity.TestStatusCreated

	// 優先度の設定
	priority := entity.Priority(createDTO.Priority)
	if priority == "" {
		priority = entity.PriorityMedium // デフォルト値
	}

	// エンティティの作成
	now := time.Now()
	testCase := &entity.TestCase{
		ID:            id,
		GroupID:       createDTO.GroupID,
		Title:         createDTO.Title,
		Description:   createDTO.Description,
		Status:        status,
		Priority:      priority,
		PlannedEffort: createDTO.PlannedEffort,
		ActualEffort:  0, // 新規作成時は0
		IsDelayed:     false,
		DelayDays:     0,
		CurrentEditor: "",
		IsLocked:      false,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	// リポジトリに保存
	err = i.testCaseRepo.Create(ctx, testCase)
	if err != nil {
		if errors.IsDomainError(err) {
			return nil, err
		}
		return nil, errors.NewSystemError("テストケースの作成に失敗しました", err)
	}

	// レスポンスDTOの作成
	return &dto.TestCaseResponseDTO{
		ID:            testCase.ID,
		GroupID:       testCase.GroupID,
		Title:         testCase.Title,
		Description:   testCase.Description,
		Status:        string(testCase.Status),
		Priority:      string(testCase.Priority),
		PlannedEffort: testCase.PlannedEffort,
		ActualEffort:  testCase.ActualEffort,
		IsDelayed:     testCase.IsDelayed,
		DelayDays:     testCase.DelayDays,
		CreatedAt:     testCase.CreatedAt,
		UpdatedAt:     testCase.UpdatedAt,
	}, nil
}
