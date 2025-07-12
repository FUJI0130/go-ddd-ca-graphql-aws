package interactor

import (
	"context"
	"fmt"
	"time"

	"github.com/FUJI0130/go-ddd-ca/internal/domain/entity"
	"github.com/FUJI0130/go-ddd-ca/internal/domain/repository"
	"github.com/FUJI0130/go-ddd-ca/internal/domain/valueobject"
	"github.com/FUJI0130/go-ddd-ca/internal/usecase/dto"
	"github.com/FUJI0130/go-ddd-ca/pkg/errors"
)

// TestSuiteInteractor はテストスイートのユースケースを実装します
type TestSuiteInteractor struct {
	repository  repository.TestSuiteRepository
	idGenerator repository.TestSuiteIDGenerator
}

// NewTestSuiteInteractor は新しいTestSuiteInteractorを作成します
func NewTestSuiteInteractor(repo repository.TestSuiteRepository, idGenerator repository.TestSuiteIDGenerator) *TestSuiteInteractor {
	return &TestSuiteInteractor{
		repository:  repo,
		idGenerator: idGenerator,
	}
}

// CreateTestSuite は新しいテストスイートを作成します
func (i *TestSuiteInteractor) CreateTestSuite(ctx context.Context, createDTO *dto.TestSuiteCreateDTO) (*dto.TestSuiteResponseDTO, error) {
	currentTime := time.Now()
	id := fmt.Sprintf("TS%03d-%04d%02d", 1, currentTime.Year(), currentTime.Month())

	suite := &entity.TestSuite{
		ID:                   id,
		Name:                 createDTO.Name,
		Description:          createDTO.Description,
		Status:               valueobject.SuiteStatusPreparation,
		EstimatedStartDate:   createDTO.EstimatedStartDate,
		EstimatedEndDate:     createDTO.EstimatedEndDate,
		RequireEffortComment: createDTO.RequireEffortComment,
		CreatedAt:            currentTime,
		UpdatedAt:            currentTime,
	}

	if err := i.repository.Create(ctx, suite); err != nil {
		// ドメインエラーに変換
		if errors.IsDomainError(err) {
			return nil, err
		}
		return nil, errors.NewSystemError("テストスイートの作成に失敗しました", err)
	}

	// レスポンスDTOの作成
	responseDTO := &dto.TestSuiteResponseDTO{
		ID:                   suite.ID,
		Name:                 suite.Name,
		Description:          suite.Description,
		Status:               suite.Status.String(),
		EstimatedStartDate:   suite.EstimatedStartDate,
		EstimatedEndDate:     suite.EstimatedEndDate,
		RequireEffortComment: suite.RequireEffortComment,
		Progress:             0.0,
		CreatedAt:            suite.CreatedAt,
		UpdatedAt:            suite.UpdatedAt,
	}

	return responseDTO, nil
}

// GetTestSuite は指定されたIDのテストスイートを取得します
func (i *TestSuiteInteractor) GetTestSuite(ctx context.Context, id string) (*dto.TestSuiteResponseDTO, error) {
	suite, err := i.repository.FindByID(ctx, id)
	if err != nil {
		// ドメインエラーに変換
		if errors.IsDomainError(err) {
			return nil, err
		}
		return nil, errors.NewTestSuiteNotFoundError(id)
	}

	// レスポンスDTOの作成
	responseDTO := &dto.TestSuiteResponseDTO{
		ID:                   suite.ID,
		Name:                 suite.Name,
		Description:          suite.Description,
		Status:               suite.Status.String(),
		EstimatedStartDate:   suite.EstimatedStartDate,
		EstimatedEndDate:     suite.EstimatedEndDate,
		RequireEffortComment: suite.RequireEffortComment,
		Progress:             i.calculateProgress(suite),
		CreatedAt:            suite.CreatedAt,
		UpdatedAt:            suite.UpdatedAt,
	}

	return responseDTO, nil
}

// calculateProgress はテストスイートの進捗率を計算します
func (i *TestSuiteInteractor) calculateProgress(suite *entity.TestSuite) float64 {
	// 状態に基づく基本進捗率の定義
	statusProgress := map[valueobject.SuiteStatus]float64{
		valueobject.SuiteStatusPreparation: 0.0,   // 準備中
		valueobject.SuiteStatusInProgress:  50.0,  // 実行中
		valueobject.SuiteStatusCompleted:   100.0, // 完了
		valueobject.SuiteStatusSuspended:   75.0,  // 中断
	}

	return statusProgress[suite.Status]
}

// UpdateTestSuite はテストスイートを更新します
func (i *TestSuiteInteractor) UpdateTestSuite(ctx context.Context, id string, updateDTO *dto.TestSuiteUpdateDTO) (*dto.TestSuiteResponseDTO, error) {
	// 既存のテストスイートを取得
	suite, err := i.repository.FindByID(ctx, id)
	if err != nil {
		// ドメインエラーに変換
		if errors.IsDomainError(err) {
			return nil, err
		}
		return nil, errors.NewTestSuiteNotFoundError(id)
	}

	// 更新が存在する場合のみ値を更新
	if updateDTO.Name != nil {
		suite.Name = *updateDTO.Name
	}
	if updateDTO.Description != nil {
		suite.Description = *updateDTO.Description
	}
	if updateDTO.EstimatedStartDate != nil {
		suite.EstimatedStartDate = *updateDTO.EstimatedStartDate
	}
	if updateDTO.EstimatedEndDate != nil {
		suite.EstimatedEndDate = *updateDTO.EstimatedEndDate
	}
	if updateDTO.RequireEffortComment != nil {
		suite.RequireEffortComment = *updateDTO.RequireEffortComment
	}

	suite.UpdatedAt = time.Now()

	// リポジトリで更新を実行
	if err := i.repository.Update(ctx, suite); err != nil {
		// ドメインエラーに変換
		if errors.IsDomainError(err) {
			return nil, err
		}
		return nil, errors.NewSystemError("テストスイートの更新に失敗しました", err)
	}

	// レスポンスDTOの作成
	responseDTO := &dto.TestSuiteResponseDTO{
		ID:                   suite.ID,
		Name:                 suite.Name,
		Description:          suite.Description,
		Status:               suite.Status.String(),
		EstimatedStartDate:   suite.EstimatedStartDate,
		EstimatedEndDate:     suite.EstimatedEndDate,
		RequireEffortComment: suite.RequireEffortComment,
		Progress:             i.calculateProgress(suite),
		CreatedAt:            suite.CreatedAt,
		UpdatedAt:            suite.UpdatedAt,
	}

	return responseDTO, nil
}

// UpdateTestSuiteStatus はテストスイートのステータスを更新します
func (i *TestSuiteInteractor) UpdateTestSuiteStatus(ctx context.Context, id string, statusDTO *dto.TestSuiteStatusUpdateDTO) (*dto.TestSuiteResponseDTO, error) {
	// 既存のテストスイートを取得
	suite, err := i.repository.FindByID(ctx, id)
	if err != nil {
		// ドメインエラーに変換
		if errors.IsDomainError(err) {
			return nil, err
		}
		return nil, errors.NewTestSuiteNotFoundError(id)
	}

	// 新しいステータスを値オブジェクトに変換
	newStatus, err := valueobject.NewSuiteStatus(statusDTO.Status)
	if err != nil {
		return nil, errors.NewDomainValidationError("無効なステータス値です", map[string]string{
			"status": statusDTO.Status,
			"error":  err.Error(),
		})
	}

	// ステータス遷移の検証
	if !isValidStatusTransition(suite.Status, newStatus) {
		return nil, errors.NewStatusTransitionConflictError(
			"このステータスへの変更は許可されていません",
			suite.Status.String(),
			newStatus.String(),
		)
	}

	// ステータスの更新
	suite.Status = newStatus
	suite.UpdatedAt = time.Now()

	// リポジトリで更新を実行
	if err := i.repository.Update(ctx, suite); err != nil {
		// ドメインエラーに変換
		if errors.IsDomainError(err) {
			return nil, err
		}
		return nil, errors.NewSystemError("テストスイートのステータス更新に失敗しました", err)
	}

	// レスポンスDTOの作成
	responseDTO := &dto.TestSuiteResponseDTO{
		ID:                   suite.ID,
		Name:                 suite.Name,
		Description:          suite.Description,
		Status:               suite.Status.String(),
		EstimatedStartDate:   suite.EstimatedStartDate,
		EstimatedEndDate:     suite.EstimatedEndDate,
		RequireEffortComment: suite.RequireEffortComment,
		Progress:             i.calculateProgress(suite),
		CreatedAt:            suite.CreatedAt,
		UpdatedAt:            suite.UpdatedAt,
	}

	return responseDTO, nil
}

// DeleteTestSuite は指定されたIDのテストスイートを削除します
func (i *TestSuiteInteractor) DeleteTestSuite(ctx context.Context, id string) error {
	// 存在確認
	_, err := i.repository.FindByID(ctx, id)
	if err != nil {
		if errors.IsDomainError(err) {
			return err
		}
		return errors.NewTestSuiteNotFoundError(id)
	}

	// 削除の実行
	if err := i.repository.Delete(ctx, id); err != nil {
		if errors.IsDomainError(err) {
			return err
		}
		return errors.NewSystemError("テストスイートの削除に失敗しました", err)
	}

	return nil
}

// isValidStatusTransition はステータス遷移が有効かどうかを検証します
func isValidStatusTransition(current, new valueobject.SuiteStatus) bool {
	// ステータス遷移ルールの定義
	validTransitions := map[valueobject.SuiteStatus][]valueobject.SuiteStatus{
		valueobject.SuiteStatusPreparation: {
			valueobject.SuiteStatusInProgress,
			valueobject.SuiteStatusSuspended,
		},
		valueobject.SuiteStatusInProgress: {
			valueobject.SuiteStatusCompleted,
			valueobject.SuiteStatusSuspended,
		},
		valueobject.SuiteStatusSuspended: {
			valueobject.SuiteStatusInProgress,
		},
		valueobject.SuiteStatusCompleted: {
			valueobject.SuiteStatusInProgress,
		},
	}

	// 現在のステータスから遷移可能なステータスのリストを取得
	allowedTransitions, exists := validTransitions[current]
	if !exists {
		return false
	}

	// 新しいステータスが遷移可能なリストに含まれているか確認
	for _, allowed := range allowedTransitions {
		if allowed == new {
			return true
		}
	}

	return false
}

// ListTestSuites はテストスイート一覧を取得します
func (i *TestSuiteInteractor) ListTestSuites(ctx context.Context, params *dto.TestSuiteQueryParamDTO) (*dto.TestSuiteListResponseDTO, error) {
	// リポジトリからデータを取得
	suites, total, err := i.repository.FindWithFilters(ctx, params)
	if err != nil {
		// ドメインエラーに変換
		if errors.IsDomainError(err) {
			return nil, err
		}
		return nil, errors.NewSystemError("テストスイート一覧の取得に失敗しました", err)
	}

	// エンティティをDTOに変換
	response := &dto.TestSuiteListResponseDTO{
		TestSuites: make([]dto.TestSuiteResponseDTO, 0, len(suites)),
		Total:      total,
	}

	for _, suite := range suites {
		responseDTO := dto.TestSuiteResponseDTO{
			ID:                   suite.ID,
			Name:                 suite.Name,
			Description:          suite.Description,
			Status:               suite.Status.String(),
			EstimatedStartDate:   suite.EstimatedStartDate,
			EstimatedEndDate:     suite.EstimatedEndDate,
			RequireEffortComment: suite.RequireEffortComment,
			Progress:             i.calculateProgress(suite),
			CreatedAt:            suite.CreatedAt,
			UpdatedAt:            suite.UpdatedAt,
		}
		response.TestSuites = append(response.TestSuites, responseDTO)
	}

	return response, nil
}
