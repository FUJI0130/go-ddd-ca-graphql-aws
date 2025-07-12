package customerrors

import (
	"fmt"

	"github.com/cockroachdb/errors"
)

// ForbiddenError は認可エラー
type ForbiddenError struct {
	*BaseErr
}

// NewForbiddenError は新しい ForbiddenError を作成
func NewForbiddenError(message string) *ForbiddenError {
	return &ForbiddenError{
		&BaseErr{
			MessageVal:    message,
			StatusCodeVal: StatusCodeForbidden,
			TraceVal:      errors.New(message),
		},
	}
}

// NewForbiddenErrorf はフォーマット指定付きで新しい ForbiddenError を作成
func NewForbiddenErrorf(format string, args ...any) *ForbiddenError {
	message := fmt.Sprintf(format, args...)
	return NewForbiddenError(message)
}

// WrapForbiddenError は既存のエラーをラップした ForbiddenError を作成
func WrapForbiddenError(err error, message string) *ForbiddenError {
	baseError := NewBaseError(message, StatusCodeForbidden, nil)
	wrappedError := baseError.WrapWithLocation(err, message)
	return &ForbiddenError{
		BaseErr: wrappedError,
	}
}

// WithContext はエラーにコンテキスト情報を追加
func (fe *ForbiddenError) WithContext(ctx Context) *ForbiddenError {
	return &ForbiddenError{
		BaseErr: fe.BaseErr.WithContext(ctx),
	}
}
