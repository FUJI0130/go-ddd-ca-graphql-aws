package customerrors

import (
	"fmt"

	"github.com/cockroachdb/errors"
)

// IsBaseError はエラーが BaseError インターフェースを実装しているかチェック
func IsBaseError(err error) bool {
	var baseErr BaseError
	return errors.As(err, &baseErr)
}

// AsBaseError はエラーを BaseError にキャスト
func AsBaseError(err error) (BaseError, bool) {
	var baseErr BaseError
	if errors.As(err, &baseErr) {
		return baseErr, true
	}
	return nil, false
}

// IsNotFoundError はエラーが NotFoundError かチェック
func IsNotFoundError(err error) bool {
	var notFoundErr *NotFoundError
	return errors.As(err, &notFoundErr)
}

// IsValidationError はエラーが ValidationError かチェック
func IsValidationError(err error) bool {
	var validationErr *ValidationError
	return errors.As(err, &validationErr)
}

// IsConflictError はエラーが ConflictError かチェック
func IsConflictError(err error) bool {
	var conflictErr *ConflictError
	return errors.As(err, &conflictErr)
}

// IsInternalServerError はエラーが InternalServerError かチェック
func IsInternalServerError(err error) bool {
	var internalErr *InternalServerError
	return errors.As(err, &internalErr)
}

// IsUnauthorizedError はエラーが UnauthorizedError かチェック
func IsUnauthorizedError(err error) bool {
	var unauthorizedErr *UnauthorizedError
	return errors.As(err, &unauthorizedErr)
}

// IsForbiddenError はエラーが ForbiddenError かチェック
func IsForbiddenError(err error) bool {
	var forbiddenErr *ForbiddenError
	return errors.As(err, &forbiddenErr)
}

// AsNotFoundError はエラーを NotFoundError にキャスト
func AsNotFoundError(err error) (*NotFoundError, bool) {
	var notFoundErr *NotFoundError
	if errors.As(err, &notFoundErr) {
		return notFoundErr, true
	}
	return nil, false
}

// AsValidationError はエラーを ValidationError にキャスト
func AsValidationError(err error) (*ValidationError, bool) {
	var validationErr *ValidationError
	if errors.As(err, &validationErr) {
		return validationErr, true
	}
	return nil, false
}

// AsConflictError はエラーを ConflictError にキャスト
func AsConflictError(err error) (*ConflictError, bool) {
	var conflictErr *ConflictError
	if errors.As(err, &conflictErr) {
		return conflictErr, true
	}
	return nil, false
}

// ConvertToErrorWithStatus は標準エラーをステータスコード付きのカスタムエラーに変換
func ConvertToErrorWithStatus(err error, defaultMessage string, statusCode int) error {
	if err == nil {
		return nil
	}

	// すでにBaseErrorの場合はそのまま返す
	if IsBaseError(err) {
		return err
	}

	// エラーメッセージの決定
	message := err.Error()
	if message == "" {
		message = defaultMessage
	}

	// ステータスコードに応じたエラー型の生成
	switch statusCode {
	case StatusCodeNotFound:
		return NewNotFoundError(message)
	case StatusCodeUnprocessableEntity:
		return NewValidationError(message, nil)
	case StatusCodeConflict:
		return NewConflictError(message)
	case StatusCodeUnauthorized:
		return NewUnauthorizedError(message)
	case StatusCodeForbidden:
		return NewForbiddenError(message)
	default:
		return NewInternalServerError(message)
	}
}

// WrapError は既存のエラーを適切なカスタムエラーでラップ
func WrapError(err error, message string) error {
	if err == nil {
		return nil
	}

	// すでにBaseErrorの場合は、そのタイプに応じてラップ
	if baseErr, ok := AsBaseError(err); ok {
		statusCode := baseErr.StatusCode()
		switch statusCode {
		case StatusCodeNotFound:
			return WrapNotFoundError(err, message)
		case StatusCodeUnprocessableEntity:
			return WrapValidationError(err, message, nil)
		case StatusCodeConflict:
			return WrapConflictError(err, message)
		case StatusCodeUnauthorized:
			return WrapUnauthorizedError(err, message)
		case StatusCodeForbidden:
			return WrapForbiddenError(err, message)
		default:
			return WrapInternalServerError(err, message)
		}
	}

	// デフォルトでは内部サーバーエラーとしてラップ
	return WrapInternalServerError(err, message)
}

// WrapErrorf はフォーマット指定付きで既存のエラーを適切なカスタムエラーでラップ
func WrapErrorf(err error, format string, args ...any) error {
	message := fmt.Sprintf(format, args...)
	return WrapError(err, message)
}
