// internal/infrastructure/persistence/memory/refresh_token_repository.go
package memory

import (
	"context"
	"sync"
	"time"

	"github.com/FUJI0130/go-ddd-ca/internal/domain/entity"
	"github.com/FUJI0130/go-ddd-ca/internal/domain/repository"
	"github.com/FUJI0130/go-ddd-ca/support/customerrors"
)

// MemoryRefreshTokenRepository はリフレッシュトークンをメモリ内に保存するリポジトリ実装
type MemoryRefreshTokenRepository struct {
	tokens   map[string]*entity.RefreshToken // tokenIDをキーとするマップ
	tokenMap map[string]string               // tokenString -> tokenID のマップ
	userMap  map[string][]string             // userID -> tokenIDs のマップ
	mutex    sync.RWMutex
}

// NewMemoryRefreshTokenRepository は新しいメモリ内リポジトリインスタンスを作成
func NewMemoryRefreshTokenRepository() repository.RefreshTokenRepository {
	return &MemoryRefreshTokenRepository{
		tokens:   make(map[string]*entity.RefreshToken),
		tokenMap: make(map[string]string),
		userMap:  make(map[string][]string),
	}
}

// 第1フェーズ - コア機能の実装

// Store はリフレッシュトークンを保存する
func (r *MemoryRefreshTokenRepository) Store(ctx context.Context, token *entity.RefreshToken) error {
	if token == nil {
		return customerrors.NewValidationError("token cannot be nil", nil)
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	// 既存のトークン文字列をチェック
	if _, exists := r.tokenMap[token.Token]; exists {
		return customerrors.NewConflictError("token already exists")
	}

	// トークンを保存
	r.tokens[token.ID] = token
	r.tokenMap[token.Token] = token.ID

	// ユーザーIDとトークンIDのマッピングを更新
	r.userMap[token.UserID] = append(r.userMap[token.UserID], token.ID)

	return nil
}

// GetByToken はトークン文字列からリフレッシュトークンエンティティを取得する
func (r *MemoryRefreshTokenRepository) GetByToken(ctx context.Context, tokenString string) (*entity.RefreshToken, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	tokenID, exists := r.tokenMap[tokenString]
	if !exists {
		return nil, customerrors.NewNotFoundError("token not found")
	}

	token, exists := r.tokens[tokenID]
	if !exists {
		return nil, customerrors.NewNotFoundError("token not found")
	}

	// コピーを返して、呼び出し元による変更を防止
	tokenCopy := *token
	return &tokenCopy, nil
}

// Revoke はトークンを無効化する
func (r *MemoryRefreshTokenRepository) Revoke(ctx context.Context, tokenID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	token, exists := r.tokens[tokenID]
	if !exists {
		return customerrors.NewNotFoundError("token not found")
	}

	// トークンを無効化
	token.IsRevoked = true

	return nil
}

// 第2フェーズ - 拡張機能の実装

// GetByUserID はユーザーIDに関連付けられたすべてのトークンを取得する
func (r *MemoryRefreshTokenRepository) GetByUserID(ctx context.Context, userID string) ([]*entity.RefreshToken, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	tokenIDs, exists := r.userMap[userID]
	if !exists {
		return []*entity.RefreshToken{}, nil
	}

	var tokens []*entity.RefreshToken
	for _, tokenID := range tokenIDs {
		if token, ok := r.tokens[tokenID]; ok {
			// コピーを追加
			tokenCopy := *token
			tokens = append(tokens, &tokenCopy)
		}
	}

	return tokens, nil
}

// UpdateLastUsed はトークンの最終使用日時を更新する
func (r *MemoryRefreshTokenRepository) UpdateLastUsed(ctx context.Context, tokenID string, lastUsedAt time.Time) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	token, exists := r.tokens[tokenID]
	if !exists {
		return customerrors.NewNotFoundError("token not found")
	}

	token.LastUsedAt = lastUsedAt
	return nil
}

// RevokeAllForUser はユーザーの全トークンを無効化する
func (r *MemoryRefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	tokenIDs, exists := r.userMap[userID]
	if !exists {
		return nil // ユーザーのトークンが存在しない場合は成功とみなす
	}

	for _, tokenID := range tokenIDs {
		if token, ok := r.tokens[tokenID]; ok {
			token.IsRevoked = true
		}
	}

	return nil
}

// DeleteExpired は期限切れトークンを削除する（クリーンアップ用）
func (r *MemoryRefreshTokenRepository) DeleteExpired(ctx context.Context) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	now := time.Now()
	var expiredTokenIDs []string
	var expiredTokenStrings []string

	// 期限切れトークンを特定
	for id, token := range r.tokens {
		if now.After(token.ExpiresAt) {
			expiredTokenIDs = append(expiredTokenIDs, id)
			expiredTokenStrings = append(expiredTokenStrings, token.Token)
		}
	}

	// ユーザーマップからも削除
	for userID, tokenIDs := range r.userMap {
		var updatedTokenIDs []string
		for _, tokenID := range tokenIDs {
			// 期限切れでないトークンIDのみ残す
			expired := false
			for _, expiredID := range expiredTokenIDs {
				if tokenID == expiredID {
					expired = true
					break
				}
			}
			if !expired {
				updatedTokenIDs = append(updatedTokenIDs, tokenID)
			}
		}
		// 更新されたリストを設定（空の場合はマップからエントリを削除）
		if len(updatedTokenIDs) == 0 {
			delete(r.userMap, userID)
		} else {
			r.userMap[userID] = updatedTokenIDs
		}
	}

	// トークンマップからトークン文字列のマッピングを削除
	for _, tokenString := range expiredTokenStrings {
		delete(r.tokenMap, tokenString)
	}

	// トークンマップから実際のトークンを削除
	for _, tokenID := range expiredTokenIDs {
		delete(r.tokens, tokenID)
	}

	return nil
}

// Count はユーザーのアクティブトークン数を返す
func (r *MemoryRefreshTokenRepository) Count(ctx context.Context, userID string) (int, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	tokenIDs, exists := r.userMap[userID]
	if !exists {
		return 0, nil
	}

	count := 0
	now := time.Now()
	for _, tokenID := range tokenIDs {
		if token, ok := r.tokens[tokenID]; ok {
			if !token.IsRevoked && !now.After(token.ExpiresAt) {
				count++
			}
		}
	}

	return count, nil
}
