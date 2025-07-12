package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"

	"github.com/FUJI0130/go-ddd-ca/internal/interface/api/handler"
	"github.com/FUJI0130/go-ddd-ca/pkg/config"
	"github.com/FUJI0130/go-ddd-ca/pkg/errors"
)

var db *sql.DB

// initDB はデータベース接続を初期化します
func initDB(dbConfig config.DatabaseConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		dbConfig.Host, dbConfig.Port, dbConfig.User, dbConfig.Password, dbConfig.DBName, dbConfig.SSLMode,
	)
	db, err := sql.Open(dbConfig.Driver, dsn)
	if err != nil {
		return nil, err
	}

	// コネクションプールの設定
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// 接続確認
	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func main() {
	// 設定の読み込み
	cfg, err := config.LoadConfig("./configs")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// 環境情報のログ出力
	log.Printf("Starting API server in %s environment", cfg.Environment)

	// デバッグ情報の追加 - サーバー起動直前
	if os.Getenv("DEBUG") == "true" || os.Getenv("APP_DEBUG") == "true" {
		log.Println("-------------------- 接続情報 --------------------")
		log.Printf("データベース接続文字列: %s (非表示のパスワード)",
			fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=%s",
				cfg.Database.Host, cfg.Database.Port, cfg.Database.User,
				cfg.Database.DBName, cfg.Database.SSLMode))
		log.Printf("サーバーアドレス: :%d", cfg.Server.Port)
		log.Printf("タイムアウト設定: 読込=%s, 書込=%s",
			cfg.Server.ReadTimeout, cfg.Server.WriteTimeout)
		log.Println("---------------------------------------------------")
	}

	// データベース接続の初期化（新しいファクトリを使用）
	db, err = cfg.NewDatabaseConnection()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	log.Println("Successfully connected to database")

	// バージョン情報（環境変数から取得またはデフォルト値）
	version := os.Getenv("APP_VERSION")
	if version == "" {
		version = "dev"
	}

	// ルーターの初期化
	router := mux.NewRouter()

	// ヘルスチェックハンドラーをルートレベルに追加
	healthHandler := handler.NewHealthHandler(version)
	router.HandleFunc("/health", healthHandler.Check).Methods(http.MethodGet)

	// APIバージョンのサブルーター
	apiRouter := router.PathPrefix("/api/v1").Subrouter()

	// ルートハンドラーの登録
	registerHandlers(apiRouter)

	// サーバーの設定
	srv := &http.Server{
		Handler:      router,
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		WriteTimeout: cfg.Server.WriteTimeout,
		ReadTimeout:  cfg.Server.ReadTimeout,
	}

	// サーバーを別のゴルーチンで起動
	go func() {
		log.Printf("Starting server on port %d", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// シグナル処理のためのチャネル
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// グレースフルシャットダウン
	log.Println("Server is shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}

// ルートハンドラーの登録
func registerHandlers(router *mux.Router) {
	// グローバルミドルウェアの設定
	router.Use(errors.ErrorHandler)
	router.Use(commonMiddleware)

	// バージョン情報（環境変数から取得またはデフォルト値）
	version := os.Getenv("APP_VERSION")
	if version == "" {
		version = "dev"
	}

	// ヘルスチェックハンドラー（ルートレベルに配置）
	// healthHandler := handler.NewHealthHandler(version)
	// router.HandleFunc("/health", healthHandler.Check).Methods(http.MethodGet)

	// 依存関係の注入は後で実装
	testSuiteHandler := handler.NewTestSuiteHandler(nil)
	router.HandleFunc("/test-suites", testSuiteHandler.Create).Methods(http.MethodPost)
}

// 共通のミドルウェア設定
func commonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
