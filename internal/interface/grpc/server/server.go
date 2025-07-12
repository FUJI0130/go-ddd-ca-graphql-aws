package server

import (
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/FUJI0130/go-ddd-ca/internal/interface/grpc/handler"
	pb "github.com/FUJI0130/go-ddd-ca/proto/testsuite/v1"
)

// GrpcServer はgRPCサーバーを表します
type GrpcServer struct {
	server   *grpc.Server
	listener net.Listener
}

// NewGrpcServer は新しいGrpcServerインスタンスを作成します
func NewGrpcServer(port int, testSuiteServer *handler.TestSuiteServer) (*GrpcServer, error) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %w", err)
	}

	server := grpc.NewServer()
	pb.RegisterTestSuiteServiceServer(server, testSuiteServer)

	// 開発用にリフレクションサービスを有効化
	reflection.Register(server)

	return &GrpcServer{
		server:   server,
		listener: lis,
	}, nil
}

// Start はgRPCサーバーを起動します
func (s *GrpcServer) Start() error {
	return s.server.Serve(s.listener)
}

// Stop はgRPCサーバーを停止します
func (s *GrpcServer) Stop() {
	s.server.GracefulStop()
}

// GetServer はgrpc.Serverインスタンスを返します
func (s *GrpcServer) GetServer() *grpc.Server {
	return s.server
}
