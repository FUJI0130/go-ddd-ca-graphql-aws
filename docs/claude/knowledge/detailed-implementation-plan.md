# テストケース管理システム 実装方針詳細（2024-01-05）

## 1. アーキテクチャ実装計画

### 1.1 REST API実装（フェーズ1）
#### エンドポイント設計
```go
// テストスイート管理
POST   /api/v1/test-suites               // スイート作成
GET    /api/v1/test-suites/:id           // スイート取得
PUT    /api/v1/test-suites/:id           // スイート更新
PATCH  /api/v1/test-suites/:id/status    // ステータス更新
GET    /api/v1/test-suites?status=:status // ステータスによる検索

// テストグループ管理
POST   /api/v1/test-suites/:suiteId/groups       // グループ作成
GET    /api/v1/test-suites/:suiteId/groups       // グループ一覧取得
PUT    /api/v1/test-groups/:groupId              // グループ更新
PATCH  /api/v1/test-groups/:groupId/order        // 表示順序更新

// テストケース管理
POST   /api/v1/test-groups/:groupId/cases        // テストケース作成
GET    /api/v1/test-groups/:groupId/cases        // テストケース一覧取得
PUT    /api/v1/test-cases/:caseId               // テストケース更新
PATCH  /api/v1/test-cases/:caseId/status        // ステータス更新
POST   /api/v1/test-cases/:caseId/efforts       // 工数記録
```

#### DTOの構造
```go
// 作成用DTO
type TestSuiteCreateDTO struct {
    Name                 string    `json:"name" validate:"required"`
    Description          string    `json:"description"`
    EstimatedStartDate   time.Time `json:"estimatedStartDate" validate:"required"`
    EstimatedEndDate     time.Time `json:"estimatedEndDate" validate:"required,gtfield=EstimatedStartDate"`
    RequireEffortComment bool      `json:"requireEffortComment"`
}

// 更新用DTO
type TestSuiteUpdateDTO struct {
    Name                 string    `json:"name,omitempty"`
    Description          string    `json:"description,omitempty"`
    EstimatedStartDate   time.Time `json:"estimatedStartDate,omitempty"`
    EstimatedEndDate     time.Time `json:"estimatedEndDate,omitempty"`
    RequireEffortComment *bool     `json:"requireEffortComment,omitempty"`
}

// レスポンス用DTO
type TestSuiteResponseDTO struct {
    ID                   string             `json:"id"`
    Name                 string             `json:"name"`
    Description          string             `json:"description"`
    Status               string             `json:"status"`
    EstimatedStartDate   time.Time          `json:"estimatedStartDate"`
    EstimatedEndDate     time.Time          `json:"estimatedEndDate"`
    RequireEffortComment bool               `json:"requireEffortComment"`
    Progress             float64            `json:"progress"`
    Groups              []TestGroupSummaryDTO `json:"groups,omitempty"`
    CreatedAt           time.Time           `json:"createdAt"`
    UpdatedAt           time.Time           `json:"updatedAt"`
}
```

#### バリデーション戦略
```go
// バリデータの実装
type TestSuiteValidator struct {
    validator *validator.Validate
}

func (v *TestSuiteValidator) Validate(dto interface{}) error {
    if err := v.validator.Struct(dto); err != nil {
        return v.translateError(err)
    }
    return v.validateBusinessRules(dto)
}

// カスタムバリデーションルール
func (v *TestSuiteValidator) validateBusinessRules(dto interface{}) error {
    switch d := dto.(type) {
    case *TestSuiteCreateDTO:
        return v.validateCreateDTO(d)
    case *TestSuiteUpdateDTO:
        return v.validateUpdateDTO(d)
    default:
        return errors.New("unknown dto type")
    }
}
```

### 1.2 Protocol Buffers実装（フェーズ2）
#### メッセージ定義
```protobuf
syntax = "proto3";

package testsuite.v1;

import "google/protobuf/timestamp.proto";

// テストスイート定義
message TestSuite {
    string id = 1;
    string name = 2;
    string description = 3;
    TestSuiteStatus status = 4;
    google.protobuf.Timestamp estimated_start_date = 5;
    google.protobuf.Timestamp estimated_end_date = 6;
    bool require_effort_comment = 7;
    repeated TestGroup groups = 8;
    google.protobuf.Timestamp created_at = 9;
    google.protobuf.Timestamp updated_at = 10;
}

// ステータス定義
enum TestSuiteStatus {
    TEST_SUITE_STATUS_UNSPECIFIED = 0;
    TEST_SUITE_STATUS_PREPARATION = 1;
    TEST_SUITE_STATUS_IN_PROGRESS = 2;
    TEST_SUITE_STATUS_COMPLETED = 3;
    TEST_SUITE_STATUS_SUSPENDED = 4;
}

// サービス定義
service TestSuiteService {
    rpc CreateTestSuite(CreateTestSuiteRequest) returns (TestSuite);
    rpc GetTestSuite(GetTestSuiteRequest) returns (TestSuite);
    rpc UpdateTestSuite(UpdateTestSuiteRequest) returns (TestSuite);
    rpc UpdateTestSuiteStatus(UpdateTestSuiteStatusRequest) returns (TestSuite);
    rpc ListTestSuitesByStatus(ListTestSuitesByStatusRequest) returns (ListTestSuitesResponse);
}
```

#### gRPCサービス実装
```go
type TestSuiteServer struct {
    pb.UnimplementedTestSuiteServiceServer
    useCase usecase.TestSuiteUseCase
}

func (s *TestSuiteServer) CreateTestSuite(ctx context.Context, req *pb.CreateTestSuiteRequest) (*pb.TestSuite, error) {
    dto := &dto.TestSuiteCreateDTO{
        Name:               req.GetName(),
        Description:        req.GetDescription(),
        EstimatedStartDate: req.GetEstimatedStartDate().AsTime(),
        EstimatedEndDate:   req.GetEstimatedEndDate().AsTime(),
    }
    
    suite, err := s.useCase.CreateTestSuite(ctx, dto)
    if err != nil {
        return nil, s.translateError(err)
    }
    
    return s.toProto(suite), nil
}
```

### 1.3 GraphQL実装（フェーズ3）
#### スキーマ定義
```graphql
type TestSuite {
    id: ID!
    name: String!
    description: String
    status: TestSuiteStatus!
    estimatedStartDate: DateTime!
    estimatedEndDate: DateTime!
    requireEffortComment: Boolean!
    progress: Float!
    groups: [TestGroup!]
    createdAt: DateTime!
    updatedAt: DateTime!
}

type TestGroup {
    id: ID!
    name: String!
    description: String
    displayOrder: Int!
    status: TestGroupStatus!
    cases: [TestCase!]
}

type TestCase {
    id: ID!
    title: String!
    description: String
    status: TestCaseStatus!
    priority: Priority!
    plannedEffort: Float
    actualEffort: Float
    isDelayed: Boolean!
    delayDays: Int
}

input CreateTestSuiteInput {
    name: String!
    description: String
    estimatedStartDate: DateTime!
    estimatedEndDate: DateTime!
    requireEffortComment: Boolean
}

type Query {
    testSuite(id: ID!): TestSuite
    testSuites(status: TestSuiteStatus): [TestSuite!]!
    testGroup(id: ID!): TestGroup
    testCase(id: ID!): TestCase
}

type Mutation {
    createTestSuite(input: CreateTestSuiteInput!): TestSuite!
    updateTestSuite(id: ID!, input: UpdateTestSuiteInput!): TestSuite!
    updateTestSuiteStatus(id: ID!, status: TestSuiteStatus!): TestSuite!
}

type Subscription {
    testSuiteStatusChanged(id: ID!): TestSuite!
    testCaseStatusChanged(suiteId: ID!): TestCase!
}
```

#### リゾルバー実装
```go
type testSuiteResolver struct {
    useCase usecase.TestSuiteUseCase
}

func (r *testSuiteResolver) Groups(ctx context.Context, obj *model.TestSuite) ([]*model.TestGroup, error) {
    // N+1問題を回避するためのDataloaderの使用
    if loader, err := getGroupLoader(ctx); err == nil {
        return loader.LoadMany(obj.ID)
    }
    return r.useCase.GetTestGroups(ctx, obj.ID)
}

func (r *mutationResolver) CreateTestSuite(ctx context.Context, input model.CreateTestSuiteInput) (*model.TestSuite, error) {
    dto := &dto.TestSuiteCreateDTO{
        Name:               input.Name,
        Description:        input.Description,
        EstimatedStartDate: input.EstimatedStartDate,
        EstimatedEndDate:   input.EstimatedEndDate,
    }
    
    return r.useCase.CreateTestSuite(ctx, dto)
}
```

## 2. 技術スタック詳細

### 2.1 共通基盤
- Go 1.21以上
- PostgreSQL 14.13
- Docker 27.3.1
- Docker Compose v2.29.7

### 2.2 REST API関連
- 標準ライブラリnet/http
- gorilla/mux（ルーティング）
- validator/v10（バリデーション）

### 2.3 Protocol Buffers関連
- protoc（Protocol Buffers compiler）v3.17.3以上
- golang/protobuf v1.5.3
- grpc-go v1.56.1
- protoc-gen-go v1.28.1
- protoc-gen-go-grpc v1.2.0

### 2.4 GraphQL関連
- gqlgen v0.17.38
- graphql-go/graphql v0.8.1
- gorilla/websocket（サブスクリプション用）
- dataloaden（N+1問題対策）

## 3. 段階的な実装計画

### 3.1 フェーズ1：REST API（3週間）
1. Week 1: DTOとバリデーション
   - DTO構造の実装
   - バリデーションルールの実装
   - テストコードの作成

2. Week 2: エンドポイント実装
   - ルーティングの設定
   - ハンドラーの実装
   - エラーハンドリングの実装

3. Week 3: テストと改善
   - 統合テストの作成
   - パフォーマンステスト
   - ドキュメント作成

### 3.2 フェーズ2：Protocol Buffers（2週間）
1. Week 1: 基盤整備
   - Proto定義の作成
   - gRPCサービスの実装
   - クライアントの生成

2. Week 2: 機能拡張
   - ストリーミング機能の追加
   - エラーハンドリングの実装
   - テストの作成

### 3.3 フェーズ3：GraphQL（2週間）
1. Week 1: スキーマとリゾルバー
   - スキーマ定義
   - ベースリゾルバーの実装
   - DataLoaderの実装

2. Week 2: 機能拡張
   - サブスクリプションの実装
   - N+1対策の実装
   - テストの作成

## 4. 学習計画

### 4.1 Protocol Buffers / gRPC
1. 基礎学習
   - Protocol Buffersの基本概念
   - gRPCの仕組みと利点
   - メッセージ定義の書き方

2. 実装練習
   - シンプルなgRPCサービスの作成
   - 各種RPCパターンの実装
   - エラーハンドリング

3. 応用学習
   - ストリーミングの実装
   - セキュリティ設定
   - 負荷テスト

### 4.2 GraphQL
1. 基礎学習
   - GraphQLの基本概念
   - スキーマ定義言語
   - リゾルバーの役割

2. 実装練習
   - シンプルなエンドポイントの作成
   - クエリとミューテーション
   - データローダーの使用

3. 応用学習
   - サブスクリプションの実装
   - N+1問題の解決
   - キャッシュ戦略

## 5. 技術選定の根拠

### 5.1 Protocol Buffersを選択した理由
1. パフォーマンス
   - バイナリ形式による効率的なデータ転送
   - 厳密な型チェック
   - コード生成による型安全性

2. マイクロサービス対応
   - サービス定義の標準化
   - 言語間の相互運用性
   - スケーラビリティ

3. 開発効率
   - コード生成による実装の自動化
   - クライアントライブラリの自動生成
   - バージョニングの容易さ

### 5.2 GraphQLを選択した理由
1. クライアント側の柔軟性
   - 必要なデータのみを取得可能
   - オーバーフェッチの防止
   - アンダーフェッチの防止

2. 開発効率
   - 単一エンドポイント
   - 型システムによる安全性
   - インタラクティブなドキュメント

3. リアルタイム機能
   - サブスクリプションによる更新通知
   - リアルタイムダッシュボード対応
   - WebSocket統合

# 6. 想定される課題と対応方針

## 6.1 技術的課題

### 6.1.1 複数APIの整合性維持
#### 課題
- 異なるAPIエンドポイント間でのデータ構造の一貫性
- レスポンス形式の統一
- エラー処理の一貫性

#### 対応方針
- 共通のドメインモデルを基盤とした設計
- 統一されたDTO変換レイヤーの実装
- 共通のエラー型定義とマッピング処理の実装

### 6.1.2 パフォーマンス最適化
#### 課題
- N+1問題（特にGraphQL）
- 大量データ転送時の効率
- レスポンス時間の管理

#### 対応方針
- DataLoaderパターンの実装
- キャッシュ戦略の策定
- クエリの最適化とモニタリング
- ページネーションの実装

### 6.1.3 トランザクション管理
#### 課題
- 分散トランザクション
- ロング・トランザクションの処理
- 楽観的ロックと悲観的ロック

#### 対応方針
- トランザクションスコープの明確な定義
- デッドロック対策の実装
- 競合解決戦略の実装

## 6.2 運用上の課題

### 6.2.1 バージョン管理
#### 課題
- APIバージョンの管理
- スキーマの進化
- 後方互換性の維持

#### 対応方針
- セマンティックバージョニングの採用
- スキーマ変更の履歴管理
- マイグレーション戦略の策定

### 6.2.2 監視とログ管理
#### 課題
- 複数APIエンドポイントの統合監視
- パフォーマンスメトリクスの収集
- エラー追跡と分析
- デバッグ情報の収集

#### 対応方針
- 統合ログ収集システムの導入
- トレースIDによる追跡
- メトリクス収集基盤の整備
- アラート閾値の設定

### 6.2.3 セキュリティ
#### 課題
- 認証・認可の統合管理
- APIアクセス制御
- データ暗号化
- レート制限

#### 対応方針
- JWTベースの認証基盤
- RBACによるアクセス制御
- TLS/SSL対応
- レート制限ミドルウェアの実装

## 6.3 開発プロセス上の課題

### 6.3.1 テスト戦略
#### 課題
- 複数API形式のテストカバレッジ
- 統合テストの複雑さ
- パフォーマンステストの実施

#### 対応方針
- テストカバレッジ目標の設定（85%以上）
- モック/スタブの適切な使用
- 自動テストパイプラインの構築
- 負荷テストシナリオの作成

### 6.3.2 ドキュメント管理
#### 課題
- API仕様書の統合管理
- クライアントライブラリのドキュメント
- 内部実装ドキュメント
- 変更履歴の管理

#### 対応方針
- OpenAPI（Swagger）による REST API ドキュメント
- Protocol Buffersの自動生成ドキュメント
- GraphQL Playgroundの活用
- ドキュメント自動生成の仕組み構築

### 6.3.3 チーム開発体制
#### 課題
- 技術スタックの学習曲線
- コードレビューの効率
- 知識共有

#### 対応方針
- 段階的な導入と学習計画
- レビューガイドラインの策定
- 定期的な技術共有セッションの実施
- ナレッジベースの整備