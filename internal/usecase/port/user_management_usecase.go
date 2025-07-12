// internal/usecase/port/user_management_usecase.go
package port

import (
	"context"

	"github.com/FUJI0130/go-ddd-ca/internal/domain/entity"
)

// CreateUserRequest はユーザー作成リクエストのデータを表します
type CreateUserRequest struct {
	Username string
	Password string
	Role     string
}

// UpdateUserRequest はユーザー更新リクエストデータ
type UpdateUserRequest struct {
	Username string
	Role     string
}

// UserManagementUseCase はユーザー管理機能のインターフェースを定義します
type UserManagementUseCase interface {
	// CreateUser は新しいユーザーを作成します（管理者権限が必要）
	CreateUser(ctx context.Context, request *CreateUserRequest) (*entity.User, error)

	// ChangePassword はユーザー自身のパスワードを変更します
	ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error

	// ResetPassword は管理者が他のユーザーのパスワードをリセットします
	ResetPassword(ctx context.Context, userID, newPassword string) error

	// 🆕 DeleteUser はユーザーを削除します（管理者権限が必要）
	DeleteUser(ctx context.Context, userID string) error

	// 🆕 FindAllUsers は全ユーザーの一覧を取得します（管理者権限が必要）
	FindAllUsers(ctx context.Context) ([]*entity.User, error)

	// 🆕 FindUserByID は指定されたIDのユーザーを取得します（管理者権限が必要）
	FindUserByID(ctx context.Context, userID string) (*entity.User, error)

	UpdateUser(ctx context.Context, userID string, request *UpdateUserRequest) (*entity.User, error)
}
