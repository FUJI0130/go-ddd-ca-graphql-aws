package errors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBaseDomainError(t *testing.T) {
	// 基本エラーの作成
	err := &BaseDomainError{
		Code:    "TEST_ERROR",
		Message: "テストエラーメッセージ",
	}

	// メソッドのテスト
	assert.Equal(t, "TEST_ERROR", err.ErrorCode())
	assert.Equal(t, "テストエラーメッセージ", err.ErrorMessage())
	assert.Equal(t, "TEST_ERROR: テストエラーメッセージ", err.Error())
	assert.Empty(t, err.DeveloperMessage())
	assert.Empty(t, err.Details())

	// 詳細情報の追加
	details := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
	}
	err.WithDetails(details)

	// 開発者メッセージの追加
	err.WithDevMessage("開発者向け詳細情報")

	// 追加後の確認
	assert.Equal(t, "開発者向け詳細情報", err.DeveloperMessage())
	assert.Equal(t, "value1", err.Details()["key1"])
	assert.Equal(t, 123, err.Details()["key2"])
}

func TestValidationError(t *testing.T) {
	// バリデーションエラーの作成
	fieldErrors := map[string]string{
		"name":      "名前は必須です",
		"startDate": "開始日は必須です",
	}

	err := NewDomainValidationError("入力内容に誤りがあります", fieldErrors)

	// メソッドのテスト
	assert.Equal(t, "VALIDATION_ERROR", err.ErrorCode())
	assert.Equal(t, "入力内容に誤りがあります", err.ErrorMessage())
	assert.Equal(t, fieldErrors, err.Details()["fieldErrors"])

	// IsDomainErrorのテスト
	assert.True(t, IsDomainError(err))
}

func TestEntityNotFoundError(t *testing.T) {
	// EntityNotFoundErrorの作成
	err := NewTestSuiteNotFoundError("TS001-202501")

	// メソッドのテスト
	assert.Equal(t, "TEST_SUITE_NOT_FOUND", err.ErrorCode())
	assert.Contains(t, err.ErrorMessage(), "TS001-202501")

	details := err.Details()
	// 実際の実装ではid キーに値が格納されていることを確認
	assert.Equal(t, "TS001-202501", details["id"])

	// IsDomainErrorのテスト
	assert.True(t, IsDomainError(err))
}

func TestSystemError(t *testing.T) {
	// SystemErrorの作成
	originalErr := assert.AnError
	err := NewSystemError("システムエラーが発生しました", originalErr)

	// メソッドのテスト
	assert.Equal(t, "SYSTEM_ERROR", err.ErrorCode())
	assert.Equal(t, "システムエラーが発生しました", err.ErrorMessage())
	assert.Equal(t, originalErr.Error(), err.DeveloperMessage())

	// IsDomainErrorのテスト
	assert.True(t, IsDomainError(err))
}

func TestErrorConversion(t *testing.T) {
	// 通常のエラー（DomainErrorではない）
	stdErr := assert.AnError
	assert.False(t, IsDomainError(stdErr))

	// nilエラー
	assert.False(t, IsDomainError(nil))
}
