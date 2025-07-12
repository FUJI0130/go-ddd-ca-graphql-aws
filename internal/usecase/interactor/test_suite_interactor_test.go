package interactor

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/FUJI0130/go-ddd-ca/internal/domain/entity"
	"github.com/FUJI0130/go-ddd-ca/internal/domain/valueobject"
	"github.com/FUJI0130/go-ddd-ca/internal/usecase/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTestSuiteRepository はテスト用のモックリポジトリ
type MockTestSuiteRepository struct {
	mock.Mock
}

// MockTestSuiteIDGenerator はテスト用のモックIDジェネレーター
type MockTestSuiteIDGenerator struct {
	mock.Mock
}

func (m *MockTestSuiteIDGenerator) GenerateID() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockTestSuiteRepository) FindByID(ctx context.Context, id string) (*entity.TestSuite, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.TestSuite), args.Error(1)
}

func (m *MockTestSuiteRepository) Create(ctx context.Context, suite *entity.TestSuite) error {
	args := m.Called(ctx, suite)
	return args.Error(0)
}

func (m *MockTestSuiteRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTestSuiteRepository) Update(ctx context.Context, suite *entity.TestSuite) error {
	args := m.Called(ctx, suite)
	return args.Error(0)
}

func (m *MockTestSuiteRepository) UpdateStatus(ctx context.Context, id string, status valueobject.SuiteStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockTestSuiteRepository) FindByStatus(ctx context.Context, status valueobject.SuiteStatus) ([]*entity.TestSuite, error) {
	args := m.Called(ctx, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.TestSuite), args.Error(1)
}

// MockTestSuiteRepositoryにFindWithFiltersメソッドを追加
func (m *MockTestSuiteRepository) FindWithFilters(ctx context.Context, params *dto.TestSuiteQueryParamDTO) ([]*entity.TestSuite, int, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*entity.TestSuite), args.Int(1), args.Error(2)
}

func TestListTestSuites(t *testing.T) {
	testCases := []struct {
		name          string
		setupMock     func(*MockTestSuiteRepository)
		inputParams   *dto.TestSuiteQueryParamDTO
		expectedLen   int
		expectedTotal int
		expectedError bool
	}{
		{
			name: "全件取得",
			setupMock: func(r *MockTestSuiteRepository) {
				r.On("FindWithFilters", mock.Anything, mock.AnythingOfType("*dto.TestSuiteQueryParamDTO")).
					Return([]*entity.TestSuite{
						{
							ID:                   "TS001-202501",
							Name:                 "テストスイート1",
							Status:               valueobject.SuiteStatusPreparation,
							RequireEffortComment: true,
							CreatedAt:            time.Now(),
							UpdatedAt:            time.Now(),
						},
						{
							ID:                   "TS002-202501",
							Name:                 "テストスイート2",
							Status:               valueobject.SuiteStatusInProgress,
							RequireEffortComment: false,
							CreatedAt:            time.Now(),
							UpdatedAt:            time.Now(),
						},
					}, 2, nil)
			},
			inputParams:   &dto.TestSuiteQueryParamDTO{},
			expectedLen:   2,
			expectedTotal: 2,
			expectedError: false,
		},
		{
			name: "ステータスによるフィルタリング",
			setupMock: func(r *MockTestSuiteRepository) {
				status := "実行中"
				r.On("FindWithFilters", mock.Anything, &dto.TestSuiteQueryParamDTO{
					Status: &status,
				}).Return([]*entity.TestSuite{
					{
						ID:                   "TS002-202501",
						Name:                 "テストスイート2",
						Status:               valueobject.SuiteStatusInProgress,
						RequireEffortComment: false,
						CreatedAt:            time.Now(),
						UpdatedAt:            time.Now(),
					},
				}, 1, nil)
			},
			inputParams: &dto.TestSuiteQueryParamDTO{
				Status: func(s string) *string { return &s }("実行中"),
			},
			expectedLen:   1,
			expectedTotal: 1,
			expectedError: false,
		},
		{
			name: "エラーケース",
			setupMock: func(r *MockTestSuiteRepository) {
				r.On("FindWithFilters", mock.Anything, mock.AnythingOfType("*dto.TestSuiteQueryParamDTO")).
					Return(nil, 0, fmt.Errorf("database error"))
			},
			inputParams:   &dto.TestSuiteQueryParamDTO{},
			expectedLen:   0,
			expectedTotal: 0,
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// モックリポジトリのセットアップ
			mockRepo := new(MockTestSuiteRepository)
			mockIDGen := new(MockTestSuiteIDGenerator)
			tc.setupMock(mockRepo)
			mockIDGen.On("GenerateID").Return("TS001-202501", nil) // IDの期待値を設定

			// インタラクターの作成
			interactor := NewTestSuiteInteractor(mockRepo, mockIDGen)

			// テストの実行
			result, err := interactor.ListTestSuites(context.Background(), tc.inputParams)

			// 検証
			if tc.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tc.expectedLen, len(result.TestSuites))
				assert.Equal(t, tc.expectedTotal, result.Total)
			}

			// モックの呼び出しを検証
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetTestSuite(t *testing.T) {
	// テストケースの構造体
	testCases := []struct {
		name             string
		setupMock        func(*MockTestSuiteRepository)
		inputID          string
		expectedError    bool
		expectedStatus   valueobject.SuiteStatus
		expectedProgress float64
	}{
		{
			name: "準備中のテストスイート取得",
			setupMock: func(r *MockTestSuiteRepository) {
				r.On("FindByID", mock.Anything, "TS001-202501").Return(&entity.TestSuite{
					ID:        "TS001-202501",
					Status:    valueobject.SuiteStatusPreparation,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}, nil)
			},
			inputID:          "TS001-202501",
			expectedError:    false,
			expectedStatus:   valueobject.SuiteStatusPreparation,
			expectedProgress: 0.0,
		},
		{
			name: "実行中のテストスイート取得",
			setupMock: func(r *MockTestSuiteRepository) {
				r.On("FindByID", mock.Anything, "TS002-202501").Return(&entity.TestSuite{
					ID:        "TS002-202501",
					Status:    valueobject.SuiteStatusInProgress,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}, nil)
			},
			inputID:          "TS002-202501",
			expectedError:    false,
			expectedStatus:   valueobject.SuiteStatusInProgress,
			expectedProgress: 50.0,
		},
		{
			name: "完了のテストスイート取得",
			setupMock: func(r *MockTestSuiteRepository) {
				r.On("FindByID", mock.Anything, "TS003-202501").Return(&entity.TestSuite{
					ID:        "TS003-202501",
					Status:    valueobject.SuiteStatusCompleted,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}, nil)
			},
			inputID:          "TS003-202501",
			expectedError:    false,
			expectedStatus:   valueobject.SuiteStatusCompleted,
			expectedProgress: 100.0,
		},
		{
			name: "中断のテストスイート取得",
			setupMock: func(r *MockTestSuiteRepository) {
				r.On("FindByID", mock.Anything, "TS004-202501").Return(&entity.TestSuite{
					ID:        "TS004-202501",
					Status:    valueobject.SuiteStatusSuspended,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}, nil)
			},
			inputID:          "TS004-202501",
			expectedError:    false,
			expectedStatus:   valueobject.SuiteStatusSuspended,
			expectedProgress: 75.0,
		},
		{
			name: "存在しないIDのテストスイート取得",
			setupMock: func(r *MockTestSuiteRepository) {
				r.On("FindByID", mock.Anything, "invalid-id").Return(nil, entity.ErrTestSuiteNotFound)
			},
			inputID:        "invalid-id",
			expectedError:  true,
			expectedStatus: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// モックリポジトリのセットアップ
			mockRepo := new(MockTestSuiteRepository)
			mockIDGen := new(MockTestSuiteIDGenerator)
			tc.setupMock(mockRepo)
			mockIDGen.On("GenerateID").Return("TS001-202501", nil) // IDの期待値を設定

			// インタラクターの作成
			interactor := NewTestSuiteInteractor(mockRepo, mockIDGen)

			// テストの実行
			result, err := interactor.GetTestSuite(context.Background(), tc.inputID)

			// 検証
			if tc.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, string(tc.expectedStatus), result.Status)
				assert.Equal(t, tc.expectedProgress, result.Progress)
			}

			// モックの呼び出しを検証
			mockRepo.AssertExpectations(t)
		})
	}
}
