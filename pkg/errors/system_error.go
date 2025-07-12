package errors

import (
	"fmt"
)

// SystemError はシステム関連エラー
type SystemError struct {
	BaseDomainError
	OriginalError error
}

// NewSystemError は新しいSystemErrorを生成
func NewSystemError(message string, err error) *SystemError {
	if message == "" {
		message = "システム内部でエラーが発生しました"
	}

	sysErr := &SystemError{
		BaseDomainError: BaseDomainError{
			Code:    "SYSTEM_ERROR",
			Message: message,
		},
		OriginalError: err,
	}

	if err != nil {
		sysErr.DevMessage = err.Error()
	}

	return sysErr
}

// DatabaseError はデータベース関連エラー
type DatabaseError struct {
	SystemError
	Operation string
	Table     string
}

// NewDatabaseError は新しいDatabaseErrorを生成
func NewDatabaseError(operation, table string, err error) *DatabaseError {
	return &DatabaseError{
		SystemError: *NewSystemError(
			fmt.Sprintf("データベース操作中にエラーが発生しました"),
			err,
		),
		Operation: operation,
		Table:     table,
	}
}

// Details はデータベース操作情報を含む詳細情報を返す
func (e *DatabaseError) Details() map[string]interface{} {
	details := e.SystemError.Details()
	details["operation"] = e.Operation
	details["table"] = e.Table
	return details
}

// ExternalServiceError は外部サービス関連エラー
type ExternalServiceError struct {
	SystemError
	ServiceName string
	Endpoint    string
	StatusCode  int
}

// NewExternalServiceError は新しいExternalServiceErrorを生成
func NewExternalServiceError(serviceName, endpoint string, statusCode int, err error) *ExternalServiceError {
	return &ExternalServiceError{
		SystemError: *NewSystemError(
			fmt.Sprintf("外部サービス呼び出し中にエラーが発生しました: %s", serviceName),
			err,
		),
		ServiceName: serviceName,
		Endpoint:    endpoint,
		StatusCode:  statusCode,
	}
}

// Details は外部サービス情報を含む詳細情報を返す
func (e *ExternalServiceError) Details() map[string]interface{} {
	details := e.SystemError.Details()
	details["serviceName"] = e.ServiceName
	details["endpoint"] = e.Endpoint
	if e.StatusCode > 0 {
		details["statusCode"] = e.StatusCode
	}
	return details
}

// InternalServerError は内部サーバーエラー
type InternalServerError struct {
	SystemError
}

// NewInternalServerError は新しいInternalServerErrorを生成
func NewDomainInternalServerError(err error) *InternalServerError {
	return &InternalServerError{
		SystemError: *NewSystemError(
			"内部サーバーエラーが発生しました",
			err,
		),
	}
}
