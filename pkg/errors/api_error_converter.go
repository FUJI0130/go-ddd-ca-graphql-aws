// pkg/errors/api_error_converter.go
package errors

import (
	"net/http"
	"strings"
)

// ConvertToAPIError はドメインエラーをAPIエラーに変換します
// エラーのタイプや内容に基づいて適切なAPIエラーを返します
func ConvertToAPIError(err error, resourceID ...string) *APIError {
	// エラーがnilの場合はnilを返す
	if err == nil {
		return nil
	}

	// すでにAPIErrorの場合はそのまま返す
	if apiErr, ok := err.(*APIError); ok {
		return apiErr
	}

	// エラーメッセージに基づく特定のエラーの識別
	// これはエンティティへの直接参照を避けるための方法
	id := ""
	if len(resourceID) > 0 {
		id = resourceID[0]
	}

	errMsg := err.Error()

	// エラーメッセージに基づいた変換
	if strings.Contains(errMsg, "テストスイートが見つかりません") {
		return NewNotFoundError("テストスイート", id)
	}
	if strings.Contains(errMsg, "既に更新されています") ||
		strings.Contains(errMsg, "同時更新が検出されました") {
		return NewConflictError("テストスイートは既に更新されています")
	}
	if strings.Contains(errMsg, "無効なステータス遷移") {
		return NewBadRequestError("無効なステータス遷移です")
	}

	// DomainErrorインターフェースを実装している場合
	if domainErr, ok := err.(DomainError); ok {
		switch domainErr.ErrorCode() {
		// NotFoundError
		case "NOT_FOUND":
			details := domainErr.Details()
			resourceType, _ := details["resourceType"].(string)
			resourceID, _ := details["resourceID"].(string)
			return NewNotFoundError(resourceType, resourceID)

		// ValidationError
		case "VALIDATION_ERROR":
			// 詳細情報をstring型に変換
			strDetails := make(map[string]string)
			for k, v := range domainErr.Details() {
				if str, ok := v.(string); ok {
					strDetails[k] = str
				} else {
					// 文字列に変換できない場合は一般的な文字列に
					strDetails[k] = "無効な値"
				}
			}
			return NewValidationError(domainErr.ErrorMessage(), strDetails)

		// ConflictError
		case "CONFLICT":
			return NewConflictError(domainErr.ErrorMessage())

		// PermissionError
		case "PERMISSION_DENIED":
			return NewForbiddenError()

		// SystemError
		case "SYSTEM_ERROR":
			return &APIError{
				Status:     http.StatusInternalServerError,
				Code:       "INTERNAL_SERVER_ERROR",
				Message:    "内部サーバーエラーが発生しました",
				DevMessage: domainErr.DeveloperMessage(),
			}
		}
	}

	// それ以外のエラーは内部サーバーエラーとして扱う
	return NewInternalServerError(err)
}
