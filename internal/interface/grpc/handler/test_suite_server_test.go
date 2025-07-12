package handler

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/FUJI0130/go-ddd-ca/internal/usecase/dto"
	pb "github.com/FUJI0130/go-ddd-ca/proto/testsuite/v1"
)

// MockTestSuiteInteractor はテスト用のモックインタラクター
type MockTestSuiteInteractor struct {
	mock.Mock
}

func (m *MockTestSuiteInteractor) ListTestSuites(ctx context.Context, params *dto.TestSuiteQueryParamDTO) (*dto.TestSuiteListResponseDTO, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TestSuiteListResponseDTO), args.Error(1)
}

// すでにあるListTestSuitesの上に以下を追加
func (m *MockTestSuiteInteractor) CreateTestSuite(ctx context.Context, createDTO *dto.TestSuiteCreateDTO) (*dto.TestSuiteResponseDTO, error) {
	args := m.Called(ctx, createDTO)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TestSuiteResponseDTO), args.Error(1)
}

func (m *MockTestSuiteInteractor) GetTestSuite(ctx context.Context, id string) (*dto.TestSuiteResponseDTO, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TestSuiteResponseDTO), args.Error(1)
}

func (m *MockTestSuiteInteractor) UpdateTestSuite(ctx context.Context, id string, updateDTO *dto.TestSuiteUpdateDTO) (*dto.TestSuiteResponseDTO, error) {
	args := m.Called(ctx, id, updateDTO)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TestSuiteResponseDTO), args.Error(1)
}

func (m *MockTestSuiteInteractor) UpdateTestSuiteStatus(ctx context.Context, id string, statusDTO *dto.TestSuiteStatusUpdateDTO) (*dto.TestSuiteResponseDTO, error) {
	args := m.Called(ctx, id, statusDTO)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TestSuiteResponseDTO), args.Error(1)
}

func TestListTestSuites(t *testing.T) {
	testCases := []struct {
		name          string
		setupMock     func(*MockTestSuiteInteractor)
		request       *pb.ListTestSuitesRequest
		expectedLen   int
		expectedTotal int32
		expectedError bool
	}{
		{
			name: "全件取得",
			setupMock: func(m *MockTestSuiteInteractor) {
				m.On("ListTestSuites", mock.Anything, mock.AnythingOfType("*dto.TestSuiteQueryParamDTO")).
					Return(&dto.TestSuiteListResponseDTO{
						TestSuites: []dto.TestSuiteResponseDTO{
							{
								ID:     "TS001-202501",
								Name:   "テストスイート1",
								Status: "準備中",
							},
							{
								ID:     "TS002-202501",
								Name:   "テストスイート2",
								Status: "実行中",
							},
						},
						Total: 2,
					}, nil)
			},
			request:       &pb.ListTestSuitesRequest{},
			expectedLen:   2,
			expectedTotal: 2,
			expectedError: false,
		},
		{
			name: "ステータスフィルター",
			setupMock: func(m *MockTestSuiteInteractor) {
				m.On("ListTestSuites",
					mock.Anything, // mock.AnythingOfType("*context.emptyCtx") から修正
					mock.MatchedBy(func(params *dto.TestSuiteQueryParamDTO) bool {
						if params.Status == nil {
							return false
						}
						return *params.Status == "実行中"
					}),
				).Return(&dto.TestSuiteListResponseDTO{
					TestSuites: []dto.TestSuiteResponseDTO{
						{
							ID:     "TS002-202501",
							Name:   "テストスイート2",
							Status: "実行中",
						},
					},
					Total: 1,
				}, nil)
			},
			request: &pb.ListTestSuitesRequest{
				Status: pb.SuiteStatus_SUITE_STATUS_IN_PROGRESS.Enum(), // .Enum()を使用してポインタを取得
			},
			expectedLen:   1,
			expectedTotal: 1,
			expectedError: false,
		},
		{
			name: "日付範囲指定",
			setupMock: func(m *MockTestSuiteInteractor) {
				m.On("ListTestSuites", mock.Anything, mock.MatchedBy(func(params *dto.TestSuiteQueryParamDTO) bool {
					return params.StartDate != nil && params.EndDate != nil
				})).Return(&dto.TestSuiteListResponseDTO{
					TestSuites: []dto.TestSuiteResponseDTO{
						{
							ID:     "TS001-202501",
							Name:   "テストスイート1",
							Status: "準備中",
						},
					},
					Total: 1,
				}, nil)
			},
			request: &pb.ListTestSuitesRequest{
				StartDate: timestamppb.New(time.Now().AddDate(0, -1, 0)),
				EndDate:   timestamppb.New(time.Now()),
			},
			expectedLen:   1,
			expectedTotal: 1,
			expectedError: false,
		},
		{
			name: "エラーケース",
			setupMock: func(m *MockTestSuiteInteractor) {
				m.On("ListTestSuites", mock.Anything, mock.AnythingOfType("*dto.TestSuiteQueryParamDTO")).
					Return(nil, assert.AnError)
			},
			request:       &pb.ListTestSuitesRequest{},
			expectedLen:   0,
			expectedTotal: 0,
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// モックインタラクターのセットアップ
			mockInteractor := new(MockTestSuiteInteractor)
			tc.setupMock(mockInteractor)

			// サーバーの作成
			server := NewTestSuiteServer(mockInteractor)

			// テストの実行
			response, err := server.ListTestSuites(context.Background(), tc.request)

			// 検証
			if tc.expectedError {
				assert.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.Equal(t, tc.expectedLen, len(response.TestSuites))
				assert.Equal(t, tc.expectedTotal, response.Total)
			}

			// モックの呼び出しを検証
			mockInteractor.AssertExpectations(t)
		})
	}
}

// mockStreamはテスト用のストリームモック
type mockStream struct {
	mock.Mock
	ctx context.Context
}

func (m *mockStream) Send(ts *pb.TestSuite) error {
	args := m.Called(ts)
	return args.Error(0)
}

func (m *mockStream) Context() context.Context {
	return m.ctx
}

// grpc.ServerStreamインターフェースの実装
func (m *mockStream) SetHeader(md metadata.MD) error {
	args := m.Called(md)
	return args.Error(0)
}

func (m *mockStream) SendHeader(md metadata.MD) error {
	args := m.Called(md)
	return args.Error(0)
}

func (m *mockStream) SetTrailer(md metadata.MD) {
	m.Called(md)
}

func (m *mockStream) SendMsg(msg interface{}) error {
	args := m.Called(msg)
	return args.Error(0)
}

func (m *mockStream) RecvMsg(msg interface{}) error {
	args := m.Called(msg)
	return args.Error(0)
}

func TestWatchTestSuite(t *testing.T) {
	testCases := []struct {
		name          string
		setupMock     func(*MockTestSuiteInteractor, *mockStream)
		setupContext  func() (context.Context, context.CancelFunc)
		request       *pb.GetTestSuiteRequest
		expectedError bool
		interval      time.Duration
	}{
		{
			name: "正常系：初期データの送信成功",
			setupMock: func(m *MockTestSuiteInteractor, stream *mockStream) {
				initialSuite := &dto.TestSuiteResponseDTO{
					ID:        "TS001-202501",
					Name:      "テストスイート1",
					Status:    "準備中",
					UpdatedAt: time.Now(),
				}
				m.On("GetTestSuite", mock.Anything, "TS001-202501").
					Return(initialSuite, nil)

				stream.On("Send", mock.MatchedBy(func(ts *pb.TestSuite) bool {
					return ts.Id == "TS001-202501" && ts.Name == "テストスイート1"
				})).Return(nil)
			},
			setupContext: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), 50*time.Millisecond)
			},
			request: &pb.GetTestSuiteRequest{
				Id: "TS001-202501",
			},
			expectedError: false,
			interval:      10 * time.Millisecond,
		},
		{
			name: "異常系：GetTestSuiteエラー",
			setupMock: func(m *MockTestSuiteInteractor, stream *mockStream) {
				m.On("GetTestSuite", mock.Anything, "TS001-202501").
					Return(nil, assert.AnError)
			},
			setupContext: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), 50*time.Millisecond)
			},
			request: &pb.GetTestSuiteRequest{
				Id: "TS001-202501",
			},
			expectedError: true,
			interval:      10 * time.Millisecond,
		},
		{
			name: "異常系：Stream.Sendエラー",
			setupMock: func(m *MockTestSuiteInteractor, stream *mockStream) {
				initialSuite := &dto.TestSuiteResponseDTO{
					ID:        "TS001-202501",
					Name:      "テストスイート1",
					Status:    "準備中",
					UpdatedAt: time.Now(),
				}
				m.On("GetTestSuite", mock.Anything, "TS001-202501").
					Return(initialSuite, nil)

				stream.On("Send", mock.Anything).Return(assert.AnError)
			},
			setupContext: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), 50*time.Millisecond)
			},
			request: &pb.GetTestSuiteRequest{
				Id: "TS001-202501",
			},
			expectedError: true,
			interval:      10 * time.Millisecond,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockInteractor := new(MockTestSuiteInteractor)
			ctx, cancel := tc.setupContext()
			defer cancel()

			mockStream := &mockStream{ctx: ctx}
			tc.setupMock(mockInteractor, mockStream)

			// オプション付きでサーバーを作成
			server := NewTestSuiteServerWithOption(mockInteractor, &TestSuiteServerOption{
				watchInterval: tc.interval,
			})

			err := server.WatchTestSuite(tc.request, mockStream)

			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockInteractor.AssertExpectations(t)
			mockStream.AssertExpectations(t)
		})
	}
}
