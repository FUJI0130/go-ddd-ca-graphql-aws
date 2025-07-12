// internal/domain/entity/refresh_token.go
package entity

import (
	"strconv"
	"time"
)

// RefreshToken はリフレッシュトークンを表すエンティティ
type RefreshToken struct {
	ID         string    // トークン識別子（データベース内部用）- DBではSERIAL型だがここでは互換性のためstring型
	Token      string    // 実際のJWTトークン文字列
	UserID     string    // トークン所有者のユーザーID
	IssuedAt   time.Time // 発行日時
	ExpiresAt  time.Time // 有効期限
	LastUsedAt time.Time // 最終使用日時（オプション）
	IsRevoked  bool      // 無効化フラグ - DBではrevokedカラム
	ClientInfo string    // クライアント情報（オプション、第2フェーズ）
	IP         string    // 発行元IPアドレス（オプション、第2フェーズ）
}

// IsExpired はトークンが期限切れかどうかを確認
func (r *RefreshToken) IsExpired() bool {
	return time.Now().After(r.ExpiresAt)
}

// IsValid はトークンが有効かどうかを確認（期限内で無効化されていない）
func (r *RefreshToken) IsValid() bool {
	return !r.IsRevoked && !r.IsExpired()
}

// IDAsInt はID文字列を整数として返す。変換エラー時は0を返す
func (r *RefreshToken) IDAsInt() int {
	id, err := strconv.Atoi(r.ID)
	if err != nil {
		return 0
	}
	return id
}

// SetIDFromInt は整数IDを文字列に変換してセットする
func (r *RefreshToken) SetIDFromInt(id int) {
	r.ID = strconv.Itoa(id)
}
