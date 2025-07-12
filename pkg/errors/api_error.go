package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// APIError はAPI全体で使用するエラー構造体です
type APIError struct {
	Status     int               `json:"-"`                 // HTTPステータスコード
	Code       string            `json:"code"`              // エラーコード
	Message    string            `json:"message"`           // エラーメッセージ
	Details    map[string]string `json:"details,omitempty"` // 詳細情報
	DevMessage string            `json:"-"`                 // 開発者向けメッセージ
}

// Error はerrorインターフェースを満たすためのメソッドです
func (e *APIError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// WriteToResponse はエラーレスポンスをHTTPレスポンスとして書き出します
func (e *APIError) WriteToResponse(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.Status)
	json.NewEncoder(w).Encode(e)
}

var (
	// ErrInvalidDateRange は開始日が終了日より後の場合のエラー
	ErrInvalidDateRange = &APIError{
		Code:    "INVALID_DATE_RANGE",
		Message: "end date must be after start date",
		Status:  http.StatusBadRequest,
	}
)

// 一般的なエラーを生成する関数群
func NewValidationError(message string, details map[string]string) *APIError {
	return &APIError{
		Status:  http.StatusBadRequest,
		Code:    "VALIDATION_ERROR",
		Message: message,
		Details: details,
	}
}

func NewNotFoundError(resource string, id string) *APIError {
	return &APIError{
		Status:  http.StatusNotFound,
		Code:    "NOT_FOUND",
		Message: fmt.Sprintf("%s (ID: %s) が見つかりません", resource, id),
	}
}

func NewInternalServerError(err error) *APIError {
	return &APIError{
		Status:     http.StatusInternalServerError,
		Code:       "INTERNAL_SERVER_ERROR",
		Message:    "内部サーバーエラーが発生しました",
		DevMessage: err.Error(),
	}
}

func NewConflictError(message string) *APIError {
	return &APIError{
		Status:  http.StatusConflict,
		Code:    "CONFLICT",
		Message: message,
	}
}

func NewBadRequestError(message string) *APIError {
	return &APIError{
		Status:  http.StatusBadRequest,
		Code:    "BAD_REQUEST",
		Message: message,
	}
}

func NewUnauthorizedError() *APIError {
	return &APIError{
		Status:  http.StatusUnauthorized,
		Code:    "UNAUTHORIZED",
		Message: "認証が必要です",
	}
}

func NewForbiddenError() *APIError {
	return &APIError{
		Status:  http.StatusForbidden,
		Code:    "FORBIDDEN",
		Message: "この操作を実行する権限がありません",
	}
}

// エラーハンドリング用のミドルウェア
func ErrorHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				apiErr := NewInternalServerError(fmt.Errorf("%v", err))
				apiErr.WriteToResponse(w)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
