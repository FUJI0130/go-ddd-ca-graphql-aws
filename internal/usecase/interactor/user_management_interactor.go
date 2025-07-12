// internal/usecase/interactor/user_management_interactor.go
package interactor

import (
	"context"
	"fmt"

	"github.com/FUJI0130/go-ddd-ca/internal/domain/entity"
	"github.com/FUJI0130/go-ddd-ca/internal/domain/repository"
	"github.com/FUJI0130/go-ddd-ca/internal/infrastructure/auth"
	infraAuth "github.com/FUJI0130/go-ddd-ca/internal/infrastructure/auth"  // パスワードサービス用
	ctxAuth "github.com/FUJI0130/go-ddd-ca/internal/interface/graphql/auth" // コンテキスト認証用

	"github.com/FUJI0130/go-ddd-ca/internal/usecase/port"
	"github.com/FUJI0130/go-ddd-ca/support/customerrors"
)

// UserManagementInteractor はユーザー管理機能を実装します
type UserManagementInteractor struct {
	userRepository  repository.UserRepository
	userIDGenerator repository.UserIDGenerator
	passwordService infraAuth.PasswordService
}

// NewUserManagementInteractor は新しいUserManagementInteractorを作成します
func NewUserManagementInteractor(
	userRepository repository.UserRepository,
	userIDGenerator repository.UserIDGenerator,
	passwordService auth.PasswordService,
) *UserManagementInteractor {
	return &UserManagementInteractor{
		userRepository:  userRepository,
		userIDGenerator: userIDGenerator,
		passwordService: passwordService,
	}
}

// CreateUser は新しいユーザーを作成します
func (i *UserManagementInteractor) CreateUser(ctx context.Context, request *port.CreateUserRequest) (*entity.User, error) {
	// 🆕 デバッグログ追加
	fmt.Printf("CreateUser リクエスト: username=[%s](len=%d), password=[%s](len=%d), role=[%s]\n",
		request.Username, len(request.Username),
		"*****", len(request.Password),
		request.Role)

	// ユーザー名のバリデーション
	if request.Username == "" {
		fmt.Println("バリデーション: ユーザー名が空です")
		return nil, customerrors.NewValidationError("ユーザー名は必須です", map[string]string{
			"username": "ユーザー名を入力してください",
		})
	}
	if len(request.Username) < 3 {
		fmt.Printf("バリデーション: ユーザー名が短すぎます [%s](len=%d)\n", request.Username, len(request.Username))
		return nil, customerrors.NewValidationError("ユーザー名が短すぎます", map[string]string{
			"username": "ユーザー名は3文字以上で入力してください",
		})
	}

	// ここで重複チェック前にデバッグ出力
	fmt.Printf("重複チェック開始: username=[%s]\n", request.Username)

	// ユーザー名が既に存在するか確認
	existingUser, err := i.userRepository.FindByUsername(ctx, request.Username)
	if err != nil && !customerrors.IsNotFoundError(err) {
		fmt.Printf("FindByUsername エラー: %v\n", err)
		return nil, err
	}
	if existingUser != nil {
		fmt.Printf("重複ユーザー検出: id=[%s], username=[%s]\n", existingUser.ID, existingUser.Username)
		return nil, customerrors.NewConflictError("ユーザー名が既に使用されています")
	}

	// パスワードのバリデーション
	if request.Password == "" {
		return nil, customerrors.NewValidationError("パスワードは必須です", map[string]string{
			"password": "パスワードを入力してください",
		})
	}
	if len(request.Password) < 6 {
		return nil, customerrors.NewValidationError("パスワードが短すぎます", map[string]string{
			"password": "パスワードは6文字以上で入力してください",
		})
	}

	// パスワードのハッシュ化
	passwordHash, err := i.passwordService.HashPassword(request.Password)
	if err != nil {
		return nil, customerrors.WrapInternalServerError(err, "パスワードのハッシュ化に失敗しました")
	}

	// ユーザーロールの検証
	var userRole entity.UserRole
	switch request.Role {
	case "Admin":
		userRole = entity.RoleAdmin
	case "Manager":
		userRole = entity.RoleManager
	case "Tester":
		userRole = entity.RoleTester
	default:
		return nil, customerrors.NewValidationError("無効なユーザーロールです", map[string]string{
			"role": "Admin, Manager, Tester のいずれかを指定してください",
		})
	}

	// ユーザーIDの生成
	userID, err := i.userIDGenerator.Generate(ctx)
	if err != nil {
		return nil, customerrors.WrapInternalServerError(err, "ユーザーID生成に失敗しました")
	}

	// ユーザーエンティティの作成
	user, err := entity.NewUser(userID, request.Username, passwordHash, userRole)
	if err != nil {
		return nil, customerrors.WrapValidationError(err, "ユーザー作成に失敗しました", nil)
	}

	// ユーザーの保存
	if err := i.userRepository.Create(ctx, user); err != nil {
		return nil, customerrors.WrapInternalServerError(err, "ユーザーの保存に失敗しました")
	}

	return user, nil
}

// ChangePassword はユーザー自身のパスワードを変更します
func (i *UserManagementInteractor) ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	// ユーザーを取得
	user, err := i.userRepository.FindByID(ctx, userID)
	if err != nil {
		if customerrors.IsNotFoundError(err) {
			return customerrors.NewNotFoundError("ユーザーが見つかりません")
		}
		return customerrors.WrapInternalServerError(err, "ユーザー情報の取得に失敗しました")
	}

	// 古いパスワードを検証
	if err := i.passwordService.VerifyPassword(oldPassword, user.PasswordHash); err != nil {
		return customerrors.NewUnauthorizedError("現在のパスワードが一致しません")
	}

	// 新しいパスワードをハッシュ化
	newPasswordHash, err := i.passwordService.HashPassword(newPassword)
	if err != nil {
		return customerrors.WrapInternalServerError(err, "パスワードのハッシュ化に失敗しました")
	}

	// ユーザー情報の更新
	user.PasswordHash = newPasswordHash
	if err := i.userRepository.Update(ctx, user); err != nil {
		return customerrors.WrapInternalServerError(err, "パスワード更新に失敗しました")
	}

	return nil
}

// ResetPassword は管理者が他のユーザーのパスワードをリセットします
func (i *UserManagementInteractor) ResetPassword(ctx context.Context, userID, newPassword string) error {
	// ユーザーを取得
	user, err := i.userRepository.FindByID(ctx, userID)
	if err != nil {
		if customerrors.IsNotFoundError(err) {
			return customerrors.NewNotFoundError("ユーザーが見つかりません")
		}
		return customerrors.WrapInternalServerError(err, "ユーザー情報の取得に失敗しました")
	}

	// 新しいパスワードをハッシュ化
	newPasswordHash, err := i.passwordService.HashPassword(newPassword)
	if err != nil {
		return customerrors.WrapInternalServerError(err, "パスワードのハッシュ化に失敗しました")
	}

	// ユーザー情報の更新
	user.PasswordHash = newPasswordHash
	if err := i.userRepository.Update(ctx, user); err != nil {
		return customerrors.WrapInternalServerError(err, "パスワード更新に失敗しました")
	}

	return nil
}

// DeleteUser はユーザーを削除します
// internal/usecase/interactor/user_management_interactor.go のDeleteUserメソッド修正
func (i *UserManagementInteractor) DeleteUser(ctx context.Context, userID string) error {
	// 削除対象ユーザーの取得
	user, err := i.userRepository.FindByID(ctx, userID)
	if err != nil {
		if customerrors.IsNotFoundError(err) {
			return customerrors.NewNotFoundError("ユーザーが見つかりません")
		}
		return customerrors.WrapInternalServerError(err, "ユーザー情報の取得に失敗しました")
	}

	// Adminの場合、最後の一人かチェック
	if user.Role == entity.RoleAdmin {
		adminCount, err := i.userRepository.CountByRole(ctx, entity.RoleAdmin)
		if err != nil {
			return customerrors.WrapInternalServerError(err, "Admin数の確認に失敗しました")
		}
		if adminCount <= 1 {
			return customerrors.NewValidationError("最後のAdminユーザーは削除できません", map[string]string{
				"reason": "システムには最低1人のAdminユーザーが必要です",
			})
		}
	}

	// ユーザーの削除
	if err := i.userRepository.Delete(ctx, userID); err != nil {
		return customerrors.WrapInternalServerError(err, "ユーザーの削除に失敗しました")
	}

	return nil
}

// FindAllUsers は全ユーザーの一覧を取得します
func (i *UserManagementInteractor) FindAllUsers(ctx context.Context) ([]*entity.User, error) {
	users, err := i.userRepository.FindAll(ctx)
	if err != nil {
		return nil, customerrors.WrapInternalServerError(err, "ユーザー一覧の取得に失敗しました")
	}

	return users, nil
}

// FindUserByID は指定されたIDのユーザーを取得します
func (i *UserManagementInteractor) FindUserByID(ctx context.Context, userID string) (*entity.User, error) {
	user, err := i.userRepository.FindByID(ctx, userID)
	if err != nil {
		if customerrors.IsNotFoundError(err) {
			return nil, customerrors.NewNotFoundError("ユーザーが見つかりません")
		}
		return nil, customerrors.WrapInternalServerError(err, "ユーザー情報の取得に失敗しました")
	}

	return user, nil
}

func (i *UserManagementInteractor) UpdateUser(ctx context.Context, userID string, request *port.UpdateUserRequest) (*entity.User, error) {
	// 更新対象ユーザーの取得
	user, err := i.userRepository.FindByID(ctx, userID)
	if err != nil {
		if customerrors.IsNotFoundError(err) {
			return nil, customerrors.NewNotFoundError("ユーザーが見つかりません")
		}
		return nil, customerrors.WrapInternalServerError(err, "ユーザー情報の取得に失敗しました")
	}

	// ユーザー名のバリデーション
	if request.Username != user.Username {
		// ユーザー名変更時は重複チェック
		existingUser, err := i.userRepository.FindByUsername(ctx, request.Username)
		if err != nil && !customerrors.IsNotFoundError(err) {
			return nil, customerrors.WrapInternalServerError(err, "ユーザー検索に失敗しました")
		}
		if existingUser != nil {
			return nil, customerrors.NewConflictError("ユーザー名が既に使用されています")
		}
	}

	// ロールのバリデーション
	var userRole entity.UserRole
	switch request.Role {
	case "Admin":
		userRole = entity.RoleAdmin
	case "Manager":
		userRole = entity.RoleManager
	case "Tester":
		userRole = entity.RoleTester
	default:
		return nil, customerrors.NewValidationError("無効なユーザーロールです", map[string]string{
			"role": "Admin, Manager, Testerのいずれかを指定してください",
		})
	}

	// 最後のAdminロール変更防止
	if user.Role == entity.RoleAdmin && userRole != entity.RoleAdmin {
		adminCount, err := i.userRepository.CountByRole(ctx, entity.RoleAdmin)
		if err != nil {
			return nil, customerrors.WrapInternalServerError(err, "Admin数の確認に失敗しました")
		}
		if adminCount <= 1 {
			return nil, customerrors.NewValidationError("最後のAdminユーザーのロールを変更することはできません", nil)
		}
	}

	// 自分自身のロール変更を防止
	authUser := ctxAuth.GetUserFromContext(ctx)
	if authUser != nil && authUser.ID == userID && user.Role != userRole {
		return nil, customerrors.NewValidationError("自分自身のロールを変更することはできません", nil)
	}

	// ユーザー情報の更新
	user.Username = request.Username
	user.Role = userRole

	if err := i.userRepository.Update(ctx, user); err != nil {
		return nil, customerrors.WrapInternalServerError(err, "ユーザー情報の更新に失敗しました")
	}

	return user, nil
}
