// internal/interface/graphql/auth/middleware.go
package auth

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/FUJI0130/go-ddd-ca/internal/domain/entity"
	"github.com/FUJI0130/go-ddd-ca/internal/usecase/port"
)

// AuthMiddleware はHTTPリクエストからJWTトークンを抽出し、
// トークンを検証してユーザー情報をコンテキストに設定するミドルウェアです
func AuthMiddleware(authUseCase port.AuthUseCase) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// リクエストヘッダーからトークンを抽出
			token := extractTokenFromRequest(r)

			// トークンが存在する場合のみ検証を行う
			if token != "" {
				// トークンを検証してユーザー情報を取得
				user, err := authUseCase.ValidateToken(r.Context(), token)
				if err != nil {
					// トークン検証エラーはログに記録するが、
					// リクエスト自体は拒否せず、非認証状態として処理を続行
					log.Printf("Token validation failed: %v", err)
				} else if user != nil {
					// 有効なトークンの場合、認証情報をコンテキストに設定
					authInfo := &AuthInfo{
						User:         user,
						IsAuthorized: true,
					}
					ctx := SetAuthInfo(r.Context(), authInfo)
					r = r.WithContext(ctx)
				}
			}

			// 次のハンドラーを呼び出す
			next.ServeHTTP(w, r)
		})
	}
}

// extractTokenFromHeader はHTTPリクエストヘッダーからBearerトークンを抽出します
func extractTokenFromHeader(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}

	return parts[1]
}

func extractTokenFromRequest(r *http.Request) string {
	// 既存：Authorizationヘッダー確認
	if token := extractTokenFromHeader(r); token != "" {
		return token
	}

	// 追加：Cookie確認
	if cookie, err := r.Cookie("auth_token"); err == nil {
		return cookie.Value
	}

	return ""
}

// WithUser はテスト用のヘルパー関数で、指定したユーザーで認証済みとしてコンテキストを設定
func WithUser(ctx context.Context, user interface{}) context.Context {
	// 型アサーションでエンティティユーザー型に変換
	entityUser, ok := user.(*entity.User)
	if !ok {
		log.Printf("Failed to assert user type in WithUser")
		return ctx
	}

	// 認証情報を設定
	authInfo := &AuthInfo{
		User:         entityUser,
		IsAuthorized: true,
	}
	return SetAuthInfo(ctx, authInfo)
}
