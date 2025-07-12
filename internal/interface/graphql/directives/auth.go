// internal/interface/graphql/directives/auth.go
package directives

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/FUJI0130/go-ddd-ca/internal/domain/entity"
	"github.com/FUJI0130/go-ddd-ca/internal/interface/graphql/auth"
	"github.com/FUJI0130/go-ddd-ca/support/customerrors"
)

// Auth はGQLGen用の@authディレクティブの実装で、
// リクエストが認証済みかどうかを確認します
func Auth(ctx context.Context, obj interface{}, next graphql.Resolver) (interface{}, error) {
	// コンテキストから認証情報を取得
	if !auth.IsAuthenticated(ctx) {
		// 認証されていない場合はエラーを返す
		return nil, customerrors.NewUnauthorizedError("認証が必要です")
	}

	// 認証されている場合は、次のリゾルバーを実行
	return next(ctx)
}

// HasRole はGQLGen用の@hasRoleディレクティブの実装で、
// 認証済みユーザーが指定されたロールを持っているかを確認します
func HasRole(ctx context.Context, obj interface{}, next graphql.Resolver, role string) (interface{}, error) {
	// まず認証されているか確認
	if !auth.IsAuthenticated(ctx) {
		return nil, customerrors.NewUnauthorizedError("認証が必要です")
	}

	// ユーザー情報を取得
	user := auth.GetUserFromContext(ctx)
	if user == nil {
		// これは通常起こらないはずだが、念のためチェック
		return nil, customerrors.NewUnauthorizedError("ユーザー情報が取得できません")
	}

	// ロールの検証
	// ここでは文字列比較だが、より複雑なロール階層なども実装可能
	userRole := user.Role.String()
	if userRole != role {
		return nil, customerrors.NewUnauthorizedError("このアクションを実行する権限がありません").WithContext(
			customerrors.Context{
				"required_role": role,
				"user_role":     userRole,
			},
		)
	}

	// 認証・認可に成功した場合は、次のリゾルバーを実行
	return next(ctx)
}

// ValidateRoles はロール文字列が有効かどうかを検証するヘルパー関数
func ValidateRoles(role string) bool {
	validRoles := map[string]bool{
		entity.RoleAdmin.String():   true,
		entity.RoleManager.String(): true,
		entity.RoleTester.String():  true,
	}

	return validRoles[role]
}
