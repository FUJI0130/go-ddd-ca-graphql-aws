package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"time"

	"github.com/FUJI0130/go-ddd-ca/internal/domain/repository"
)

// TestEnvironmentUserIDGenerator はテスト環境用のユーザーIDジェネレーター
type TestEnvironmentUserIDGenerator struct {
	db *sql.DB
}

// NewTestEnvironmentUserIDGenerator は新しいTestEnvironmentUserIDGeneratorインスタンスを作成する
func NewTestEnvironmentUserIDGenerator(db *sql.DB) repository.UserIDGenerator {
	// 乱数生成器の初期化
	rand.Seed(time.Now().UnixNano())
	return &TestEnvironmentUserIDGenerator{db: db}
}

// Generate はテスト用のユーザーIDを生成する
// テスト環境では "test_user_" + ランダム数字 の形式でIDを生成
func (g *TestEnvironmentUserIDGenerator) Generate(ctx context.Context) (string, error) {
	var id int64
	err := g.db.QueryRowContext(ctx, "SELECT nextval('user_seq')").Scan(&id)
	if err != nil {
		// シーケンスが存在しない場合はランダムIDを使用
		randID := rand.Intn(10000)
		return fmt.Sprintf("test_user_%d", randID), nil
	}

	return fmt.Sprintf("test_user_%d", id), nil
}
