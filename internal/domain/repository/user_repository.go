package repository

import (
	"context"

	"github.com/FUJI0130/go-ddd-ca/internal/domain/entity"
)

// UserRepository はユーザーの永続化を担当するリポジトリインターフェース
type UserRepository interface {
	Repository[entity.User]

	// FindByUsername は指定されたユーザー名のユーザーを取得する
	FindByUsername(ctx context.Context, username string) (*entity.User, error)

	// UpdateLastLogin はユーザーの最終ログイン日時を更新する
	UpdateLastLogin(ctx context.Context, id string) error

	// FindAll は全ユーザーの一覧を取得する
	FindAll(ctx context.Context) ([]*entity.User, error)

	// internal/domain/repository/user_repository.go
	// CountByRole は指定されたロールを持つユーザーの数を取得する
	CountByRole(ctx context.Context, role entity.UserRole) (int, error)
}
