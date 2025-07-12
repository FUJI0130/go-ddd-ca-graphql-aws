package customerrors

import (
	"fmt"

	"github.com/cockroachdb/errors"
)

// ConflictError はリソースの競合エラー
type ConflictError struct {
	*BaseErr
}

// NewConflictError は新しい ConflictError を作成
func NewConflictError(message string) *ConflictError {
	return &ConflictError{
		&BaseErr{
			MessageVal:    message,
			StatusCodeVal: StatusCodeConflict,
			TraceVal:      errors.New(message),
			ContextVal:    nil,
		},
	}
}

// WithContext はエラーにコンテキスト情報を追加
func (ce *ConflictError) WithContext(ctx Context) *ConflictError {
	return &ConflictError{
		BaseErr: ce.BaseErr.WithContext(ctx),
	}
}

// NewConflictErrorf はフォーマット指定付きで新しい ConflictError を作成
func NewConflictErrorf(format string, args ...any) *ConflictError {
	message := fmt.Sprintf(format, args...)
	return NewConflictError(message)
}

// WrapConflictError は既存のエラーをラップした ConflictError を作成
func WrapConflictError(err error, message string) *ConflictError {
	baseError := NewBaseError(message, StatusCodeConflict, nil)
	wrappedError := baseError.WrapWithLocation(err, message)
	return &ConflictError{
		BaseErr: wrappedError,
	}
}
