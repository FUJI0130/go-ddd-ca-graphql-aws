package errors

// PermissionError は権限関連エラー
type PermissionError struct {
	BaseDomainError
	UserID string
}

// NewDomainPermissionError は新しいPermissionErrorを生成
func NewDomainPermissionError(message string, userID string) *PermissionError {
	if message == "" {
		message = "操作を実行する権限がありません"
	}

	return &PermissionError{
		BaseDomainError: BaseDomainError{
			Code:    "PERMISSION_ERROR",
			Message: message,
		},
		UserID: userID,
	}
}

// Details はユーザーIDを含む詳細情報を返す
func (e *PermissionError) Details() map[string]interface{} {
	details := e.BaseDomainError.Details()
	if e.UserID != "" {
		details["userId"] = e.UserID
	}
	return details
}

// UnauthorizedError は認証エラー
type UnauthorizedError struct {
	BaseDomainError
}

// NewDomainUnauthorizedError は新しいUnauthorizedErrorを生成
func NewDomainUnauthorizedError() *UnauthorizedError {
	return &UnauthorizedError{
		BaseDomainError: BaseDomainError{
			Code:    "UNAUTHORIZED",
			Message: "認証が必要です",
		},
	}
}

// ForbiddenError は権限不足エラー
type ForbiddenError struct {
	PermissionError
	Resource string
	Action   string
}

// NewDomainForbiddenError は新しいForbiddenErrorを生成
func NewDomainForbiddenError(userID, resource, action string) *ForbiddenError {
	return &ForbiddenError{
		PermissionError: *NewDomainPermissionError(
			"この操作を実行する権限がありません",
			userID,
		),
		Resource: resource,
		Action:   action,
	}
}

// Details はアクセス情報を含む詳細情報を返す
func (e *ForbiddenError) Details() map[string]interface{} {
	details := e.PermissionError.Details()
	details["resource"] = e.Resource
	details["action"] = e.Action
	return details
}
