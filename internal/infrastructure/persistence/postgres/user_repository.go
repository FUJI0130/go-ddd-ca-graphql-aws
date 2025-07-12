package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/FUJI0130/go-ddd-ca/internal/domain/entity"
	"github.com/FUJI0130/go-ddd-ca/internal/domain/repository"
	"github.com/FUJI0130/go-ddd-ca/support/customerrors"
)

// PostgresUserRepository はPostgreSQLを使用したユーザーリポジトリの実装
type PostgresUserRepository struct {
	db *sql.DB
}

// NewUserRepository は新しいUserRepositoryインスタンスを作成する
func NewUserRepository(db *sql.DB) repository.UserRepository {
	return &PostgresUserRepository{
		db: db,
	}
}

// Create は新しいユーザーをデータベースに作成する
func (r *PostgresUserRepository) Create(ctx context.Context, user *entity.User) error {
	query := `
		INSERT INTO users (
			id, username, password_hash, role, 
			created_at, updated_at, last_login_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		user.ID,
		user.Username,
		user.PasswordHash,
		user.Role,
		user.CreatedAt,
		user.UpdatedAt,
		user.LastLoginAt,
	)

	if err != nil {
		return customerrors.ConvertDBError(err, "create", "User", user.ID)
	}

	return nil
}

// FindByID は指定されたIDのユーザーを取得する
func (r *PostgresUserRepository) FindByID(ctx context.Context, id string) (*entity.User, error) {
	query := `
		SELECT 
			id, username, password_hash, role,
			created_at, updated_at, last_login_at
		FROM users
		WHERE id = $1
	`
	user := &entity.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLoginAt,
	)

	if err != nil {
		return nil, customerrors.ConvertDBError(err, "find", "User", id)
	}

	return user, nil
}

// FindByUsername は指定されたユーザー名のユーザーを取得する
func (r *PostgresUserRepository) FindByUsername(ctx context.Context, username string) (*entity.User, error) {
	query := `
		SELECT 
			id, username, password_hash, role,
			created_at, updated_at, last_login_at
		FROM users
		WHERE username = $1
	`
	user := &entity.User{}
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLoginAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, customerrors.NotFound("User", "").WithContext(customerrors.Context{
				"username": username,
			})
		}
		return nil, customerrors.DBError("find_by_username", "users", err).WithContext(customerrors.Context{
			"username": username,
		})
	}

	return user, nil
}

// Update はユーザー情報を更新する
func (r *PostgresUserRepository) Update(ctx context.Context, user *entity.User) error {
	query := `
		UPDATE users
		SET 
			username = $1,
			password_hash = $2,
			role = $3,
			updated_at = $4,
			last_login_at = $5
		WHERE id = $6
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		user.Username,
		user.PasswordHash,
		user.Role,
		time.Now(),
		user.LastLoginAt,
		user.ID,
	)

	if err != nil {
		return customerrors.ConvertDBError(err, "update", "User", user.ID)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return customerrors.NewInternalServerError(
			"更新結果の取得に失敗しました",
		).WithContext(customerrors.Context{
			"id":    user.ID,
			"error": err.Error(),
		})
	}

	if rowsAffected == 0 {
		return customerrors.NotFound("User", user.ID)
	}

	return nil
}

// UpdateLastLogin はユーザーの最終ログイン日時を更新する
func (r *PostgresUserRepository) UpdateLastLogin(ctx context.Context, id string) error {
	query := `
		UPDATE users
		SET 
			last_login_at = $1,
			updated_at = $1
		WHERE id = $2
	`

	now := time.Now()
	result, err := r.db.ExecContext(
		ctx,
		query,
		now,
		id,
	)

	if err != nil {
		return customerrors.ConvertDBError(err, "update_last_login", "User", id)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return customerrors.NewInternalServerError(
			"更新結果の取得に失敗しました",
		).WithContext(customerrors.Context{
			"id":    id,
			"error": err.Error(),
		})
	}

	if rowsAffected == 0 {
		return customerrors.NotFound("User", id)
	}

	return nil
}

// Delete は指定されたIDのユーザーを削除する
func (r *PostgresUserRepository) Delete(ctx context.Context, id string) error {
	query := "DELETE FROM users WHERE id = $1"

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return customerrors.ConvertDBError(err, "delete", "User", id)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return customerrors.NewInternalServerError(
			"削除結果の取得に失敗しました",
		).WithContext(customerrors.Context{
			"id":    id,
			"error": err.Error(),
		})
	}

	if rowsAffected == 0 {
		return customerrors.NotFound("User", id)
	}

	return nil
}

// FindAll は全ユーザーの一覧を取得する
func (r *PostgresUserRepository) FindAll(ctx context.Context) ([]*entity.User, error) {
	query := `
		SELECT 
			id, username, password_hash, role,
			created_at, updated_at, last_login_at
		FROM users
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, customerrors.ConvertDBError(err, "find_all", "User", "")
	}
	defer rows.Close()

	var users []*entity.User
	for rows.Next() {
		user := &entity.User{}
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.PasswordHash,
			&user.Role,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.LastLoginAt,
		)
		if err != nil {
			return nil, customerrors.ConvertDBError(err, "scan", "User", "")
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, customerrors.ConvertDBError(err, "iterate", "User", "")
	}

	return users, nil
}

// internal/infrastructure/persistence/postgres/user_repository.go
// CountByRole は指定されたロールを持つユーザーの数を取得する
func (r *PostgresUserRepository) CountByRole(ctx context.Context, role entity.UserRole) (int, error) {
	query := "SELECT COUNT(*) FROM users WHERE role = $1"
	var count int
	err := r.db.QueryRowContext(ctx, query, role).Scan(&count)
	if err != nil {
		return 0, customerrors.ConvertDBError(err, "count_by_role", "User", "")
	}
	return count, nil
}
