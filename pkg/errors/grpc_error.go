package errors

import (
	"fmt"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// convertDetailsToStringMap はmap[string]interface{}をmap[string]stringに変換
func convertDetailsToStringMap(details map[string]interface{}) map[string]string {
	result := make(map[string]string)
	for k, v := range details {
		switch val := v.(type) {
		case string:
			result[k] = val
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
			result[k] = toString(val)
		case bool:
			if val {
				result[k] = "true"
			} else {
				result[k] = "false"
			}
		default:
			// その他の型は文字列表現を使用
			result[k] = toString(val)
		}
	}
	return result
}

// toString は任意の値を文字列に変換
func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	// 無限再帰を避ける修正
	return fmt.Sprintf("%v", v)
}

// determineGRPCCode はドメインエラーからgRPCのステータスコードを決定
func determineGRPCCode(err DomainError) codes.Code {
	switch err.(type) {
	case *ValidationError, *InvalidDateRangeError, *InvalidInputError:
		return codes.InvalidArgument
	case *NotFoundError, *EntityNotFoundError, *TestSuiteNotFoundError, *TestGroupNotFoundError, *TestCaseNotFoundError:
		return codes.NotFound
	case *ConflictError, *ConcurrentModificationError, *AlreadyExistsError:
		return codes.Aborted
	case *UnauthorizedError:
		return codes.Unauthenticated
	case *PermissionError, *ForbiddenError:
		return codes.PermissionDenied
	case *SystemError, *DatabaseError, *ExternalServiceError, *InternalServerError:
		return codes.Internal
	default:
		return codes.Unknown
	}
}

// ToGRPCError はドメインエラーをgRPC用のステータスエラーに変換
func ToGRPCError(err error) error {
	// nil エラーの場合はそのまま返す
	if err == nil {
		return nil
	}

	// すでにgRPCステータスエラーの場合はそのまま返す
	if _, ok := status.FromError(err); ok {
		return err
	}

	// DomainErrorインターフェースを実装しているか確認
	domainErr, ok := err.(DomainError)
	if !ok {
		// 通常のエラーは内部エラーとして扱う
		return status.Error(codes.Internal, err.Error())
	}

	// gRPCステータスコードを決定
	code := determineGRPCCode(domainErr)

	// 基本的なステータスを作成
	st := status.New(code, domainErr.ErrorMessage())

	// 詳細情報をProtobuf形式に変換
	details := &errdetails.ErrorInfo{
		Reason:   domainErr.ErrorCode(),
		Domain:   "testsuite.service",
		Metadata: convertDetailsToStringMap(domainErr.Details()),
	}

	// 開発者向けメッセージ（存在する場合）
	if devMsg := domainErr.DeveloperMessage(); devMsg != "" {
		debugInfo := &errdetails.DebugInfo{
			Detail: devMsg,
		}

		// 詳細情報を追加
		var err error
		st, err = st.WithDetails(details, debugInfo)
		if err != nil {
			// 詳細情報の追加に失敗した場合は元のステータスを使用
			return status.Error(code, domainErr.ErrorMessage())
		}
	} else {
		// 開発者向けメッセージがない場合は基本情報のみ追加
		var err error
		st, err = st.WithDetails(details)
		if err != nil {
			return status.Error(code, domainErr.ErrorMessage())
		}
	}

	return st.Err()
}

// FromGRPCError はgRPCステータスエラーからドメインエラーへの変換を試みる
// クライアント側での使用を想定
func FromGRPCError(err error) error {
	// nil エラーの場合はそのまま返す
	if err == nil {
		return nil
	}

	// gRPCステータスエラーの場合は変換を試みる
	st, ok := status.FromError(err)
	if !ok {
		// gRPCステータスエラーでない場合はそのまま返す
		return err
	}

	// ステータスコードに基づいて適切なドメインエラーを生成
	switch st.Code() {
	case codes.InvalidArgument:
		return &ValidationError{
			BaseDomainError: BaseDomainError{
				Code:    "VALIDATION_ERROR",
				Message: st.Message(),
			},
		}
	case codes.NotFound:
		return &NotFoundError{
			BaseDomainError: BaseDomainError{
				Code:    "NOT_FOUND",
				Message: st.Message(),
			},
		}
	case codes.Aborted:
		return &ConflictError{
			BaseDomainError: BaseDomainError{
				Code:    "CONFLICT",
				Message: st.Message(),
			},
		}
	case codes.Unauthenticated:
		return &UnauthorizedError{
			BaseDomainError: BaseDomainError{
				Code:    "UNAUTHORIZED",
				Message: st.Message(),
			},
		}
	case codes.PermissionDenied:
		return &PermissionError{
			BaseDomainError: BaseDomainError{
				Code:    "PERMISSION_ERROR",
				Message: st.Message(),
			},
		}
	default:
		return &SystemError{
			BaseDomainError: BaseDomainError{
				Code:    "SYSTEM_ERROR",
				Message: st.Message(),
			},
		}
	}
}
