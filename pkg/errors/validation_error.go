package errors

import (
	"fmt"
)

// ValidationError は入力検証エラー
type ValidationError struct {
	BaseDomainError
	FieldErrors map[string]string // フィールド名 -> エラーメッセージ
}

// NewDomainValidationError は新しいValidationErrorを生成
func NewDomainValidationError(message string, fieldErrors map[string]string) *ValidationError {
	return &ValidationError{
		BaseDomainError: BaseDomainError{
			Code:       "VALIDATION_ERROR",
			Message:    message,
			DetailInfo: make(map[string]interface{}),
		},
		FieldErrors: fieldErrors,
	}
}

// Details はフィールドエラーを含む詳細情報を返す
func (e *ValidationError) Details() map[string]interface{} {
	details := e.BaseDomainError.Details()
	details["fieldErrors"] = e.FieldErrors
	return details
}

// InvalidDateRangeError は日付範囲不正エラー
type InvalidDateRangeError struct {
	BaseDomainError
	StartDate string
	EndDate   string
}

// NewInvalidDateRangeError は新しいInvalidDateRangeErrorを生成
func NewInvalidDateRangeError(startDate, endDate string) *InvalidDateRangeError {
	return &InvalidDateRangeError{
		BaseDomainError: BaseDomainError{
			Code:    "INVALID_DATE_RANGE",
			Message: "終了日は開始日より後である必要があります",
		},
		StartDate: startDate,
		EndDate:   endDate,
	}
}

// Details は日付情報を含む詳細情報を返す
func (e *InvalidDateRangeError) Details() map[string]interface{} {
	details := e.BaseDomainError.Details()
	details["startDate"] = e.StartDate
	details["endDate"] = e.EndDate
	return details
}

// InvalidInputError は入力値不正エラー
type InvalidInputError struct {
	BaseDomainError
	Field      string
	Value      interface{}
	Constraint string
}

// NewInvalidInputError は新しいInvalidInputErrorを生成
func NewInvalidInputError(field string, value interface{}, constraint string) *InvalidInputError {
	return &InvalidInputError{
		BaseDomainError: BaseDomainError{
			Code:    "INVALID_INPUT",
			Message: fmt.Sprintf("%s フィールドの値が不正です", field),
		},
		Field:      field,
		Value:      value,
		Constraint: constraint,
	}
}

// Details はフィールド情報を含む詳細情報を返す
func (e *InvalidInputError) Details() map[string]interface{} {
	details := e.BaseDomainError.Details()
	details["field"] = e.Field
	details["value"] = fmt.Sprintf("%v", e.Value)
	details["constraint"] = e.Constraint
	return details
}
