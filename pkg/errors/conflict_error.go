package errors

import (
	"fmt"
)

// ConflictError は競合エラー
type ConflictError struct {
	BaseDomainError
	Resource string
	ID       string
}

// NewConflictError は新しいConflictErrorを生成
func NewDomainConflictError(resource, id string, message string) *ConflictError {
	if message == "" {
		message = fmt.Sprintf("%s (ID: %s) のデータに競合が発生しました", resource, id)
	}

	return &ConflictError{
		BaseDomainError: BaseDomainError{
			Code:    "CONFLICT",
			Message: message,
		},
		Resource: resource,
		ID:       id,
	}
}

// Details はリソース情報を含む詳細情報を返す
func (e *ConflictError) Details() map[string]interface{} {
	details := e.BaseDomainError.Details()
	details["resource"] = e.Resource
	details["id"] = e.ID
	return details
}

// ConcurrentModificationError は同時編集衝突エラー
type ConcurrentModificationError struct {
	ConflictError
	CurrentVersion   int64
	RequestedVersion int64
}

// NewConcurrentModificationError は新しいConcurrentModificationErrorを生成
func NewConcurrentModificationError(resource, id string, current, requested int64) *ConcurrentModificationError {
	return &ConcurrentModificationError{
		ConflictError: *NewDomainConflictError(
			resource,
			id,
			fmt.Sprintf("%s (ID: %s) は別のユーザーによって更新されています", resource, id),
		),
		CurrentVersion:   current,
		RequestedVersion: requested,
	}
}

// Details はバージョン情報を含む詳細情報を返す
func (e *ConcurrentModificationError) Details() map[string]interface{} {
	details := e.ConflictError.Details()
	details["currentVersion"] = e.CurrentVersion
	details["requestedVersion"] = e.RequestedVersion
	return details
}

// AlreadyExistsError は既存リソースエラー
type AlreadyExistsError struct {
	ConflictError
}

// NewAlreadyExistsError は新しいAlreadyExistsErrorを生成
func NewAlreadyExistsError(resource, id string) *AlreadyExistsError {
	return &AlreadyExistsError{
		ConflictError: *NewDomainConflictError(
			resource,
			id,
			fmt.Sprintf("%s (ID: %s) は既に存在しています", resource, id),
		),
	}
}

// StatusTransitionConflictError はステータス遷移の競合エラー
type StatusTransitionConflictError struct {
	ConflictError
	CurrentStatus string
	NewStatus     string
}

// NewStatusTransitionConflictError は新しいStatusTransitionConflictErrorを生成
func NewStatusTransitionConflictError(message string, currentStatus, newStatus string) *StatusTransitionConflictError {
	return &StatusTransitionConflictError{
		ConflictError: *NewDomainConflictError(
			"TestSuite",
			"",
			message,
		),
		CurrentStatus: currentStatus,
		NewStatus:     newStatus,
	}
}

// Details はステータス情報を含む詳細情報を返す
func (e *StatusTransitionConflictError) Details() map[string]interface{} {
	details := e.ConflictError.Details()
	details["current_status"] = e.CurrentStatus
	details["new_status"] = e.NewStatus
	return details
}
