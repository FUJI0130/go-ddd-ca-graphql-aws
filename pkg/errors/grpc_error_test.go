package errors

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestToGRPCError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode codes.Code
	}{
		{
			name:     "ValidationError",
			err:      NewDomainValidationError("バリデーションエラー", nil),
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "EntityNotFoundError",
			err:      NewTestSuiteNotFoundError("TS001"),
			wantCode: codes.NotFound,
		},
		{
			name:     "ConflictError",
			err:      NewDomainConflictError("TestSuite", "TS001", ""),
			wantCode: codes.Aborted,
		},
		{
			name:     "UnauthorizedError",
			err:      NewDomainUnauthorizedError(),
			wantCode: codes.Unauthenticated,
		},
		{
			name:     "SystemError",
			err:      NewSystemError("システムエラー", nil),
			wantCode: codes.Internal,
		},
		{
			name:     "StandardError",
			err:      assert.AnError,
			wantCode: codes.Internal,
		},
		{
			name:     "NilError",
			err:      nil,
			wantCode: codes.OK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grpcErr := ToGRPCError(tt.err)

			if tt.err == nil {
				assert.Nil(t, grpcErr)
				return
			}

			st, ok := status.FromError(grpcErr)
			assert.True(t, ok, "Expected gRPC status error")
			assert.Equal(t, tt.wantCode, st.Code())

			// ドメインエラーの場合は詳細情報が含まれているか確認
			if IsDomainError(tt.err) {
				domainErr := tt.err.(DomainError)
				assert.Contains(t, st.Message(), domainErr.ErrorMessage())
			}
		})
	}
}

func TestFromGRPCError(t *testing.T) {
	tests := []struct {
		name     string
		code     codes.Code
		message  string
		wantType string
	}{
		{
			name:     "InvalidArgument",
			code:     codes.InvalidArgument,
			message:  "入力が不正です",
			wantType: "*errors.ValidationError",
		},
		{
			name:     "NotFound",
			code:     codes.NotFound,
			message:  "見つかりません",
			wantType: "*errors.NotFoundError",
		},
		{
			name:     "Aborted",
			code:     codes.Aborted,
			message:  "競合が発生しました",
			wantType: "*errors.ConflictError",
		},
		{
			name:     "Unauthenticated",
			code:     codes.Unauthenticated,
			message:  "認証が必要です",
			wantType: "*errors.UnauthorizedError",
		},
		{
			name:     "PermissionDenied",
			code:     codes.PermissionDenied,
			message:  "権限がありません",
			wantType: "*errors.PermissionError",
		},
		{
			name:     "Internal",
			code:     codes.Internal,
			message:  "内部エラー",
			wantType: "*errors.SystemError",
		},
		{
			name:     "NilError",
			code:     codes.OK,
			message:  "",
			wantType: "<nil>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var grpcErr error
			if tt.code != codes.OK {
				grpcErr = status.Error(tt.code, tt.message)
			}

			domainErr := FromGRPCError(grpcErr)

			if tt.code == codes.OK {
				assert.Nil(t, domainErr)
				return
			}

			assert.NotNil(t, domainErr)
			if domainErr != nil {
				// 型情報を文字列として取得して比較
				assert.Equal(t, tt.wantType, fmt.Sprintf("%T", domainErr))

				if de, ok := domainErr.(DomainError); ok {
					assert.Equal(t, tt.message, de.ErrorMessage())
				}
			}
		})
	}
}

// 通常のエラーをgRPC変換した後、再度ドメインエラーに戻せるかテスト
func TestRoundTripConversion(t *testing.T) {
	// オリジナルのドメインエラー
	original := NewTestSuiteNotFoundError("TS001-202501")

	// gRPCエラーに変換
	grpcErr := ToGRPCError(original)

	// 再度ドメインエラーに変換
	reconverted := FromGRPCError(grpcErr)

	// 型と基本情報の確認
	assert.IsType(t, &NotFoundError{}, reconverted)

	if de, ok := reconverted.(DomainError); ok {
		assert.Equal(t, "NOT_FOUND", de.ErrorCode())
		assert.Contains(t, de.ErrorMessage(), "TS001-202501")
	} else {
		t.Fail()
	}
}
