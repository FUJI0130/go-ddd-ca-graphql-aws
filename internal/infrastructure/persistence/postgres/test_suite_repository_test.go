// test/integration/postgres/test_suite_repository_test.go
// 注意　テスト実行時にはテスト用のコンテナを立ち上げること make test-integration
package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/FUJI0130/go-ddd-ca/internal/domain/entity"
	"github.com/FUJI0130/go-ddd-ca/internal/domain/valueobject"
	"github.com/FUJI0130/go-ddd-ca/internal/infrastructure/persistence/postgres"
	"github.com/FUJI0130/go-ddd-ca/internal/usecase/dto"
	"github.com/FUJI0130/go-ddd-ca/support/customerrors"
)

func TestPostgresTestSuiteRepository_Create(t *testing.T) {
	tests := []struct {
		name    string
		input   *entity.TestSuite
		wantErr bool
		setup   func(t *testing.T, db *sql.DB) // テストデータセットアップ
		verify  func(t *testing.T, db *sql.DB) // 結果検証
	}{
		{
			name: "正常系：新規テストスイート作成",
			input: &entity.TestSuite{
				ID:                   "TS001-202412",
				Name:                 "商品管理システムテストスイート",
				Description:          "商品管理システムの結合テスト",
				Status:               valueobject.SuiteStatusPreparation,
				EstimatedStartDate:   time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
				EstimatedEndDate:     time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
				RequireEffortComment: true,
			},
			wantErr: false,
			setup: func(t *testing.T, db *sql.DB) {
				// 必要に応じてテストデータをセットアップ
			},
			verify: func(t *testing.T, db *sql.DB) {
				var count int
				err := db.QueryRow("SELECT COUNT(*) FROM test_suites WHERE id = $1", "TS001-202412").Scan(&count)
				if err != nil {
					t.Errorf("verify failed: %v", err)
				}
				if count != 1 {
					t.Errorf("expected 1 record, got %d", count)
				}
			},
		},
		{
			name: "異常系：重複するID",
			input: &entity.TestSuite{
				ID: "TS001-202412", // 既に存在するID
				// ... その他のフィールド
			},
			wantErr: true,
			setup: func(t *testing.T, db *sql.DB) {
				// 重複するレコードを事前に挿入
				_, err := db.Exec(`
                    INSERT INTO test_suites (id, name, status)
                    VALUES ($1, $2, $3)
                `, "TS001-202412", "既存テストスイート", "準備中")
				if err != nil {
					t.Fatalf("setup failed: %v", err)
				}
			},
			verify: func(t *testing.T, db *sql.DB) {
				var count int
				err := db.QueryRow("SELECT COUNT(*) FROM test_suites WHERE id = $1", "TS001-202412").Scan(&count)
				if err != nil {
					t.Errorf("verify failed: %v", err)
				}
				if count != 1 { // 重複せず1件のまま
					t.Errorf("expected 1 record, got %d", count)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テスト用DBセットアップ
			db, cleanup := setupTestDB(t)
			defer cleanup()

			if tt.setup != nil {
				tt.setup(t, db)
			}

			repo := postgres.NewTestSuiteRepository(db)
			err := repo.Create(context.Background(), tt.input)

			// エラー検証
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}

			// 結果検証
			if tt.verify != nil {
				tt.verify(t, db)
			}
		})
	}
}

func TestPostgresTestSuiteRepository_FindByStatus(t *testing.T) {
	tests := []struct {
		name      string
		status    valueobject.SuiteStatus
		wantCount int
		setup     func(t *testing.T, db *sql.DB)
		verify    func(t *testing.T, results []*entity.TestSuite)
	}{
		{
			name:      "準備中のテストスイートを全て取得",
			status:    valueobject.SuiteStatusPreparation,
			wantCount: 2,
			setup: func(t *testing.T, db *sql.DB) {
				testData := []struct {
					id                 string
					status             valueobject.SuiteStatus
					estimatedStartDate time.Time
					estimatedEndDate   time.Time
				}{
					{"TS001-202412", valueobject.SuiteStatusPreparation, time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC), time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)},
					{"TS002-202412", valueobject.SuiteStatusPreparation, time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC), time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)},
					{"TS003-202412", valueobject.SuiteStatusInProgress, time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC), time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)},
				}

				for _, td := range testData {
					_, err := db.Exec(`
                        INSERT INTO test_suites (id, name, description, status, estimated_start_date, estimated_end_date)
                        VALUES ($1, $2, $3, $4, $5, $6)
                    `, td.id, "テストスイート"+td.id, "テストスイートの説明", td.status, td.estimatedStartDate, td.estimatedEndDate)
					if err != nil {
						t.Fatalf("setup failed: %v", err)
					}
				}
			},
			verify: func(t *testing.T, results []*entity.TestSuite) {
				if len(results) != 2 {
					t.Errorf("expected 2 results, got %d", len(results))
				}
				for _, r := range results {
					if r.Status != valueobject.SuiteStatusPreparation {
						t.Errorf("expected status %v, got %v", valueobject.SuiteStatusPreparation, r.Status)
					}
				}
			},
		},
		// 他のテストケース
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, cleanup := setupTestDB(t)
			defer cleanup()

			if tt.setup != nil {
				tt.setup(t, db)
			}

			repo := postgres.NewTestSuiteRepository(db)
			results, err := repo.FindByStatus(context.Background(), tt.status)

			if err != nil {
				t.Errorf("FindByStatus() error = %v", err)
				return
			}

			if tt.verify != nil {
				tt.verify(t, results)
			}
		})
	}
}

func TestPostgresTestSuiteRepository_FindByID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		setup   func(t *testing.T, db *sql.DB)
		wantErr bool
		errType error
		verify  func(t *testing.T, suite *entity.TestSuite, err error)
	}{
		{
			name: "正常系：存在するIDのテストスイート取得",
			id:   "TS001-202412",
			setup: func(t *testing.T, db *sql.DB) {
				// テスト用データ作成
				_, err := db.Exec(`
					INSERT INTO test_suites (
						id, name, description, status, 
						estimated_start_date, estimated_end_date,
						require_effort_comment, created_at, updated_at
					) VALUES (
						$1, $2, $3, $4, $5, $6, $7, $8, $9
					)
				`,
					"TS001-202412",
					"商品管理システムテスト",
					"商品管理機能のテスト",
					valueobject.SuiteStatusPreparation,
					time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
					true,
					time.Now(),
					time.Now(),
				)
				if err != nil {
					t.Fatalf("setup failed: %v", err)
				}
			},
			wantErr: false,
			verify: func(t *testing.T, suite *entity.TestSuite, err error) {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
					return
				}
				if suite == nil {
					t.Errorf("expected suite, got nil")
					return
				}
				if suite.ID != "TS001-202412" {
					t.Errorf("expected ID TS001-202412, got %s", suite.ID)
				}
				if suite.Name != "商品管理システムテスト" {
					t.Errorf("expected Name 商品管理システムテスト, got %s", suite.Name)
				}
				if suite.Status != valueobject.SuiteStatusPreparation {
					t.Errorf("expected Status %v, got %v", valueobject.SuiteStatusPreparation, suite.Status)
				}
			},
		},
		{
			name: "異常系：存在しないIDのテストスイート取得",
			id:   "TS999-202412",
			setup: func(t *testing.T, db *sql.DB) {
				// 特に何もセットアップしない - 存在しないIDをテスト
			},
			wantErr: true,
			errType: &customerrors.NotFoundError{}, // 変更
			verify: func(t *testing.T, suite *entity.TestSuite, err error) {
				if err == nil {
					t.Errorf("expected error, got nil")
					return
				}

				// NotFoundError型の検証
				if !customerrors.IsNotFoundError(err) {
					t.Errorf("expected NotFoundError, got %T", err)
					return
				}

				// 直接型アサーションでエラー情報を取得
				var notFoundErr *customerrors.NotFoundError
				if !errors.As(err, &notFoundErr) {
					t.Errorf("failed to cast to NotFoundError")
					return
				}

				// ステータスコードの検証
				if notFoundErr.StatusCode() != customerrors.StatusCodeNotFound {
					t.Errorf("expected status code %d, got %d", customerrors.StatusCodeNotFound, notFoundErr.StatusCode())
				}
				// エラーメッセージに該当IDが含まれていることを確認
				if msg := notFoundErr.Error(); !strings.Contains(msg, "TS999-202412") {
					t.Errorf("expected error message to contain ID TS999-202412, got: %s", msg)
				}

				// スイートがnilであることの検証
				if suite != nil {
					t.Errorf("expected nil suite, got %+v", suite)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テスト用DBセットアップ
			db, cleanup := setupTestDB(t)
			defer func() {
				// セットアップ内でDBを閉じる場合があるため、panic回避のためにリカバリーを実装
				defer func() {
					if r := recover(); r != nil {
						t.Logf("Recovered from panic in cleanup: %v", r)
					}
				}()
				cleanup()
			}()

			if tt.setup != nil {
				tt.setup(t, db)
			}

			repo := postgres.NewTestSuiteRepository(db)
			suite, err := repo.FindByID(context.Background(), tt.id)

			// エラー検証
			if (err != nil) != tt.wantErr {
				t.Errorf("FindByID() error = %v, wantErr %v", err, tt.wantErr)
			}

			// 期待するエラー型の検証
			if tt.errType != nil && err != nil {
				if errType := tt.errType; errType != nil {
					expectedType := reflect.TypeOf(errType)
					actualType := reflect.TypeOf(err)
					if expectedType != actualType {
						t.Errorf("FindByID() error type = %v, want %v", actualType, expectedType)
					}
				}
			}

			// 結果検証
			if tt.verify != nil {
				tt.verify(t, suite, err)
			}
		})
	}
}

func TestPostgresTestSuiteRepository_Update(t *testing.T) {
	tests := []struct {
		name    string
		input   *entity.TestSuite
		setup   func(t *testing.T, db *sql.DB)
		wantErr bool
		verify  func(t *testing.T, err error, db *sql.DB)
	}{
		{
			name: "正常系：テストスイート更新",
			input: &entity.TestSuite{
				ID:                   "TS001-202412",
				Name:                 "更新後の商品管理システムテスト",
				Description:          "更新後の説明文",
				Status:               valueobject.SuiteStatusInProgress,
				EstimatedStartDate:   time.Date(2024, 12, 10, 0, 0, 0, 0, time.UTC),
				EstimatedEndDate:     time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
				RequireEffortComment: false,
				UpdatedAt:            time.Now(),
			},
			setup: func(t *testing.T, db *sql.DB) {
				// 更新前のデータを作成
				_, err := db.Exec(`
					INSERT INTO test_suites (
						id, name, description, status, 
						estimated_start_date, estimated_end_date,
						require_effort_comment, created_at, updated_at
					) VALUES (
						$1, $2, $3, $4, $5, $6, $7, $8, $9
					)
				`,
					"TS001-202412",
					"元の商品管理システムテスト",
					"元の説明文",
					valueobject.SuiteStatusPreparation,
					time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
					true,
					time.Now(),
					time.Now(),
				)
				if err != nil {
					t.Fatalf("setup failed: %v", err)
				}
			},
			wantErr: false,
			verify: func(t *testing.T, err error, db *sql.DB) {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
					return
				}

				// 更新後のデータを確認
				var name, description string
				var status valueobject.SuiteStatus
				var requireEffortComment bool
				var startDate, endDate time.Time

				err = db.QueryRow(`
					SELECT name, description, status, 
						   estimated_start_date, estimated_end_date,
						   require_effort_comment
					FROM test_suites 
					WHERE id = $1
				`, "TS001-202412").Scan(
					&name, &description, &status,
					&startDate, &endDate, &requireEffortComment,
				)

				if err != nil {
					t.Errorf("failed to query updated suite: %v", err)
					return
				}

				// 更新されたフィールドの確認
				if name != "更新後の商品管理システムテスト" {
					t.Errorf("name not updated, got: %s", name)
				}
				if description != "更新後の説明文" {
					t.Errorf("description not updated, got: %s", description)
				}
				if status != valueobject.SuiteStatusInProgress {
					t.Errorf("status not updated, got: %v", status)
				}
				if requireEffortComment != false {
					t.Errorf("requireEffortComment not updated, got: %v", requireEffortComment)
				}
			},
		},
		{
			name: "異常系：存在しないIDの更新",
			input: &entity.TestSuite{
				ID:                   "TS999-202412", // 存在しないID
				Name:                 "存在しないテストスイート",
				Description:          "説明文",
				Status:               valueobject.SuiteStatusInProgress,
				EstimatedStartDate:   time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
				EstimatedEndDate:     time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
				RequireEffortComment: true,
				UpdatedAt:            time.Now(),
			},
			setup: func(t *testing.T, db *sql.DB) {
				// 何もセットアップしない（存在しないIDをテスト）
			},
			wantErr: true,
			verify: func(t *testing.T, err error, db *sql.DB) {
				if err == nil {
					t.Errorf("expected error, got nil")
					return
				}

				// NotFoundエラー型の検証 - この部分を修正
				var notFoundErr *customerrors.NotFoundError
				if !errors.As(err, &notFoundErr) {
					t.Errorf("expected NotFoundError, got %T", err)
					return
				}

				// ステータスコードの検証
				if notFoundErr.StatusCode() != customerrors.StatusCodeNotFound {
					t.Errorf("expected status code %d, got %d", customerrors.StatusCodeNotFound, notFoundErr.StatusCode())
				}

				// エラーメッセージに該当IDが含まれていることを確認
				if msg := notFoundErr.Error(); !strings.Contains(msg, "TS999-202412") {
					t.Errorf("expected error message to contain ID TS999-202412, got: %s", msg)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テスト用DBセットアップ
			db, cleanup := setupTestDB(t)
			defer cleanup()

			if tt.setup != nil {
				tt.setup(t, db)
			}

			repo := postgres.NewTestSuiteRepository(db)
			err := repo.Update(context.Background(), tt.input)

			// エラー検証
			if (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
			}

			// 結果検証
			if tt.verify != nil {
				tt.verify(t, err, db)
			}
		})
	}
}

func TestPostgresTestSuiteRepository_UpdateStatus(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		newStatus valueobject.SuiteStatus
		setup     func(t *testing.T, db *sql.DB)
		wantErr   bool
		verify    func(t *testing.T, err error, db *sql.DB)
	}{
		{
			name:      "正常系：テストスイートのステータス更新",
			id:        "TS001-202412",
			newStatus: valueobject.SuiteStatusInProgress,
			setup: func(t *testing.T, db *sql.DB) {
				// テスト用データ作成
				_, err := db.Exec(`
					INSERT INTO test_suites (
						id, name, description, status, 
						estimated_start_date, estimated_end_date,
						require_effort_comment, created_at, updated_at
					) VALUES (
						$1, $2, $3, $4, $5, $6, $7, $8, $9
					)
				`,
					"TS001-202412",
					"商品管理システムテスト",
					"商品管理機能のテスト",
					valueobject.SuiteStatusPreparation,
					time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
					true,
					time.Now(),
					time.Now(),
				)
				if err != nil {
					t.Fatalf("setup failed: %v", err)
				}
			},
			wantErr: false,
			verify: func(t *testing.T, err error, db *sql.DB) {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
					return
				}

				// 更新後のステータスを確認
				var status valueobject.SuiteStatus
				err = db.QueryRow("SELECT status FROM test_suites WHERE id = $1", "TS001-202412").Scan(&status)
				if err != nil {
					t.Errorf("failed to query updated status: %v", err)
					return
				}

				if status != valueobject.SuiteStatusInProgress {
					t.Errorf("status not updated, expected: %v, got: %v", valueobject.SuiteStatusInProgress, status)
				}
			},
		},
		{
			name:      "異常系：存在しないIDのステータス更新",
			id:        "TS999-202412", // 存在しないID
			newStatus: valueobject.SuiteStatusCompleted,
			setup: func(t *testing.T, db *sql.DB) {
				// 何もセットアップしない（存在しないIDをテスト）
			},
			wantErr: true,
			// 異常系：存在しないIDのステータス更新のテストケース内の検証部分
			verify: func(t *testing.T, err error, db *sql.DB) {
				if err == nil {
					t.Errorf("expected error, got nil")
					return
				}

				// NotFoundエラー型の検証 - この部分を修正
				var notFoundErr *customerrors.NotFoundError
				if !errors.As(err, &notFoundErr) {
					t.Errorf("expected NotFoundError, got %T", err)
					return
				}

				// ステータスコードの検証
				if notFoundErr.StatusCode() != customerrors.StatusCodeNotFound {
					t.Errorf("expected status code %d, got %d", customerrors.StatusCodeNotFound, notFoundErr.StatusCode())
				}

				// エラーメッセージに該当IDが含まれていることを確認
				if msg := notFoundErr.Error(); !strings.Contains(msg, "TS999-202412") {
					t.Errorf("expected error message to contain ID TS999-202412, got: %s", msg)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テスト用DBセットアップ
			db, cleanup := setupTestDB(t)
			defer cleanup()

			if tt.setup != nil {
				tt.setup(t, db)
			}

			repo := postgres.NewTestSuiteRepository(db)
			err := repo.UpdateStatus(context.Background(), tt.id, tt.newStatus)

			// エラー検証
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateStatus() error = %v, wantErr %v", err, tt.wantErr)
			}

			// 結果検証
			if tt.verify != nil {
				tt.verify(t, err, db)
			}
		})
	}
}

func TestPostgresTestSuiteRepository_Delete(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		setup   func(t *testing.T, db *sql.DB)
		wantErr bool
		verify  func(t *testing.T, err error, db *sql.DB)
	}{
		{
			name: "正常系：テストスイート削除",
			id:   "TS001-202412",
			setup: func(t *testing.T, db *sql.DB) {
				// テスト用データ作成
				_, err := db.Exec(`
					INSERT INTO test_suites (
						id, name, description, status, 
						estimated_start_date, estimated_end_date,
						require_effort_comment, created_at, updated_at
					) VALUES (
						$1, $2, $3, $4, $5, $6, $7, $8, $9
					)
				`,
					"TS001-202412",
					"商品管理システムテスト",
					"商品管理機能のテスト",
					valueobject.SuiteStatusPreparation,
					time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
					true,
					time.Now(),
					time.Now(),
				)
				if err != nil {
					t.Fatalf("setup failed: %v", err)
				}
			},
			wantErr: false,
			verify: func(t *testing.T, err error, db *sql.DB) {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
					return
				}

				// 削除されたことを確認
				var count int
				err = db.QueryRow("SELECT COUNT(*) FROM test_suites WHERE id = $1", "TS001-202412").Scan(&count)
				if err != nil {
					t.Errorf("failed to count records: %v", err)
					return
				}

				if count != 0 {
					t.Errorf("expected 0 records after deletion, got %d", count)
				}
			},
		},
		{
			name: "異常系：存在しないIDの削除",
			id:   "TS999-202412", // 存在しないID
			setup: func(t *testing.T, db *sql.DB) {
				// 何もセットアップしない（存在しないIDをテスト）
			},
			wantErr: true,
			// 異常系：存在しないIDの削除のテストケース内の検証部分
			verify: func(t *testing.T, err error, db *sql.DB) {
				if err == nil {
					t.Errorf("expected error, got nil")
					return
				}

				// NotFoundエラー型の検証 - この部分を修正
				var notFoundErr *customerrors.NotFoundError
				if !errors.As(err, &notFoundErr) {
					t.Errorf("expected NotFoundError, got %T", err)
					return
				}

				// ステータスコードの検証
				if notFoundErr.StatusCode() != customerrors.StatusCodeNotFound {
					t.Errorf("expected status code %d, got %d", customerrors.StatusCodeNotFound, notFoundErr.StatusCode())
				}

				// エラーメッセージに該当IDが含まれていることを確認
				if msg := notFoundErr.Error(); !strings.Contains(msg, "TS999-202412") {
					t.Errorf("expected error message to contain ID TS999-202412, got: %s", msg)
				}
			},
		},
		{
			name: "異常系：関連するテストグループが存在する場合の削除",
			id:   "TS002-202412",
			setup: func(t *testing.T, db *sql.DB) {
				// テストスイート作成
				_, err := db.Exec(`
					INSERT INTO test_suites (
						id, name, description, status, 
						estimated_start_date, estimated_end_date
					) VALUES (
						$1, $2, $3, $4, $5, $6
					)
				`,
					"TS002-202412",
					"削除テスト用スイート",
					"削除テスト",
					valueobject.SuiteStatusPreparation,
					time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
				)
				if err != nil {
					t.Fatalf("failed to create test suite: %v", err)
				}

				// 関連するテストグループを作成
				_, err = db.Exec(`
					INSERT INTO test_groups (
						id, suite_id, name, display_order, status
					) VALUES (
						$1, $2, $3, $4, $5
					)
				`,
					"TS002TG01-202412",
					"TS002-202412",
					"関連グループ",
					1,
					valueobject.SuiteStatusPreparation,
				)
				if err != nil {
					t.Fatalf("failed to create test group: %v", err)
				}
			},
			wantErr: true,
			// 異常系：関連するテストグループが存在する場合の削除のテストケース内の検証部分
			verify: func(t *testing.T, err error, db *sql.DB) {
				if err == nil {
					t.Errorf("expected error, got nil")
					return
				}

				// ConflictError型の検証 - この部分を修正
				var conflictErr *customerrors.ConflictError
				if !errors.As(err, &conflictErr) {
					t.Errorf("expected ConflictError, got %T", err)
					return
				}

				// ステータスコードの検証
				if conflictErr.StatusCode() != customerrors.StatusCodeConflict {
					t.Errorf("expected status code %d, got %d", customerrors.StatusCodeConflict, conflictErr.StatusCode())
				}

				// スイートが削除されていないことを確認
				var count int
				err = db.QueryRow("SELECT COUNT(*) FROM test_suites WHERE id = $1", "TS002-202412").Scan(&count)
				if err != nil {
					t.Errorf("failed to count records: %v", err)
					return
				}

				if count != 1 {
					t.Errorf("expected suite to still exist, but got %d records", count)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テスト用DBセットアップ
			db, cleanup := setupTestDB(t)
			defer cleanup()

			if tt.setup != nil {
				tt.setup(t, db)
			}

			repo := postgres.NewTestSuiteRepository(db)
			err := repo.Delete(context.Background(), tt.id)

			// エラー検証
			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}

			// 結果検証
			if tt.verify != nil {
				tt.verify(t, err, db)
			}
		})
	}
}

func TestPostgresTestSuiteRepository_FindWithFilters(t *testing.T) {
	tests := []struct {
		name       string
		filters    *dto.TestSuiteQueryParamDTO
		setup      func(t *testing.T, db *sql.DB)
		wantErr    bool
		wantCount  int
		totalCount int
		verify     func(t *testing.T, suites []*entity.TestSuite, total int, err error)
	}{
		// 正常系のテストケースはそのまま

		{
			name: "異常系：無効なステータス値",
			filters: &dto.TestSuiteQueryParamDTO{
				Status:   stringPtr("無効なステータス"),
				Page:     intPtr(1),
				PageSize: intPtr(10),
			},
			setup: func(t *testing.T, db *sql.DB) {
				// 特に何もセットアップしない
			},
			wantErr: true,
			// 異常系：無効なステータス値のテストケース内の検証部分
			verify: func(t *testing.T, suites []*entity.TestSuite, total int, err error) {
				if err == nil {
					t.Errorf("expected error, got nil")
					return
				}

				// ValidationError型の検証 - この部分を修正
				var validationErr *customerrors.ValidationError
				if !errors.As(err, &validationErr) {
					t.Errorf("expected ValidationError, got %T", err)
					return
				}

				// ステータスコードの検証
				if validationErr.StatusCode() != customerrors.StatusCodeUnprocessableEntity {
					t.Errorf("expected status code %d, got %d", customerrors.StatusCodeUnprocessableEntity, validationErr.StatusCode())
				}

				// 結果がnilであることを確認
				if suites != nil {
					t.Errorf("expected nil suites, got %+v", suites)
				}

				if total != 0 {
					t.Errorf("expected total count 0, got %d", total)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テスト用DBセットアップ
			db, cleanup := setupTestDB(t)
			defer cleanup()

			if tt.setup != nil {
				tt.setup(t, db)
			}

			repo := postgres.NewTestSuiteRepository(db)
			suites, total, err := repo.FindWithFilters(context.Background(), tt.filters)

			// エラー検証
			if (err != nil) != tt.wantErr {
				t.Errorf("FindWithFilters() error = %v, wantErr %v", err, tt.wantErr)
			}

			// 結果検証
			if !tt.wantErr {
				if len(suites) != tt.wantCount {
					t.Errorf("expected %d suites, got %d", tt.wantCount, len(suites))
				}

				if total != tt.totalCount {
					t.Errorf("expected total count %d, got %d", tt.totalCount, total)
				}
			}

			if tt.verify != nil {
				tt.verify(t, suites, total, err)
			}
		})
	}
}

// ヘルパー関数
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func timePtr(t time.Time) *time.Time {
	return &t
}
