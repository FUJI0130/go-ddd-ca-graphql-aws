package customerrors

import (
	"fmt"

	"github.com/cockroachdb/errors"
)

// NotFoundError は要求されたリソースが見つからない場合のエラー
type NotFoundError struct {
	*BaseErr
}

// NewNotFoundError は新しい NotFoundError を作成
func NewNotFoundError(message string) *NotFoundError {
	return &NotFoundError{
		&BaseErr{
			MessageVal:    message,
			StatusCodeVal: StatusCodeNotFound,
			TraceVal:      errors.New(message),
		},
	}
}

// NewNotFoundErrorf はフォーマット指定付きで新しい NotFoundError を作成
func NewNotFoundErrorf(format string, args ...any) *NotFoundError {
	message := fmt.Sprintf(format, args...)
	return NewNotFoundError(message)
}

// WrapNotFoundError は既存のエラーをラップした NotFoundError を作成
func WrapNotFoundError(err error, message string) *NotFoundError {
	baseError := NewBaseError(message, StatusCodeNotFound, nil)
	wrappedError := baseError.WrapWithLocation(err, message)
	return &NotFoundError{
		BaseErr: wrappedError,
	}
}

// WrapNotFoundErrorf はフォーマット指定付きで既存のエラーをラップした NotFoundError を作成
func WrapNotFoundErrorf(err error, format string, args ...any) *NotFoundError {
	message := fmt.Sprintf(format, args...)
	return WrapNotFoundError(err, message)
}

// WithContext はエラーにコンテキスト情報を追加
func (nfe *NotFoundError) WithContext(ctx Context) *NotFoundError {
	return &NotFoundError{
		BaseErr: nfe.BaseErr.WithContext(ctx),
	}
}
