package customerrors

import (
	"fmt"

	"github.com/cockroachdb/errors"
)

// UnauthorizedError は認証エラー
type UnauthorizedError struct {
	*BaseErr
}

// NewUnauthorizedError は新しい UnauthorizedError を作成
func NewUnauthorizedError(message string) *UnauthorizedError {
	return &UnauthorizedError{
		&BaseErr{
			MessageVal:    message,
			StatusCodeVal: StatusCodeUnauthorized,
			TraceVal:      errors.New(message),
		},
	}
}

// NewUnauthorizedErrorf はフォーマット指定付きで新しい UnauthorizedError を作成
func NewUnauthorizedErrorf(format string, args ...any) *UnauthorizedError {
	message := fmt.Sprintf(format, args...)
	return NewUnauthorizedError(message)
}

// WrapUnauthorizedError は既存のエラーをラップした UnauthorizedError を作成
func WrapUnauthorizedError(err error, message string) *UnauthorizedError {
	baseError := NewBaseError(message, StatusCodeUnauthorized, nil)
	wrappedError := baseError.WrapWithLocation(err, message)
	return &UnauthorizedError{
		BaseErr: wrappedError,
	}
}

// WithContext はエラーにコンテキスト情報を追加
func (ue *UnauthorizedError) WithContext(ctx Context) *UnauthorizedError {
	return &UnauthorizedError{
		BaseErr: ue.BaseErr.WithContext(ctx),
	}
}
