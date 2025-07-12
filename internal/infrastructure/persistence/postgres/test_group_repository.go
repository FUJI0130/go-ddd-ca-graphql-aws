package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/FUJI0130/go-ddd-ca/internal/domain/entity"
	"github.com/FUJI0130/go-ddd-ca/internal/domain/repository"
	"github.com/FUJI0130/go-ddd-ca/internal/domain/valueobject"
	"github.com/FUJI0130/go-ddd-ca/pkg/errors"
	"github.com/lib/pq"
)

// PostgresTestGroupRepository はテストグループのPostgreSQL実装
type PostgresTestGroupRepository struct {
	db *sql.DB
}

// NewTestGroupRepository は新しいTestGroupRepositoryを作成します
func NewTestGroupRepository(db *sql.DB) repository.TestGroupRepository {
	return &PostgresTestGroupRepository{
		db: db,
	}
}

// Create は新しいテストグループをデータベースに作成します
func (r *PostgresTestGroupRepository) Create(ctx context.Context, group *entity.TestGroup) error {
	query := `
        INSERT INTO test_groups (
            id, suite_id, name, description, display_order, 
            status, created_at, updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `

	_, err := r.db.ExecContext(
		ctx,
		query,
		group.ID,
		group.SuiteID,
		group.Name,
		group.Description,
		group.DisplayOrder,
		group.Status,
		group.CreatedAt,
		group.UpdatedAt,
	)

	if err != nil {
		// 一意性制約違反の検出
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case PgUniqueViolationCode:
				return errors.NewAlreadyExistsError("TestGroup", group.ID)
			case PgForeignKeyViolationCode:
				return errors.NewDomainValidationError("関連するリソースが存在しません", map[string]string{
					"id":         group.ID,
					"suiteId":    group.SuiteID,
					"constraint": pqErr.Constraint,
				})
			}
		}

		// その他のデータベースエラー
		return errors.NewDatabaseError("create", "test_groups", err).WithDetails(map[string]interface{}{
			"id": group.ID,
		})
	}

	return nil
}

// FindByID は指定されたIDのテストグループを取得します
func (r *PostgresTestGroupRepository) FindByID(ctx context.Context, id string) (*entity.TestGroup, error) {
	query := `
        SELECT 
            id, suite_id, name, description, display_order,
            status, created_at, updated_at
        FROM test_groups
        WHERE id = $1
    `
	group := &entity.TestGroup{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&group.ID,
		&group.SuiteID,
		&group.Name,
		&group.Description,
		&group.DisplayOrder,
		&group.Status,
		&group.CreatedAt,
		&group.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError("TestGroup", id)
		}
		return nil, errors.NewSystemError(
			"テストグループの検索中にエラーが発生しました",
			err,
		).WithDetails(map[string]interface{}{
			"id": id,
		})
	}

	return group, nil
}

// Update はテストグループの情報を更新します
func (r *PostgresTestGroupRepository) Update(ctx context.Context, group *entity.TestGroup) error {
	query := `
        UPDATE test_groups
        SET 
            name = $1,
            description = $2,
            display_order = $3,
            status = $4,
            updated_at = $5
        WHERE id = $6
    `

	result, err := r.db.ExecContext(
		ctx,
		query,
		group.Name,
		group.Description,
		group.DisplayOrder,
		group.Status,
		time.Now(),
		group.ID,
	)

	if err != nil {
		return errors.NewDatabaseError("update", "test_groups", err).WithDetails(map[string]interface{}{
			"id": group.ID,
		})
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewSystemError("更新結果の取得に失敗しました", err).WithDetails(map[string]interface{}{
			"id": group.ID,
		})
	}

	if rowsAffected == 0 {
		return errors.NewNotFoundError("TestGroup", group.ID)
	}

	return nil
}

// Delete は指定されたIDのテストグループを削除します
func (r *PostgresTestGroupRepository) Delete(ctx context.Context, id string) error {
	query := "DELETE FROM test_groups WHERE id = $1"

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case PgForeignKeyViolationCode:
				return errors.NewDomainConflictError(
					"TestGroup",
					id,
					"関連するテストケースが存在するため削除できません",
				)
			}
		}

		return errors.NewDatabaseError("delete", "test_groups", err).WithDetails(map[string]interface{}{
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
		return errors.NewNotFoundError("TestGroup", id)
	}

	return nil
}

// FindBySuiteID は指定されたスイートIDに属するテストグループ一覧を取得します
func (r *PostgresTestGroupRepository) FindBySuiteID(ctx context.Context, suiteID string) ([]*entity.TestGroup, error) {
	query := `
        SELECT 
            id, suite_id, name, description, display_order,
            status, created_at, updated_at
        FROM test_groups
        WHERE suite_id = $1
        ORDER BY display_order ASC
    `

	rows, err := r.db.QueryContext(ctx, query, suiteID)
	if err != nil {
		return nil, errors.NewDatabaseError("query", "test_groups", err).WithDetails(map[string]interface{}{
			"suiteId": suiteID,
		})
	}
	defer rows.Close()

	var groups []*entity.TestGroup
	for rows.Next() {
		group := &entity.TestGroup{}
		err := rows.Scan(
			&group.ID,
			&group.SuiteID,
			&group.Name,
			&group.Description,
			&group.DisplayOrder,
			&group.Status,
			&group.CreatedAt,
			&group.UpdatedAt,
		)
		if err != nil {
			return nil, errors.NewSystemError("テストグループデータの読み取りに失敗しました", err)
		}
		groups = append(groups, group)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.NewDatabaseError("iterate", "test_groups", err)
	}

	return groups, nil
}

// UpdateStatus は指定されたテストグループのステータスを更新します
func (r *PostgresTestGroupRepository) UpdateStatus(ctx context.Context, id string, status valueobject.SuiteStatus) error {
	query := `
        UPDATE test_groups
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
		return errors.NewDatabaseError("update_status", "test_groups", err).WithDetails(map[string]interface{}{
			"id":     id,
			"status": status.String(),
		})
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewSystemError("ステータス更新結果の取得に失敗しました", err).WithDetails(map[string]interface{}{
			"id":     id,
			"status": status.String(),
		})
	}

	if rowsAffected == 0 {
		return errors.NewNotFoundError("TestGroup", id)
	}

	return nil
}

// UpdateDisplayOrder は指定されたテストグループの表示順序を更新します
func (r *PostgresTestGroupRepository) UpdateDisplayOrder(ctx context.Context, id string, displayOrder int) error {
	query := `
        UPDATE test_groups
        SET 
            display_order = $1,
            updated_at = $2
        WHERE id = $3
    `

	result, err := r.db.ExecContext(
		ctx,
		query,
		displayOrder,
		time.Now(),
		id,
	)

	if err != nil {
		return errors.NewDatabaseError("update_display_order", "test_groups", err).WithDetails(map[string]interface{}{
			"id":           id,
			"displayOrder": displayOrder,
		})
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewSystemError("表示順序更新結果の取得に失敗しました", err).WithDetails(map[string]interface{}{
			"id":           id,
			"displayOrder": displayOrder,
		})
	}

	if rowsAffected == 0 {
		return errors.NewNotFoundError("TestGroup", id)
	}

	return nil
}
