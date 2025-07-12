// internal/domain/entity/user.go
package entity

import (
	"time"

	"github.com/cockroachdb/errors"

	domainerrors "github.com/FUJI0130/go-ddd-ca/internal/domain/errors"
)

// String はUserRoleを文字列に変換する
func (r UserRole) String() string {
	return string(r)
}

// UserRole はユーザーの役割を表す
type UserRole string

const (
	RoleAdmin   UserRole = "Admin"
	RoleManager UserRole = "Manager"
	RoleTester  UserRole = "Tester"
)

// ドメインエラー定義のための関数
// エラー作成はファクトリ関数として、ユースケース層でエラー実装を作成
type userErrorFactory interface {
	EmptyUserID() domainerrors.DomainError
	EmptyUsername() domainerrors.DomainError
	EmptyPasswordHash() domainerrors.DomainError
	InvalidUserRole() domainerrors.DomainError
}

var userErrors userErrorFactory

// SetUserErrorFactory はユーザーエンティティで使用するエラーファクトリを設定
func SetUserErrorFactory(factory userErrorFactory) {
	userErrors = factory
}

// User はシステムユーザーを表す
type User struct {
	ID           string
	Username     string
	PasswordHash string
	Role         UserRole
	CreatedAt    time.Time
	UpdatedAt    time.Time
	LastLoginAt  *time.Time
}

// NewUser は新しいユーザーエンティティを作成する
func NewUser(id, username, passwordHash string, role UserRole) (*User, error) {
	if userErrors == nil {
		// エラーファクトリが設定されていない場合は標準エラーを使用
		if id == "" {
			return nil, ErrEmptyUserID
		}
		// 他のチェックも同様...
	} else {
		if id == "" {
			return nil, userErrors.EmptyUserID()
		}
		if username == "" {
			return nil, userErrors.EmptyUsername()
		}
		if passwordHash == "" {
			return nil, userErrors.EmptyPasswordHash()
		}
		if !isValidRole(role) {
			return nil, userErrors.InvalidUserRole()
		}
	}

	now := time.Now()
	return &User{
		ID:           id,
		Username:     username,
		PasswordHash: passwordHash,
		Role:         role,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// 標準エラーのフォールバック
var (
	ErrEmptyUserID       = errors.New("ユーザーIDは必須です")
	ErrEmptyUsername     = errors.New("ユーザー名は必須です")
	ErrEmptyPasswordHash = errors.New("パスワードハッシュは必須です")
	ErrInvalidUserRole   = errors.New("無効なユーザーロールです")
)

// 残りのメソッド実装...

// UpdateLastLogin はユーザーの最終ログイン日時を更新する
func (u *User) UpdateLastLogin() {
	now := time.Now()
	u.LastLoginAt = &now
	u.UpdatedAt = now
}

// isValidRole はロールが有効かどうかを検証する
func isValidRole(role UserRole) bool {
	return role == RoleAdmin || role == RoleManager || role == RoleTester
}

// CanCreateTestSuite はテストスイート作成権限を持つかチェック
func (u *User) CanCreateTestSuite() bool {
	return u.Role == RoleAdmin || u.Role == RoleManager
}

// CanUpdateTestSuite はテストスイート更新権限を持つかチェック
func (u *User) CanUpdateTestSuite() bool {
	return u.Role == RoleAdmin || u.Role == RoleManager
}

// CanViewTestSuite はテストスイート閲覧権限を持つかチェック
func (u *User) CanViewTestSuite() bool {
	return true // すべてのユーザーが閲覧可能
}

// CanUpdateTestCase はテストケース更新権限を持つかチェック
func (u *User) CanUpdateTestCase() bool {
	return true // すべてのユーザーがテストケースを更新可能
}

// CanRecordEffort は工数記録権限を持つかチェック
func (u *User) CanRecordEffort() bool {
	return true // すべてのユーザーが工数を記録可能
}
