package auth

import (
	"github.com/FUJI0130/go-ddd-ca/support/customerrors"
	"golang.org/x/crypto/bcrypt"
)

// PasswordService はパスワードのハッシュ化と検証を行うインターフェース
type PasswordService interface {
	// HashPassword はパスワードをハッシュ化する
	HashPassword(password string) (string, error)

	// VerifyPassword はパスワードとハッシュが一致するか検証する
	VerifyPassword(password, hash string) error
}

// BCryptPasswordService はbcryptを使用したパスワードサービスの実装
type BCryptPasswordService struct {
	cost int
}

// NewBCryptPasswordService は新しいBCryptPasswordServiceを作成する
func NewBCryptPasswordService(cost int) PasswordService {
	// コストの範囲をチェックし、無効な場合はデフォルト値を使用
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		cost = bcrypt.DefaultCost
	}
	return &BCryptPasswordService{cost: cost}
}

// HashPassword はパスワードをbcryptを使用してハッシュ化する
func (s *BCryptPasswordService) HashPassword(password string) (string, error) {
	if password == "" {
		return "", customerrors.NewValidationError("password cannot be empty", nil)
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), s.cost)
	if err != nil {
		return "", customerrors.WrapInternalServerError(err, "failed to hash password")
	}

	return string(hashedBytes), nil
}

// VerifyPassword はパスワードとハッシュが一致するかbcryptを使用して検証する
func (s *BCryptPasswordService) VerifyPassword(password, hash string) error {
	if password == "" {
		return customerrors.NewValidationError("password cannot be empty", nil)
	}

	if hash == "" {
		return customerrors.NewValidationError("hash cannot be empty", nil)
	}

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return customerrors.NewUnauthorizedError("password does not match")
		}
		return customerrors.WrapInternalServerError(err, "failed to verify password")
	}

	return nil
}
