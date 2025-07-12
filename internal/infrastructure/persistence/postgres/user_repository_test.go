package postgres

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/FUJI0130/go-ddd-ca/internal/domain/entity"
	"github.com/FUJI0130/go-ddd-ca/support/customerrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgresUserRepository_Create(t *testing.T) {
	// テスト用のデータベース接続を取得
	db, err := setupTestDB()
	require.NoError(t, err)
	defer db.Close()

	// リポジトリとIDジェネレーターを作成
	repo := NewUserRepository(db)
	idGen := NewTestEnvironmentUserIDGenerator(db)

	// テスト用ユーザーを作成
	ctx := context.Background()
	id, err := idGen.Generate(ctx)
	require.NoError(t, err)

	now := time.Now()
	user := &entity.User{
		ID:           id,
		Username:     "testuser",
		PasswordHash: "hashedpassword",
		Role:         entity.RoleManager,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// ユーザーを作成
	err = repo.Create(ctx, user)
	assert.NoError(t, err)

	// 一意性制約違反テスト
	duplicateUser := &entity.User{
		ID:           "duplicate_id",
		Username:     "testuser", // 既存のユーザー名を使用
		PasswordHash: "hashedpassword",
		Role:         entity.RoleManager,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	err = repo.Create(ctx, duplicateUser)
	assert.Error(t, err)
	assert.True(t, customerrors.IsConflictError(err))

	// テスト後のクリーンアップ
	cleanupUser(t, db, id)
}

func TestPostgresUserRepository_FindByID(t *testing.T) {
	db, err := setupTestDB()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	// テスト用ユーザーを挿入
	id := insertTestUser(t, db, "findbyid_user", entity.RoleTester)

	// ユーザーを検索
	user, err := repo.FindByID(ctx, id)
	assert.NoError(t, err)
	assert.Equal(t, id, user.ID)
	assert.Equal(t, "findbyid_user", user.Username)
	assert.Equal(t, entity.RoleTester, user.Role)

	// 存在しないIDでの検索
	_, err = repo.FindByID(ctx, "nonexistent_id")
	assert.Error(t, err)
	assert.True(t, customerrors.IsNotFoundError(err))

	// クリーンアップ
	cleanupUser(t, db, id)
}

func TestPostgresUserRepository_FindByUsername(t *testing.T) {
	db, err := setupTestDB()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	// テスト用ユーザーを挿入
	id := insertTestUser(t, db, "findbyusername_user", entity.RoleAdmin)

	// ユーザー名で検索
	user, err := repo.FindByUsername(ctx, "findbyusername_user")
	assert.NoError(t, err)
	assert.Equal(t, id, user.ID)
	assert.Equal(t, "findbyusername_user", user.Username)
	assert.Equal(t, entity.RoleAdmin, user.Role)

	// 存在しないユーザー名での検索
	_, err = repo.FindByUsername(ctx, "nonexistent_user")
	assert.Error(t, err)
	assert.True(t, customerrors.IsNotFoundError(err))

	// クリーンアップ
	cleanupUser(t, db, id)
}

func TestPostgresUserRepository_Update(t *testing.T) {
	db, err := setupTestDB()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	// テスト用ユーザーを挿入
	id := insertTestUser(t, db, "update_user", entity.RoleManager)

	// ユーザーを取得して更新
	user, err := repo.FindByID(ctx, id)
	require.NoError(t, err)

	// ユーザー情報を更新
	user.Username = "updated_username"
	user.Role = entity.RoleAdmin
	user.UpdatedAt = time.Now()

	err = repo.Update(ctx, user)
	assert.NoError(t, err)

	// 更新後のユーザーを取得して確認
	updatedUser, err := repo.FindByID(ctx, id)
	assert.NoError(t, err)
	assert.Equal(t, "updated_username", updatedUser.Username)
	assert.Equal(t, entity.RoleAdmin, updatedUser.Role)

	// 存在しないIDでの更新
	nonExistentUser := &entity.User{
		ID:           "nonexistent_id",
		Username:     "nonexistent",
		PasswordHash: "hash",
		Role:         entity.RoleTester,
		UpdatedAt:    time.Now(),
	}
	err = repo.Update(ctx, nonExistentUser)
	assert.Error(t, err)
	assert.True(t, customerrors.IsNotFoundError(err))

	// クリーンアップ
	cleanupUser(t, db, id)
}

func TestPostgresUserRepository_UpdateLastLogin(t *testing.T) {
	db, err := setupTestDB()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	// テスト用ユーザーを挿入（初期状態ではlast_login_atはNULL）
	id := insertTestUser(t, db, "lastlogin_user", entity.RoleTester)

	// 最終ログイン日時を更新
	err = repo.UpdateLastLogin(ctx, id)
	assert.NoError(t, err)

	// 更新後のユーザーを取得して確認
	updatedUser, err := repo.FindByID(ctx, id)
	assert.NoError(t, err)
	assert.NotNil(t, updatedUser.LastLoginAt) // 最終ログイン日時が設定されていることを確認

	// 存在しないIDでの更新
	err = repo.UpdateLastLogin(ctx, "nonexistent_id")
	assert.Error(t, err)
	assert.True(t, customerrors.IsNotFoundError(err))

	// クリーンアップ
	cleanupUser(t, db, id)
}

func TestPostgresUserRepository_Delete(t *testing.T) {
	db, err := setupTestDB()
	require.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	// テスト用ユーザーを挿入
	id := insertTestUser(t, db, "delete_user", entity.RoleTester)

	// ユーザーを削除
	err = repo.Delete(ctx, id)
	assert.NoError(t, err)

	// 削除されたことを確認
	_, err = repo.FindByID(ctx, id)
	assert.Error(t, err)
	assert.True(t, customerrors.IsNotFoundError(err))

	// 存在しないIDでの削除
	err = repo.Delete(ctx, "nonexistent_id")
	assert.Error(t, err)
	assert.True(t, customerrors.IsNotFoundError(err))
}

// テスト用ヘルパー関数

func setupTestDB() (*sql.DB, error) {
	// テスト用の接続文字列（環境変数から取得することが望ましい）
	dsn := "postgresql://test_user:test_pass@localhost:5433/test_db?sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	// 接続確認
	if err = db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func insertTestUser(t *testing.T, db *sql.DB, username string, role entity.UserRole) string {
	// ユニークなIDを生成
	id := "test_user_" + username

	// ユーザーをデータベースに直接挿入
	query := `
		INSERT INTO users (
			id, username, password_hash, role, 
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6)
	`

	now := time.Now()
	_, err := db.Exec(
		query,
		id,
		username,
		"test_password_hash",
		role,
		now,
		now,
	)
	require.NoError(t, err)

	return id
}

func cleanupUser(t *testing.T, db *sql.DB, id string) {
	_, err := db.Exec("DELETE FROM users WHERE id = $1", id)
	require.NoError(t, err)
}
