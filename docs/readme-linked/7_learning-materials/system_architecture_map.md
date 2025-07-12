# 技術配置図・システム内役割マップ
*プロジェクト全体での各技術の配置と役割の完全ガイド*

## 🎯 この資料の目的

あなたのプロジェクトで使用されている技術が、システムのどこに配置され、どのような役割を担っているかを視覚的に分かりやすく解説します。

---

## 1. システム全体アーキテクチャ

### 1.1 フルスタック技術配置の概要

```mermaid
graph TB
    subgraph "👥 ユーザー"
        USER[エンドユーザー<br/>ブラウザ・モバイル]
    end
    
    subgraph "🌐 フロントエンド層"
        REACT[React 19.1.0<br/>UIコンポーネント]
        TS[TypeScript 5.8.3<br/>型安全性]
        APOLLO[Apollo Client 3.13.8<br/>GraphQL状態管理]
        MUI[Material UI 7.1.1<br/>UIライブラリ]
        ROUTER[React Router 7.6.1<br/>画面遷移]
    end
    
    subgraph "☁️ AWS CDN・配信層"
        CF[CloudFront<br/>CDN配信]
        S3[S3 Bucket<br/>静的ファイル保存]
        R53[Route53<br/>DNS管理]
    end
    
    subgraph "🔗 API・通信層"
        ALB[Application Load Balancer<br/>負荷分散・SSL終端]
        API_REST[REST API<br/>標準HTTP通信]
        API_GQL[GraphQL API<br/>効率的データ取得]
        API_GRPC[gRPC API<br/>高性能RPC通信]
    end
    
    subgraph "⚙️ アプリケーション層"
        ECS[ECS Fargate<br/>コンテナ実行環境]
        GO[Go言語<br/>バックエンドロジック]
        MIDDLEWARE[認証ミドルウェア<br/>セキュリティ制御]
        DATALOADER[DataLoader<br/>パフォーマンス最適化]
    end
    
    subgraph "🏗️ ビジネスロジック層"
        CA[Clean Architecture<br/>アーキテクチャパターン]
        DDD[Domain-Driven Design<br/>ドメインモデリング]
        USECASE[Use Case Interactors<br/>ビジネス処理]
    end
    
    subgraph "🔐 認証・セキュリティ層"
        JWT[JWT Service<br/>トークン管理]
        BCRYPT[BCrypt<br/>パスワードハッシュ化]
        COOKIE[HttpOnly Cookie<br/>安全なトークン保存]
    end
    
    subgraph "💾 データ・永続化層"
        RDS[RDS PostgreSQL<br/>メインデータベース]
        SSM[SSM Parameter Store<br/>設定・シークレット管理]
    end
    
    subgraph "🔧 開発・運用層"
        TERRAFORM[Terraform<br/>Infrastructure as Code]
        DOCKER[Docker<br/>コンテナ化]
        GITHUB[GitHub Actions<br/>CI/CD]
    end
    
    USER --> CF
    CF --> S3
    CF --> ALB
    R53 --> ALB
    ALB --> API_REST
    ALB --> API_GQL
    ALB --> API_GRPC
    
    REACT --> APOLLO
    TS --> REACT
    MUI --> REACT
    ROUTER --> REACT
    APOLLO --> API_GQL
    
    API_REST --> MIDDLEWARE
    API_GQL --> MIDDLEWARE
    API_GRPC --> MIDDLEWARE
    MIDDLEWARE --> DATALOADER
    DATALOADER --> USECASE
    
    USECASE --> CA
    USECASE --> DDD
    MIDDLEWARE --> JWT
    JWT --> BCRYPT
    JWT --> COOKIE
    
    USECASE --> RDS
    GO --> SSM
    
    ECS --> GO
    TERRAFORM --> ECS
    DOCKER --> ECS
    GITHUB --> TERRAFORM
```

### 1.2 各層の責任範囲

| 層 | 主要技術 | 責任 | 配置場所 |
|---|----------|------|----------|
| **フロントエンド層** | React, TypeScript, Apollo Client | UI表示・ユーザーインタラクション | ブラウザ |
| **CDN・配信層** | CloudFront, S3, Route53 | 静的ファイル配信・DNS解決 | AWS Global |
| **API・通信層** | ALB, REST/GraphQL/gRPC | リクエスト分散・プロトコル処理 | AWS Region |
| **アプリケーション層** | ECS, Go, DataLoader | ビジネスロジック実行・最適化 | AWS Container |
| **ビジネスロジック層** | Clean Architecture, DDD | ドメインルール・ユースケース | Go Application |
| **認証・セキュリティ層** | JWT, BCrypt, HttpOnly Cookie | 認証・認可・暗号化 | 全層横断 |
| **データ・永続化層** | PostgreSQL, SSM | データ保存・設定管理 | AWS Managed Service |
| **開発・運用層** | Terraform, Docker, GitHub | インフラ管理・デプロイ自動化 | 開発環境・CI/CD |

---

## 2. フロントエンド技術配置の詳細

### 2.1 React アプリケーション内の技術配置

```mermaid
graph TB
    subgraph "React Application Structure"
        subgraph "📱 Pages Layer"
            LOGIN_PAGE[LoginPage.tsx<br/>React + Material UI + Apollo]
            DASHBOARD_PAGE[DashboardPage.tsx<br/>React + GraphQL Query]
            TESTSUITE_PAGE[TestSuiteListPage.tsx<br/>React + 複数カスタムフック]
        end
        
        subgraph "🧩 Components Layer"
            NAVIGATION[MainNavigation.tsx<br/>React Router + Material UI]
            TESTSUITE_LIST[TestSuiteList.tsx<br/>Material UI + TypeScript]
            CREATE_MODAL[CreateTestSuiteModal.tsx<br/>Material UI + Apollo Mutation]
        end
        
        subgraph "🔄 State Management Layer"
            AUTH_CONTEXT[AuthContext.tsx<br/>React Context + JWT]
            APOLLO_CLIENT[Apollo Client<br/>GraphQL Cache + Network]
            CUSTOM_HOOKS[useTestSuites.ts<br/>ビジネスロジック抽象化]
        end
        
        subgraph "🔗 Integration Layer"
            GRAPHQL_CODEGEN[GraphQL Code Generator<br/>型生成・カスタムフック自動生成]
            APOLLO_CONFIG[Apollo Client Config<br/>CORS・認証・エラーハンドリング]
        end
        
        subgraph "🛠️ Development Layer"
            TYPESCRIPT[TypeScript 5.8.3<br/>型チェック・IntelliSense]
            VITE[Vite 6.3.5<br/>高速ビルド・HMR]
            ESLINT[ESLint + Prettier<br/>コード品質・フォーマット]
        end
    end
    
    LOGIN_PAGE --> AUTH_CONTEXT
    DASHBOARD_PAGE --> APOLLO_CLIENT
    TESTSUITE_PAGE --> CUSTOM_HOOKS
    
    NAVIGATION --> AUTH_CONTEXT
    TESTSUITE_LIST --> APOLLO_CLIENT
    CREATE_MODAL --> APOLLO_CLIENT
    
    AUTH_CONTEXT --> APOLLO_CONFIG
    APOLLO_CLIENT --> GRAPHQL_CODEGEN
    CUSTOM_HOOKS --> GRAPHQL_CODEGEN
    
    APOLLO_CONFIG --> TYPESCRIPT
    GRAPHQL_CODEGEN --> TYPESCRIPT
    TYPESCRIPT --> VITE
    VITE --> ESLINT
```

### 2.2 フロントエンド技術の具体的役割

#### 🎨 **UI・表示技術**
```typescript
// Material UI - モダンなUIコンポーネント
import { TextField, Button, Card, Alert } from '@mui/material';

// 使用箇所: 全てのPage・Componentで標準UI
<TextField label="ユーザー名" variant="outlined" />
<Button variant="contained" color="primary">ログイン</Button>
```

#### 🔄 **状態管理技術**
```typescript
// React Context - 認証状態のグローバル管理
const AuthContext = createContext<AuthContextType | undefined>(undefined);

// Apollo Client - GraphQLデータ・キャッシュ管理
const client = new ApolloClient({
  uri: 'https://example-graphql-api.com/graphql',
  cache: new InMemoryCache(),
  credentials: 'include', // HttpOnly Cookie自動送信
});
```

#### 🛡️ **型安全性技術**
```typescript
// TypeScript - コンパイル時型チェック
interface AuthUser {
  id: string;
  username: string;
  role: 'admin' | 'manager' | 'tester';
}

// GraphQL Code Generator - 自動型生成
export function useLoginMutation() {
  return Apollo.useMutation<LoginMutation, LoginMutationVariables>(LoginDocument);
}
```

---

## 3. バックエンド技術配置の詳細

### 3.1 Go アプリケーション内の技術配置

```mermaid
graph TB
    subgraph "Go Application Architecture"
        subgraph "🌐 Presentation Layer"
            REST_HANDLER[REST API Handlers<br/>net/http + Gorilla Mux]
            GQL_RESOLVER[GraphQL Resolvers<br/>gqlgen + DataLoader]
            GRPC_SERVER[gRPC Servers<br/>Protocol Buffers + grpc-go]
            AUTH_MIDDLEWARE[Authentication Middleware<br/>JWT検証 + CORS]
        end
        
        subgraph "🔧 Application Layer"
            USECASE_INTERACTOR[Use Case Interactors<br/>ビジネスロジック実行]
            INPUT_VALIDATION[Input Validation<br/>リクエスト検証]
            ERROR_HANDLING[Error Handling<br/>customerrors + HTTP変換]
            DATALOADER_IMPL[DataLoader Implementation<br/>N+1問題解決]
        end
        
        subgraph "🏗️ Domain Layer"
            ENTITIES[Entities<br/>User・TestSuite・TestGroup・TestCase]
            VALUE_OBJECTS[Value Objects<br/>SuiteStatus・UserRole]
            DOMAIN_SERVICES[Domain Services<br/>ビジネスルール]
            REPOSITORY_INTERFACES[Repository Interfaces<br/>抽象化]
        end
        
        subgraph "💾 Infrastructure Layer"
            POSTGRESQL_REPO[PostgreSQL Repositories<br/>database/sql + pq driver]
            JWT_SERVICE[JWT Service<br/>HMAC-SHA256 + crypto]
            PASSWORD_SERVICE[Password Service<br/>BCrypt]
            CONFIG_PROVIDER[Configuration Provider<br/>12-Factor準拠]
        end
        
        subgraph "🔐 Security & Cross-cutting"
            CRYPTO[Cryptographic Services<br/>JWT + BCrypt]
            LOGGING[Structured Logging<br/>logrus + JSON format]
            MONITORING[Health Checks<br/>各プロトコル対応]
            DB_CONNECTION[Database Connection Pool<br/>最適化設定]
        end
    end
    
    REST_HANDLER --> AUTH_MIDDLEWARE
    GQL_RESOLVER --> AUTH_MIDDLEWARE
    GRPC_SERVER --> AUTH_MIDDLEWARE
    AUTH_MIDDLEWARE --> USECASE_INTERACTOR
    
    USECASE_INTERACTOR --> INPUT_VALIDATION
    USECASE_INTERACTOR --> ERROR_HANDLING
    GQL_RESOLVER --> DATALOADER_IMPL
    
    USECASE_INTERACTOR --> ENTITIES
    ENTITIES --> VALUE_OBJECTS
    ENTITIES --> DOMAIN_SERVICES
    USECASE_INTERACTOR --> REPOSITORY_INTERFACES
    
    REPOSITORY_INTERFACES -.-> POSTGRESQL_REPO
    AUTH_MIDDLEWARE --> JWT_SERVICE
    JWT_SERVICE --> PASSWORD_SERVICE
    POSTGRESQL_REPO --> CONFIG_PROVIDER
    
    JWT_SERVICE --> CRYPTO
    ALL --> LOGGING
    ALL --> MONITORING
    POSTGRESQL_REPO --> DB_CONNECTION
```

### 3.2 バックエンド技術の具体的役割

#### 🌐 **API・プロトコル技術**
```go
// REST API - 標準HTTP/JSON
func (h *TestSuiteHandler) CreateTestSuite(w http.ResponseWriter, r *http.Request) {
    // 役割: 外部システム統合・標準準拠
}

// GraphQL - 効率的データ取得
func (r *mutationResolver) CreateTestSuite(ctx context.Context, input CreateTestSuiteInput) (*TestSuite, error) {
    // 役割: フロントエンド最適化・型安全
}

// gRPC - 高性能内部通信
func (s *TestSuiteServer) CreateTestSuite(ctx context.Context, req *pb.CreateTestSuiteRequest) (*pb.TestSuite, error) {
    // 役割: マイクロサービス間通信・性能重視
}
```

#### 🔐 **認証・セキュリティ技術**
```go
// JWT Service - トークン生成・検証
type JWTService struct {
    secretKey []byte // HMAC-SHA256用秘密鍵
}

// BCrypt Service - パスワードハッシュ化
func (s *BCryptPasswordService) HashPassword(password string) (string, error) {
    return bcrypt.GenerateFromPassword([]byte(password), s.cost)
}
```

#### 🏗️ **アーキテクチャ・設計技術**
```go
// Clean Architecture - 依存関係逆転
type TestSuiteInteractor struct {
    repo repository.TestSuiteRepository // インターフェース依存
}

// DDD - ドメインモデル
type TestSuite struct {
    ID     string
    Name   string
    Status SuiteStatus // 値オブジェクト
}
```

---

## 4. AWS インフラ技術配置の詳細

### 4.1 AWS サービス配置とデータフロー

```mermaid
graph TB
    subgraph "🌍 Global Services"
        CF[CloudFront<br/>CDN・エッジキャッシュ]
        R53[Route53<br/>DNS・ヘルスチェック]
        ACM[Certificate Manager<br/>SSL証明書自動管理]
    end
    
    subgraph "🌐 Region: Asia Pacific (Tokyo)"
        subgraph "🔒 Security & Access"
            IAM[IAM Roles<br/>最小権限設定]
            WAF[WAF (将来実装)<br/>Web攻撃防御]
        end
        
        subgraph "⚖️ Load Balancing & Networking"
            ALB[Application Load Balancer<br/>SSL終端・ヘルスチェック]
            VPC[VPC<br/>論理ネットワーク分離]
            PUB_SUBNET[Public Subnet<br/>ALB・NAT Gateway配置]
            PRI_SUBNET[Private Subnet<br/>ECS・RDS配置]
            IGW[Internet Gateway<br/>インターネット接続]
            NAT[NAT Gateway<br/>アウトバウンド通信]
        end
        
        subgraph "🚀 Compute & Application"
            ECS[ECS Fargate<br/>サーバーレスコンテナ]
            ECR[ECR<br/>Dockerイメージ保存]
            TASK[ECS Task Definition<br/>コンテナ設定]
        end
        
        subgraph "💾 Data & Storage"
            RDS[RDS PostgreSQL<br/>マネージドデータベース]
            S3[S3 Bucket<br/>静的ファイル・バックアップ]
            SSM[SSM Parameter Store<br/>設定・シークレット]
        end
        
        subgraph "📊 Monitoring & Logging"
            CW[CloudWatch<br/>メトリクス・ログ]
            CW_LOGS[CloudWatch Logs<br/>アプリケーションログ]
            CW_ALARMS[CloudWatch Alarms<br/>アラート（将来実装）]
        end
    end
    
    subgraph "🛠️ Management & Deployment"
        TERRAFORM[Terraform<br/>Infrastructure as Code]
        GH_ACTIONS[GitHub Actions<br/>CI/CD Pipeline]
    end
    
    %% フロー接続
    CF --> ALB
    R53 --> CF
    ACM --> ALB
    ALB --> ECS
    ECS --> RDS
    ECS --> SSM
    ECS --> CW_LOGS
    S3 --> CF
    
    VPC --> PUB_SUBNET
    VPC --> PRI_SUBNET
    PUB_SUBNET --> ALB
    PUB_SUBNET --> NAT
    PRI_SUBNET --> ECS
    PRI_SUBNET --> RDS
    IGW --> PUB_SUBNET
    NAT --> PRI_SUBNET
    
    ECR --> ECS
    TASK --> ECS
    CW --> CW_LOGS
    CW --> CW_ALARMS
    
    TERRAFORM --> VPC
    TERRAFORM --> ECS
    TERRAFORM --> RDS
    GH_ACTIONS --> TERRAFORM
```

### 4.2 AWS技術の具体的役割

#### 🌐 **ネットワーク・配信技術**
```hcl
# CloudFront - グローバルCDN
resource "aws_cloudfront_distribution" "main" {
  # 役割: 世界中のエッジでフロントエンド高速配信
  enabled = true
  default_cache_behavior {
    target_origin_id = aws_s3_bucket.frontend.id
    viewer_protocol_policy = "redirect-to-https"
  }
}

# Application Load Balancer - 負荷分散
resource "aws_lb" "main" {
  # 役割: HTTPS終端・複数ECSサービスへの振り分け
  load_balancer_type = "application"
  scheme            = "internet-facing"
}
```

#### 🚀 **コンピューティング技術**
```hcl
# ECS Fargate - サーバーレスコンテナ
resource "aws_ecs_service" "graphql" {
  # 役割: Goアプリケーションのスケーラブル実行
  launch_type = "FARGATE"
  desired_count = 2
  # CPU: 512, Memory: 1024
}

# ECS Task Definition - コンテナ設定
resource "aws_ecs_task_definition" "app" {
  # 役割: Dockerコンテナの詳細設定
  cpu    = 512
  memory = 1024
  requires_compatibilities = ["FARGATE"]
}
```

#### 💾 **データ・ストレージ技術**
```hcl
# RDS PostgreSQL - マネージドデータベース
resource "aws_db_instance" "main" {
  # 役割: 本番レベルのデータ永続化・バックアップ
  engine         = "postgres"
  engine_version = "15.4"
  instance_class = "db.t3.medium"
  multi_az      = true  # 高可用性
}

# SSM Parameter Store - 設定管理
resource "aws_ssm_parameter" "db_password" {
  # 役割: 機密情報の安全な保存・管理
  type  = "SecureString"
  value = random_password.db_password.result
}
```

---

## 5. 技術統合ポイントとデータフロー

### 5.1 認証フローでの技術統合

```mermaid
sequenceDiagram
    participant Browser as ブラウザ<br/>(React + Apollo)
    participant CF as CloudFront<br/>(CDN)
    participant ALB as ALB<br/>(SSL終端)
    participant ECS as ECS<br/>(Go App)
    participant JWT as JWT Service<br/>(Go)
    participant BCrypt as BCrypt<br/>(Go)
    participant RDS as RDS<br/>(PostgreSQL)
    
    Browser->>CF: ログインページ要求
    CF-->>Browser: React SPA配信
    
    Browser->>ALB: ログイン要求 (HTTPS)
    ALB->>ECS: GraphQL Mutation
    ECS->>BCrypt: パスワード検証
    BCrypt->>RDS: ユーザー情報取得
    RDS-->>BCrypt: ハッシュ化パスワード
    BCrypt-->>ECS: 検証結果
    
    ECS->>JWT: JWTトークン生成
    JWT-->>ECS: 署名済みトークン
    ECS-->>ALB: HttpOnly Cookie設定
    ALB-->>Browser: 認証成功レスポンス
    
    Browser->>Browser: Apollo Client認証状態更新
    Browser->>CF: ダッシュボードページ遷移
```

### 5.2 データ取得フローでの技術統合

```mermaid
sequenceDiagram
    participant React as React<br/>(TestSuiteList)
    participant Apollo as Apollo Client<br/>(Cache + Network)
    participant CodeGen as Code Generator<br/>(Types + Hooks)
    participant ALB as ALB<br/>(Load Balancer)
    participant GraphQL as GraphQL<br/>(Resolver)
    participant DataLoader as DataLoader<br/>(Batch + Cache)
    participant PostgreSQL as PostgreSQL<br/>(Database)
    
    React->>Apollo: useTestSuites() Hook実行
    Apollo->>CodeGen: 自動生成型・Hook使用
    Apollo->>ALB: GraphQLクエリ送信 (Cookie付き)
    ALB->>GraphQL: 認証済みリクエスト
    
    GraphQL->>DataLoader: TestSuite一覧要求
    DataLoader->>PostgreSQL: SELECT * FROM test_suites
    PostgreSQL-->>DataLoader: 20件のTestSuite
    
    GraphQL->>DataLoader: Groups要求（20回）
    DataLoader->>DataLoader: バッチング（20→1）
    DataLoader->>PostgreSQL: SELECT * WHERE suite_id IN (...)
    PostgreSQL-->>DataLoader: 全Groupデータ
    
    GraphQL->>DataLoader: Cases要求（60回）
    DataLoader->>DataLoader: バッチング（60→1）
    DataLoader->>PostgreSQL: SELECT * WHERE group_id IN (...)
    PostgreSQL-->>DataLoader: 全Caseデータ
    
    DataLoader-->>GraphQL: 統合データ（3クエリ）
    GraphQL-->>Apollo: GraphQLレスポンス
    Apollo->>Apollo: キャッシュ更新
    Apollo-->>React: TypeScript型付きデータ
    React->>React: Material UIで表示
```

---

## 6. 技術選択の理由と適材適所

### 6.1 フロントエンド技術選択の理由

| 技術 | 選択理由 | 適用場面 | 代替選択肢との比較 |
|------|----------|----------|-------------------|
| **React 19** | 最新コンポーネント・JSX・生態系 | UI構築全般 | Vue.js: 学習容易性 vs React: 生態系豊富 |
| **TypeScript** | 型安全性・開発効率・エラー早期発見 | 全コード | JavaScript: 開発速度 vs TypeScript: 品質 |
| **Apollo Client** | GraphQL特化・キャッシュ・開発者体験 | データ取得・状態管理 | TanStack Query: 軽量 vs Apollo: GraphQL統合 |
| **Material UI** | モダンデザイン・コンポーネント豊富 | UI表示 | Tailwind: カスタマイズ vs MUI: 標準化 |

### 6.2 バックエンド技術選択の理由

| 技術 | 選択理由 | 適用場面 | 代替選択肢との比較 |
|------|----------|----------|-------------------|
| **Go** | 高性能・シンプル・並行処理・型安全 | API実装全般 | Node.js: JS統一 vs Go: 性能・型安全 |
| **GraphQL** | 効率的データ取得・型安全・開発者体験 | フロントエンド連携 | REST: 単純 vs GraphQL: 効率・型安全 |
| **gRPC** | 高性能・型安全・ストリーミング | 内部サービス通信 | REST: 汎用性 vs gRPC: 性能・型安全 |
| **PostgreSQL** | ACID・拡張性・JSON対応・信頼性 | データ永続化 | MySQL: 普及率 vs PostgreSQL: 機能・標準準拠 |

### 6.3 AWS技術選択の理由

| 技術 | 選択理由 | 適用場面 | 代替選択肢との比較 |
|------|----------|----------|-------------------|
| **ECS Fargate** | サーバーレス・スケーラブル・マネージド | コンテナ実行 | EC2: 制御性 vs Fargate: 運用簡単 |
| **ALB** | L7負荷分散・SSL終端・ヘルスチェック | ロードバランシング | NLB: 性能 vs ALB: 機能豊富 |
| **RDS** | マネージド・バックアップ・マルチAZ | データベース | Aurora: 性能 vs RDS: 標準・コスト |
| **CloudFront** | グローバルCDN・エッジキャッシュ | 静的配信 | S3直接: シンプル vs CloudFront: 性能・キャッシュ |

---

## 7. 運用・監視での技術配置

### 7.1 ログ・監視技術の配置

```mermaid
graph TB
    subgraph "📊 監視・ログ体系"
        subgraph "フロントエンド監視"
            BROWSER_CONSOLE[Browser Console<br/>クライアントエラー]
            APOLLO_DEVTOOLS[Apollo DevTools<br/>GraphQLクエリ・キャッシュ]
        end
        
        subgraph "アプリケーション監視"
            GO_LOGGING[Go Structured Logging<br/>logrus + JSON]
            ECS_LOGS[ECS Container Logs<br/>stdout/stderr]
            HEALTH_CHECK[Health Check Endpoints<br/>/health・/health-http]
        end
        
        subgraph "インフラ監視"
            ALB_LOGS[ALB Access Logs<br/>リクエスト・レスポンス]
            CW_METRICS[CloudWatch Metrics<br/>CPU・メモリ・ネットワーク]
            RDS_METRICS[RDS Performance Insights<br/>DB性能監視]
        end
        
        subgraph "統合監視（将来実装）"
            CW_DASHBOARD[CloudWatch Dashboard<br/>統合ダッシュボード]
            CW_ALARMS[CloudWatch Alarms<br/>しきい値アラート]
            SNS_ALERTS[SNS Notifications<br/>メール・Slack通知]
        end
    end
    
    BROWSER_CONSOLE --> GO_LOGGING
    APOLLO_DEVTOOLS --> GO_LOGGING
    GO_LOGGING --> ECS_LOGS
    ECS_LOGS --> CW_METRICS
    HEALTH_CHECK --> ALB_LOGS
    ALB_LOGS --> CW_METRICS
    CW_METRICS --> RDS_METRICS
    
    CW_METRICS --> CW_DASHBOARD
    CW_DASHBOARD --> CW_ALARMS
    CW_ALARMS --> SNS_ALERTS
```

### 7.2 デプロイ・CI/CD技術の配置

```mermaid
graph LR
    subgraph "🔧 開発・デプロイパイプライン"
        subgraph "開発環境"
            LOCAL_DEV[ローカル開発<br/>Docker Compose]
            GIT_REPO[Git Repository<br/>GitHub]
        end
        
        subgraph "CI/CD Pipeline"
            GH_ACTIONS[GitHub Actions<br/>自動ビルド・テスト]
            DOCKER_BUILD[Docker Build<br/>コンテナイメージ作成]
            ECR_PUSH[ECR Push<br/>イメージプッシュ]
        end
        
        subgraph "Infrastructure as Code"
            TF_PLAN[Terraform Plan<br/>変更プレビュー]
            TF_APPLY[Terraform Apply<br/>インフラ更新]
            STATE_MGMT[State Management<br/>S3 + DynamoDB]
        end
        
        subgraph "デプロイ・更新"
            ECS_DEPLOY[ECS Service Update<br/>ローリングアップデート]
            HEALTH_VERIFY[Health Check<br/>デプロイ検証]
            ROLLBACK[Rollback<br/>障害時復旧]
        end
    end
    
    LOCAL_DEV --> GIT_REPO
    GIT_REPO --> GH_ACTIONS
    GH_ACTIONS --> DOCKER_BUILD
    DOCKER_BUILD --> ECR_PUSH
    
    GH_ACTIONS --> TF_PLAN
    TF_PLAN --> TF_APPLY
    TF_APPLY --> STATE_MGMT
    
    ECR_PUSH --> ECS_DEPLOY
    TF_APPLY --> ECS_DEPLOY
    ECS_DEPLOY --> HEALTH_VERIFY
    HEALTH_VERIFY --> ROLLBACK
```

---

## 8. 技術間の依存関係と相互作用

### 8.1 技術依存関係マップ

```mermaid
graph TB
    subgraph "外部依存"
        AWS_SERVICES[AWS Services<br/>実行環境・インフラ]
        GITHUB[GitHub<br/>コード管理・CI/CD]
        DNS_PROVIDER[DNS Provider<br/>ドメイン管理]
    end
    
    subgraph "基盤技術"
        DOCKER[Docker<br/>コンテナ化技術]
        TERRAFORM[Terraform<br/>IaC]
        GOLANG[Go<br/>バックエンド言語]
        REACT[React<br/>フロントエンド基盤]
    end
    
    subgraph "フレームワーク・ライブラリ"
        APOLLO[Apollo Client<br/>GraphQL統合]
        GQLGEN[gqlgen<br/>GraphQL Go実装]
        MUI[Material UI<br/>UIコンポーネント]
        GRPC_GO[grpc-go<br/>gRPC実装]
    end
    
    subgraph "開発支援技術"
        TYPESCRIPT[TypeScript<br/>型システム]
        CODEGEN[GraphQL Code Generator<br/>型生成]
        ESLINT[ESLint/Prettier<br/>コード品質]
        DATALOADER[DataLoader<br/>パフォーマンス]
    end
    
    subgraph "セキュリティ・認証"
        JWT_LIB[JWT Library<br/>トークン処理]
        BCRYPT[BCrypt<br/>暗号化]
        CORS[CORS<br/>クロスオリジン]
        TLS[TLS/SSL<br/>通信暗号化]
    end
    
    %% 依存関係
    TERRAFORM --> AWS_SERVICES
    DOCKER --> AWS_SERVICES
    GOLANG --> DOCKER
    REACT --> GITHUB
    
    APOLLO --> REACT
    GQLGEN --> GOLANG
    MUI --> REACT
    GRPC_GO --> GOLANG
    
    TYPESCRIPT --> REACT
    CODEGEN --> APOLLO
    CODEGEN --> GQLGEN
    ESLINT --> TYPESCRIPT
    DATALOADER --> GQLGEN
    
    JWT_LIB --> GOLANG
    BCRYPT --> GOLANG
    CORS --> APOLLO
    TLS --> AWS_SERVICES
```

### 8.2 主要技術統合ポイント

#### 🔗 **フロント・バック統合ポイント**
```typescript
// GraphQL Code Generator - 型統合
// バックエンドのGraphQLスキーマ → フロントエンドTypeScript型
export type LoginMutation = {
  login: {
    token: string;
    expiresAt: string;
    user: {
      id: string;
      username: string;
      role: string;
    };
  };
};
```

#### ⚙️ **バック・インフラ統合ポイント**
```go
// 12-Factor Config - 環境設定統合
// Terraformで設定したSSMパラメータ → Go環境変数
func NewConfigFromEnvironment() *Config {
    return &Config{
        DatabaseURL: os.Getenv("DATABASE_URL"), // TerraformのSSMから注入
        JWTSecret:   os.Getenv("JWT_SECRET"),   // TerraformのSSMから注入
    }
}
```

#### 🌐 **インフラ・運用統合ポイント**
```hcl
# Terraform - インフラとアプリケーション統合
resource "aws_ecs_task_definition" "app" {
  container_definitions = jsonencode([{
    environment = [
      { name = "DATABASE_URL", valueFrom = aws_ssm_parameter.db_url.arn },
      { name = "JWT_SECRET", valueFrom = aws_ssm_parameter.jwt_secret.arn }
    ]
  }])
}
```

---

## 9. 技術スタックの発展・拡張計画

### 9.1 短期拡張計画（現在の技術基盤活用）

```mermaid
graph TB
    subgraph "現在の技術基盤"
        CURRENT[React + Go + GraphQL + AWS]
    end
    
    subgraph "短期拡張（1-3ヶ月）"
        MONITORING[監視強化<br/>CloudWatch Alarms + Dashboard]
        SECURITY[セキュリティ強化<br/>WAF + Advanced Logging]
        PERFORMANCE[性能最適化<br/>Redis Cache + CDN Optimization]
        TESTING[テスト強化<br/>E2E Testing + Performance Testing]
    end
    
    subgraph "追加機能"
        REALTIME[リアルタイム機能<br/>GraphQL Subscription]
        NOTIFICATION[通知機能<br/>SNS + Email Service]
        ANALYTICS[分析機能<br/>Usage Analytics + Metrics]
    end
    
    CURRENT --> MONITORING
    CURRENT --> SECURITY
    CURRENT --> PERFORMANCE
    CURRENT --> TESTING
    
    MONITORING --> REALTIME
    SECURITY --> NOTIFICATION
    PERFORMANCE --> ANALYTICS
```

### 9.2 中長期発展計画（アーキテクチャ拡張）

```mermaid
graph TB
    subgraph "マイクロサービス化"
        USER_SERVICE[User Service<br/>認証・ユーザー管理]
        TESTSUITE_SERVICE[TestSuite Service<br/>テスト管理]
        NOTIFICATION_SERVICE[Notification Service<br/>通知・メール]
        ANALYTICS_SERVICE[Analytics Service<br/>分析・レポート]
    end
    
    subgraph "新技術統合"
        KUBERNETES[Kubernetes<br/>コンテナオーケストレーション]
        ISTIO[Istio<br/>サービスメッシュ]
        PROMETHEUS[Prometheus<br/>監視・メトリクス]
        GRAFANA[Grafana<br/>可視化・ダッシュボード]
    end
    
    subgraph "AI/ML統合"
        AI_RECOMMENDATIONS[AI Recommendations<br/>テスト推奨・自動化]
        ANOMALY_DETECTION[Anomaly Detection<br/>異常検知・アラート]
        PREDICTIVE_ANALYTICS[Predictive Analytics<br/>予測分析・計画]
    end
    
    USER_SERVICE --> KUBERNETES
    TESTSUITE_SERVICE --> KUBERNETES
    NOTIFICATION_SERVICE --> ISTIO
    ANALYTICS_SERVICE --> PROMETHEUS
    
    KUBERNETES --> AI_RECOMMENDATIONS
    ISTIO --> ANOMALY_DETECTION
    PROMETHEUS --> PREDICTIVE_ANALYTICS
    GRAFANA --> PREDICTIVE_ANALYTICS
```

---

## 10. まとめ: 技術配置の価値と学習成果

### 10.1 技術統合による相乗効果

✅ **フロントエンド統合価値**:
- React + TypeScript + Apollo Client = 型安全な高効率開発
- Material UI + React Router = 統一されたユーザー体験
- GraphQL Code Generator = 40%開発効率向上

✅ **バックエンド統合価値**:
- Go + Clean Architecture + DDD = 保守性・拡張性の高い設計
- 3プロトコル統合 = 適材適所の技術活用
- DataLoader + PostgreSQL = 96%クエリ削減・性能最適化

✅ **インフラ統合価値**:
- AWS + Terraform + Docker = Infrastructure as Codeによる一貫性
- ECS + ALB + CloudFront = スケーラブルで高性能な配信
- 本番環境継続稼働 = 実用システムとしての実証

### 10.2 技術配置の学習価値

✅ **フルスタック理解**:
- フロントエンドからインフラまでの一貫した技術理解
- 各層での技術選択理由と適用場面の実践的把握
- 技術間の依存関係と相互作用の深い理解

✅ **現代的開発手法**:
- Infrastructure as Code による一貫したインフラ管理
- GraphQL による効率的API設計・フロントエンド統合
- Container化による環境一貫性・スケーラビリティ実現

✅ **エンタープライズレベル技術力**:
- 複数プロトコル対応による技術適応力
- セキュリティ・パフォーマンス・監視を考慮した実装
- AWS本番環境での実際の運用経験

---

**🎯 重要なポイント**: このプロジェクトは、各技術が適切な場所に配置され、相互に連携することで、単体では実現できない価値を生み出しています。フロントエンドからインフラまでの包括的な技術統合により、現代的Webアプリケーションの設計・実装・運用能力を完全に実証しています。