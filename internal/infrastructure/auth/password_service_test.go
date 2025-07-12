package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestNewBCryptPasswordService(t *testing.T) {
	tests := []struct {
		name         string
		cost         int
		expectedCost int
	}{
		{
			name:         "valid cost",
			cost:         bcrypt.MinCost + 1,
			expectedCost: bcrypt.MinCost + 1,
		},
		{
			name:         "cost too low",
			cost:         bcrypt.MinCost - 1,
			expectedCost: bcrypt.DefaultCost,
		},
		{
			name:         "cost too high",
			cost:         bcrypt.MaxCost + 1,
			expectedCost: bcrypt.DefaultCost,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewBCryptPasswordService(tt.cost)
			// インターフェース型として返されるため、型アサーションで内部実装にアクセス
			if impl, ok := service.(*BCryptPasswordService); ok {
				assert.Equal(t, tt.expectedCost, impl.cost)
			} else {
				t.Errorf("Expected *BCryptPasswordService, got %T", service)
			}
		})
	}
}

func TestBCryptPasswordService_HashPassword(t *testing.T) {
	// bcrypt.MinCostを使用してテストを高速化
	service := NewBCryptPasswordService(bcrypt.MinCost)

	tests := []struct {
		name         string
		password     string
		shouldError  bool
		errorMessage string
	}{
		{
			name:        "valid password",
			password:    "secret123",
			shouldError: false,
		},
		{
			name:         "empty password",
			password:     "",
			shouldError:  true,
			errorMessage: "password cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := service.HashPassword(tt.password)

			if tt.shouldError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMessage)
				assert.Empty(t, hash)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, hash)

				// 同じパスワードで2回ハッシュを生成し、異なることを確認（ソルトの効果）
				hash2, err := service.HashPassword(tt.password)
				assert.NoError(t, err)
				assert.NotEqual(t, hash, hash2, "ハッシュは毎回異なるべき（ソルトの効果）")

				// bcryptの形式に従っていることを確認
				assert.True(t, len(hash) > 20)
				assert.Contains(t, hash, "$2a$")
			}
		})
	}
}

func TestBCryptPasswordService_VerifyPassword(t *testing.T) {
	service := NewBCryptPasswordService(bcrypt.MinCost)

	// テスト用のパスワードとハッシュ
	validPassword := "secret123"
	hash, err := service.HashPassword(validPassword)
	require.NoError(t, err)
	require.NotEmpty(t, hash)

	tests := []struct {
		name         string
		password     string
		hash         string
		shouldError  bool
		errorMessage string
	}{
		{
			name:        "valid password and hash",
			password:    validPassword,
			hash:        hash,
			shouldError: false,
		},
		{
			name:         "invalid password",
			password:     "wrongpassword",
			hash:         hash,
			shouldError:  true,
			errorMessage: "password does not match",
		},
		{
			name:         "empty password",
			password:     "",
			hash:         hash,
			shouldError:  true,
			errorMessage: "password cannot be empty",
		},
		{
			name:         "empty hash",
			password:     validPassword,
			hash:         "",
			shouldError:  true,
			errorMessage: "hash cannot be empty",
		},
		{
			name:         "invalid hash format",
			password:     validPassword,
			hash:         "invalidhashformat",
			shouldError:  true,
			errorMessage: "failed to verify password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.VerifyPassword(tt.password, tt.hash)

			if tt.shouldError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMessage)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
