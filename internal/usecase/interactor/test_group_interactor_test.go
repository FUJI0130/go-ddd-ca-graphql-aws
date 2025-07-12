package interactor

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/FUJI0130/go-ddd-ca/internal/domain/entity"
	"github.com/FUJI0130/go-ddd-ca/internal/domain/valueobject"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTestGroupRepository はテスト用のモックリポジトリ
type MockTestGroupRepository struct {
	mock.Mock
}

func (m *MockTestGroupRepository) FindByID(ctx context.Context, id string) (*entity.TestGroup, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.TestGroup), args.Error(1)
}

func (m *MockTestGroupRepository) Create(ctx context.Context, group *entity.TestGroup) error {
	args := m.Called(ctx, group)
	return args.Error(0)
}

func (m *MockTestGroupRepository) Update(ctx context.Context, group *entity.TestGroup) error {
	args := m.Called(ctx, group)
	return args.Error(0)
}

func (m *MockTestGroupRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTestGroupRepository) FindBySuiteID(ctx context.Context, suiteID string) ([]*entity.TestGroup, error) {
	args := m.Called(ctx, suiteID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.TestGroup), args.Error(1)
}

func (m *MockTestGroupRepository) UpdateStatus(ctx context.Context, id string, status valueobject.SuiteStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockTestGroupRepository) UpdateDisplayOrder(ctx context.Context, id string, displayOrder int) error {
	args := m.Called(ctx, id, displayOrder)
	return args.Error(0)
}

// MockTestGroupIDGenerator はテスト用のモックIDジェネレーター
type MockTestGroupIDGenerator struct {
	mock.Mock
}

func (m *MockTestGroupIDGenerator) GenerateID(suiteID string) (string, error) {
	args := m.Called(suiteID)
	return args.String(0), args.Error(1)
}

func TestGetGroupsBySuiteID(t *testing.T) {
	testCases := []struct {
		name           string
		setupMock      func(*MockTestGroupRepository)
		suiteID        string
		expectedGroups int
		expectedError  bool
	}{
		{
			name: "正常系：グループの取得成功",
			setupMock: func(r *MockTestGroupRepository) {
				r.On("FindBySuiteID", mock.Anything, "TS001-202501").Return([]*entity.TestGroup{
					{
						ID:           "TS001TG01-202501",
						SuiteID:      "TS001-202501",
						Name:         "機能テスト",
						Description:  "基本機能のテスト",
						DisplayOrder: 1,
						Status:       valueobject.SuiteStatusInProgress,
						CreatedAt:    time.Now(),
						UpdatedAt:    time.Now(),
					},
					{
						ID:           "TS001TG02-202501",
						SuiteID:      "TS001-202501",
						Name:         "パフォーマンステスト",
						Description:  "負荷テスト",
						DisplayOrder: 2,
						Status:       valueobject.SuiteStatusPreparation,
						CreatedAt:    time.Now(),
						UpdatedAt:    time.Now(),
					},
				}, nil)
			},
			suiteID:        "TS001-202501",
			expectedGroups: 2,
			expectedError:  false,
		},
		{
			name: "正常系：グループが0件",
			setupMock: func(r *MockTestGroupRepository) {
				r.On("FindBySuiteID", mock.Anything, "TS002-202501").Return([]*entity.TestGroup{}, nil)
			},
			suiteID:        "TS002-202501",
			expectedGroups: 0,
			expectedError:  false,
		},
		{
			name: "異常系：リポジトリエラー",
			setupMock: func(r *MockTestGroupRepository) {
				r.On("FindBySuiteID", mock.Anything, "TS003-202501").Return(nil, fmt.Errorf("repository error"))
			},
			suiteID:        "TS003-202501",
			expectedGroups: 0,
			expectedError:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// モックのセットアップ
			mockRepo := new(MockTestGroupRepository)
			mockIDGen := new(MockTestGroupIDGenerator)
			tc.setupMock(mockRepo)

			// インタラクターの作成
			interactor := NewTestGroupInteractor(mockRepo, mockIDGen)

			// テスト実行
			groups, err := interactor.GetGroupsBySuiteID(context.Background(), tc.suiteID)

			// 検証
			if tc.expectedError {
				assert.Error(t, err)
				assert.Nil(t, groups)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedGroups, len(groups))
			}

			// モックの呼び出し検証
			mockRepo.AssertExpectations(t)
		})
	}
}
