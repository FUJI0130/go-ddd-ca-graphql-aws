// support/customerrors/domain_bridge.go の修正案
package customerrors

import (
	domainerrors "github.com/FUJI0130/go-ddd-ca/internal/domain/errors"
	"github.com/cockroachdb/errors"
)

// ToDomainError は BaseError を domainerrors.DomainError に変換する
func ToDomainError(err error) domainerrors.DomainError {
	if err == nil {
		return nil
	}

	// すでに DomainError ならそのまま返す
	if de, ok := err.(domainerrors.DomainError); ok {
		return de
	}

	// BaseError かチェック
	if baseErr, ok := AsBaseError(err); ok {
		// メッセージとステータスコードから適切なドメインエラーを生成
		message := baseErr.Message()
		statusCode := baseErr.StatusCode()

		// ここでは直接domainerrors.DomainErrorを実装する構造体を返す
		switch statusCode {
		case StatusCodeNotFound:
			return newDomainNotFoundError(message)
		case StatusCodeUnprocessableEntity:
			return newDomainValidationError(message)
		case StatusCodeConflict:
			return newDomainConflictError(message)
		case StatusCodeForbidden:
			return newDomainPermissionError(message)
		default:
			return newDomainSystemError(message)
		}
	}

	// その他のエラーはSystemErrorに変換
	return newDomainSystemError(err.Error())
}

// ドメインエラーの実装 (domainerrors.DomainError インターフェースを満たす)
type domainErrorImpl struct {
	code    string
	message string
}

func (e *domainErrorImpl) Error() string {
	return e.message
}

func (e *domainErrorImpl) Code() string {
	return e.code
}

func (e *domainErrorImpl) Message() string {
	return e.message
}

func (e *domainErrorImpl) IsDomainError() bool {
	return true
}

// 各エラータイプのファクトリメソッド
func newDomainValidationError(message string) domainerrors.DomainError {
	return &domainErrorImpl{
		code:    string(domainerrors.ValidationErrorType),
		message: message,
	}
}

func newDomainNotFoundError(message string) domainerrors.DomainError {
	return &domainErrorImpl{
		code:    string(domainerrors.NotFoundErrorType),
		message: message,
	}
}

func newDomainConflictError(message string) domainerrors.DomainError {
	return &domainErrorImpl{
		code:    string(domainerrors.ConflictErrorType),
		message: message,
	}
}

func newDomainPermissionError(message string) domainerrors.DomainError {
	return &domainErrorImpl{
		code:    string(domainerrors.PermissionErrorType),
		message: message,
	}
}

func newDomainSystemError(message string) domainerrors.DomainError {
	return &domainErrorImpl{
		code:    string(domainerrors.SystemErrorType),
		message: message,
	}
}

// FromDomainError は domainerrors.DomainError から適切な BaseError を生成する
func FromDomainError(err domainerrors.DomainError) error {
	if err == nil {
		return nil
	}

	message := err.Message()
	code := err.Code()

	// エラーの種類に基づいて適切な BaseError を返す
	switch code {
	case string(domainerrors.ValidationErrorType):
		return NewValidationError(message, nil)
	case string(domainerrors.NotFoundErrorType):
		return NewNotFoundError(message)
	case string(domainerrors.ConflictErrorType):
		return NewConflictError(message)
	case string(domainerrors.PermissionErrorType):
		return NewForbiddenError(message)
	default:
		return NewInternalServerError(message)
	}
}

// WrapWithDomainError は既存のエラーをドメインエラーでラップする
func WrapWithDomainError(err error, message string) error {
	if err == nil {
		return nil
	}

	// すでに BaseError の場合
	if baseErr, ok := AsBaseError(err); ok {
		statusCode := baseErr.StatusCode()

		switch statusCode {
		case StatusCodeNotFound:
			return WrapNotFoundError(err, message)
		case StatusCodeUnprocessableEntity:
			return WrapValidationError(err, message, nil)
		case StatusCodeConflict:
			return WrapConflictError(err, message)
		case StatusCodeForbidden:
			return WrapForbiddenError(err, message)
		default:
			return WrapInternalServerError(err, message)
		}
	}

	// すでに DomainError の場合、適切な BaseError に変換してからラップ
	if de, ok := err.(domainerrors.DomainError); ok {
		baseErr := FromDomainError(de)
		return errors.Wrap(baseErr, message)
	}

	// その他のエラーは InternalServerError でラップ
	return WrapInternalServerError(err, message)
}
