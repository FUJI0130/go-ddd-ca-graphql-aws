package customerrors

import (
	"fmt"

	"github.com/cockroachdb/errors"
)

// InternalServerError はサーバー内部エラー
type InternalServerError struct {
	*BaseErr
}

// NewInternalServerError は新しい InternalServerError を作成
func NewInternalServerError(message string) *InternalServerError {
	return &InternalServerError{
		&BaseErr{
			MessageVal:    message,
			StatusCodeVal: StatusCodeInternalServerError,
			TraceVal:      errors.New(message),
		},
	}
}

// NewInternalServerErrorf はフォーマット指定付きで新しい InternalServerError を作成
func NewInternalServerErrorf(format string, args ...any) *InternalServerError {
	message := fmt.Sprintf(format, args...)
	return NewInternalServerError(message)
}

// WrapInternalServerError は既存のエラーをラップした InternalServerError を作成
func WrapInternalServerError(err error, message string) *InternalServerError {
	baseError := NewBaseError(message, StatusCodeInternalServerError, nil)
	wrappedError := baseError.WrapWithLocation(err, message)
	return &InternalServerError{
		BaseErr: wrappedError,
	}
}

// WrapInternalServerErrorf はフォーマット指定付きで既存のエラーをラップした InternalServerError を作成
func WrapInternalServerErrorf(err error, format string, args ...any) *InternalServerError {
	message := fmt.Sprintf(format, args...)
	return WrapInternalServerError(err, message)
}

// WithContext はエラーにコンテキスト情報を追加
func (ise *InternalServerError) WithContext(ctx Context) *InternalServerError {
	return &InternalServerError{
		BaseErr: ise.BaseErr.WithContext(ctx),
	}
}
