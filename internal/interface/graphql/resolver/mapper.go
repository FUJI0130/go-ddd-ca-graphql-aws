// internal/interface/graphql/resolver/mapper.go
package resolver

import (
	"github.com/FUJI0130/go-ddd-ca/internal/domain/entity"
	"github.com/FUJI0130/go-ddd-ca/internal/interface/graphql/model"
)

// mapUserEntityToModel はユーザーエンティティをGraphQLモデルに変換するヘルパー関数
// この関数はauth.resolvers.goで使用されているが、コード生成の影響を受けないよう
// 別ファイルに実装している
func mapUserEntityToModel(user *entity.User) *model.User {
	if user == nil {
		return nil
	}

	return &model.User{
		ID:          user.ID,
		Username:    user.Username,
		Role:        string(user.Role),
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		LastLoginAt: user.LastLoginAt,
	}
}
