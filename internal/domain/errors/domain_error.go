// internal/domain/errors/domain_error.go
package errors

// DomainError はドメインロジックから発生するエラーのインターフェース
type DomainError interface {
	error
	Code() string
	Message() string
	Error() string
	IsDomainError() bool
}

// ErrorType はドメインエラーの種類を表す
type ErrorType string

const (
	ValidationErrorType ErrorType = "VALIDATION_ERROR"
	NotFoundErrorType   ErrorType = "NOT_FOUND_ERROR"
	ConflictErrorType   ErrorType = "CONFLICT_ERROR"
	PermissionErrorType ErrorType = "PERMISSION_ERROR"
	SystemErrorType     ErrorType = "SYSTEM_ERROR"
)
