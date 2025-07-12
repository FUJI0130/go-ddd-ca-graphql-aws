package resolver

import (
	"context"
	"testing"
	"time"

	"github.com/FUJI0130/go-ddd-ca/internal/interface/graphql/model"
	"github.com/FUJI0130/go-ddd-ca/internal/usecase/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTestSuiteUseCase はテストスイートユースケースのモック
type MockTestSuiteUseCase struct {
	mock.Mock
}

func (m *MockTestSuiteUseCase) CreateTestSuite(ctx context.Context, createDTO *dto.TestSuiteCreateDTO) (*dto.TestSuiteResponseDTO, error) {
	args := m.Called(ctx, createDTO)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TestSuiteResponseDTO), args.Error(1)
}

func (m *MockTestSuiteUseCase) GetTestSuite(ctx context.Context, id string) (*dto.TestSuiteResponseDTO, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TestSuiteResponseDTO), args.Error(1)
}

func (m *MockTestSuiteUseCase) ListTestSuites(ctx context.Context, params *dto.TestSuiteQueryParamDTO) (*dto.TestSuiteListResponseDTO, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TestSuiteListResponseDTO), args.Error(1)
}

func (m *MockTestSuiteUseCase) UpdateTestSuite(ctx context.Context, id string, updateDTO *dto.TestSuiteUpdateDTO) (*dto.TestSuiteResponseDTO, error) {
	args := m.Called(ctx, id, updateDTO)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TestSuiteResponseDTO), args.Error(1)
}

func (m *MockTestSuiteUseCase) UpdateTestSuiteStatus(ctx context.Context, id string, statusDTO *dto.TestSuiteStatusUpdateDTO) (*dto.TestSuiteResponseDTO, error) {
	args := m.Called(ctx, id, statusDTO)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TestSuiteResponseDTO), args.Error(1)
}

// MockTestGroupUseCase はテストグループユースケースのモック
type MockTestGroupUseCase struct {
	mock.Mock
}

func (m *MockTestGroupUseCase) GetGroupsBySuiteID(ctx context.Context, suiteID string) ([]*dto.TestGroupResponseDTO, error) {
	args := m.Called(ctx, suiteID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*dto.TestGroupResponseDTO), args.Error(1)
}

// CreateTestGroup はテストグループを作成するモックメソッド
func (m *MockTestGroupUseCase) CreateTestGroup(ctx context.Context, input *dto.TestGroupCreateDTO) (*dto.TestGroupResponseDTO, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TestGroupResponseDTO), args.Error(1)
}

// その他TestGroupUseCaseメソッドを実装...

// MockTestCaseUseCase はテストケースユースケースのモック
type MockTestCaseUseCase struct {
	mock.Mock
}

func (m *MockTestCaseUseCase) GetCasesByGroupID(ctx context.Context, groupID string) ([]*dto.TestCaseResponseDTO, error) {
	args := m.Called(ctx, groupID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*dto.TestCaseResponseDTO), args.Error(1)
}

// その他TestCaseUseCaseメソッドを実装...

// テストヘルパー関数
func createTestDate(year, month, day int) time.Time {
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}

func createTestSuiteResponse() *dto.TestSuiteResponseDTO {
	return &dto.TestSuiteResponseDTO{
		ID:                   "TS001-202501",
		Name:                 "テストスイート1",
		Description:          "テスト用スイート",
		Status:               "準備中",
		EstimatedStartDate:   createTestDate(2025, 1, 1),
		EstimatedEndDate:     createTestDate(2025, 1, 31),
		RequireEffortComment: true,
		Progress:             0.0,
		CreatedAt:            time.Now().Add(-24 * time.Hour),
		UpdatedAt:            time.Now(),
	}
}

func createTestGroupResponses() []*dto.TestGroupResponseDTO {
	return []*dto.TestGroupResponseDTO{
		{
			ID:           "TS001TG01-202501",
			SuiteID:      "TS001-202501",
			Name:         "基本機能",
			Description:  "基本機能のテスト",
			DisplayOrder: 1,
			Status:       "準備中",
			CreatedAt:    time.Now().Add(-24 * time.Hour),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           "TS001TG02-202501",
			SuiteID:      "TS001-202501",
			Name:         "拡張機能",
			Description:  "拡張機能のテスト",
			DisplayOrder: 2,
			Status:       "準備中",
			CreatedAt:    time.Now().Add(-24 * time.Hour),
			UpdatedAt:    time.Now(),
		},
	}
}

func createTestCaseResponses() []*dto.TestCaseResponseDTO {
	return []*dto.TestCaseResponseDTO{
		{
			ID:            "TS001TG01TC001-202501",
			GroupID:       "TS001TG01-202501",
			Title:         "ログイン機能",
			Description:   "ログイン機能のテスト",
			Status:        "作成",
			Priority:      "High",
			PlannedEffort: 2.0,
			ActualEffort:  0.0,
			IsDelayed:     false,
			DelayDays:     0,
			CreatedAt:     time.Now().Add(-24 * time.Hour),
			UpdatedAt:     time.Now(),
		},
	}
}

// テストケース
func TestQueryResolver_TestSuite(t *testing.T) {
	// テストケース定義
	testCases := []struct {
		name           string
		id             string
		setupMock      func(*MockTestSuiteUseCase)
		expectedResult *model.TestSuite
		expectError    bool
	}{
		{
			name: "正常系：存在するIDのテストスイート取得",
			id:   "TS001-202501",
			setupMock: func(m *MockTestSuiteUseCase) {
				m.On("GetTestSuite", mock.Anything, "TS001-202501").Return(createTestSuiteResponse(), nil)
			},
			expectedResult: &model.TestSuite{
				ID:                   "TS001-202501",
				Name:                 "テストスイート1",
				Status:               model.SuiteStatusPreparation,
				RequireEffortComment: true,
			},
			expectError: false,
		},
		{
			name: "異常系：存在しないIDのテストスイート取得",
			id:   "INVALID-ID",
			setupMock: func(m *MockTestSuiteUseCase) {
				m.On("GetTestSuite", mock.Anything, "INVALID-ID").Return(nil, assert.AnError)
			},
			expectedResult: nil,
			expectError:    true,
		},
	}

	// 各テストケースを実行
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// モックのセットアップ
			mockSuiteUseCase := new(MockTestSuiteUseCase)
			mockGroupUseCase := new(MockTestGroupUseCase)
			mockCaseUseCase := new(MockTestCaseUseCase)
			tc.setupMock(mockSuiteUseCase)

			// リゾルバーの作成
			resolver := NewResolver(mockSuiteUseCase, mockGroupUseCase, mockCaseUseCase)
			queryResolver := resolver.Query()

			// テスト実行
			result, err := queryResolver.TestSuite(context.Background(), tc.id)

			// 結果の検証
			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tc.expectedResult.ID, result.ID)
				assert.Equal(t, tc.expectedResult.Name, result.Name)
				assert.Equal(t, tc.expectedResult.Status, result.Status)
			}

			// モックの検証
			mockSuiteUseCase.AssertExpectations(t)
		})
	}
}

func TestMutationResolver_CreateTestSuite(t *testing.T) {
	// テストケース定義
	testCases := []struct {
		name           string
		input          model.CreateTestSuiteInput
		setupMock      func(*MockTestSuiteUseCase)
		expectedResult *model.TestSuite
		expectError    bool
	}{
		{
			name: "正常系：テストスイート作成",
			input: model.CreateTestSuiteInput{
				Name:               "新規テストスイート",
				EstimatedStartDate: createTestDate(2025, 4, 1),
				EstimatedEndDate:   createTestDate(2025, 4, 30),
			},
			setupMock: func(m *MockTestSuiteUseCase) {
				m.On("CreateTestSuite", mock.Anything, mock.MatchedBy(func(dto *dto.TestSuiteCreateDTO) bool {
					return dto.Name == "新規テストスイート" &&
						dto.EstimatedStartDate.Equal(createTestDate(2025, 4, 1)) &&
						dto.EstimatedEndDate.Equal(createTestDate(2025, 4, 30))
				})).Return(&dto.TestSuiteResponseDTO{
					ID:                   "TS002-202504",
					Name:                 "新規テストスイート",
					Status:               "準備中",
					EstimatedStartDate:   createTestDate(2025, 4, 1),
					EstimatedEndDate:     createTestDate(2025, 4, 30),
					RequireEffortComment: false,
					Progress:             0.0,
					CreatedAt:            time.Now(),
					UpdatedAt:            time.Now(),
				}, nil)
			},
			expectedResult: &model.TestSuite{
				ID:     "TS002-202504",
				Name:   "新規テストスイート",
				Status: model.SuiteStatusPreparation,
			},
			expectError: false,
		},
	}

	// 各テストケースを実行
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// モックのセットアップ
			mockSuiteUseCase := new(MockTestSuiteUseCase)
			mockGroupUseCase := new(MockTestGroupUseCase)
			mockCaseUseCase := new(MockTestCaseUseCase)
			tc.setupMock(mockSuiteUseCase)

			// リゾルバーの作成
			resolver := NewResolver(mockSuiteUseCase, mockGroupUseCase, mockCaseUseCase)
			mutationResolver := resolver.Mutation()

			// テスト実行
			result, err := mutationResolver.CreateTestSuite(context.Background(), tc.input)

			// 結果の検証
			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tc.expectedResult.ID, result.ID)
				assert.Equal(t, tc.expectedResult.Name, result.Name)
				assert.Equal(t, tc.expectedResult.Status, result.Status)
			}

			// モックの検証
			mockSuiteUseCase.AssertExpectations(t)
		})
	}
}

func TestTestSuiteResolver_Groups(t *testing.T) {
	// テストケース定義
	testCases := []struct {
		name              string
		testSuite         *model.TestSuite
		setupMock         func(*MockTestGroupUseCase)
		expectedGroupsLen int
		expectError       bool
	}{
		{
			name: "正常系：グループ取得",
			testSuite: &model.TestSuite{
				ID:     "TS001-202501",
				Name:   "テストスイート1",
				Status: model.SuiteStatusPreparation,
			},
			setupMock: func(m *MockTestGroupUseCase) {
				m.On("GetGroupsBySuiteID", mock.Anything, "TS001-202501").Return(createTestGroupResponses(), nil)
			},
			expectedGroupsLen: 2,
			expectError:       false,
		},
		{
			name: "異常系：グループ取得エラー",
			testSuite: &model.TestSuite{
				ID:     "INVALID-ID",
				Name:   "無効なテストスイート",
				Status: model.SuiteStatusPreparation,
			},
			setupMock: func(m *MockTestGroupUseCase) {
				m.On("GetGroupsBySuiteID", mock.Anything, "INVALID-ID").Return(nil, assert.AnError)
			},
			expectedGroupsLen: 0,
			expectError:       true,
		},
	}

	// 各テストケースを実行
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// モックのセットアップ
			mockSuiteUseCase := new(MockTestSuiteUseCase)
			mockGroupUseCase := new(MockTestGroupUseCase)
			mockCaseUseCase := new(MockTestCaseUseCase)
			tc.setupMock(mockGroupUseCase)

			// リゾルバーの作成
			resolver := NewResolver(mockSuiteUseCase, mockGroupUseCase, mockCaseUseCase)
			testSuiteResolver := resolver.TestSuite()

			// テスト実行
			groups, err := testSuiteResolver.Groups(context.Background(), tc.testSuite)

			// 結果の検証
			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, groups)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, groups)
				assert.Len(t, groups, tc.expectedGroupsLen)
				if len(groups) > 0 {
					assert.Equal(t, "TS001TG01-202501", groups[0].ID)
					assert.Equal(t, "基本機能", groups[0].Name)
				}
			}

			// モックの検証
			mockGroupUseCase.AssertExpectations(t)
		})
	}
}

func TestTestGroupResolver_Cases(t *testing.T) {
	// テストケース定義
	testCases := []struct {
		name             string
		testGroup        *model.TestGroup
		setupMock        func(*MockTestCaseUseCase)
		expectedCasesLen int
		expectError      bool
	}{
		{
			name: "正常系：ケース取得",
			testGroup: &model.TestGroup{
				ID:     "TS001TG01-202501",
				Name:   "基本機能",
				Status: model.SuiteStatusPreparation,
			},
			setupMock: func(m *MockTestCaseUseCase) {
				m.On("GetCasesByGroupID", mock.Anything, "TS001TG01-202501").Return(createTestCaseResponses(), nil)
			},
			expectedCasesLen: 1,
			expectError:      false,
		},
		{
			name: "異常系：ケース取得エラー",
			testGroup: &model.TestGroup{
				ID:     "INVALID-ID",
				Name:   "無効なテストグループ",
				Status: model.SuiteStatusPreparation,
			},
			setupMock: func(m *MockTestCaseUseCase) {
				m.On("GetCasesByGroupID", mock.Anything, "INVALID-ID").Return(nil, assert.AnError)
			},
			expectedCasesLen: 0,
			expectError:      true,
		},
	}

	// 各テストケースを実行
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// モックのセットアップ
			mockSuiteUseCase := new(MockTestSuiteUseCase)
			mockGroupUseCase := new(MockTestGroupUseCase)
			mockCaseUseCase := new(MockTestCaseUseCase)
			tc.setupMock(mockCaseUseCase)

			// リゾルバーの作成
			resolver := NewResolver(mockSuiteUseCase, mockGroupUseCase, mockCaseUseCase)
			testGroupResolver := resolver.TestGroup()

			// テスト実行
			cases, err := testGroupResolver.Cases(context.Background(), tc.testGroup)

			// 結果の検証
			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, cases)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cases)
				assert.Len(t, cases, tc.expectedCasesLen)
				if len(cases) > 0 {
					assert.Equal(t, "TS001TG01TC001-202501", cases[0].ID)
					assert.Equal(t, "ログイン機能", cases[0].Title)
				}
			}

			// モックの検証
			mockCaseUseCase.AssertExpectations(t)
		})
	}
}
