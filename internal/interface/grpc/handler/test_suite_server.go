package handler

import (
	"context"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/FUJI0130/go-ddd-ca/internal/usecase/dto"
	"github.com/FUJI0130/go-ddd-ca/pkg/errors"
	pb "github.com/FUJI0130/go-ddd-ca/proto/testsuite/v1"
)

// TestSuiteServerの前に追加
type TestSuiteInteractorInterface interface {
	CreateTestSuite(ctx context.Context, createDTO *dto.TestSuiteCreateDTO) (*dto.TestSuiteResponseDTO, error)
	GetTestSuite(ctx context.Context, id string) (*dto.TestSuiteResponseDTO, error)
	UpdateTestSuite(ctx context.Context, id string, updateDTO *dto.TestSuiteUpdateDTO) (*dto.TestSuiteResponseDTO, error)
	UpdateTestSuiteStatus(ctx context.Context, id string, statusDTO *dto.TestSuiteStatusUpdateDTO) (*dto.TestSuiteResponseDTO, error)
	ListTestSuites(ctx context.Context, params *dto.TestSuiteQueryParamDTO) (*dto.TestSuiteListResponseDTO, error)
}

type TestSuiteServer struct {
	pb.UnimplementedTestSuiteServiceServer
	interactor TestSuiteInteractorInterface
	option     TestSuiteServerOption
}

// TestSuiteServerにオプション構造体を追加
type TestSuiteServerOption struct {
	watchInterval time.Duration
}

// デフォルト値の設定
var defaultTestSuiteServerOption = TestSuiteServerOption{
	watchInterval: 10 * time.Second,
}

// NewTestSuiteServerWithOptionを追加
func NewTestSuiteServerWithOption(interactor TestSuiteInteractorInterface, opt *TestSuiteServerOption) *TestSuiteServer {
	option := defaultTestSuiteServerOption
	if opt != nil {
		option = *opt
	}
	return &TestSuiteServer{
		interactor: interactor,
		option:     option,
	}
}

func NewTestSuiteServer(interactor TestSuiteInteractorInterface) *TestSuiteServer {
	return &TestSuiteServer{
		interactor: interactor,
		option:     defaultTestSuiteServerOption,
	}
}

func (s *TestSuiteServer) CreateTestSuite(ctx context.Context, req *pb.CreateTestSuiteRequest) (*pb.TestSuite, error) {
	// プロトコルバッファからDTOへの変換
	createDTO := &dto.TestSuiteCreateDTO{
		Name:                 req.GetName(),
		Description:          req.GetDescription(),
		EstimatedStartDate:   req.GetEstimatedStartDate().AsTime(),
		EstimatedEndDate:     req.GetEstimatedEndDate().AsTime(),
		RequireEffortComment: req.GetRequireEffortComment(),
	}

	// インタラクターを呼び出し
	result, err := s.interactor.CreateTestSuite(ctx, createDTO)
	if err != nil {
		// 標準化されたエラー変換を使用
		return nil, errors.ToGRPCError(err)
	}

	// DTOからプロトコルバッファへの変換
	return s.toProtoTestSuite(result), nil
}

// toProtoTestSuite はDTOからプロトコルバッファへの変換
func (s *TestSuiteServer) toProtoTestSuite(dto *dto.TestSuiteResponseDTO) *pb.TestSuite {
	status := pb.SuiteStatus_SUITE_STATUS_UNSPECIFIED
	switch dto.Status {
	case "準備中":
		status = pb.SuiteStatus_SUITE_STATUS_PREPARATION
	case "実行中":
		status = pb.SuiteStatus_SUITE_STATUS_IN_PROGRESS
	case "完了":
		status = pb.SuiteStatus_SUITE_STATUS_COMPLETED
	case "中断":
		status = pb.SuiteStatus_SUITE_STATUS_SUSPENDED
	}

	return &pb.TestSuite{
		Id:                   dto.ID,
		Name:                 dto.Name,
		Description:          dto.Description,
		Status:               status,
		EstimatedStartDate:   timestamppb.New(dto.EstimatedStartDate),
		EstimatedEndDate:     timestamppb.New(dto.EstimatedEndDate),
		RequireEffortComment: dto.RequireEffortComment,
		Progress:             float32(dto.Progress),
		CreatedAt:            timestamppb.New(dto.CreatedAt),
		UpdatedAt:            timestamppb.New(dto.UpdatedAt),
	}
}

// GetTestSuite は指定されたIDのテストスイートを取得します
func (s *TestSuiteServer) GetTestSuite(ctx context.Context, req *pb.GetTestSuiteRequest) (*pb.TestSuite, error) {
	result, err := s.interactor.GetTestSuite(ctx, req.GetId())
	if err != nil {
		// ドメインエラーをgRPCエラーに変換
		return nil, errors.ToGRPCError(err)
	}

	return s.toProtoTestSuite(result), nil
}

// UpdateTestSuite はテストスイートを更新します
func (s *TestSuiteServer) UpdateTestSuite(ctx context.Context, req *pb.UpdateTestSuiteRequest) (*pb.TestSuite, error) {
	params := req.GetParams()
	updateDTO := &dto.TestSuiteUpdateDTO{}

	if params.Name != nil {
		name := params.GetName()
		updateDTO.Name = &name
	}

	if params.Description != nil {
		desc := params.GetDescription()
		updateDTO.Description = &desc
	}

	if params.EstimatedStartDate != nil {
		date := params.GetEstimatedStartDate().AsTime()
		updateDTO.EstimatedStartDate = &date
	}

	if params.EstimatedEndDate != nil {
		date := params.GetEstimatedEndDate().AsTime()
		updateDTO.EstimatedEndDate = &date
	}

	if params.RequireEffortComment != nil {
		reqComment := params.GetRequireEffortComment()
		updateDTO.RequireEffortComment = &reqComment
	}

	result, err := s.interactor.UpdateTestSuite(ctx, req.GetId(), updateDTO)
	if err != nil {
		// ドメインエラーをgRPCエラーに変換
		return nil, errors.ToGRPCError(err)
	}

	return s.toProtoTestSuite(result), nil
}

// UpdateTestSuiteStatus はテストスイートのステータスを更新します
func (s *TestSuiteServer) UpdateTestSuiteStatus(ctx context.Context, req *pb.UpdateTestSuiteStatusRequest) (*pb.TestSuite, error) {
	statusMap := map[pb.SuiteStatus]string{
		pb.SuiteStatus_SUITE_STATUS_PREPARATION: "準備中",
		pb.SuiteStatus_SUITE_STATUS_IN_PROGRESS: "実行中",
		pb.SuiteStatus_SUITE_STATUS_COMPLETED:   "完了",
		pb.SuiteStatus_SUITE_STATUS_SUSPENDED:   "中断",
	}

	statusDTO := &dto.TestSuiteStatusUpdateDTO{
		Status: statusMap[req.GetStatus()],
	}

	result, err := s.interactor.UpdateTestSuiteStatus(ctx, req.GetId(), statusDTO)
	if err != nil {
		// ドメインエラーをgRPCエラーに変換
		return nil, errors.ToGRPCError(err)
	}

	return s.toProtoTestSuite(result), nil
}

// ListTestSuites はテストスイート一覧を取得します
func (s *TestSuiteServer) ListTestSuites(ctx context.Context, req *pb.ListTestSuitesRequest) (*pb.ListTestSuitesResponse, error) {
	// Protocol BuffersのリクエストをDTOに変換
	queryDTO := &dto.TestSuiteQueryParamDTO{}

	// ステータスの変換（ポインタの比較に修正）
	if req.GetStatus() != pb.SuiteStatus_SUITE_STATUS_UNSPECIFIED {
		status := statusProtoToString(req.GetStatus())
		queryDTO.Status = &status
	}

	if req.StartDate != nil {
		startDate := req.GetStartDate().AsTime()
		queryDTO.StartDate = &startDate
	}

	if req.EndDate != nil {
		endDate := req.GetEndDate().AsTime()
		queryDTO.EndDate = &endDate
	}

	// ページネーション情報の設定
	page := int(req.GetPage())
	if page > 0 {
		queryDTO.Page = &page
	}

	pageSize := int(req.GetPageSize())
	if pageSize > 0 {
		queryDTO.PageSize = &pageSize
	}

	// インタラクターを呼び出し
	result, err := s.interactor.ListTestSuites(ctx, queryDTO)
	if err != nil {
		// ドメインエラーをgRPCエラーに変換
		return nil, errors.ToGRPCError(err)
	}

	// レスポンスの変換と作成
	response := &pb.ListTestSuitesResponse{
		TestSuites: make([]*pb.TestSuite, 0, len(result.TestSuites)),
		Total:      int32(result.Total),
	}

	for _, suite := range result.TestSuites {
		pbSuite := s.toProtoTestSuite(&suite)
		response.TestSuites = append(response.TestSuites, pbSuite)
	}

	return response, nil
}

// statusProtoToString はProtocol BuffersのステータスをString型に変換します
func statusProtoToString(status pb.SuiteStatus) string {
	switch status {
	case pb.SuiteStatus_SUITE_STATUS_PREPARATION:
		return "準備中"
	case pb.SuiteStatus_SUITE_STATUS_IN_PROGRESS:
		return "実行中"
	case pb.SuiteStatus_SUITE_STATUS_COMPLETED:
		return "完了"
	case pb.SuiteStatus_SUITE_STATUS_SUSPENDED:
		return "中断"
	default:
		return "準備中" // デフォルト値
	}
}

// WatchTestSuite はテストスイートの変更を監視します
func (s *TestSuiteServer) WatchTestSuite(req *pb.GetTestSuiteRequest, stream pb.TestSuiteService_WatchTestSuiteServer) error {
	ctx := stream.Context()
	id := req.GetId()

	// 初期データの送信
	initialSuite, err := s.interactor.GetTestSuite(ctx, id)
	if err != nil {
		// ドメインエラーをgRPCエラーに変換
		return errors.ToGRPCError(err)
	}

	if err := stream.Send(s.toProtoTestSuite(initialSuite)); err != nil {
		return errors.ToGRPCError(
			errors.NewSystemError("初期状態の送信に失敗しました", err),
		)
	}

	// ここでコンテキストのキャンセルをチェック
	if ctx.Err() != nil {
		return nil // 初期データ送信後にキャンセルされた場合は正常終了
	}

	ticker := time.NewTicker(s.option.watchInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil // キャンセルは正常終了として扱う
		case <-ticker.C:
			currentSuite, err := s.interactor.GetTestSuite(ctx, id)
			if err != nil {
				return errors.ToGRPCError(err)
			}

			if currentSuite.UpdatedAt.After(initialSuite.UpdatedAt) {
				if err := stream.Send(s.toProtoTestSuite(currentSuite)); err != nil {
					return errors.ToGRPCError(
						errors.NewSystemError("更新状態の送信に失敗しました", err),
					)
				}
				initialSuite = currentSuite
			}
		}
	}
}
