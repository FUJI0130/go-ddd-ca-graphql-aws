// internal/interface/graphql/auth/context.go
package auth

import (
	"context"
	"net/http"

	"github.com/FUJI0130/go-ddd-ca/internal/domain/entity"
)

// AuthInfo は認証情報を保持する構造体
type AuthInfo struct {
	User         *entity.User // 認証済みユーザー
	IsAuthorized bool         // 認証済みかどうか
}

// contextキーの型を定義（string型のキーのみでなく型の異なるキーを使うことでコンテキストのキー衝突を避ける）
type contextKey string

// 認証情報用のコンテキストキー
const authInfoKey contextKey = "auth_info"

// SetAuthInfo はコンテキストに認証情報を設定する
func SetAuthInfo(ctx context.Context, authInfo *AuthInfo) context.Context {
	return context.WithValue(ctx, authInfoKey, authInfo)
}

// GetAuthInfo はコンテキストから認証情報を取得する
func GetAuthInfo(ctx context.Context) *AuthInfo {
	value := ctx.Value(authInfoKey)
	if value == nil {
		return &AuthInfo{
			IsAuthorized: false,
			User:         nil,
		}
	}

	authInfo, ok := value.(*AuthInfo)
	if !ok {
		return &AuthInfo{
			IsAuthorized: false,
			User:         nil,
		}
	}

	return authInfo
}

// GetUserFromContext はコンテキストからユーザー情報を取得する便利関数
func GetUserFromContext(ctx context.Context) *entity.User {
	authInfo := GetAuthInfo(ctx)
	if authInfo == nil || !authInfo.IsAuthorized {
		return nil
	}
	return authInfo.User
}

// IsAuthenticated はコンテキストから認証済みかどうかを確認する便利関数
func IsAuthenticated(ctx context.Context) bool {
	authInfo := GetAuthInfo(ctx)
	return authInfo != nil && authInfo.IsAuthorized
}

// HasRole はユーザーが指定したロールを持っているか確認する便利関数
func HasRole(ctx context.Context, role entity.UserRole) bool {
	user := GetUserFromContext(ctx)
	if user == nil {
		return false
	}
	return user.Role == role
}

// 既存コードはそのまま維持し、以下を追加

// HTTPレスポンスライター用のコンテキストキー（既存のものと重複しないよう）
const responseWriterKey contextKey = "response_writer"

// WithResponseWriter はコンテキストにHTTPレスポンスライターを設定する
func WithResponseWriter(ctx context.Context, w http.ResponseWriter) context.Context {
	return context.WithValue(ctx, responseWriterKey, w)
}

// GetResponseWriterFromContext はコンテキストからHTTPレスポンスライターを取得する
func GetResponseWriterFromContext(ctx context.Context) http.ResponseWriter {
	if w, ok := ctx.Value(responseWriterKey).(http.ResponseWriter); ok {
		return w
	}
	return nil
}
