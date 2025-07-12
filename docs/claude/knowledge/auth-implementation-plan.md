# 認証機能実装計画書

## 1. 概要
テストケース管理システムに必要な認証・認可機能の実装計画について記述します。AWS環境での運用を見据えた最小限の実装と、将来的な拡張性を考慮した設計を提案します。

## 2. 認証機能要件

### 2.1 基本要件
- ユーザー認証（ログイン/ログアウト）
- APIアクセス制御
- トークンベースの認証（JWT）
- セッション管理

### 2.2 ユーザー情報
- ユーザーID
- ユーザー名
- パスワード（ハッシュ化）
- ロール（Admin/Manager/Tester）
- 最終ログイン日時

### 2.3 アクセス制御
| ロール | 権限 |
|-------|------|
| Admin | すべての操作が可能 |
| Manager | テストスイート作成・更新、テストケース管理、レポート閲覧 |
| Tester | テストケース更新、工数記録、コメント追加 |

## 3. 技術設計

### 3.1 認証アーキテクチャ
```
クライアント側                        サーバー側
┌─────────────┐       JWT Token       ┌─────────────┐
│ フロントエンド├─────────────────────►│  API Layer  │
└─────────────┘                       └──────┬──────┘
       ▲                                    │
       │                                    ▼
       │                              ┌─────────────┐
       └──────────────────────────────┤  Auth Service│
                  JWT Token           └──────┬──────┘
                                            │
                                            ▼
                                      ┌─────────────┐
                                      │  Database   │
                                      └─────────────┘
```

### 3.2 認証フロー
1. ユーザーがログインフォームに認証情報を入力
2. サーバーが認証情報を検証
3. 認証成功時、JWTトークンを生成して返却
4. フロントエンドがトークンを保存（localStorage）
5. 以降のAPIリクエストにトークンを付与
6. サーバー側でトークンを検証してアクセス制御

### 3.3 JWTトークン設計
```json
{
  "header": {
    "alg": "HS256",
    "typ": "JWT"
  },
  "payload": {
    "sub": "user123",
    "name": "テストユーザー",
    "role": "Manager",
    "iat": 1678861200,
    "exp": 1678947600
  },
  "signature": "..."
}
```

### 3.4 技術スタック
- JWT実装ライブラリ: `github.com/golang-jwt/jwt/v5`
- パスワードハッシュ: bcrypt
- ミドルウェア実装: カスタム認証ミドルウェア

## 4. データベース設計

### 4.1 ユーザーテーブル
```sql
CREATE TABLE users (
    id VARCHAR(50) PRIMARY KEY,
    username VARCHAR(100) NOT NULL UNIQUE,
    password_hash VARCHAR(100) NOT NULL,
    role VARCHAR(20) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP
);
```

### 4.2 セッション管理（オプション）
```sql
CREATE TABLE sessions (
    id VARCHAR(100) PRIMARY KEY,
    user_id VARCHAR(50) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
```

## 5. 実装計画

### 5.1 インターフェース層

#### ドメインモデル (internal/domain/entity/user.go)
```go
package entity

type UserRole string

const (
    RoleAdmin   UserRole = "Admin"
    RoleManager UserRole = "Manager"
    RoleTester  UserRole = "Tester"
)

type User struct {
    ID           string
    Username     string
    PasswordHash string
    Role         UserRole
    CreatedAt    time.Time
    UpdatedAt    time.Time
    LastLoginAt  *time.Time
}

func (u *User) CanCreateTestSuite() bool {
    return u.Role == RoleAdmin || u.Role == RoleManager
}

func (u *User) CanUpdateTestCase() bool {
    return u.Role == RoleAdmin || u.Role == RoleManager || u.Role == RoleTester
}

// 他の権限チェックメソッド...
```

#### リポジトリ (internal/domain/repository/user_repository.go)
```go
package repository

type UserRepository interface {
    FindByID(ctx context.Context, id string) (*entity.User, error)
    FindByUsername(ctx context.Context, username string) (*entity.User, error)
    Create(ctx context.Context, user *entity.User) error
    Update(ctx context.Context, user *entity.User) error
    UpdateLastLogin(ctx context.Context, id string) error
}
```

### 5.2 ユースケース層

#### DTO (internal/usecase/dto/auth.go)
```go
package dto

type LoginRequestDTO struct {
    Username string `json:"username" validate:"required"`
    Password string `json:"password" validate:"required"`
}

type LoginResponseDTO struct {
    Token     string    `json:"token"`
    ExpiresAt time.Time `json:"expiresAt"`
    User      UserDTO   `json:"user"`
}

type UserDTO struct {
    ID       string `json:"id"`
    Username string `json:"username"`
    Role     string `json:"role"`
}
```

#### ユースケース (internal/usecase/interactor/auth_interactor.go)
```go
package interactor

type AuthInteractor struct {
    userRepository repository.UserRepository
    jwtService     service.JWTService
}

func (a *AuthInteractor) Login(ctx context.Context, dto *dto.LoginRequestDTO) (*dto.LoginResponseDTO, error) {
    // 実装詳細...
}

func (a *AuthInteractor) VerifyToken(ctx context.Context, token string) (*entity.User, error) {
    // 実装詳細...
}
```

### 5.3 インフラストラクチャ層

#### JWTサービス (internal/infrastructure/auth/jwt_service.go)
```go
package auth

type JWTService struct {
    secretKey []byte
    issuer    string
}

func (j *JWTService) GenerateToken(user *entity.User) (string, time.Time, error) {
    // 実装詳細...
}

func (j *JWTService) ValidateToken(tokenString string) (*jwt.Token, error) {
    // 実装詳細...
}

func (j *JWTService) ExtractUserID(token *jwt.Token) (string, error) {
    // 実装詳細...
}
```

#### Postgresリポジトリ (internal/infrastructure/persistence/postgres/user_repository.go)
```go
package postgres

type UserRepository struct {
    db *sql.DB
}

func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*entity.User, error) {
    // 実装詳細...
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, id string) error {
    // 実装詳細...
}

// 他のメソッド実装...
```

### 5.4 インターフェース層

#### 認証ミドルウェア (internal/interface/api/middleware/auth_middleware.go)
```go
package middleware

func AuthMiddleware(authUseCase usecase.AuthUseCase) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // トークン検証ロジック実装...
        })
    }
}
```

#### 認証ハンドラー (internal/interface/api/handler/auth_handler.go)
```go
package handler

type AuthHandler struct {
    authUseCase usecase.AuthUseCase
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
    // ログイン処理実装...
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
    // ユーザー登録処理実装...
}
```

## 6. REST API, GraphQL, gRPCの統合

### 6.1 REST API
- JWT認証ミドルウェアをハンドラーに適用
- Authorizationヘッダーによるトークン検証

### 6.2 GraphQL
- コンテキストにユーザー情報を追加
- リゾルバー内での権限チェック実装

### 6.3 gRPC
- インターセプターによる認証実装
- メタデータからトークン取得と検証

## 7. 作業スケジュール
1. **基本設計と準備** (1-2日)
   - インターフェース設計
   - 必要なライブラリの選定と導入

2. **基本認証機能の実装** (2-3日)
   - ユーザーモデルとリポジトリ
   - JWTサービス
   - ログイン/ログアウト機能

3. **アクセス制御の実装** (1-2日)
   - ロールベースの権限制御
   - 各APIへの認証適用

4. **統合とテスト** (1-2日)
   - REST/GraphQL/gRPCへの統合
   - テストコードの作成と実行

## 8. 今後の拡張ポイント
- 多要素認証（MFA）
- OAuth/OpenID連携
- パスワードリセット機能
- セッション管理の高度化
- ユーザー管理画面の実装
- 監査ログの記録と分析

## 9. まとめ
認証機能の実装は、AWS環境への移行において重要な要素です。本計画書で定義した最小限の認証機能を実装することで、セキュアなアプリケーション運用の基盤を整えることができます。JWTベースの認証により、ステートレスでスケーラブルな認証システムを実現し、将来的な機能拡張にも対応可能な設計となっています。