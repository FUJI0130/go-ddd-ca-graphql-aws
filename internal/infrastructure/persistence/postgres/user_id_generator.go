package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/FUJI0130/go-ddd-ca/internal/domain/repository"
	"github.com/FUJI0130/go-ddd-ca/support/customerrors"
)

// UserIDGenerator はPostgreSQL用のユーザーIDジェネレーター実装
type UserIDGenerator struct {
	db *sql.DB
}

// NewUserIDGenerator は新しいUserIDGeneratorインスタンスを作成する
func NewUserIDGenerator(db *sql.DB) repository.UserIDGenerator {
	return &UserIDGenerator{db: db}
}

// Generate は新しいユーザーIDを生成する
// PostgreSQLのシーケンスuser_seqを使用して採番し、"user_"プレフィックスを付ける
func (g *UserIDGenerator) Generate(ctx context.Context) (string, error) {
	var id int64
	err := g.db.QueryRowContext(ctx, "SELECT nextval('user_seq')").Scan(&id)
	if err != nil {
		return "", customerrors.DBError("generate_id", "user", err).WithContext(customerrors.Context{
			"sequence": "user_seq",
		})
	}

	// "user_"プレフィックス + シーケンス番号の形式でIDを生成
	return fmt.Sprintf("user_%d", id), nil
}
