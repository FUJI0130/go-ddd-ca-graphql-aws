// internal/domain/repository/refresh_token_repository.go
package repository

import (
	"context"
	"time"

	"github.com/FUJI0130/go-ddd-ca/internal/domain/entity"
)

// RefreshTokenRepository はリフレッシュトークン管理のリポジトリインターフェース
type RefreshTokenRepository interface {
	// 第1フェーズ - コア機能

	// Store はリフレッシュトークンを保存する
	Store(ctx context.Context, token *entity.RefreshToken) error

	// GetByToken はトークン文字列からリフレッシュトークンエンティティを取得する
	GetByToken(ctx context.Context, tokenString string) (*entity.RefreshToken, error)

	// Revoke はトークンを無効化する
	Revoke(ctx context.Context, tokenID string) error

	// 第2フェーズ - 拡張機能

	// GetByUserID はユーザーIDに関連付けられたすべてのトークンを取得する
	GetByUserID(ctx context.Context, userID string) ([]*entity.RefreshToken, error)

	// UpdateLastUsed はトークンの最終使用日時を更新する
	UpdateLastUsed(ctx context.Context, tokenID string, lastUsedAt time.Time) error

	// RevokeAllForUser はユーザーの全トークンを無効化する
	RevokeAllForUser(ctx context.Context, userID string) error

	// DeleteExpired は期限切れトークンを削除する（クリーンアップ用）
	DeleteExpired(ctx context.Context) error

	// Count はユーザーのアクティブトークン数を返す
	Count(ctx context.Context, userID string) (int, error)
}
