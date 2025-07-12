package customerrors

import (
	"fmt"
	"strings"

	"github.com/cockroachdb/errors"
)

// Context はエラーのコンテキスト情報を保持する型
type Context map[string]interface{}

// BaseError はすべてのカスタムエラーが実装すべきインターフェース
type BaseError interface {
	Error() string
	StatusCode() int
	Trace() error
	Message() string
	WithContext(ctx Context) BaseError
}

// BaseErr は BaseError インターフェースの基本実装
type BaseErr struct {
	MessageVal    string
	StatusCodeVal int
	TraceVal      error
	ContextVal    Context
}

// NewBaseError は BaseErr の新しいインスタンスを作成
func NewBaseError(message string, statusCode int, trace error) *BaseErr {
	return &BaseErr{
		MessageVal:    message,
		StatusCodeVal: statusCode,
		TraceVal:      trace,
		ContextVal:    nil,
	}
}

// Error は error インターフェースの実装
func (be *BaseErr) Error() string {
	if be.TraceVal != nil {
		traceString := fmt.Sprintf("%+v", be.TraceVal)
		return fmt.Sprintf("%s ### \n%s", be.MessageVal, traceString)
	}
	return be.MessageVal
}

// Message はエラーメッセージのみを返す
func (be *BaseErr) Message() string {
	return be.MessageVal
}

// StatusCode はエラーのステータスコードを返す
func (be *BaseErr) StatusCode() int {
	return be.StatusCodeVal
}

// Trace はエラーのスタックトレースを返す
func (be *BaseErr) Trace() error {
	return be.TraceVal
}

// GetContext はエラーのコンテキスト情報を返す
func (be *BaseErr) GetContext() Context {
	return be.ContextVal
}

// WithContext はエラーにコンテキスト情報を追加
func (be *BaseErr) WithContext(ctx Context) *BaseErr {
	newBe := &BaseErr{
		MessageVal:    be.MessageVal,
		StatusCodeVal: be.StatusCodeVal,
		TraceVal:      be.TraceVal,
		ContextVal:    ctx,
	}

	// コンテキスト情報をTraceValに追加（オプション）
	if len(ctx) > 0 && be.TraceVal != nil {
		contextStr := fmt.Sprintf("Context: %v", ctx)
		newBe.TraceVal = errors.WithHint(be.TraceVal, contextStr)
	}

	return newBe
}

// WrapWithLocation は既存のエラーをラップし、新しいメッセージとスタックトレースを追加
func (be *BaseErr) WrapWithLocation(err error, message string) *BaseErr {
	wrappedError := &BaseErr{
		MessageVal:    message,
		StatusCodeVal: be.StatusCodeVal,
		TraceVal:      errors.Wrap(err, message),
		ContextVal:    be.ContextVal,
	}
	return wrappedError
}

// SplitMessageAndTrace はエラー文字列からメッセージとスタックトレースを分離
func SplitMessageAndTrace(errStr string) (string, string) {
	parts := strings.SplitN(errStr, " ### ", 2)
	if len(parts) < 2 {
		return errStr, ""
	}
	message := parts[0]
	stackTrace := parts[1]
	return message, stackTrace
}
