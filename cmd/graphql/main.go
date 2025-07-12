// cmd/graphql/main.go の修正版（エラー修正済み）

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/FUJI0130/go-ddd-ca/internal/infrastructure/auth"
	"github.com/FUJI0130/go-ddd-ca/internal/infrastructure/persistence/postgres"
	graphqlauth "github.com/FUJI0130/go-ddd-ca/internal/interface/graphql/auth"
	"github.com/FUJI0130/go-ddd-ca/internal/interface/graphql/dataloader"
	"github.com/FUJI0130/go-ddd-ca/internal/interface/graphql/generated"
	"github.com/FUJI0130/go-ddd-ca/internal/interface/graphql/resolver"
	"github.com/FUJI0130/go-ddd-ca/internal/usecase/interactor"
	"github.com/FUJI0130/go-ddd-ca/pkg/config"
	"github.com/FUJI0130/go-ddd-ca/support/customerrors"
	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
	"golang.org/x/crypto/bcrypt" // bcrypt.DefaultCostに必要
)

const defaultPort = "8080"

func main() {
	// 設定の読み込み
	cfg, err := config.LoadConfig("./configs")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// 環境情報のログ出力
	log.Printf("Starting GraphQL server in %s environment", cfg.Environment)

	// ポート設定（環境変数 > 設定ファイル > デフォルト値）
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	// 環境変数があれば優先（既存の挙動を維持）
	if envUser := os.Getenv("DB_USER"); envUser != "" {
		cfg.Database.User = envUser
	}
	if envPass := os.Getenv("DB_PASS"); envPass != "" {
		cfg.Database.Password = envPass
	}
	if envHost := os.Getenv("DB_HOST"); envHost != "" {
		cfg.Database.Host = envHost
	}
	if envPort := os.Getenv("DB_PORT"); envPort != "" {
		portNum, err := strconv.Atoi(envPort)
		if err == nil {
			cfg.Database.Port = portNum
		}
	}
	if envName := os.Getenv("DB_NAME"); envName != "" {
		cfg.Database.DBName = envName
	}

	// データベース接続（新しいファクトリを使用）
	db, err := cfg.NewDatabaseConnection()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// 接続確認のログ（既存コードを維持）
	log.Println("Successfully connected to database")

	// コネクションプールの設定
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// 接続確認
	if err = db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Successfully connected to database")

	// リポジトリの初期化
	testSuiteRepo := postgres.NewTestSuiteRepository(db)
	testGroupRepo := postgres.NewTestGroupRepository(db)
	testCaseRepo := postgres.NewTestCaseRepository(db)
	userRepo := postgres.NewUserRepository(db)
	refreshTokenRepo := postgres.NewPostgresRefreshTokenRepository(db)

	// IDジェネレーターの初期化
	testSuiteIDGenerator := postgres.NewTestSuiteIDGenerator(db)
	testGroupIDGenerator := postgres.NewTestGroupIDGenerator(db)
	testCaseIDGenerator := postgres.NewTestCaseIDGenerator(db)
	userIDGenerator := postgres.NewUserIDGenerator(db)

	// 認証関連サービスの初期化
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "default-jwt-secret" // 開発環境用デフォルト値
		log.Println("WARNING: Using default JWT secret. Set JWT_SECRET environment variable in production!")
	}

	accessTokenDuration := 24 * time.Hour      // アクセストークンの有効期間: 1日
	refreshTokenDuration := 7 * 24 * time.Hour // リフレッシュトークンの有効期間: 1週間

	// パスワードサービスとJWTサービスの初期化
	passwordService := auth.NewBCryptPasswordService(bcrypt.DefaultCost)
	jwtService := auth.NewJWTServiceWithRepo(
		jwtSecret,
		accessTokenDuration,
		refreshTokenDuration,
		refreshTokenRepo,
	)

	// ユースケースの初期化
	testSuiteUseCase := interactor.NewTestSuiteInteractor(testSuiteRepo, testSuiteIDGenerator)
	testGroupUseCase := interactor.NewTestGroupInteractor(testGroupRepo, testGroupIDGenerator)
	testCaseUseCase := interactor.NewTestCaseInteractor(testCaseRepo, testCaseIDGenerator)
	authUseCase := interactor.NewAuthInteractor(
		userRepo,
		jwtService,
		passwordService,
		refreshTokenRepo,
	)

	// UserManagementUseCaseのインスタンス作成
	userManagementInteractor := interactor.NewUserManagementInteractor(
		userRepo,
		userIDGenerator,
		passwordService,
	)

	// DataLoaderの初期化
	loaders := dataloader.NewDataLoaders(testGroupUseCase, testCaseUseCase)

	// リゾルバーの初期化（AuthUseCaseとUserManagementUseCaseを追加）
	resolverObj := resolver.NewResolver(
		testSuiteUseCase,
		testGroupUseCase,
		testCaseUseCase,
		authUseCase,
		userManagementInteractor,
	)

	// GraphQLサーバーの設定
	c := generated.Config{
		Resolvers: resolverObj,
	}

	// ディレクティブを登録（コードが生成された後のDirectiveRoot構造体に合わせる）
	c.Directives.Auth = func(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error) {
		// 認証済みかチェック
		user := graphqlauth.GetUserFromContext(ctx)
		if user == nil {
			return nil, customerrors.NewUnauthorizedError("認証が必要です")
		}
		return next(ctx)
	}

	c.Directives.HasRole = func(ctx context.Context, obj interface{}, next graphql.Resolver, role string) (res interface{}, err error) {
		// ロールをチェック
		user := graphqlauth.GetUserFromContext(ctx)
		if user == nil {
			return nil, customerrors.NewUnauthorizedError("認証が必要です")
		}
		if user.Role.String() != role {
			return nil, customerrors.NewForbiddenError("アクセス権限がありません")
		}
		return next(ctx)
	}

	// サーバーの作成
	srv := handler.New(generated.NewExecutableSchema(c))

	// トランスポート設定
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.MultipartForm{})
	srv.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	})

	// クエリの複雑さ制限などの拡張機能
	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100), // 型パラメータを明示
	})

	// 認証ミドルウェアの初期化
	authMiddleware := graphqlauth.AuthMiddleware(authUseCase)

	// ヘルスチェックエンドポイントを追加 - ALBのヘルスチェック用
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"graphql"}`))
		log.Printf("Health check requested from %s", r.RemoteAddr)
	})

	// GraphQL PlaygroundのUIを設定（DataLoaderミドルウェアと認証ミドルウェアを追加）
	// http.Handle("/", dataloader.Middleware(loaders, playground.Handler("GraphQL playground", "/query")))
	// http.Handle("/query", authMiddleware(dataloader.Middleware(loaders, srv)))

	// CORS設定
	corsHandler := cors.New(cors.Options{
		AllowedOrigins: []string{
			"http://localhost:3000",                 // 開発環境
			"https://example-frontend.cloudfront.net", // 本番フロントエンド
		},
		AllowCredentials: true,
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
	})

	// レスポンスライターミドルウェア
	responseWriterMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := graphqlauth.WithResponseWriter(r.Context(), w)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}

	// ハンドラー設定
	http.Handle("/", dataloader.Middleware(loaders, playground.Handler("GraphQL playground", "/query")))
	http.Handle("/query", corsHandler.Handler(responseWriterMiddleware(authMiddleware(dataloader.Middleware(loaders, srv)))))

	log.Fatal(http.ListenAndServe(":"+port, nil))

	log.Printf("Connect to http://localhost:%s/ for GraphQL playground", port)
	log.Printf("Health check endpoint available at http://localhost:%s/health", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
