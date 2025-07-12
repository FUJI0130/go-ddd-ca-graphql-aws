package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/FUJI0130/go-ddd-ca/internal/domain/entity"
	"github.com/FUJI0130/go-ddd-ca/internal/domain/repository"
	"github.com/FUJI0130/go-ddd-ca/internal/domain/valueobject"
	"github.com/FUJI0130/go-ddd-ca/internal/usecase/dto"
	"github.com/FUJI0130/go-ddd-ca/support/customerrors"
)

type PostgresTestSuiteRepository struct {
	db *sql.DB
}

func NewTestSuiteRepository(db *sql.DB) repository.TestSuiteRepository {
	return &PostgresTestSuiteRepository{
		db: db,
	}
}

// Create は新しいテストスイートをデータベースに作成します
func (r *PostgresTestSuiteRepository) Create(ctx context.Context, suite *entity.TestSuite) error {
	query := `
        INSERT INTO test_suites (
            id, name, description, status, 
            estimated_start_date, estimated_end_date,
            require_effort_comment, created_at, updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
    `

	_, err := r.db.ExecContext(
		ctx,
		query,
		suite.ID,
		suite.Name,
		suite.Description,
		suite.Status,
		suite.EstimatedStartDate,
		suite.EstimatedEndDate,
		suite.RequireEffortComment,
		suite.CreatedAt,
		suite.UpdatedAt,
	)

	if err != nil {
		return customerrors.ConvertDBError(err, "create", "TestSuite", suite.ID)
	}

	return nil
}

// FindByID は指定されたIDのテストスイートを取得します
func (r *PostgresTestSuiteRepository) FindByID(ctx context.Context, id string) (*entity.TestSuite, error) {
	query := `
        SELECT 
            id, name, description, status,
            estimated_start_date, estimated_end_date,
            require_effort_comment, created_at, updated_at
        FROM test_suites
        WHERE id = $1
    `
	suite := &entity.TestSuite{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&suite.ID,
		&suite.Name,
		&suite.Description,
		&suite.Status,
		&suite.EstimatedStartDate,
		&suite.EstimatedEndDate,
		&suite.RequireEffortComment,
		&suite.CreatedAt,
		&suite.UpdatedAt,
	)

	if err != nil {
		return nil, customerrors.ConvertDBError(err, "find", "TestSuite", id)
	}

	return suite, nil
}

// Update はテストスイートの情報を更新します
func (r *PostgresTestSuiteRepository) Update(ctx context.Context, suite *entity.TestSuite) error {
	query := `
        UPDATE test_suites
        SET 
            name = $1,
            description = $2,
            status = $3,
            estimated_start_date = $4,
            estimated_end_date = $5,
            require_effort_comment = $6,
            updated_at = $7
        WHERE id = $8
    `

	result, err := r.db.ExecContext(
		ctx,
		query,
		suite.Name,
		suite.Description,
		suite.Status,
		suite.EstimatedStartDate,
		suite.EstimatedEndDate,
		suite.RequireEffortComment,
		time.Now(),
		suite.ID,
	)

	if err != nil {
		return customerrors.ConvertDBError(err, "update", "TestSuite", suite.ID)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return customerrors.NewInternalServerError(
			"更新結果の取得に失敗しました",
		).WithContext(customerrors.Context{
			"id":    suite.ID,
			"error": err.Error(),
		})
	}

	if rowsAffected == 0 {
		return customerrors.NotFound("TestSuite", suite.ID)
	}

	return nil
}

// Delete は指定されたIDのテストスイートを削除します
func (r *PostgresTestSuiteRepository) Delete(ctx context.Context, id string) error {
	query := "DELETE FROM test_suites WHERE id = $1"

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return customerrors.ConvertDBError(err, "delete", "TestSuite", id)
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
		return customerrors.NotFound("TestSuite", id)
	}

	return nil
}

// UpdateStatus はテストスイートの状態を更新します
func (r *PostgresTestSuiteRepository) UpdateStatus(ctx context.Context, id string, status valueobject.SuiteStatus) error {
	query := `
        UPDATE test_suites
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
		return customerrors.ConvertDBError(err, "update_status", "TestSuite", id)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return customerrors.NewInternalServerError(
			"ステータス更新結果の取得に失敗しました",
		).WithContext(customerrors.Context{
			"id":     id,
			"status": status.String(),
			"error":  err.Error(),
		})
	}

	if rowsAffected == 0 {
		return customerrors.NotFound("TestSuite", id)
	}

	return nil
}

// FindByStatus は指定された状態のテストスイート一覧を取得します
func (r *PostgresTestSuiteRepository) FindByStatus(ctx context.Context, status valueobject.SuiteStatus) ([]*entity.TestSuite, error) {
	query := `
        SELECT 
            id, name, description, status,
            estimated_start_date, estimated_end_date,
            require_effort_comment, created_at, updated_at
        FROM test_suites
        WHERE status = $1
        ORDER BY created_at DESC
    `
	rows, err := r.db.QueryContext(ctx, query, status)
	if err != nil {
		return nil, customerrors.DBError("query", "test_suites", err).WithContext(customerrors.Context{
			"status": status.String(),
		})
	}
	defer rows.Close()

	var suites []*entity.TestSuite
	for rows.Next() {
		suite := &entity.TestSuite{}
		err := rows.Scan(
			&suite.ID,
			&suite.Name,
			&suite.Description,
			&suite.Status,
			&suite.EstimatedStartDate,
			&suite.EstimatedEndDate,
			&suite.RequireEffortComment,
			&suite.CreatedAt,
			&suite.UpdatedAt,
		)
		if err != nil {
			return nil, customerrors.NewInternalServerError(
				"テストスイートデータの読み取りに失敗しました",
			).WithContext(customerrors.Context{
				"error":  err.Error(),
				"status": status.String(),
			})
		}
		suites = append(suites, suite)
	}

	if err = rows.Err(); err != nil {
		return nil, customerrors.DBError("iterate", "test_suites", err).WithContext(customerrors.Context{
			"status": status.String(),
		})
	}

	return suites, nil
}

// FindWithFilters はフィルター条件に基づいてテストスイート一覧を取得します
func (r *PostgresTestSuiteRepository) FindWithFilters(ctx context.Context, params *dto.TestSuiteQueryParamDTO) ([]*entity.TestSuite, int, error) {
	baseQuery := `
        SELECT 
            id, name, description, status,
            estimated_start_date, estimated_end_date,
            require_effort_comment, created_at, updated_at
        FROM test_suites
        WHERE 1=1
    `
	countQuery := "SELECT COUNT(*) FROM test_suites WHERE 1=1"

	var queryParams []interface{}
	paramCount := 1

	whereClause := ""
	if params.Status != nil {
		status, err := valueobject.NewSuiteStatus(*params.Status)
		if err != nil {
			return nil, 0, customerrors.Validation(
				"無効なステータス値です",
				map[string]string{
					"status": *params.Status,
					"error":  err.Error(),
				},
			)
		}

		whereClause += fmt.Sprintf(" AND status = $%d", paramCount)
		queryParams = append(queryParams, status)
		paramCount++
	}
	if params.StartDate != nil {
		whereClause += fmt.Sprintf(" AND estimated_start_date >= $%d", paramCount)
		queryParams = append(queryParams, params.StartDate)
		paramCount++
	}
	if params.EndDate != nil {
		whereClause += fmt.Sprintf(" AND estimated_end_date <= $%d", paramCount)
		queryParams = append(queryParams, params.EndDate)
		paramCount++
	}

	var total int
	err := r.db.QueryRowContext(ctx, countQuery+whereClause, queryParams...).Scan(&total)
	if err != nil {
		return nil, 0, customerrors.DBError("count", "test_suites", err).WithContext(customerrors.Context{
			"filters": fmt.Sprintf("%+v", params),
		})
	}

	limit := 10 // デフォルト値
	offset := 0
	if params.PageSize != nil {
		limit = *params.PageSize
	}
	if params.Page != nil {
		offset = (*params.Page - 1) * limit
	}

	query := baseQuery + whereClause + " ORDER BY created_at DESC LIMIT $" +
		fmt.Sprintf("%d OFFSET $%d", paramCount, paramCount+1)
	queryParams = append(queryParams, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, queryParams...)
	if err != nil {
		return nil, 0, customerrors.DBError("query", "test_suites", err).WithContext(customerrors.Context{
			"filters": fmt.Sprintf("%+v", params),
			"limit":   limit,
			"offset":  offset,
		})
	}
	defer rows.Close()

	var suites []*entity.TestSuite
	for rows.Next() {
		suite := &entity.TestSuite{}
		err := rows.Scan(
			&suite.ID,
			&suite.Name,
			&suite.Description,
			&suite.Status,
			&suite.EstimatedStartDate,
			&suite.EstimatedEndDate,
			&suite.RequireEffortComment,
			&suite.CreatedAt,
			&suite.UpdatedAt,
		)
		if err != nil {
			return nil, 0, customerrors.NewInternalServerError(
				"テストスイートデータの読み取りに失敗しました",
			).WithContext(customerrors.Context{
				"error":   err.Error(),
				"filters": fmt.Sprintf("%+v", params),
			})
		}
		suites = append(suites, suite)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, customerrors.DBError("iterate", "test_suites", err).WithContext(customerrors.Context{
			"filters": fmt.Sprintf("%+v", params),
		})
	}

	return suites, total, nil
}
