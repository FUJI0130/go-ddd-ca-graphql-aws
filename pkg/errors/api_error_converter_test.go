package errors

import (
	"errors"
	"net/http"
	"testing"

	"github.com/FUJI0130/go-ddd-ca/internal/domain/entity"
)

func TestConvertToAPIError(t *testing.T) {
	// テストケース
	testCases := []struct {
		name           string
		err            error
		resourceID     string
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "Nilエラー",
			err:            nil,
			expectedStatus: 0,
			expectedCode:   "",
		},
		{
			name:           "通常のエラー",
			err:            errors.New("一般的なエラー"),
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   "INTERNAL_SERVER_ERROR",
		},
		{
			name:           "エンティティ定義のNotFoundエラー",
			err:            entity.ErrTestSuiteNotFound,
			resourceID:     "TS001",
			expectedStatus: http.StatusNotFound,
			expectedCode:   "NOT_FOUND",
		},
		{
			name:           "エンティティ定義のConcurrentModificationエラー",
			err:            entity.ErrConcurrentModification,
			expectedStatus: http.StatusConflict,
			expectedCode:   "CONFLICT",
		},
		{
			name:           "エンティティ定義のInvalidStatusTransitionエラー",
			err:            entity.ErrInvalidStatusTransition,
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "BAD_REQUEST",
		},
		{
			name: "NotFoundエラー",
			err: &NotFoundError{
				BaseDomainError: BaseDomainError{
					Code:    "NOT_FOUND",
					Message: "リソースが見つかりません",
					DetailInfo: map[string]interface{}{
						"resourceType": "テストスイート",
						"resourceID":   "123",
					},
				},
			},
			expectedStatus: http.StatusNotFound,
			expectedCode:   "NOT_FOUND",
		},
		{
			name: "ValidationErrorエラー",
			err: &ValidationError{
				BaseDomainError: BaseDomainError{
					Code:    "VALIDATION_ERROR",
					Message: "入力値が不正です",
					DetailInfo: map[string]interface{}{
						"field1": "必須項目です",
					},
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "VALIDATION_ERROR",
		},
		{
			name: "ConflictErrorエラー",
			err: &ConflictError{
				BaseDomainError: BaseDomainError{
					Code:    "CONFLICT",
					Message: "リソースが競合しています",
				},
			},
			expectedStatus: http.StatusConflict,
			expectedCode:   "CONFLICT",
		},
		{
			name: "PermissionErrorエラー",
			err: &PermissionError{
				BaseDomainError: BaseDomainError{
					Code:    "PERMISSION_DENIED",
					Message: "権限がありません",
				},
			},
			expectedStatus: http.StatusForbidden,
			expectedCode:   "FORBIDDEN",
		},
		{
			name: "SystemErrorエラー",
			err: &SystemError{
				BaseDomainError: BaseDomainError{
					Code:       "SYSTEM_ERROR",
					Message:    "システムエラーが発生しました",
					DevMessage: "テスト用のシステムエラー",
				},
			},
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   "INTERNAL_SERVER_ERROR",
		},
		{
			name:           "既存APIエラー",
			err:            NewBadRequestError("不正なリクエストです"),
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "BAD_REQUEST",
		},
	}

	// テスト実行
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var result *APIError
			if tc.resourceID != "" {
				result = ConvertToAPIError(tc.err, tc.resourceID)
			} else {
				result = ConvertToAPIError(tc.err)
			}

			// nilエラーの場合
			if tc.err == nil {
				if result != nil {
					t.Errorf("期待: nil, 実際: %v", result)
				}
				return
			}

			// ステータスコードの検証
			if result.Status != tc.expectedStatus {
				t.Errorf("ステータスコードが一致しません。期待: %d, 実際: %d", tc.expectedStatus, result.Status)
			}

			// エラーコードの検証
			if result.Code != tc.expectedCode {
				t.Errorf("エラーコードが一致しません。期待: %s, 実際: %s", tc.expectedCode, result.Code)
			}
		})
	}
}
