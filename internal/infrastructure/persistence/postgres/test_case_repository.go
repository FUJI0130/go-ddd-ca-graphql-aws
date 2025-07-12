package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/FUJI0130/go-ddd-ca/internal/domain/entity"
	"github.com/FUJI0130/go-ddd-ca/internal/domain/repository"
	"github.com/FUJI0130/go-ddd-ca/pkg/errors"
	"github.com/lib/pq"
)

// PostgresTestCaseRepository はテストケースのPostgreSQL実装
type PostgresTestCaseRepository struct {
	db *sql.DB
}

// NewTestCaseRepository は新しいTestCaseRepositoryを作成します
func NewTestCaseRepository(db *sql.DB) repository.TestCaseRepository {
	return &PostgresTestCaseRepository{
		db: db,
	}
}

// Create は新しいテストケースをデータベースに作成します
func (r *PostgresTestCaseRepository) Create(ctx context.Context, tc *entity.TestCase) error {
	query := `
        INSERT INTO test_cases (
            id, group_id, title, description, status, 
            priority, planned_effort, actual_effort, is_delayed, 
            delay_days, current_editor, is_locked, created_at, updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
    `

	_, err := r.db.ExecContext(
		ctx,
		query,
		tc.ID,
		tc.GroupID,
		tc.Title,
		tc.Description,
		tc.Status,
		tc.Priority,
		tc.PlannedEffort,
		tc.ActualEffort,
		tc.IsDelayed,
		tc.DelayDays,
		tc.CurrentEditor,
		tc.IsLocked,
		tc.CreatedAt,
		tc.UpdatedAt,
	)

	if err != nil {
		// 一意性制約違反の検出
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case PgUniqueViolationCode:
				return errors.NewAlreadyExistsError("TestCase", tc.ID)
			case PgForeignKeyViolationCode:
				return errors.NewDomainValidationError("関連するリソースが存在しません", map[string]string{
					"id":         tc.ID,
					"groupId":    tc.GroupID,
					"constraint": pqErr.Constraint,
				})
			}
		}

		// その他のデータベースエラー
		return errors.NewDatabaseError("create", "test_cases", err).WithDetails(map[string]interface{}{
			"id": tc.ID,
		})
	}

	return nil
}

// FindByID は指定されたIDのテストケースを取得します
func (r *PostgresTestCaseRepository) FindByID(ctx context.Context, id string) (*entity.TestCase, error) {
	query := `
        SELECT 
            id, group_id, title, description, status, 
            priority, planned_effort, actual_effort, is_delayed, 
            delay_days, current_editor, is_locked, created_at, updated_at
        FROM test_cases
        WHERE id = $1
    `
	tc := &entity.TestCase{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&tc.ID,
		&tc.GroupID,
		&tc.Title,
		&tc.Description,
		&tc.Status,
		&tc.Priority,
		&tc.PlannedEffort,
		&tc.ActualEffort,
		&tc.IsDelayed,
		&tc.DelayDays,
		&tc.CurrentEditor,
		&tc.IsLocked,
		&tc.CreatedAt,
		&tc.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError("TestCase", id)
		}
		return nil, errors.NewSystemError(
			"テストケースの検索中にエラーが発生しました",
			err,
		).WithDetails(map[string]interface{}{
			"id": id,
		})
	}

	return tc, nil
}

// Update はテストケースの情報を更新します
func (r *PostgresTestCaseRepository) Update(ctx context.Context, tc *entity.TestCase) error {
	query := `
        UPDATE test_cases
        SET 
            title = $1,
            description = $2,
            status = $3,
            priority = $4,
            planned_effort = $5,
            actual_effort = $6,
            is_delayed = $7,
            delay_days = $8,
            current_editor = $9,
            is_locked = $10,
            updated_at = $11
        WHERE id = $12
    `

	result, err := r.db.ExecContext(
		ctx,
		query,
		tc.Title,
		tc.Description,
		tc.Status,
		tc.Priority,
		tc.PlannedEffort,
		tc.ActualEffort,
		tc.IsDelayed,
		tc.DelayDays,
		tc.CurrentEditor,
		tc.IsLocked,
		time.Now(),
		tc.ID,
	)

	if err != nil {
		return errors.NewDatabaseError("update", "test_cases", err).WithDetails(map[string]interface{}{
			"id": tc.ID,
		})
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewSystemError("更新結果の取得に失敗しました", err).WithDetails(map[string]interface{}{
			"id": tc.ID,
		})
	}

	if rowsAffected == 0 {
		return errors.NewNotFoundError("TestCase", tc.ID)
	}

	return nil
}

// Delete は指定されたIDのテストケースを削除します
func (r *PostgresTestCaseRepository) Delete(ctx context.Context, id string) error {
	query := "DELETE FROM test_cases WHERE id = $1"

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case PgForeignKeyViolationCode:
				return errors.NewDomainConflictError(
					"TestCase",
					id,
					"関連する工数記録が存在するため削除できません",
				)
			}
		}

		return errors.NewDatabaseError("delete", "test_cases", err).WithDetails(map[string]interface{}{
			"id": id,
		})
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewSystemError("削除結果の取得に失敗しました", err).WithDetails(map[string]interface{}{
			"id": id,
		})
	}

	if rowsAffected == 0 {
		return errors.NewNotFoundError("TestCase", id)
	}

	return nil
}

// FindByGroupID は指定されたグループIDに属するテストケース一覧を取得します
func (r *PostgresTestCaseRepository) FindByGroupID(ctx context.Context, groupID string) ([]*entity.TestCase, error) {
	query := `
        SELECT 
            id, group_id, title, description, status, 
            priority, planned_effort, actual_effort, is_delayed, 
            delay_days, current_editor, is_locked, created_at, updated_at
        FROM test_cases
        WHERE group_id = $1
        ORDER BY id ASC
    `

	rows, err := r.db.QueryContext(ctx, query, groupID)
	if err != nil {
		return nil, errors.NewDatabaseError("query", "test_cases", err).WithDetails(map[string]interface{}{
			"groupId": groupID,
		})
	}
	defer rows.Close()

	var cases []*entity.TestCase
	for rows.Next() {
		tc := &entity.TestCase{}
		err := rows.Scan(
			&tc.ID,
			&tc.GroupID,
			&tc.Title,
			&tc.Description,
			&tc.Status,
			&tc.Priority,
			&tc.PlannedEffort,
			&tc.ActualEffort,
			&tc.IsDelayed,
			&tc.DelayDays,
			&tc.CurrentEditor,
			&tc.IsLocked,
			&tc.CreatedAt,
			&tc.UpdatedAt,
		)
		if err != nil {
			return nil, errors.NewSystemError("テストケースデータの読み取りに失敗しました", err)
		}
		cases = append(cases, tc)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.NewDatabaseError("iterate", "test_cases", err)
	}

	return cases, nil
}

// UpdateStatus は指定されたテストケースのステータスを更新します
func (r *PostgresTestCaseRepository) UpdateStatus(ctx context.Context, id string, status entity.TestStatus) error {
	query := `
        UPDATE test_cases
        SET 
            status = $1,
            updated_at = $2
        WHERE id = $3
    `

	result, err := r.db.ExecContext(
		ctx,
		query,
		status,
		time.Now(),
		id,
	)

	if err != nil {
		return errors.NewDatabaseError("update_status", "test_cases", err).WithDetails(map[string]interface{}{
			"id":     id,
			"status": string(status),
		})
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewSystemError("ステータス更新結果の取得に失敗しました", err).WithDetails(map[string]interface{}{
			"id":     id,
			"status": string(status),
		})
	}

	if rowsAffected == 0 {
		return errors.NewNotFoundError("TestCase", id)
	}

	return nil
}

// AddEffort は指定されたテストケースに工数を追加します
func (r *PostgresTestCaseRepository) AddEffort(ctx context.Context, id string, effort float64) error {
	query := `
        UPDATE test_cases
        SET 
            actual_effort = actual_effort + $1,
            updated_at = $2
        WHERE id = $3
    `

	result, err := r.db.ExecContext(
		ctx,
		query,
		effort,
		time.Now(),
		id,
	)

	if err != nil {
		return errors.NewDatabaseError("add_effort", "test_cases", err).WithDetails(map[string]interface{}{
			"id":     id,
			"effort": effort,
		})
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewSystemError("工数追加結果の取得に失敗しました", err).WithDetails(map[string]interface{}{
			"id":     id,
			"effort": effort,
		})
	}

	if rowsAffected == 0 {
		return errors.NewNotFoundError("TestCase", id)
	}

	return nil
}

// FindByStatus は指定されたステータスのテストケース一覧を取得します
func (r *PostgresTestCaseRepository) FindByStatus(ctx context.Context, status entity.TestStatus) ([]*entity.TestCase, error) {
	query := `
        SELECT 
            id, group_id, title, description, status, 
            priority, planned_effort, actual_effort, is_delayed, 
            delay_days, current_editor, is_locked, created_at, updated_at
        FROM test_cases
        WHERE status = $1
        ORDER BY updated_at DESC
    `

	rows, err := r.db.QueryContext(ctx, query, status)
	if err != nil {
		return nil, errors.NewDatabaseError("query", "test_cases", err).WithDetails(map[string]interface{}{
			"status": string(status),
		})
	}
	defer rows.Close()

	var cases []*entity.TestCase
	for rows.Next() {
		tc := &entity.TestCase{}
		err := rows.Scan(
			&tc.ID,
			&tc.GroupID,
			&tc.Title,
			&tc.Description,
			&tc.Status,
			&tc.Priority,
			&tc.PlannedEffort,
			&tc.ActualEffort,
			&tc.IsDelayed,
			&tc.DelayDays,
			&tc.CurrentEditor,
			&tc.IsLocked,
			&tc.CreatedAt,
			&tc.UpdatedAt,
		)
		if err != nil {
			return nil, errors.NewSystemError("テストケースデータの読み取りに失敗しました", err)
		}
		cases = append(cases, tc)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.NewDatabaseError("iterate", "test_cases", err)
	}

	return cases, nil
}
