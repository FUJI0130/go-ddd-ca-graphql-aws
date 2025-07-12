package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/FUJI0130/go-ddd-ca/internal/domain/repository"
	"github.com/FUJI0130/go-ddd-ca/internal/infrastructure/persistence/postgres"
	"github.com/FUJI0130/go-ddd-ca/internal/interface/grpc/handler"
	"github.com/FUJI0130/go-ddd-ca/internal/interface/grpc/server"
	"github.com/FUJI0130/go-ddd-ca/internal/usecase/interactor"
	"github.com/FUJI0130/go-ddd-ca/pkg/config"

	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	// 設定の読み込み
	cfg, err := config.LoadConfig("./configs")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// 環境情報のログ出力
	log.Printf("Starting gRPC server in %s environment", cfg.Environment)

	// データベース接続（新しいファクトリを使用）
	db, err := cfg.NewDatabaseConnection()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// 接続確認のログ
	log.Println("Successfully connected to database")

	// リポジトリの作成
	var testSuiteRepo repository.TestSuiteRepository
	testSuiteRepo = postgres.NewTestSuiteRepository(db)

	// IDジェネレーターの初期化
	testSuiteIDGenerator := postgres.NewTestSuiteIDGenerator(db)

	// インタラクターの作成（IDジェネレーターを追加）
	testSuiteInteractor := interactor.NewTestSuiteInteractor(testSuiteRepo, testSuiteIDGenerator)

	// gRPCハンドラーの作成
	testSuiteServer := handler.NewTestSuiteServer(testSuiteInteractor)

	// gRPCサーバーの設定
	grpcServer, err := server.NewGrpcServer(50051, testSuiteServer)
	if err != nil {
		log.Fatalf("Failed to create gRPC server: %v", err)
	}

	// 追加: gRPC Health Serviceの登録
	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer.GetServer(), healthServer)

	// サービス全体の状態を「稼働中」に設定
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	// 特定のサービスの状態を設定（必要に応じてカスタマイズ）
	healthServer.SetServingStatus("testsuite.v1.TestSuiteService", healthpb.HealthCheckResponse_SERVING)

	// 追加: ALBのgRPCネイティブヘルスチェック用のサービス名を設定
	healthServer.SetServingStatus("grpc.health.v1.Health", healthpb.HealthCheckResponse_SERVING)

	log.Println("Registered gRPC Health Service")

	// HTTPヘルスチェックサーバーの追加
	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8080"
	}

	// 簡易的なHTTPサーバーを起動
	go func() {
		// 既存の/healthエンドポイント
		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok","service":"grpc"}`))
			log.Printf("Health check requested from %s", r.RemoteAddr)
		})

		// 追加: /health-httpエンドポイント（Terraformの設定に合わせる）
		http.HandleFunc("/health-http", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok","service":"grpc"}`))
			log.Printf("Health check requested from %s (health-http)", r.RemoteAddr)
		})

		log.Printf("Starting HTTP health check server on port %s", httpPort)
		if err := http.ListenAndServe(":"+httpPort, nil); err != nil {
			log.Printf("HTTP health check server error: %v", err)
		}
	}()

	// シグナルハンドリングの設定
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// サーバー起動
	go func() {
		log.Println("Starting gRPC server on port 50051")
		if err := grpcServer.Start(); err != nil {
			log.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()

	// シグナルを待つ
	<-ctx.Done()
	log.Println("Shutting down gRPC server...")
	grpcServer.Stop()
	log.Println("gRPC server stopped")
}
