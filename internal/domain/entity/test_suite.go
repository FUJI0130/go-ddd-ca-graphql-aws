package entity

import (
	"time"

	"github.com/FUJI0130/go-ddd-ca/internal/domain/valueobject"
	"github.com/FUJI0130/go-ddd-ca/support/customerrors"
)

// TestSuite はテストスイートを表すエンティティ
type TestSuite struct {
	ID                   string
	Name                 string
	Description          string
	Status               valueobject.SuiteStatus
	EstimatedStartDate   time.Time
	EstimatedEndDate     time.Time
	RequireEffortComment bool
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// エラーメッセージ定数
const (
	MsgTestSuiteIDRequired       = "テストスイートIDは必須です"
	MsgTestSuiteNameRequired     = "テストスイート名は必須です"
	MsgTestSuiteInvalidDateRange = "終了予定日は開始予定日より後である必要があります"
	MsgTestSuiteInvalidStatus    = "無効なステータスです: %s"
	MsgTestSuiteNotFound         = "テストスイートが見つかりません"
	MsgTestSuiteConcurrentUpdate = "同時更新が検出されました"
)

var (
	ErrTestSuiteNotFound       = customerrors.NewNotFoundError(MsgTestSuiteNotFound)
	ErrConcurrentModification  = customerrors.NewConflictError(MsgTestSuiteConcurrentUpdate)
	ErrInvalidStatusTransition = customerrors.NewValidationError("無効なステータス遷移です", nil)
)

// NewTestSuite は新しいTestSuiteエンティティを生成する
func NewTestSuite(
	id string,
	name string,
	description string,
	estimatedStartDate time.Time,
	estimatedEndDate time.Time,
	requireEffortComment bool,
) (*TestSuite, error) {
	// 基本的なバリデーション
	if id == "" {
		return nil, customerrors.NewValidationError(MsgTestSuiteIDRequired, nil)
	}
	if name == "" {
		return nil, customerrors.NewValidationError(MsgTestSuiteNameRequired, nil)
	}
	if estimatedEndDate.Before(estimatedStartDate) {
		return nil, customerrors.NewValidationError(MsgTestSuiteInvalidDateRange, nil)
	}

	now := time.Now()
	return &TestSuite{
		ID:                   id,
		Name:                 name,
		Description:          description,
		Status:               valueobject.SuiteStatusPreparation, // 初期状態は準備中
		EstimatedStartDate:   estimatedStartDate,
		EstimatedEndDate:     estimatedEndDate,
		RequireEffortComment: requireEffortComment,
		CreatedAt:            now,
		UpdatedAt:            now,
	}, nil
}

// UpdateStatus はテストスイートのステータスを更新する
func (ts *TestSuite) UpdateStatus(newStatus valueobject.SuiteStatus) error {
	if !newStatus.IsValid() {
		return customerrors.NewValidationErrorf(MsgTestSuiteInvalidStatus, newStatus)
	}
	ts.Status = newStatus
	ts.UpdatedAt = time.Now()
	return nil
}
