package customerrors

import (
	"fmt"

	"github.com/cockroachdb/errors"
)

// ValidationError はバリデーションエラー
type ValidationError struct {
	*BaseErr
	Details map[string]string
}

// 既存の関数シグネチャを維持（互換性のため）
func NewValidationError(message string, details map[string]string) *ValidationError {
	ve := &ValidationError{
		BaseErr: &BaseErr{
			MessageVal:    message,
			StatusCodeVal: StatusCodeUnprocessableEntity,
			TraceVal:      errors.New(message),
			ContextVal:    nil,
		},
		Details: details,
	}
	return ve
}

// 新しいフルイドインターフェースのための関数
func NewSimpleValidationError(message string) *ValidationError {
	return &ValidationError{
		BaseErr: &BaseErr{
			MessageVal:    message,
			StatusCodeVal: StatusCodeUnprocessableEntity,
			TraceVal:      errors.New(message),
			ContextVal:    nil,
		},
		Details: nil,
	}
}

// WithDetails はバリデーションエラーに詳細情報を追加
func (ve *ValidationError) WithDetails(details map[string]string) *ValidationError {
	ve.Details = details
	return ve
}

// WithContext はベースのWithContextをオーバーライド
func (ve *ValidationError) WithContext(ctx Context) *ValidationError {
	return &ValidationError{
		BaseErr: ve.BaseErr.WithContext(ctx),
		Details: ve.Details,
	}
}

// NewValidationErrorf はフォーマット指定付きで新しい ValidationError を作成
func NewValidationErrorf(format string, args ...any) *ValidationError {
	message := fmt.Sprintf(format, args...)
	return NewValidationError(message, nil)
}

// WrapValidationError は既存のエラーをラップした ValidationError を作成
func WrapValidationError(err error, message string, details map[string]string) *ValidationError {
	baseError := NewBaseError(message, StatusCodeUnprocessableEntity, nil)
	wrappedError := baseError.WrapWithLocation(err, message)
	return &ValidationError{
		BaseErr: wrappedError,
		Details: details,
	}
}

// GetDetails はバリデーションエラーの詳細を返す
func (ve *ValidationError) GetDetails() map[string]string {
	return ve.Details
}
