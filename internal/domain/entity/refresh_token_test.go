// internal/domain/entity/refresh_token_test.go
package entity

import (
	"testing"
	"time"
)

func TestRefreshToken_IsExpired(t *testing.T) {
	tests := []struct {
		name     string
		token    RefreshToken
		expected bool
	}{
		{
			name: "未期限のトークン",
			token: RefreshToken{
				ExpiresAt: time.Now().Add(time.Hour),
			},
			expected: false,
		},
		{
			name: "期限切れのトークン",
			token: RefreshToken{
				ExpiresAt: time.Now().Add(-time.Hour),
			},
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.token.IsExpired()
			if result != test.expected {
				t.Errorf("IsExpired() = %v, expected %v", result, test.expected)
			}
		})
	}
}

func TestRefreshToken_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		token    RefreshToken
		expected bool
	}{
		{
			name: "有効なトークン",
			token: RefreshToken{
				ExpiresAt: time.Now().Add(time.Hour),
				IsRevoked: false,
			},
			expected: true,
		},
		{
			name: "無効化されたトークン",
			token: RefreshToken{
				ExpiresAt: time.Now().Add(time.Hour),
				IsRevoked: true,
			},
			expected: false,
		},
		{
			name: "期限切れのトークン",
			token: RefreshToken{
				ExpiresAt: time.Now().Add(-time.Hour),
				IsRevoked: false,
			},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.token.IsValid()
			if result != test.expected {
				t.Errorf("IsValid() = %v, expected %v", result, test.expected)
			}
		})
	}
}

func TestRefreshToken_IDAsInt(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		expected int
	}{
		{
			name:     "正の整数ID",
			id:       "123",
			expected: 123,
		},
		{
			name:     "0のID",
			id:       "0",
			expected: 0,
		},
		{
			name:     "数値でないID",
			id:       "abc",
			expected: 0, // 変換エラー時は0を返す
		},
		{
			name:     "空のID",
			id:       "",
			expected: 0, // 変換エラー時は0を返す
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			token := RefreshToken{ID: test.id}
			result := token.IDAsInt()
			if result != test.expected {
				t.Errorf("IDAsInt() = %v, expected %v", result, test.expected)
			}
		})
	}
}

func TestRefreshToken_SetIDFromInt(t *testing.T) {
	tests := []struct {
		name     string
		id       int
		expected string
	}{
		{
			name:     "正の整数ID",
			id:       123,
			expected: "123",
		},
		{
			name:     "0のID",
			id:       0,
			expected: "0",
		},
		{
			name:     "負の整数ID",
			id:       -1,
			expected: "-1",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			token := RefreshToken{}
			token.SetIDFromInt(test.id)
			if token.ID != test.expected {
				t.Errorf("SetIDFromInt(%d) sets ID to %v, expected %v", test.id, token.ID, test.expected)
			}
		})
	}
}
