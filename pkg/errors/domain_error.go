package errors

import (
	"fmt"
)

// DomainError はドメイン層のエラーインターフェース
type DomainError interface {
	error
	ErrorCode() string               // エラーコード
	ErrorMessage() string            // ユーザー向けメッセージ
	Details() map[string]interface{} // 詳細情報
	DeveloperMessage() string        // 開発者向け詳細メッセージ
}

// BaseDomainError はDomainErrorを実装する基本構造体
type BaseDomainError struct {
	Code       string                 // エラーコード
	Message    string                 // ユーザー向けメッセージ
	DetailInfo map[string]interface{} // 詳細情報
	DevMessage string                 // 開発者向け詳細メッセージ
}

// Error はerrorインターフェースを満たすためのメソッド
func (e *BaseDomainError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// ErrorCode はエラーコードを返す
func (e *BaseDomainError) ErrorCode() string {
	return e.Code
}

// ErrorMessage はユーザー向けメッセージを返す
func (e *BaseDomainError) ErrorMessage() string {
	return e.Message
}

// Details は詳細情報を返す
func (e *BaseDomainError) Details() map[string]interface{} {
	if e.DetailInfo == nil {
		return map[string]interface{}{}
	}
	return e.DetailInfo
}

// DeveloperMessage は開発者向け詳細メッセージを返す
func (e *BaseDomainError) DeveloperMessage() string {
	return e.DevMessage
}

// WithDevMessage は開発者向けメッセージを追加したエラーを返す
func (e *BaseDomainError) WithDevMessage(message string) *BaseDomainError {
	e.DevMessage = message
	return e
}

// WithDetails は詳細情報を追加したエラーを返す
func (e *BaseDomainError) WithDetails(details map[string]interface{}) *BaseDomainError {
	if e.DetailInfo == nil {
		e.DetailInfo = make(map[string]interface{})
	}

	for k, v := range details {
		e.DetailInfo[k] = v
	}

	return e
}

// IsDomainError は指定されたエラーがDomainErrorインターフェースを実装しているかを確認
func IsDomainError(err error) bool {
	_, ok := err.(DomainError)
	return ok
}
