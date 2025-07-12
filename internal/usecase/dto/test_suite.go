package dto

import (
	"time"

	"github.com/FUJI0130/go-ddd-ca/pkg/errors"
)

// TestSuiteCreateDTO はテストスイート作成時のリクエストDTO
type TestSuiteCreateDTO struct {
	Name                 string    `json:"name" validate:"required,min=1,max=100"`
	Description          string    `json:"description"`
	EstimatedStartDate   time.Time `json:"estimatedStartDate" validate:"required"`
	EstimatedEndDate     time.Time `json:"estimatedEndDate" validate:"required,gtfield=EstimatedStartDate"`
	RequireEffortComment bool      `json:"requireEffortComment"`
}

// TestSuiteResponseDTO はテストスイート情報を返すためのDTO
type TestSuiteResponseDTO struct {
	ID                   string    `json:"id"`
	Name                 string    `json:"name"`
	Description          string    `json:"description"`
	Status               string    `json:"status"`
	EstimatedStartDate   time.Time `json:"estimatedStartDate"`
	EstimatedEndDate     time.Time `json:"estimatedEndDate"`
	RequireEffortComment bool      `json:"requireEffortComment"`
	Progress             float64   `json:"progress"`
	CreatedAt            time.Time `json:"createdAt"`
	UpdatedAt            time.Time `json:"updatedAt"`
}

// TestSuiteListResponseDTO はテストスイート一覧を返すためのDTO
type TestSuiteListResponseDTO struct {
	TestSuites []TestSuiteResponseDTO `json:"testSuites"`
	Total      int                    `json:"total"`
}

// TestSuiteUpdateDTO は、テストスイート更新時のリクエストデータを表現します
type TestSuiteUpdateDTO struct {
	Name                 *string    `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Description          *string    `json:"description,omitempty"`
	EstimatedStartDate   *time.Time `json:"estimatedStartDate,omitempty" validate:"omitempty"`
	EstimatedEndDate     *time.Time `json:"estimatedEndDate,omitempty" validate:"omitempty,gtfield=EstimatedStartDate"`
	RequireEffortComment *bool      `json:"requireEffortComment,omitempty"`
}

// Validate は、DTOのカスタムバリデーションを実行します
func (dto *TestSuiteUpdateDTO) Validate() error {
	// 日付の相対バリデーション
	if dto.EstimatedStartDate != nil && dto.EstimatedEndDate != nil {
		if dto.EstimatedEndDate.Before(*dto.EstimatedStartDate) {
			return errors.ErrInvalidDateRange
		}
	}
	return nil
}

type TestSuiteStatusUpdateDTO struct {
	Status string `json:"status" validate:"required,oneof=準備中 実行中 完了 中断"`
}

// Validate は、ステータス更新のバリデーションを実行します
func (dto *TestSuiteStatusUpdateDTO) Validate() error {
	// 将来的にステータス遷移のバリデーションをここに実装
	return nil
}

// TestSuiteQueryParamDTO は、テストスイート一覧取得時のクエリパラメータを表現します
type TestSuiteQueryParamDTO struct {
	Status    *string    `json:"status" validate:"omitempty,oneof=準備中 実行中 完了 中断"`
	StartDate *time.Time `json:"startDate" validate:"omitempty"`
	EndDate   *time.Time `json:"endDate" validate:"omitempty"`
	Page      *int       `json:"page" validate:"omitempty,min=1"`
	PageSize  *int       `json:"pageSize" validate:"omitempty,min=1,max=100"`
}

// Validate はクエリパラメータのカスタムバリデーションを実行します
func (dto *TestSuiteQueryParamDTO) Validate() error {
	if dto.StartDate != nil && dto.EndDate != nil {
		if dto.EndDate.Before(*dto.StartDate) {
			return errors.ErrInvalidDateRange
		}
	}
	return nil
}
