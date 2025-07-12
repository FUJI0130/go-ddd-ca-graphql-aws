package interactor

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/FUJI0130/go-ddd-ca/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTestCaseRepository はテスト用のモックリポジトリ
type MockTestCaseRepository struct {
	mock.Mock
}

func (m *MockTestCaseRepository) FindByID(ctx context.Context, id string) (*entity.TestCase, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.TestCase), args.Error(1)
}

func (m *MockTestCaseRepository) Create(ctx context.Context, testCase *entity.TestCase) error {
	args := m.Called(ctx, testCase)
	return args.Error(0)
}

func (m *MockTestCaseRepository) Update(ctx context.Context, testCase *entity.TestCase) error {
	args := m.Called(ctx, testCase)
	return args.Error(0)
}

func (m *MockTestCaseRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTestCaseRepository) FindByGroupID(ctx context.Context, groupID string) ([]*entity.TestCase, error) {
	args := m.Called(ctx, groupID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.TestCase), args.Error(1)
}

func (m *MockTestCaseRepository) UpdateStatus(ctx context.Context, id string, status entity.TestStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockTestCaseRepository) AddEffort(ctx context.Context, id string, effort float64) error {
	args := m.Called(ctx, id, effort)
	return args.Error(0)
}

func (m *MockTestCaseRepository) FindByStatus(ctx context.Context, status entity.TestStatus) ([]*entity.TestCase, error) {
	args := m.Called(ctx, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.TestCase), args.Error(1)
}

// MockTestCaseIDGenerator はテスト用のモックIDジェネレーター
type MockTestCaseIDGenerator struct {
	mock.Mock
}

func (m *MockTestCaseIDGenerator) GenerateID(groupID string) (string, error) {
	args := m.Called(groupID)
	return args.String(0), args.Error(1)
}

func TestGetCasesByGroupID(t *testing.T) {
	testCases := []struct {
		name          string
		setupMock     func(*MockTestCaseRepository)
		groupID       string
		expectedCases int
		expectedError bool
	}{
		{
			name: "正常系：テストケースの取得成功",
			setupMock: func(r *MockTestCaseRepository) {
				r.On("FindByGroupID", mock.Anything, "TS001TG01-202501").Return([]*entity.TestCase{
					{
						ID:            "TS001TG01TC001-202501",
						GroupID:       "TS001TG01-202501",
						Title:         "ログイン機能のテスト",
						Description:   "ユーザーがログインできることを確認する",
						Status:        "テスト",
						Priority:      "高",
						PlannedEffort: 2.5,
						ActualEffort:  3.0,
						IsDelayed:     true,
						DelayDays:     1,
						CreatedAt:     time.Now(),
						UpdatedAt:     time.Now(),
					},
					{
						ID:            "TS001TG01TC002-202501",
						GroupID:       "TS001TG01-202501",
						Title:         "ログアウト機能のテスト",
						Description:   "ユーザーがログアウトできることを確認する",
						Status:        "完了",
						Priority:      "中",
						PlannedEffort: 1.0,
						ActualEffort:  1.5,
						IsDelayed:     false,
						DelayDays:     0,
						CreatedAt:     time.Now(),
						UpdatedAt:     time.Now(),
					},
				}, nil)
			},
			groupID:       "TS001TG01-202501",
			expectedCases: 2,
			expectedError: false,
		},
		{
			name: "正常系：テストケースが0件",
			setupMock: func(r *MockTestCaseRepository) {
				r.On("FindByGroupID", mock.Anything, "TS001TG02-202501").Return([]*entity.TestCase{}, nil)
			},
			groupID:       "TS001TG02-202501",
			expectedCases: 0,
			expectedError: false,
		},
		{
			name: "異常系：リポジトリエラー",
			setupMock: func(r *MockTestCaseRepository) {
				r.On("FindByGroupID", mock.Anything, "TS001TG03-202501").Return(nil, fmt.Errorf("repository error"))
			},
			groupID:       "TS001TG03-202501",
			expectedCases: 0,
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// モックのセットアップ
			mockRepo := new(MockTestCaseRepository)
			mockIDGen := new(MockTestCaseIDGenerator)
			tc.setupMock(mockRepo)

			// インタラクターの作成
			interactor := NewTestCaseInteractor(mockRepo, mockIDGen)

			// テスト実行
			cases, err := interactor.GetCasesByGroupID(context.Background(), tc.groupID)

			// 検証
			if tc.expectedError {
				assert.Error(t, err)
				assert.Nil(t, cases)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedCases, len(cases))

				if tc.expectedCases > 0 {
					// レスポンスDTOの変換が正しく行われているかの確認
					assert.Equal(t, tc.groupID, cases[0].GroupID)
				}
			}

			// モックの呼び出し検証
			mockRepo.AssertExpectations(t)
		})
	}
}
