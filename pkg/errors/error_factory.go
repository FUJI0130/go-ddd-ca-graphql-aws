// pkg/errors/error_factory.go - 修正版
package errors

import (
	domainerrors "github.com/FUJI0130/go-ddd-ca/internal/domain/errors"
)

// DomainErrorAdapterは内部ドメインエラーインターフェースを実装する構造体
type DomainErrorAdapter struct {
	code    string
	message string
}

// Code はドメインエラーコードを返す
func (e *DomainErrorAdapter) Code() string {
	return e.code
}

// Message はエラーメッセージを返す
func (e *DomainErrorAdapter) Message() string {
	return e.message
}

// Error はエラー文字列を返す
func (e *DomainErrorAdapter) Error() string {
	return e.message
}

// IsDomainError はドメインエラーであることを示す
func (e *DomainErrorAdapter) IsDomainError() bool {
	return true
}

// simpleValidationError は簡易的なバリデーションエラーを返す
// 他のエラータイプは既存の関数を使用するため、この関数はprivateにする
func simpleValidationError(message string) domainerrors.DomainError {
	return &DomainErrorAdapter{
		code:    string(domainerrors.ValidationErrorType),
		message: message,
	}
}

// ErrorFactory は汎用的なエラーファクトリ
type ErrorFactory struct{}

func NewErrorFactory() *ErrorFactory {
	return &ErrorFactory{}
}

// UserEntityErrorFactory の実装
type UserEntityErrorFactory struct{}

func NewUserEntityErrorFactory() *UserEntityErrorFactory {
	return &UserEntityErrorFactory{}
}

// UserEntityErrorFactoryメソッドの実装
func (f *UserEntityErrorFactory) EmptyUserID() domainerrors.DomainError {
	return simpleValidationError("ユーザーIDは必須です")
}

func (f *UserEntityErrorFactory) EmptyUsername() domainerrors.DomainError {
	return simpleValidationError("ユーザー名は必須です")
}

func (f *UserEntityErrorFactory) EmptyPasswordHash() domainerrors.DomainError {
	return simpleValidationError("パスワードハッシュは必須です")
}

func (f *UserEntityErrorFactory) InvalidUserRole() domainerrors.DomainError {
	return simpleValidationError("無効なユーザーロールです")
}
