package errors

import (
	"fmt"
)

// NotFoundError は基本的な未検出エラー
type NotFoundError struct {
	BaseDomainError
	Resource string
	ID       string
}

// NewDomainNotFoundError は新しいNotFoundErrorを生成
func NewDomainNotFoundError(resource, id string) *NotFoundError {
	return &NotFoundError{
		BaseDomainError: BaseDomainError{
			Code:    "NOT_FOUND",
			Message: fmt.Sprintf("%s (ID: %s) が見つかりません", resource, id),
		},
		Resource: resource,
		ID:       id,
	}
}

// Details はリソース情報を含む詳細情報を返す
func (e *NotFoundError) Details() map[string]interface{} {
	details := e.BaseDomainError.Details()
	details["resource"] = e.Resource
	details["id"] = e.ID
	return details
}

// EntityNotFoundError はエンティティ未検出エラー
type EntityNotFoundError struct {
	NotFoundError
	EntityType string
}

// NewEntityNotFoundError は新しいEntityNotFoundErrorを生成
func NewEntityNotFoundError(entityType, id string) *EntityNotFoundError {
	return &EntityNotFoundError{
		NotFoundError: *NewDomainNotFoundError(entityType, id),
		EntityType:    entityType,
	}
}

// TestSuiteNotFoundError はテストスイート未検出エラー
type TestSuiteNotFoundError struct {
	BaseDomainError
	ID string
}

// NewTestSuiteNotFoundError は新しいTestSuiteNotFoundErrorを生成
func NewTestSuiteNotFoundError(id string) *TestSuiteNotFoundError {
	return &TestSuiteNotFoundError{
		BaseDomainError: BaseDomainError{
			Code:    "TEST_SUITE_NOT_FOUND",
			Message: fmt.Sprintf("ID %s のテストスイートが見つかりません", id),
			DetailInfo: map[string]interface{}{
				"id": id,
			},
		},
		ID: id,
	}
}

// TestGroupNotFoundError はテストグループ未検出エラー
type TestGroupNotFoundError struct {
	EntityNotFoundError
}

// NewTestGroupNotFoundError は新しいTestGroupNotFoundErrorを生成
func NewTestGroupNotFoundError(id string) *TestGroupNotFoundError {
	return &TestGroupNotFoundError{
		EntityNotFoundError: *NewEntityNotFoundError("TestGroup", id),
	}
}

// TestCaseNotFoundError はテストケース未検出エラー
type TestCaseNotFoundError struct {
	EntityNotFoundError
}

// NewTestCaseNotFoundError は新しいTestCaseNotFoundErrorを生成
func NewTestCaseNotFoundError(id string) *TestCaseNotFoundError {
	return &TestCaseNotFoundError{
		EntityNotFoundError: *NewEntityNotFoundError("TestCase", id),
	}
}
