// internal/infrastructure/persistence/postgres/refresh_token_repository.go
package postgres

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"github.com/FUJI0130/go-ddd-ca/internal/domain/entity"
	"github.com/FUJI0130/go-ddd-ca/internal/domain/repository"
	"github.com/FUJI0130/go-ddd-ca/internal/infrastructure/persistence/common"
	"github.com/FUJI0130/go-ddd-ca/support/customerrors"
)

// PostgresRefreshTokenRepository はPostgreSQLを使用したリフレッシュトークンリポジトリ
type PostgresRefreshTokenRepository struct {
	db common.SQLExecutor // *sql.DB または *sql.Tx を受け入れ可能
}

// NewPostgresRefreshTokenRepository は新しいPostgreSQLリポジトリインスタンスを作成
func NewPostgresRefreshTokenRepository(executor common.SQLExecutor) repository.RefreshTokenRepository {
	return &PostgresRefreshTokenRepository{
		db: executor,
	}
}

// 第1フェーズ - コア機能の実装

// Store はリフレッシュトークンをデータベースに保存する
func (r *PostgresRefreshTokenRepository) Store(ctx context.Context, token *entity.RefreshToken) error {
	if token == nil {
		return customerrors.NewValidationError("token cannot be nil", nil)
	}

	// マイグレーション適用前の場合
	query := `
		INSERT INTO refresh_tokens (
			token, user_id, expires_at, revoked, created_at, issued_at, updated_at
		) VALUES (
			$1, $2, $3, $4, CURRENT_TIMESTAMP, $5, CURRENT_TIMESTAMP
		) RETURNING id
	`

	var id int
	err := r.db.QueryRowContext(
		ctx,
		query,
		token.Token,
		token.UserID,
		token.ExpiresAt,
		token.IsRevoked,
		token.IssuedAt, // 追加されたパラメータ
	).Scan(&id)

	if err != nil {
		// PostgreSQLエラーを適切なドメインエラーに変換
		if _, ok := customerrors.IsPgUniqueViolation(err); ok {
			return customerrors.NewConflictError("token already exists")
		}
		return customerrors.WrapInternalServerError(err, "failed to store refresh token")
	}

	// 生成されたIDを設定
	token.SetIDFromInt(id)

	return nil
}

// GetByToken はトークン文字列からリフレッシュトークンエンティティを取得する
func (r *PostgresRefreshTokenRepository) GetByToken(ctx context.Context, tokenString string) (*entity.RefreshToken, error) {
	// 基本クエリ - マイグレーション適用前でも動作する最小限のフィールド
	query := `
		SELECT id, token, user_id, expires_at, revoked, created_at
		FROM refresh_tokens
		WHERE token = $1
	`

	var token entity.RefreshToken
	var id int
	var createdAt time.Time

	// 基本フィールドをスキャン
	err := r.db.QueryRowContext(ctx, query, tokenString).Scan(
		&id,
		&token.Token,
		&token.UserID,
		&token.ExpiresAt,
		&token.IsRevoked,
		&createdAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, customerrors.NewNotFoundError("token not found")
		}
		return nil, customerrors.WrapInternalServerError(err, "failed to get refresh token")
	}

	// 整数IDを文字列に変換
	token.SetIDFromInt(id)

	// マイグレーション実行前の場合、IssuedAtにcreated_atを設定
	token.IssuedAt = createdAt

	// 追加フィールドの取得を試みる（マイグレーション適用後のみ動作）
	// ここでエラーが発生しても無視（カラムが存在しない可能性があるため）
	additionalQuery := `
		SELECT issued_at, last_used_at, client_info, ip_address
		FROM refresh_tokens
		WHERE id = $1
	`
	row := r.db.QueryRowContext(ctx, additionalQuery, id)
	// エラーは無視して、返されたデータがあれば使用
	row.Scan(&token.IssuedAt, &token.LastUsedAt, &token.ClientInfo, &token.IP)

	return &token, nil
}

// Revoke はトークンを無効化する
func (r *PostgresRefreshTokenRepository) Revoke(ctx context.Context, tokenID string) error {
	// 文字列IDを整数に変換
	id, err := strconv.Atoi(tokenID)
	if err != nil {
		return customerrors.NewValidationError("invalid token id format", map[string]string{
			"tokenID": "must be a valid integer",
		})
	}

	query := `
		UPDATE refresh_tokens
		SET revoked = true
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return customerrors.WrapInternalServerError(err, "failed to revoke refresh token")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return customerrors.WrapInternalServerError(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return customerrors.NewNotFoundError("token not found")
	}

	return nil
}

// GetByUserID はユーザーIDに関連付けられたすべてのトークンを取得する
func (r *PostgresRefreshTokenRepository) GetByUserID(ctx context.Context, userID string) ([]*entity.RefreshToken, error) {
	query := `
		SELECT id, token, user_id, expires_at, revoked, created_at
		FROM refresh_tokens
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, customerrors.WrapInternalServerError(err, "failed to get refresh tokens for user")
	}
	defer rows.Close()

	var tokens []*entity.RefreshToken

	for rows.Next() {
		var token entity.RefreshToken
		var id int
		var createdAt time.Time

		err := rows.Scan(
			&id,
			&token.Token,
			&token.UserID,
			&token.ExpiresAt,
			&token.IsRevoked,
			&createdAt,
		)
		if err != nil {
			return nil, customerrors.WrapInternalServerError(err, "failed to scan refresh token")
		}

		token.SetIDFromInt(id)
		token.IssuedAt = createdAt
		tokens = append(tokens, &token)
	}

	if err = rows.Err(); err != nil {
		return nil, customerrors.WrapInternalServerError(err, "error iterating refresh tokens")
	}

	return tokens, nil
}

// UpdateLastUsed はトークンの最終使用日時を更新する
func (r *PostgresRefreshTokenRepository) UpdateLastUsed(ctx context.Context, tokenID string, lastUsedAt time.Time) error {
	// 文字列IDを整数に変換
	id, err := strconv.Atoi(tokenID)
	if err != nil {
		return customerrors.NewValidationError("invalid token id format", map[string]string{
			"tokenID": "must be a valid integer",
		})
	}

	// まず、カラムが存在するか確認
	var columnExists bool
	err = r.db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 
			FROM information_schema.columns 
			WHERE table_name = 'refresh_tokens' AND column_name = 'last_used_at'
		)
	`).Scan(&columnExists)

	if err != nil {
		return customerrors.WrapInternalServerError(err, "failed to check column existence")
	}

	// カラムが存在しない場合は何もしない
	if !columnExists {
		return nil
	}

	// カラムが存在する場合は更新
	query := `
		UPDATE refresh_tokens
		SET last_used_at = $2
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id, lastUsedAt)
	if err != nil {
		return customerrors.WrapInternalServerError(err, "failed to update last used timestamp")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return customerrors.WrapInternalServerError(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return customerrors.NewNotFoundError("token not found")
	}

	return nil
}

// RevokeAllForUser はユーザーの全トークンを無効化する
func (r *PostgresRefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID string) error {
	query := `
		UPDATE refresh_tokens
		SET revoked = true
		WHERE user_id = $1 AND NOT revoked
	`

	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return customerrors.WrapInternalServerError(err, "failed to revoke all tokens for user")
	}

	return nil
}

// DeleteExpired は期限切れトークンを削除する（クリーンアップ用）
func (r *PostgresRefreshTokenRepository) DeleteExpired(ctx context.Context) error {
	query := `
		DELETE FROM refresh_tokens
		WHERE expires_at < CURRENT_TIMESTAMP
	`

	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return customerrors.WrapInternalServerError(err, "failed to delete expired tokens")
	}

	return nil
}

// Count はユーザーのアクティブトークン数を返す
func (r *PostgresRefreshTokenRepository) Count(ctx context.Context, userID string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM refresh_tokens
		WHERE user_id = $1 AND NOT revoked AND expires_at > CURRENT_TIMESTAMP
	`

	var count int
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, customerrors.WrapInternalServerError(err, "failed to count active tokens")
	}

	return count, nil
}
