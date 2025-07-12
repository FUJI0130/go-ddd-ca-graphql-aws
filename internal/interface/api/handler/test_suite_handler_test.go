package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/FUJI0130/go-ddd-ca/internal/domain/entity"
	"github.com/FUJI0130/go-ddd-ca/internal/usecase/dto"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// モックの定義
type mockTestSuiteUseCase struct {
	mock.Mock
}

func (m *mockTestSuiteUseCase) CreateTestSuite(createDTO *dto.TestSuiteCreateDTO) (*dto.TestSuiteResponseDTO, error) {
	args := m.Called(createDTO)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TestSuiteResponseDTO), args.Error(1)
}

func (m *mockTestSuiteUseCase) GetTestSuite(id string) (*dto.TestSuiteResponseDTO, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TestSuiteResponseDTO), args.Error(1)
}

// テストケースの構造体定義
type getTestSuiteTestCase struct {
	name           string
	setupMock      func(*mockTestSuiteUseCase)
	inputID        string
	expectedCode   int
	isValidRequest bool
	expectSuccess  bool
}

func TestGetTestSuite(t *testing.T) {
	// テストケースの定義
	fixedTime := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	tests := []getTestSuiteTestCase{
		{
			name: "既存のテストスイートの取得",
			setupMock: func(m *mockTestSuiteUseCase) {
				m.On("GetTestSuite", "TS001-202401").Return(&dto.TestSuiteResponseDTO{
					ID:                   "TS001-202401",
					Name:                 "テストスイート1",
					Description:          "説明",
					Status:               "準備中",
					EstimatedStartDate:   fixedTime,
					EstimatedEndDate:     fixedTime.AddDate(0, 1, 0),
					RequireEffortComment: true,
					Progress:             0.0,
					CreatedAt:            fixedTime,
					UpdatedAt:            fixedTime,
				}, nil)
			},
			inputID:        "TS001-202401",
			expectedCode:   http.StatusOK,
			isValidRequest: true,
			expectSuccess:  true,
		},
		{
			name: "指定IDのテストスイートが未存在",
			setupMock: func(m *mockTestSuiteUseCase) {
				m.On("GetTestSuite", "NON-EXISTENT").Return(nil, entity.ErrTestSuiteNotFound)
			},
			inputID:        "NON-EXISTENT",
			expectedCode:   http.StatusNotFound,
			isValidRequest: true,
			expectSuccess:  false,
		},
		{
			name: "データベースエラー発生",
			setupMock: func(m *mockTestSuiteUseCase) {
				m.On("GetTestSuite", "ERROR-CASE").Return(nil, errors.New("database connection error"))
			},
			inputID:        "ERROR-CASE",
			expectedCode:   http.StatusInternalServerError,
			isValidRequest: true,
			expectSuccess:  false,
		},
		{
			name: "空のIDを指定",
			setupMock: func(m *mockTestSuiteUseCase) {
				m.On("GetTestSuite", "").Return(nil, entity.ErrTestSuiteNotFound)
			},
			inputID:        "",
			expectedCode:   http.StatusNotFound,
			isValidRequest: false,
			expectSuccess:  false,
		},
		{
			name: "不正なフォーマットのID",
			setupMock: func(m *mockTestSuiteUseCase) {
				m.On("GetTestSuite", "invalid-format").Return(nil, entity.ErrTestSuiteNotFound)
			},
			inputID:        "invalid-format",
			expectedCode:   http.StatusNotFound,
			isValidRequest: false,
			expectSuccess:  false,
		},
		{
			name: "レスポンス全フィールドの検証",
			setupMock: func(m *mockTestSuiteUseCase) {
				expectedResponse := &dto.TestSuiteResponseDTO{
					ID:                   "TS002-202401",
					Name:                 "詳細テストスイート",
					Description:          "全フィールド検証用",
					Status:               "実行中",
					EstimatedStartDate:   fixedTime,
					EstimatedEndDate:     fixedTime.AddDate(0, 1, 0),
					RequireEffortComment: true,
					Progress:             50.0, // 実行中ステータスの進捗率
					CreatedAt:            fixedTime,
					UpdatedAt:            fixedTime,
				}
				m.On("GetTestSuite", "TS002-202401").Return(expectedResponse, nil)
			},
			inputID:        "TS002-202401",
			expectedCode:   http.StatusOK,
			isValidRequest: true,
			expectSuccess:  true,
		},
	}

	// テストの実行
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックの準備
			mockUseCase := new(mockTestSuiteUseCase)
			tt.setupMock(mockUseCase)
			handler := NewTestSuiteHandler(mockUseCase)

			// リクエストの作成
			req, err := http.NewRequest("GET", "/test-suites/"+tt.inputID, nil)
			assert.NoError(t, err)

			vars := map[string]string{
				"id": tt.inputID,
			}
			req = mux.SetURLVars(req, vars)

			// レスポンスの準備
			w := httptest.NewRecorder()

			// ハンドラーの実行
			handler.GetTestSuite(w, req)

			// アサーション
			assert.Equal(t, tt.expectedCode, w.Code)

			if tt.expectSuccess {
				var response dto.TestSuiteResponseDTO
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, tt.inputID, response.ID)
				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

				// 全フィールドの検証（TS002-202401の場合）
				if tt.inputID == "TS002-202401" {
					assert.Equal(t, "詳細テストスイート", response.Name)
					assert.Equal(t, "全フィールド検証用", response.Description)
					assert.Equal(t, "実行中", response.Status)
					assert.Equal(t, fixedTime, response.EstimatedStartDate)
					assert.Equal(t, fixedTime.AddDate(0, 1, 0), response.EstimatedEndDate)
					assert.True(t, response.RequireEffortComment)
					assert.Equal(t, 50.0, response.Progress)
					assert.Equal(t, fixedTime, response.CreatedAt)
					assert.Equal(t, fixedTime, response.UpdatedAt)
				}
			}

			// モックの検証
			mockUseCase.AssertExpectations(t)
		})
	}
}

type createTestSuiteTestCase struct {
	name          string
	input         *dto.TestSuiteCreateDTO
	setupMock     func(*mockTestSuiteUseCase)
	expectedCode  int
	expectSuccess bool
}

func TestCreateTestSuite(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	tests := []createTestSuiteTestCase{
		{
			name: "テストスイート作成成功",
			input: &dto.TestSuiteCreateDTO{
				Name:                 "新規テストスイート",
				Description:          "説明",
				EstimatedStartDate:   fixedTime,
				EstimatedEndDate:     fixedTime.AddDate(0, 1, 0),
				RequireEffortComment: true,
			},
			setupMock: func(m *mockTestSuiteUseCase) {
				expectedResponse := &dto.TestSuiteResponseDTO{
					ID:                   "TS001-202401",
					Name:                 "新規テストスイート",
					Description:          "説明",
					Status:               "準備中",
					EstimatedStartDate:   fixedTime,
					EstimatedEndDate:     fixedTime.AddDate(0, 1, 0),
					RequireEffortComment: true,
					Progress:             0.0,
					CreatedAt:            fixedTime,
					UpdatedAt:            fixedTime,
				}
				m.On("CreateTestSuite", mock.AnythingOfType("*dto.TestSuiteCreateDTO")).Return(expectedResponse, nil)
			},
			expectedCode:  http.StatusCreated,
			expectSuccess: true,
		},
		{
			name: "重複テストスイート作成エラー",
			input: &dto.TestSuiteCreateDTO{
				Name:                 "既存テストスイート",
				Description:          "説明",
				EstimatedStartDate:   fixedTime,
				EstimatedEndDate:     fixedTime.AddDate(0, 1, 0),
				RequireEffortComment: true,
			},
			setupMock: func(m *mockTestSuiteUseCase) {
				m.On("CreateTestSuite", mock.AnythingOfType("*dto.TestSuiteCreateDTO")).Return(nil, ErrDuplicateTestSuite)
			},
			expectedCode:  http.StatusConflict,
			expectSuccess: false,
		},
		{
			name: "必須フィールド未入力エラー",
			input: &dto.TestSuiteCreateDTO{
				Description:          "説明",
				EstimatedStartDate:   fixedTime,
				EstimatedEndDate:     fixedTime.AddDate(0, 1, 0),
				RequireEffortComment: true,
			},
			setupMock: func(m *mockTestSuiteUseCase) {
				// バリデーションはハンドラーレベルで行われるため、モックの設定は不要
			},
			expectedCode:  http.StatusBadRequest,
			expectSuccess: false,
		},
		{
			name: "データベースエラー",
			input: &dto.TestSuiteCreateDTO{
				Name:                 "テストスイート",
				Description:          "説明",
				EstimatedStartDate:   fixedTime,
				EstimatedEndDate:     fixedTime.AddDate(0, 1, 0),
				RequireEffortComment: true,
			},
			setupMock: func(m *mockTestSuiteUseCase) {
				m.On("CreateTestSuite", mock.AnythingOfType("*dto.TestSuiteCreateDTO")).Return(nil, errors.New("database error"))
			},
			expectedCode:  http.StatusInternalServerError,
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUseCase := new(mockTestSuiteUseCase)
			tt.setupMock(mockUseCase)
			handler := NewTestSuiteHandler(mockUseCase)

			jsonData, err := json.Marshal(tt.input)
			assert.NoError(t, err)

			req, err := http.NewRequest("POST", "/test-suites", bytes.NewBuffer(jsonData))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			handler.Create(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)

			if tt.expectSuccess {
				var response dto.TestSuiteResponseDTO
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, tt.input.Name, response.Name)
			}

			mockUseCase.AssertExpectations(t)
		})
	}
}

// モックメソッド追加
func (m *mockTestSuiteUseCase) ListTestSuites(params *dto.TestSuiteQueryParamDTO) (*dto.TestSuiteListResponseDTO, error) {
	args := m.Called(params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TestSuiteListResponseDTO), args.Error(1)
}

// テストケース構造体定義
type listTestSuiteTestCase struct {
	name          string
	queryParams   map[string]string
	setupMock     func(*mockTestSuiteUseCase)
	expectedCode  int
	expectSuccess bool
}

func TestListTestSuite(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	tests := []listTestSuiteTestCase{
		{
			name:        "全テストスイート取得",
			queryParams: map[string]string{},
			setupMock: func(m *mockTestSuiteUseCase) {
				expectedResponse := &dto.TestSuiteListResponseDTO{
					TestSuites: []dto.TestSuiteResponseDTO{{
						ID:                 "TS001-202401",
						Name:               "テストスイート1",
						Status:             "準備中",
						EstimatedStartDate: fixedTime,
						EstimatedEndDate:   fixedTime.AddDate(0, 1, 0),
						CreatedAt:          fixedTime,
						UpdatedAt:          fixedTime,
					}},
					Total: 1,
				}
				m.On("ListTestSuites", mock.AnythingOfType("*dto.TestSuiteQueryParamDTO")).Return(expectedResponse, nil)
			},
			expectedCode:  http.StatusOK,
			expectSuccess: true,
		},
		{
			name: "ステータスによるフィルタリング",
			queryParams: map[string]string{
				"status":   "準備中",
				"page":     "1",
				"pageSize": "10",
			},
			setupMock: func(m *mockTestSuiteUseCase) {
				expectedResponse := &dto.TestSuiteListResponseDTO{
					TestSuites: []dto.TestSuiteResponseDTO{{
						ID:                 "TS001-202401",
						Name:               "テストスイート1",
						Status:             "準備中",
						EstimatedStartDate: fixedTime,
						EstimatedEndDate:   fixedTime.AddDate(0, 1, 0),
						CreatedAt:          fixedTime,
						UpdatedAt:          fixedTime,
					}},
					Total: 1,
				}
				m.On("ListTestSuites", mock.AnythingOfType("*dto.TestSuiteQueryParamDTO")).Return(expectedResponse, nil)
			},
			expectedCode:  http.StatusOK,
			expectSuccess: true,
		},
		{
			name: "無効なページネーションパラメータ",
			queryParams: map[string]string{
				"page":     "-1",
				"pageSize": "999",
			},
			setupMock: func(m *mockTestSuiteUseCase) {
				// バリデーションエラーのためモック不要
			},
			expectedCode:  http.StatusBadRequest,
			expectSuccess: false,
		},
		{
			name: "データベースエラー",
			queryParams: map[string]string{
				"page":     "1",
				"pageSize": "10",
			},
			setupMock: func(m *mockTestSuiteUseCase) {
				m.On("ListTestSuites", mock.AnythingOfType("*dto.TestSuiteQueryParamDTO")).Return(nil, errors.New("database error"))
			},
			expectedCode:  http.StatusInternalServerError,
			expectSuccess: false,
		},
	}

	// テストの実行
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUseCase := new(mockTestSuiteUseCase)
			tt.setupMock(mockUseCase)
			handler := NewTestSuiteHandler(mockUseCase)

			// クエリパラメータ付きURLの作成
			req, err := http.NewRequest("GET", "/test-suites", nil)
			assert.NoError(t, err)

			q := req.URL.Query()
			for key, value := range tt.queryParams {
				q.Add(key, value)
			}
			req.URL.RawQuery = q.Encode()

			w := httptest.NewRecorder()
			handler.List(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)

			if tt.expectSuccess {
				var response dto.TestSuiteListResponseDTO
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response.TestSuites)
				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
			}

			mockUseCase.AssertExpectations(t)
		})
	}
}

// モックメソッド追加
func (m *mockTestSuiteUseCase) UpdateTestSuite(id string, updateDTO *dto.TestSuiteUpdateDTO) (*dto.TestSuiteResponseDTO, error) {
	args := m.Called(id, updateDTO)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TestSuiteResponseDTO), args.Error(1)
}

// テストケース構造体定義
type updateTestSuiteTestCase struct {
	name          string
	inputID       string
	input         *dto.TestSuiteUpdateDTO
	setupMock     func(*mockTestSuiteUseCase)
	expectedCode  int
	expectSuccess bool
}

func TestUpdateTestSuite(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	newName := "更新後のテストスイート名"
	endDate := fixedTime.AddDate(0, -1, 0)

	tests := []updateTestSuiteTestCase{
		{
			name:    "テストスイート更新成功",
			inputID: "TS001-202401",
			input: &dto.TestSuiteUpdateDTO{
				Name: &newName,
			},
			setupMock: func(m *mockTestSuiteUseCase) {
				expectedResponse := &dto.TestSuiteResponseDTO{
					ID:                   "TS001-202401",
					Name:                 newName,
					Status:               "準備中",
					EstimatedStartDate:   fixedTime,
					EstimatedEndDate:     fixedTime.AddDate(0, 1, 0),
					RequireEffortComment: true,
					Progress:             0.0,
					CreatedAt:            fixedTime,
					UpdatedAt:            fixedTime,
				}
				m.On("UpdateTestSuite", "TS001-202401", mock.AnythingOfType("*dto.TestSuiteUpdateDTO")).Return(expectedResponse, nil)
			},
			expectedCode:  http.StatusOK,
			expectSuccess: true,
		},
		{
			name:    "存在しないテストスイート",
			inputID: "NON-EXISTENT",
			input: &dto.TestSuiteUpdateDTO{
				Name: &newName,
			},
			setupMock: func(m *mockTestSuiteUseCase) {
				m.On("UpdateTestSuite", "NON-EXISTENT", mock.AnythingOfType("*dto.TestSuiteUpdateDTO")).Return(nil, entity.ErrTestSuiteNotFound)
			},
			expectedCode:  http.StatusNotFound,
			expectSuccess: false,
		},
		{
			name:    "楽観的ロックエラー",
			inputID: "TS001-202401",
			input: &dto.TestSuiteUpdateDTO{
				Name: &newName,
			},
			setupMock: func(m *mockTestSuiteUseCase) {
				m.On("UpdateTestSuite", "TS001-202401", mock.AnythingOfType("*dto.TestSuiteUpdateDTO")).Return(nil, entity.ErrConcurrentModification)
			},
			expectedCode:  http.StatusConflict,
			expectSuccess: false,
		},
		{
			name:    "バリデーションエラー（不正な日付範囲）",
			inputID: "TS001-202401",
			input: &dto.TestSuiteUpdateDTO{
				EstimatedStartDate: &fixedTime,
				EstimatedEndDate:   &endDate, // 開始日より前の終了日
			},
			setupMock: func(m *mockTestSuiteUseCase) {
				// バリデーションエラーのためモック不要
			},
			expectedCode:  http.StatusBadRequest,
			expectSuccess: false,
		},
		// UpdateTestSuiteのテストケースを修正
		{
			name:    "空のリクエストボディ",
			inputID: "TS001-202401",
			input:   &dto.TestSuiteUpdateDTO{}, // nilではなく空の構造体を使用
			setupMock: func(m *mockTestSuiteUseCase) {
				// モックの期待値を設定
				expectedResponse := &dto.TestSuiteResponseDTO{
					ID:                   "TS001-202401",
					Name:                 "既存のテストスイート名", // 既存の値
					Status:               "準備中",
					EstimatedStartDate:   fixedTime,
					EstimatedEndDate:     fixedTime.AddDate(0, 1, 0),
					RequireEffortComment: true,
					Progress:             0.0,
					CreatedAt:            fixedTime,
					UpdatedAt:            fixedTime,
				}
				m.On("UpdateTestSuite", "TS001-202401", mock.AnythingOfType("*dto.TestSuiteUpdateDTO")).Return(expectedResponse, nil)
			},
			expectedCode:  http.StatusOK,
			expectSuccess: true,
		},
		{
			name:    "不正なJSONフォーマット",
			inputID: "TS001-202401",
			input:   nil,
			setupMock: func(m *mockTestSuiteUseCase) {
				// バリデーションエラーのためモック不要
			},
			expectedCode:  http.StatusBadRequest,
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUseCase := new(mockTestSuiteUseCase)
			tt.setupMock(mockUseCase)
			handler := NewTestSuiteHandler(mockUseCase)

			var jsonData []byte
			var err error
			if tt.input != nil {
				jsonData, err = json.Marshal(tt.input)
				assert.NoError(t, err)
			} else {
				// 不正なJSONの場合
				jsonData = []byte(`{invalid json}`)
			}

			req, err := http.NewRequest("PUT", "/test-suites/"+tt.inputID, bytes.NewBuffer(jsonData))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			vars := map[string]string{
				"id": tt.inputID,
			}
			req = mux.SetURLVars(req, vars)

			w := httptest.NewRecorder()
			handler.Update(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)

			if tt.expectSuccess {
				var response dto.TestSuiteResponseDTO
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, tt.inputID, response.ID)
				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
			}

			mockUseCase.AssertExpectations(t)
		})
	}
}

func (m *mockTestSuiteUseCase) UpdateTestSuiteStatus(id string, statusDTO *dto.TestSuiteStatusUpdateDTO) (*dto.TestSuiteResponseDTO, error) {
	args := m.Called(id, statusDTO)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TestSuiteResponseDTO), args.Error(1)
}

// テストケース構造体定義
type updateStatusTestCase struct {
	name          string
	inputID       string
	input         *dto.TestSuiteStatusUpdateDTO
	setupMock     func(*mockTestSuiteUseCase)
	expectedCode  int
	expectSuccess bool
}

func TestUpdateStatus(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	tests := []updateStatusTestCase{
		{
			name:    "ステータス更新成功",
			inputID: "TS001-202401",
			input: &dto.TestSuiteStatusUpdateDTO{
				Status: "実行中",
			},
			setupMock: func(m *mockTestSuiteUseCase) {
				expectedResponse := &dto.TestSuiteResponseDTO{
					ID:                   "TS001-202401",
					Name:                 "テストスイート1",
					Status:               "実行中",
					EstimatedStartDate:   fixedTime,
					EstimatedEndDate:     fixedTime.AddDate(0, 1, 0),
					RequireEffortComment: true,
					Progress:             50.0,
					CreatedAt:            fixedTime,
					UpdatedAt:            fixedTime,
				}
				m.On("UpdateTestSuiteStatus", "TS001-202401", mock.AnythingOfType("*dto.TestSuiteStatusUpdateDTO")).Return(expectedResponse, nil)
			},
			expectedCode:  http.StatusOK,
			expectSuccess: true,
		},
		{
			name:    "存在しないテストスイート",
			inputID: "NON-EXISTENT",
			input: &dto.TestSuiteStatusUpdateDTO{
				Status: "実行中",
			},
			setupMock: func(m *mockTestSuiteUseCase) {
				m.On("UpdateTestSuiteStatus", "NON-EXISTENT", mock.AnythingOfType("*dto.TestSuiteStatusUpdateDTO")).Return(nil, entity.ErrTestSuiteNotFound)
			},
			expectedCode:  http.StatusNotFound,
			expectSuccess: false,
		},
		{
			name:    "無効なステータス遷移",
			inputID: "TS001-202401",
			input: &dto.TestSuiteStatusUpdateDTO{
				Status: "完了",
			},
			setupMock: func(m *mockTestSuiteUseCase) {
				m.On("UpdateTestSuiteStatus", "TS001-202401", mock.AnythingOfType("*dto.TestSuiteStatusUpdateDTO")).Return(nil, entity.ErrInvalidStatusTransition)
			},
			expectedCode:  http.StatusBadRequest,
			expectSuccess: false,
		},
		{
			name:    "不正なステータス値",
			inputID: "TS001-202401",
			input: &dto.TestSuiteStatusUpdateDTO{
				Status: "不正な値",
			},
			setupMock: func(m *mockTestSuiteUseCase) {
				// バリデーションエラーのためモック不要
			},
			expectedCode:  http.StatusBadRequest,
			expectSuccess: false,
		},
		{
			name:    "ステータス未指定",
			inputID: "TS001-202401",
			input:   &dto.TestSuiteStatusUpdateDTO{},
			setupMock: func(m *mockTestSuiteUseCase) {
				// バリデーションエラーのためモック不要
			},
			expectedCode:  http.StatusBadRequest,
			expectSuccess: false,
		},
		{
			name:    "同じステータスへの更新",
			inputID: "TS001-202401",
			input: &dto.TestSuiteStatusUpdateDTO{
				Status: "準備中",
			},
			setupMock: func(m *mockTestSuiteUseCase) {
				expectedResponse := &dto.TestSuiteResponseDTO{
					ID:     "TS001-202401",
					Status: "準備中",
					// 他のフィールドは変更なし
				}
				m.On("UpdateTestSuiteStatus", "TS001-202401", mock.AnythingOfType("*dto.TestSuiteStatusUpdateDTO")).Return(expectedResponse, nil)
			},
			expectedCode:  http.StatusOK,
			expectSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUseCase := new(mockTestSuiteUseCase)
			tt.setupMock(mockUseCase)
			handler := NewTestSuiteHandler(mockUseCase)

			jsonData, err := json.Marshal(tt.input)
			assert.NoError(t, err)

			req, err := http.NewRequest("PATCH", "/test-suites/"+tt.inputID+"/status", bytes.NewBuffer(jsonData))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			vars := map[string]string{
				"id": tt.inputID,
			}
			req = mux.SetURLVars(req, vars)

			w := httptest.NewRecorder()
			handler.UpdateStatus(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)

			if tt.expectSuccess {
				var response dto.TestSuiteResponseDTO
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, tt.inputID, response.ID)
				assert.Equal(t, tt.input.Status, response.Status)
				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
			}

			mockUseCase.AssertExpectations(t)
		})
	}
}
