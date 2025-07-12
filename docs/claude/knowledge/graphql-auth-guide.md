# GraphQL認証とユーザー管理機能ガイド

## 1. 概要

本ガイドでは、AWS環境に構築されたGraphQLサービスにおける認証システムとユーザー管理機能の実装について解説します。これらの機能はAWS ECS上でデプロイされ、JWT（JSON Web Token）を使用した認証機構と、ロールベースのアクセス制御を実現しています。

### 1.1 実装された機能の全体像

- **認証機能**：
  - JWTベースのトークン認証
  - ロールベースのアクセス制御（Admin, Manager, Tester）
  - GraphQLディレクティブによる宣言的な権限制御（@auth, @hasRole）
  - リフレッシュトークンによるセッション管理

- **ユーザー管理機能**：
  - ユーザー作成（createUser）
  - パスワード変更（changePassword）
  - パスワードリセット（resetPassword）
  - ユーザー削除（deleteUser）
  - 入力バリデーション（ユーザー名長さ等）

### 1.2 技術スタック

- **バックエンド**：
  - Go言語（1.23）
  - GraphQL（gqlgen）
  - JWT認証（BCryptパスワードハッシュ）
  - PostgreSQL（RDS）

- **インフラストラクチャ**：
  - AWS ECS (Fargate)
  - AWS RDS (PostgreSQL)
  - AWS ALB
  - AWS SSM Parameter Store

- **CI/CD**：
  - GitLab CI/CD
  - Terraform
  - Docker

### 1.3 アーキテクチャ概要

```
┌───────────────────────────────────────────────────────────────────────┐
│                            AWS Environment                             │
│                                                                       │
│   ┌───────────┐       ┌─────────────┐       ┌─────────────────────┐   │
│   │ ALB       │       │ ECS Fargate │       │ RDS PostgreSQL      │   │
│   │           │       │             │       │                     │   │
│   │  ┌─────┐  │       │  ┌───────┐  │       │  ┌───────────────┐  │   │
│   │  │ 443 ├──┼───────┼─►│GraphQL│  │       │  │ Users Table   │  │   │
│   │  └─────┘  │       │  │Service│  │       │  └───────────────┘  │   │
│   │           │       │  └───┬───┘  │       │                     │   │
│   └───────────┘       └─────┼───────┘       └─────────────────────┘   │
│                             │                          ▲               │
│                             │                          │               │
│   ┌───────────────┐         │                          │               │
│   │ SSM Parameter │         │                          │               │
│   │ Store         │         │                          │               │
│   │  ┌─────────┐  │         │                          │               │
│   │  │Secrets  │◄─┼─────────┘                          │               │
│   │  └─────────┘  │                                    │               │
│   └───────────────┘                                    │               │
│                                                        │               │
└───────────────────────────────────────────────────────────────────────┘
                                   │                      │
                                   ▼                      │
┌─────────────────────────────────────────────────────────┴─────────────┐
│                    GraphQL Service Architecture                        │
│                                                                       │
│   ┌───────────────┐       ┌───────────────┐       ┌────────────────┐  │
│   │GraphQL API    │       │Usecase Layer  │       │Domain Layer    │  │
│   │  ┌─────────┐  │       │  ┌─────────┐  │       │  ┌──────────┐  │  │
│   │  │Resolver │──┼───────┼─►│Interactor│──┼───────┼─►│Entities │  │  │
│   │  └─────────┘  │       │  └─────────┘  │       │  └──────────┘  │  │
│   │  ┌─────────┐  │       │               │       │  ┌──────────┐  │  │
│   │  │Directive│  │       │               │       │  │Repository│◄─┼──┘
│   │  └─────────┘  │       │               │       │  │Interface │  │
│   │  ┌─────────┐  │       │               │       │  └──────────┘  │
│   │  │Middleware│  │       │               │       │               │
│   │  └─────────┘  │       │               │       │               │
│   └───────────────┘       └───────────────┘       └────────────────┘
│           ▲                                                ▲          │
│           │                                                │          │
│           ▼                                                ▼          │
│   ┌───────────────┐                               ┌────────────────┐  │
│   │Infrastructure │                               │Persistence     │  │
│   │  ┌─────────┐  │                               │  ┌──────────┐  │  │
│   │  │JWT      │  │                               │  │PostgreSQL│  │  │
│   │  │Service  │  │                               │  │Repository│  │  │
│   │  └─────────┘  │                               │  └──────────┘  │  │
│   │  ┌─────────┐  │                               │               │  │
│   │  │Password │  │                               │               │  │
│   │  │Service  │  │                               │               │  │
│   │  └─────────┘  │                               │               │  │
│   └───────────────┘                               └────────────────┘  │
│                                                                       │
└───────────────────────────────────────────────────────────────────────┘
```

## 2. 認証システムの設計と実装

### 2.1 認証アーキテクチャ

認証システムは以下のコンポーネントで構成されています：

```
┌─────────────────┐     ┌───────────────┐     ┌────────────────┐
│  GraphQL Client │────▶│ Auth Middleware│────▶│ Auth Directive │
└─────────────────┘     └───────────────┘     └────────────────┘
         │                      │                      │
         │                      │                      │
         ▼                      ▼                      ▼
┌─────────────────┐     ┌───────────────┐     ┌────────────────┐
│  Context Auth   │◀───▶│  JWT Service  │     │ User Repository │
└─────────────────┘     └───────────────┘     └────────────────┘
                              │                      │
                              │                      │
                              ▼                      ▼
                        ┌───────────────┐     ┌────────────────┐
                        │Password Service│     │  PostgreSQL DB │
                        └───────────────┘     └────────────────┘
```

1. **認証ミドルウェア**（`middleware.go`）：
   - HTTPリクエストからBearerトークンを抽出
   - JWTを検証し、有効な場合はユーザー情報をコンテキストに設定

2. **認証ディレクティブ**（`auth.go`）：
   - GraphQLスキーマに定義された`@auth`と`@hasRole`ディレクティブを処理
   - リクエストコンテキストから認証情報をチェック

3. **コンテキスト管理**（`context.go`）：
   - リクエストコンテキストに認証情報を保存・取得する機能を提供

4. **JWTサービス**（`jwt_service.go`）：
   - トークンの生成と検証を担当
   - 有効期限や署名の検証

5. **パスワードサービス**（`password_service.go`）：
   - BCryptによるパスワードのハッシュ化と検証
   - セキュアなパスワード管理を提供

### 2.2 JWT認証フロー

#### 認証フローの概要：

1. **ログイン**：
   ```graphql
   mutation {
     login(username: "demo_user", password: "password") {
       token
       refreshToken
       user {
         id
         username
         role
       }
       expiresAt
     }
   }
   ```

2. **認証ヘッダーの使用**：
   ```
   Authorization: Bearer <jwt_token>
   ```

3. **認証状態の確認**：
   ```graphql
   query {
     me {
       id
       username
       role
     }
   }
   ```

4. **トークン更新**：
   ```graphql
   mutation {
     refreshToken(refreshToken: "<refresh_token>") {
       token
       refreshToken
       expiresAt
     }
   }
   ```

#### 実装詳細（`middleware.go`）：

```go
// AuthMiddleware はHTTPリクエストからJWTトークンを抽出し、
// トークンを検証してユーザー情報をコンテキストに設定するミドルウェアです
func AuthMiddleware(authUseCase port.AuthUseCase) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // リクエストヘッダーからトークンを抽出
            token := extractTokenFromHeader(r)

            // トークンが存在する場合のみ検証を行う
            if token != "" {
                // トークンを検証してユーザー情報を取得
                user, err := authUseCase.ValidateToken(r.Context(), token)
                if err != nil {
                    // トークン検証エラーはログに記録するが、
                    // リクエスト自体は拒否せず、非認証状態として処理を続行
                    log.Printf("Token validation failed: %v", err)
                } else if user != nil {
                    // 有効なトークンの場合、認証情報をコンテキストに設定
                    authInfo := &AuthInfo{
                        User:         user,
                        IsAuthorized: true,
                    }
                    ctx := SetAuthInfo(r.Context(), authInfo)
                    r = r.WithContext(ctx)
                }
            }

            // 次のハンドラーを呼び出す
            next.ServeHTTP(w, r)
        })
    }
}
```

### 2.3 認証ディレクティブ

GraphQLスキーマでは以下のディレクティブを使用して権限制御を行っています：

```graphql
# ディレクティブの定義
directive @auth on FIELD_DEFINITION
directive @hasRole(role: String!) on FIELD_DEFINITION

# 認証関連のミューテーション
extend type Mutation {
  # ログイン操作：ユーザー名とパスワードでユーザーを認証し、トークンを取得
  login(username: String!, password: String!): AuthPayload!
  
  # トークン更新：リフレッシュトークンを使用して新しいアクセストークンを取得
  refreshToken(refreshToken: String!): AuthPayload!
  
  # ログアウト：リフレッシュトークンを無効化
  logout(refreshToken: String!): Boolean! @auth
}

# 認証関連のクエリ
extend type Query {
  # 認証済みユーザー情報取得
  me: User! @auth
}
```

これらのディレクティブは、以下のようにフィールドに対して権限制御を行います：

- **@auth**：認証済みユーザーのみアクセス可能
- **@hasRole(role: "xxx")**：指定したロールを持つユーザーのみアクセス可能

### 2.4 コンテキスト管理

認証情報はリクエストのコンテキストに保存され、GraphQLリゾルバーからアクセスできます：

```go
// AuthInfo は認証情報を保持する構造体
type AuthInfo struct {
    User         *entity.User // 認証済みユーザー
    IsAuthorized bool         // 認証済みかどうか
}

// GetUserFromContext はコンテキストからユーザー情報を取得する便利関数
func GetUserFromContext(ctx context.Context) *entity.User {
    authInfo := GetAuthInfo(ctx)
    if authInfo == nil || !authInfo.IsAuthorized {
        return nil
    }
    return authInfo.User
}

// IsAuthenticated はコンテキストから認証済みかどうかを確認する便利関数
func IsAuthenticated(ctx context.Context) bool {
    authInfo := GetAuthInfo(ctx)
    return authInfo != nil && authInfo.IsAuthorized
}

// HasRole はユーザーが指定したロールを持っているか確認する便利関数
func HasRole(ctx context.Context, role entity.UserRole) bool {
    user := GetUserFromContext(ctx)
    if user == nil {
        return false
    }
    return user.Role == role
}
```

### 2.5 パスワード管理

パスワードは安全にハッシュ化されて保存されます。BCryptアルゴリズムを使用しています：

```go
// BCryptPasswordService はbcryptを使用したパスワードサービスの実装
type BCryptPasswordService struct {
    cost int
}

// HashPassword はパスワードをbcryptを使用してハッシュ化する
func (s *BCryptPasswordService) HashPassword(password string) (string, error) {
    if password == "" {
        return "", customerrors.NewValidationError("password cannot be empty", nil)
    }

    hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), s.cost)
    if err != nil {
        return "", customerrors.WrapInternalServerError(err, "failed to hash password")
    }

    return string(hashedBytes), nil
}

// VerifyPassword はパスワードとハッシュが一致するかbcryptを使用して検証する
func (s *BCryptPasswordService) VerifyPassword(password, hash string) error {
    if password == "" {
        return customerrors.NewValidationError("password cannot be empty", nil)
    }

    if hash == "" {
        return customerrors.NewValidationError("hash cannot be empty", nil)
    }

    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    if err != nil {
        if err == bcrypt.ErrMismatchedHashAndPassword {
            return customerrors.NewUnauthorizedError("password does not match")
        }
        return customerrors.WrapInternalServerError(err, "failed to verify password")
    }

    return nil
}
```

## 3. ユーザー管理機能

### 3.1 ユーザーモデル

ユーザーモデルは以下の属性を持ちます：

```go
// User はシステムユーザーを表す
type User struct {
    ID           string
    Username     string
    PasswordHash string
    Role         UserRole
    CreatedAt    time.Time
    UpdatedAt    time.Time
    LastLoginAt  *time.Time
}
```

GraphQL側では以下のように定義されています：

```graphql
type User {
  id: ID!
  username: String!
  role: String!
  createdAt: DateTime!
  updatedAt: DateTime!
  lastLoginAt: DateTime
}
```

### 3.2 ロール管理

システムでは以下のロールを定義しています：

```go
// UserRole はユーザーの役割を表す
type UserRole string

const (
    RoleAdmin   UserRole = "Admin"
    RoleManager UserRole = "Manager"
    RoleTester  UserRole = "Tester"
)
```

各ロールには以下の権限があります：

- **Admin**：すべての操作が可能（ユーザー管理含む）
- **Manager**：テストスイート作成・更新、ケース管理
- **Tester**：テストケース更新、工数記録

これらの権限は、ユーザーエンティティに以下のように実装されています：

```go
// CanCreateTestSuite はテストスイート作成権限を持つかチェック
func (u *User) CanCreateTestSuite() bool {
    return u.Role == RoleAdmin || u.Role == RoleManager
}

// CanUpdateTestSuite はテストスイート更新権限を持つかチェック
func (u *User) CanUpdateTestSuite() bool {
    return u.Role == RoleAdmin || u.Role == RoleManager
}

// CanViewTestSuite はテストスイート閲覧権限を持つかチェック
func (u *User) CanViewTestSuite() bool {
    return true // すべてのユーザーが閲覧可能
}

// CanUpdateTestCase はテストケース更新権限を持つかチェック
func (u *User) CanUpdateTestCase() bool {
    return true // すべてのユーザーがテストケースを更新可能
}

// CanRecordEffort は工数記録権限を持つかチェック
func (u *User) CanRecordEffort() bool {
    return true // すべてのユーザーが工数を記録可能
}
```

### 3.3 ユーザー管理機能の実装

GraphQLスキーマでは以下のユーザー管理ミューテーションを定義しています：

```graphql
# ユーザー管理関連のミューテーション
extend type Mutation {
  # 新規ユーザーの作成（管理者権限が必要）
  createUser(input: CreateUserInput!): User! @hasRole(role: "Admin")
  
  # ユーザー自身のパスワード変更（認証が必要）
  changePassword(oldPassword: String!, newPassword: String!): Boolean! @auth
  
  # 他のユーザーのパスワードリセット（管理者権限が必要）
  resetPassword(userId: ID!, newPassword: String!): Boolean! @hasRole(role: "Admin")

  # ユーザーの削除（管理者権限が必要）
  deleteUser(userId: ID!): Boolean! @hasRole(role: "Admin")
}

# ユーザー作成入力
input CreateUserInput {
  username: String!
  password: String!
  role: String!
}
```

#### 3.3.1 createUser実装

新規ユーザーを作成する機能です。Admin権限が必要です。

```go
// CreateUser は新規ユーザーを作成します
func (r *mutationResolver) CreateUser(ctx context.Context, input model.CreateUserInput) (*model.User, error) {
    // リクエストデータの準備
    request := &port.CreateUserRequest{
        Username: input.Username,
        Password: input.Password,
        Role:     input.Role,
    }

    // ユースケースの呼び出し
    user, err := r.UserManagementUseCase.CreateUser(ctx, request)
    if err != nil {
        return nil, err
    }

    // エンティティからGraphQLモデルへの変換
    return &model.User{
        ID:        user.ID,
        Username:  user.Username,
        Role:      user.Role.String(),
        CreatedAt: user.CreatedAt,
        UpdatedAt: user.UpdatedAt,
    }, nil
}
```

インタラクター（ビジネスロジック）側では、以下のようにバリデーションを実装しています：

```go
// CreateUser は新しいユーザーを作成します
func (i *UserManagementInteractor) CreateUser(ctx context.Context, request *port.CreateUserRequest) (*entity.User, error) {
    // ユーザー名のバリデーション
    if request.Username == "" {
        return nil, customerrors.NewValidationError("ユーザー名は必須です", map[string]string{
            "username": "ユーザー名を入力してください",
        })
    }
    if len(request.Username) < 3 {
        return nil, customerrors.NewValidationError("ユーザー名が短すぎます", map[string]string{
            "username": "ユーザー名は3文字以上で入力してください",
        })
    }

    // ユーザー名が既に存在するか確認
    existingUser, err := i.userRepository.FindByUsername(ctx, request.Username)
    if err != nil && !customerrors.IsNotFoundError(err) {
        return nil, err
    }
    if existingUser != nil {
        return nil, customerrors.NewConflictError("ユーザー名が既に使用されています")
    }

    // パスワードのバリデーション
    if request.Password == "" {
        return nil, customerrors.NewValidationError("パスワードは必須です", map[string]string{
            "password": "パスワードを入力してください",
        })
    }
    if len(request.Password) < 6 {
        return nil, customerrors.NewValidationError("パスワードが短すぎます", map[string]string{
            "password": "パスワードは6文字以上で入力してください",
        })
    }

    // パスワードのハッシュ化
    passwordHash, err := i.passwordService.HashPassword(request.Password)
    if err != nil {
        return nil, customerrors.WrapInternalServerError(err, "パスワードのハッシュ化に失敗しました")
    }

    // ユーザーロールの検証
    var userRole entity.UserRole
    switch request.Role {
    case "Admin":
        userRole = entity.RoleAdmin
    case "Manager":
        userRole = entity.RoleManager
    case "Tester":
        userRole = entity.RoleTester
    default:
        return nil, customerrors.NewValidationError("無効なユーザーロールです", map[string]string{
            "role": "Admin, Manager, Tester のいずれかを指定してください",
        })
    }

    // ユーザーIDの生成
    userID, err := i.userIDGenerator.Generate(ctx)
    if err != nil {
        return nil, customerrors.WrapInternalServerError(err, "ユーザーID生成に失敗しました")
    }

    // ユーザーエンティティの作成
    user, err := entity.NewUser(userID, request.Username, passwordHash, userRole)
    if err != nil {
        return nil, customerrors.WrapValidationError(err, "ユーザー作成に失敗しました", nil)
    }

    // ユーザーの保存
    if err := i.userRepository.Create(ctx, user); err != nil {
        return nil, customerrors.WrapInternalServerError(err, "ユーザーの保存に失敗しました")
    }

    return user, nil
}
```

#### 3.3.2 changePassword実装

ユーザー自身のパスワードを変更する機能です。認証が必要です。

```go
// ChangePassword はユーザー自身のパスワードを変更します
func (r *mutationResolver) ChangePassword(ctx context.Context, oldPassword string, newPassword string) (bool, error) {
    // 認証コンテキストからユーザーIDを取得
    user := auth.GetUserFromContext(ctx)
    if user == nil {
        return false, nil
    }

    // ユースケースの呼び出し
    err := r.UserManagementUseCase.ChangePassword(ctx, user.ID, oldPassword, newPassword)
    if err != nil {
        return false, err
    }

    return true, nil
}
```

#### 3.3.3 resetPassword実装

他のユーザーのパスワードをリセットする機能です。Admin権限が必要です。

```go
// ResetPassword は管理者が他のユーザーのパスワードをリセットします
func (r *mutationResolver) ResetPassword(ctx context.Context, userID string, newPassword string) (bool, error) {
    // ユースケースの呼び出し
    err := r.UserManagementUseCase.ResetPassword(ctx, userID, newPassword)
    if err != nil {
        return false, err
    }

    return true, nil
}
```

#### 3.3.4 deleteUser実装

ユーザーを削除する機能です。Admin権限が必要です。

```go
// DeleteUser はユーザーを削除します
func (r *mutationResolver) DeleteUser(ctx context.Context, userID string) (bool, error) {
    // ユースケースの呼び出し
    err := r.UserManagementUseCase.DeleteUser(ctx, userID)
    if err != nil {
        return false, err
    }

    return true, nil
}
```

インタラクターの実装：

```go
// DeleteUser はユーザーを削除します
func (i *UserManagementInteractor) DeleteUser(ctx context.Context, userID string) error {
    // ユーザーの存在確認
    _, err := i.userRepository.FindByID(ctx, userID)
    if err != nil {
        if customerrors.IsNotFoundError(err) {
            return customerrors.NewNotFoundError("ユーザーが見つかりません")
        }
        return customerrors.WrapInternalServerError(err, "ユーザー情報の取得に失敗しました")
    }

    // ユーザーの削除
    if err := i.userRepository.Delete(ctx, userID); err != nil {
        return customerrors.WrapInternalServerError(err, "ユーザーの削除に失敗しました")
    }

    return nil
}
```

### 3.4 バリデーション機能

ユーザー管理機能では、以下のバリデーションを実装しています：

1. **ユーザー名バリデーション**:
   - 必須項目チェック
   - 長さチェック（3文字以上）
   - 重複チェック（既存ユーザー名との比較）

2. **パスワードバリデーション**:
   - 必須項目チェック
   - 長さチェック（6文字以上）

3. **ロールバリデーション**:
   - 有効なロール値（Admin, Manager, Tester）のみ許可

これらのバリデーションは、ユースケース層（インタラクター）で実装されており、GraphQLクライアントに適切なエラーメッセージが返されます。

```go
// バリデーションエラーの例
if len(request.Username) < 3 {
    return nil, customerrors.NewValidationError("ユーザー名が短すぎます", map[string]string{
        "username": "ユーザー名は3文字以上で入力してください",
    })
}
```

### 3.5 エラーハンドリング

ユーザー管理機能では、`customerrors`パッケージを使用して一貫したエラーハンドリングを実現しています。以下のエラータイプが使用されています：

1. **ValidationError**: バリデーション失敗時のエラー
2. **NotFoundError**: リソースが見つからない場合のエラー
3. **ConflictError**: 一意性制約違反などの競合エラー
4. **UnauthorizedError**: 認証失敗時のエラー
5. **InternalServerError**: 内部エラー

これらのエラーは、GraphQLレスポンスに適切に変換され、クライアントに返されます。

## 4. テスト結果と検証

GraphQL認証およびユーザー管理機能の動作確認として、以下のテストを実施しました。

### 4.1 テスト環境

- **AWS環境**：ECSコンテナでGraphQLサービスを実行
- **テストツール**：GraphQL Playground
- **テストユーザー**：
  - demo_user (Admin権限, ID: USER001)
  - test_manager (Manager権限, ID: USER002)
  - test_tester (Tester権限, ID: USER003)

### 4.2 認証機能テスト

#### 4.2.1 ログインテスト

```graphql
mutation {
  login(username: "demo_user", password: "password") {
    token
    refreshToken
    user {
      id
      username
      role
    }
    expiresAt
  }
}
```

**結果**：正常にJWTトークンが発行され、ユーザー情報が返却されました。

```json
{
  "data": {
    "login": {
      "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
      "refreshToken": "d8f5b885-8d35-4...",
      "user": {
        "id": "USER001",
        "username": "demo_user",
        "role": "Admin"
      },
      "expiresAt": "2025-05-29T15:30:00Z"
    }
  }
}
```

#### 4.2.2 認証済みユーザー情報取得テスト

```graphql
# Authorizationヘッダー: Bearer <token>
query {
  me {
    id
    username
    role
  }
}
```

**結果**：認証トークンを使用してユーザー情報が正常に取得できました。

```json
{
  "data": {
    "me": {
      "id": "USER001",
      "username": "demo_user",
      "role": "Admin"
    }
  }
}
```

#### 4.2.3 認証失敗テスト

- 無効なトークンでのアクセス
- トークンなしでの認証要求エンドポイントアクセス

**結果**：適切なエラーメッセージ（「アクセス権限がありません」）が返却されました。

```json
{
  "errors": [
    {
      "message": "アクセス権限がありません",
      "path": ["me"],
      "extensions": {
        "code": "UNAUTHENTICATED"
      }
    }
  ],
  "data": null
}
```

### 4.3 ユーザー管理機能テスト

#### 4.3.1 ユーザー作成テスト

```graphql
# Authorizationヘッダー: Bearer <admin_token>
mutation {
  createUser(input: {
    username: "new_test_user", 
    password: "password123", 
    role: "Tester"
  }) {
    id
    username
    role
  }
}
```

**結果**：新規ユーザーが正常に作成されました。

```json
{
  "data": {
    "createUser": {
      "id": "USER004",
      "username": "new_test_user",
      "role": "Tester"
    }
  }
}
```

#### 4.3.2 バリデーションテスト

```graphql
# Authorizationヘッダー: Bearer <admin_token>
mutation {
  createUser(input: {
    username: "a",  # 短すぎるユーザー名
    password: "password123", 
    role: "Tester"
  }) {
    id
    username
    role
  }
}
```

**結果**：「ユーザー名が短すぎます」エラーが正常に返却されました。バリデーション機能が適切に動作しています。

```json
{
  "errors": [
    {
      "message": "ユーザー名が短すぎます",
      "path": ["createUser"],
      "extensions": {
        "code": "BAD_USER_INPUT",
        "field_errors": {
          "username": "ユーザー名は3文字以上で入力してください"
        }
      }
    }
  ],
  "data": null
}
```

#### 4.3.3 権限制御テスト

```graphql
# Authorizationヘッダー: Bearer <manager_token> (Managerロール)
mutation {
  deleteUser(userId: "USER003")
}
```

**結果**：「アクセス権限がありません」エラーが返却され、Admin以外のユーザーがdeleteUser機能を使用できないことが確認されました。

```json
{
  "errors": [
    {
      "message": "アクセス権限がありません",
      "path": ["deleteUser"],
      "extensions": {
        "code": "FORBIDDEN"
      }
    }
  ],
  "data": null
}
```

#### 4.3.4 ユーザー削除テスト

```graphql
# Authorizationヘッダー: Bearer <admin_token>
mutation {
  deleteUser(userId: "USER003")
}
```

**結果**：ユーザーが正常に削除されました。

```json
{
  "data": {
    "deleteUser": true
  }
}
```

#### 4.3.5 存在しないユーザー削除テスト

```graphql
# Authorizationヘッダー: Bearer <admin_token>
mutation {
  deleteUser(userId: "NONEXISTENT")
}
```

**結果**：適切なエラーメッセージが返却され、存在しないユーザーの削除処理が適切に処理されました。

```json
{
  "errors": [
    {
      "message": "ユーザーが見つかりません",
      "path": ["deleteUser"],
      "extensions": {
        "code": "NOT_FOUND"
      }
    }
  ],
  "data": null
}
```

### 4.4 テスト結果の総括

- **認証機能**：JWTトークンの生成、検証、認証ディレクティブが正常に動作
- **ユーザー管理機能**：ユーザー作成、削除が正常に動作
- **バリデーション**：入力値の検証が適切に行われ、エラーメッセージが表示される
- **権限制御**：ロールに基づいたアクセス制御が正しく機能している
- **エラーハンドリング**：存在しないリソースへのアクセス時などに適切なエラーが返却される

これらのテスト結果から、GraphQL認証機能とユーザー管理機能が期待通りに動作していることが確認できました。

## 5. デプロイメントプロセス

### 5.1 GraphQLスキーマ更新フロー

GraphQLスキーマの更新は、以下のフローで行います：

1. スキーマファイル（`*.graphqls`）を編集
2. gqlgenコマンドを実行してコードを生成
3. 生成されたコードをデプロイ

これを自動化するMakefileのコマンドが用意されています：

```makefile
# 最高速度版（検証スキップ） - gqlgen実行を含む修正版
.PHONY: fastest-update-graphql

fastest-update-graphql:
	@echo -e "${BLUE}========== 最速GraphQL更新開始 ==========${NC}"
	@START_TIME=$(date +%s); \
	\
	echo -e "${BLUE}[1/5] GraphQLスキーマを更新しています...${NC}"; \
	if command -v gqlgen >/dev/null 2>&1; then \
		echo -e "${GREEN}ローカルのgqlgenを使用します${NC}"; \
		gqlgen generate; \
	else \
		echo -e "${YELLOW}ローカルのgqlgenが見つかりません。自動セットアップします${NC}"; \
		go mod tidy > /dev/null 2>&1 || true; \
		go run github.com/99designs/gqlgen generate; \
	fi; \
	\
	echo -e "${GREEN}✓ GraphQLスキーマ更新完了${NC}"; \
	\
	# 以下、イメージビルドとデプロイ処理
```

gqlgenの設定は`gqlgen.yml`で管理されており、以下のように構成されています：

```yaml
schema:
  - internal/interface/graphql/schema/schema.graphqls
  - internal/interface/graphql/schema/auth.graphqls  # 認証スキーマを追加
  - internal/interface/graphql/schema/role_test.graphqls
  - internal/interface/graphql/schema/user_management.graphqls  # ← 追加

exec:
  filename: internal/interface/graphql/generated/generated.go
  package: generated

model:
  filename: internal/interface/graphql/model/models_gen.go
  package: model

resolver:
  layout: follow-schema
  dir: internal/interface/graphql/resolver
  package: resolver

# モデル定義
models:
  ID:
    model:
      - github.com/99designs/gqlgen/graphql.ID
  # その他のモデル定義...

# ディレクティブ設定
directives:
  auth: {}
  hasRole: {}
```

### 5.2 AWS環境へのデプロイプロセス

AWS環境へのデプロイは、以下のコマンドで行います：

```bash
# GraphQLサービスをデプロイ
make deploy-graphql-new-dev
```

このコマンドは、以下の処理を実行します：

1. GraphQLスキーマの更新（gqlgen実行）
2. Dockerイメージのビルド
3. ECRへのプッシュ
4. Terraformによるインフラ更新
5. ECSサービスの更新

Terraformによるデプロイは、以下のモジュールを使用します：

```
deployments/terraform/modules/service/graphql/
├── ecs-service/
│   ├── main.tf
│   ├── variables.tf
│   └── outputs.tf
├── load-balancer/
│   ├── main.tf
│   ├── variables.tf
│   └── outputs.tf
└── target-group/
    ├── main.tf
    ├── outputs.tf
    └── variables.tf
```

これらのモジュールにより、GraphQLサービス用のECSサービス、ロードバランサー、ターゲットグループが設定されます。

### 5.3 GitLab CI/CD統合

GitLab CI/CDでは、以下のジョブを定義しています：

```yaml
stages:
  - migrate
  - seed

migrate:database:
  stage: migrate
  script:
    - bash scripts/terraform/aws-migrate-ci.sh $TF_ENV
  when: manual
  timeout: 45m

seed:test-users:
  stage: seed
  script:
    - bash scripts/terraform/aws-seed-user.sh $TF_ENV
  dependencies:
    - migrate:database
  when: manual
  timeout: 30m
```

これらのジョブにより、以下の処理が自動化されています：

1. **migrate:database**：データベースマイグレーションを実行
2. **seed:test-users**：テストユーザーデータの投入

テストユーザー投入スクリプト（`aws-seed-user.sh`）は、以下のテストユーザーを作成します：

- demo_user (パスワード: password, ロール: Admin, ID: USER001)
- test_manager (パスワード: password, ロール: Manager, ID: USER002)
- test_tester (パスワード: password, ロール: Tester, ID: USER003)

### 5.4 デプロイフロー全体図

```
┌───────────────┐     ┌───────────────┐     ┌───────────────┐
│ ローカル開発   │     │ GitLab CI/CD  │     │ AWS環境       │
└───────┬───────┘     └───────┬───────┘     └───────┬───────┘
        │                     │                     │
        ▼                     │                     │
┌───────────────┐             │                     │
│ GraphQLスキーマ│             │                     │
│   編集        │             │                     │
└───────┬───────┘             │                     │
        │                     │                     │
        ▼                     │                     │
┌───────────────┐             │                     │
│ gqlgen実行    │             │                     │
└───────┬───────┘             │                     │
        │                     │                     │
        ▼                     │                     │
┌───────────────┐             │                     │
│ Dockerイメージ │             │                     │
│   ビルド      │             │                     │
└───────┬───────┘             │                     │
        │                     │                     │
        ▼                     │                     │
┌───────────────┐             │                     │
│ ECRプッシュ    │             │                     │
└───────┬───────┘             │                     │
        │                     ▼                     │
        │             ┌───────────────┐             │
        │             │ マイグレーション │             │
        │             │   実行        │             │
        │             └───────┬───────┘             │
        │                     │                     │
        │                     ▼                     │
        │             ┌───────────────┐             │
        │             │テストユーザー投入│             │
        │             └───────┬───────┘             │
        │                     │                     │
        ▼                     ▼                     ▼
┌──────────────────────────────────────────────────────┐
│                  Terraform適用                        │
├──────────────────────────────────────────────────────┤
│  - VPC、サブネット、セキュリティグループ               │
│  - RDSデータベース                                   │
│  - ECSクラスター                                     │
│  - ALB、ターゲットグループ                           │
│  - ECSサービス（GraphQL）                            │
└──────────────────────────────────────────────────────┘
                          │
                          ▼
┌──────────────────────────────────────────────────────┐
│                 AWS環境にデプロイ完了                  │
└──────────────────────────────────────────────────────┘
```

## 6. 課題と今後の拡張

複数のペルソナ（開発者、セキュリティ専門家、UX設計者、アーキテクト、プロジェクトマネージャー）による議論と分析の結果、以下の課題と拡張案が特定されました。

### 6.1 最優先課題

#### 6.1.1 ユーザー一覧機能の追加

現在のGraphQLスキーマには、ユーザー一覧を取得するクエリが実装されていません。ユーザー管理画面を実装する場合は、この機能が必要です。

**問題点**：
- 管理者がユーザーを管理するためには、ユーザー一覧が必要
- 現在は個々のユーザーIDを知っている場合のみ操作可能
- フロントエンド実装に必須の機能が欠如

**推奨されるスキーマ追加**：

```graphql
extend type Query {
  # ユーザー一覧取得（管理者権限が必要）
  users: [User!]! @hasRole(role: "Admin")
  
  # 特定ユーザーの取得（管理者権限が必要）
  user(id: ID!): User @hasRole(role: "Admin")
  
  # 検索条件によるユーザー一覧取得（管理者権限が必要）
  searchUsers(filter: UserFilterInput): [User!]! @hasRole(role: "Admin")
}

# ユーザー検索フィルター
input UserFilterInput {
  username: String
  role: String
  createdAfter: DateTime
  createdBefore: DateTime
}
```

**実装イメージ**：

```go
// SearchUsers はユーザーを検索します
func (r *queryResolver) SearchUsers(ctx context.Context, filter *model.UserFilterInput) ([]*model.User, error) {
    // 検索条件の構築
    criteria := &repository.UserSearchCriteria{}
    if filter != nil {
        if filter.Username != nil {
            criteria.Username = *filter.Username
        }
        if filter.Role != nil {
            criteria.Role = *filter.Role
        }
        // その他の条件...
    }
    
    // リポジトリからユーザーを検索
    users, err := r.UserRepository.Search(ctx, criteria)
    if err != nil {
        return nil, err
    }
    
    // エンティティからGraphQLモデルへの変換
    result := make([]*model.User, len(users))
    for i, user := range users {
        result[i] = &model.User{
            ID:        user.ID,
            Username:  user.Username,
            Role:      user.Role.String(),
            CreatedAt: user.CreatedAt,
            UpdatedAt: user.UpdatedAt,
        }
    }
    
    return result, nil
}
```

#### 6.1.2 最後のAdmin削除防止機能

現在の実装では、最後のAdminユーザーを削除できてしまう可能性があります。これにより、管理者権限を持つユーザーがいなくなり、管理機能が使用できなくなるリスクがあります。

**問題点**：
- システム管理上の重大なリスク
- 回復するには直接データベース操作が必要になる可能性がある
- 単一障害点（SPOF）となる

**推奨される改善策**：

```go
// DeleteUser実装例
func (i *UserManagementInteractor) DeleteUser(ctx context.Context, userID string) error {
    // 削除対象ユーザーのロールを確認
    user, err := i.userRepository.FindByID(ctx, userID)
    if err != nil {
        if customerrors.IsNotFoundError(err) {
            return customerrors.NewNotFoundError("ユーザーが見つかりません")
        }
        return customerrors.WrapInternalServerError(err, "ユーザー情報の取得に失敗しました")
    }
    
    // 最後のAdminユーザーかどうかチェック
    if user.Role == entity.RoleAdmin {
        // Adminユーザーの総数をカウント
        adminUsers, err := i.userRepository.FindByRole(ctx, entity.RoleAdmin)
        if err != nil {
            return customerrors.WrapInternalServerError(err, "管理者ユーザーの取得に失敗しました")
        }
        
        // 最後のAdminユーザーを削除しようとしている場合はエラー
        if len(adminUsers) <= 1 {
            return customerrors.NewValidationError("最後の管理者ユーザーは削除できません", nil)
        }
    }
    
    // 通常の削除処理
    if err := i.userRepository.Delete(ctx, userID); err != nil {
        return customerrors.WrapInternalServerError(err, "ユーザーの削除に失敗しました")
    }
    
    return nil
}
```

### 6.2 中優先課題

#### 6.2.1 パスワード強度検証の強化

現在のパスワードバリデーションは長さのみをチェックしていますが、セキュリティを向上させるためにはより強力な検証が必要です。

**問題点**：
- 現在は6文字以上という長さのみの検証
- 単純なパスワードを許容してしまう
- セキュリティリスクの増加

**推奨される改善策**：

```go
// パスワード強度検証関数の例
func validatePasswordStrength(password string) (bool, map[string]string) {
    errors := make(map[string]string)
    
    // 長さチェック
    if len(password) < 8 {
        errors["length"] = "パスワードは8文字以上である必要があります"
    }
    
    // 大文字含有チェック
    if !regexp.MustCompile(`[A-Z]`).MatchString(password) {
        errors["uppercase"] = "パスワードには少なくとも1つの大文字を含める必要があります"
    }
    
    // 数字含有チェック
    if !regexp.MustCompile(`[0-9]`).MatchString(password) {
        errors["digit"] = "パスワードには少なくとも1つの数字を含める必要があります"
    }
    
    // 特殊文字含有チェック
    if !regexp.MustCompile(`[!@#$%^&*]`).MatchString(password) {
        errors["special"] = "パスワードには少なくとも1つの特殊文字(!@#$%^&*)を含める必要があります"
    }
    
    return len(errors) == 0, errors
}
```

#### 6.2.2 監査ログ機能

ユーザー管理操作のログ記録機能は、セキュリティ監査やトラブルシューティングに役立ちます。

**問題点**：
- 現状ではユーザー操作の記録がない
- セキュリティインシデント調査が困難
- コンプライアンス要件を満たせない可能性がある

**推奨される改善策**：

```go
// 監査ログサービスのインターフェース
type AuditLogService interface {
    LogUserCreation(ctx context.Context, adminID, createdUserID string) error
    LogUserDeletion(ctx context.Context, adminID, deletedUserID string) error
    LogPasswordChange(ctx context.Context, userID string, byAdmin bool) error
    LogAuthFailure(ctx context.Context, username string, reason string) error
    QueryLogs(ctx context.Context, filter LogFilter) ([]*AuditLog, error)
}

// インタラクターへの統合例
func (i *UserManagementInteractor) DeleteUser(ctx context.Context, userID string) error {
    // 認証済みユーザーの取得
    authUser := auth.GetUserFromContext(ctx)
    
    // ... 既存の処理 ...
    
    // 削除処理
    if err := i.userRepository.Delete(ctx, userID); err != nil {
        return err
    }
    
    // 監査ログの記録
    if err := i.auditLogService.LogUserDeletion(ctx, authUser.ID, userID); err != nil {
        // ログ記録エラーは処理を継続させる（警告のみ）
        log.Printf("Warning: Failed to log user deletion: %v", err)
    }
    
    return nil
}
```

### 6.3 低優先課題

#### 6.3.1 アカウントロックアウト機能

複数回のログイン失敗によるアカウントロック機能は、ブルートフォース攻撃から保護するために有効です。

**問題点**：
- 現状では連続ログイン失敗の制限がない
- ブルートフォース攻撃に脆弱
- アカウント保護機能の欠如

**推奨される改善策**：

```go
// ログイン試行管理サービス
type LoginAttemptService interface {
    RecordAttempt(ctx context.Context, username string, success bool) error
    GetFailedAttempts(ctx context.Context, username string, duration time.Duration) (int, error)
    ResetAttempts(ctx context.Context, username string) error
}

// ログイン処理の改善例
func (i *AuthInteractor) Login(ctx context.Context, username, password string) (*LoginResponse, error) {
    // 失敗回数の確認
    failedAttempts, err := i.loginAttemptService.GetFailedAttempts(ctx, username, time.Hour)
    if err != nil {
        return nil, err
    }
    
    // ロックアウト確認（例: 1時間以内に5回失敗でロック）
    if failedAttempts >= 5 {
        return nil, customerrors.NewUnauthorizedError("アカウントがロックされています。しばらくしてから再試行してください")
    }
    
    // ... 通常のログイン処理 ...
    
    // 結果の記録
    if err != nil {
        i.loginAttemptService.RecordAttempt(ctx, username, false)
        return nil, err
    }
    
    // 成功時は試行回数をリセット
    i.loginAttemptService.ResetAttempts(ctx, username)
    
    return response, nil
}
```

#### 6.3.2 セルフサービスパスワードリセット

ユーザーが自分でパスワードをリセットできる機能は、管理者の負担を減らし、ユーザー体験を向上させます。

**問題点**：
- 現状ではパスワードリセットは管理者のみが実行可能
- ユーザー体験の低下
- 管理者の負担増加

**推奨される改善策**：

```graphql
extend type Mutation {
  # パスワードリセットリクエスト（メールアドレスによる）
  requestPasswordReset(email: String!): Boolean!
  
  # パスワードリセットトークンによるリセット実行
  resetPasswordWithToken(token: String!, newPassword: String!): Boolean!
}
```

### 6.4 フロントエンド連携のガイドライン

フロントエンド（React）とGraphQL APIを連携する際のガイドラインです。

#### 6.4.1 認証フロー実装

```typescript
// AuthContext.tsx
import { createContext, useState, useContext, ReactNode } from 'react';

type AuthContextType = {
  token: string | null;
  login: (token: string) => void;
  logout: () => void;
  isAuthenticated: boolean;
};

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider = ({ children }: { children: ReactNode }) => {
  const [token, setToken] = useState<string | null>(
    localStorage.getItem('authToken')
  );

  const login = (newToken: string) => {
    localStorage.setItem('authToken', newToken);
    setToken(newToken);
  };

  const logout = () => {
    localStorage.removeItem('authToken');
    setToken(null);
  };

  return (
    <AuthContext.Provider value={{ 
      token, 
      login, 
      logout, 
      isAuthenticated: !!token 
    }}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};
```

#### 6.4.2 Apollo Client設定

```typescript
// apolloClient.ts
import { ApolloClient, InMemoryCache, createHttpLink } from '@apollo/client';
import { setContext } from '@apollo/client/link/context';

// GraphQL APIエンドポイント設定
const httpLink = createHttpLink({
  uri: 'http://your-graphql-endpoint/query',
});

// 認証ヘッダー追加
const authLink = setContext((_, { headers }) => {
  const token = localStorage.getItem('authToken');
  return {
    headers: {
      ...headers,
      authorization: token ? `Bearer ${token}` : "",
    }
  };
});

// Apolloクライアント生成
export const client = new ApolloClient({
  link: authLink.concat(httpLink),
  cache: new InMemoryCache()
});
```

#### 6.4.3 GraphQL Code Generator設定

TypeScript型定義を自動生成するための設定：

```yaml
# codegen.yml
schema: http://your-graphql-endpoint/query
documents: './src/**/*.graphql'
generates:
  src/generated/graphql.tsx:
    plugins:
      - typescript
      - typescript-operations
      - typescript-react-apollo
    config:
      withHooks: true
```

#### 6.4.4 エラーハンドリングの実装

```typescript
// GraphQLエラー処理の共通コンポーネント
const ErrorDisplay: React.FC<{ error?: ApolloError }> = ({ error }) => {
  if (!error) return null;
  
  // GraphQLエラーの解析
  const validationErrors = error.graphQLErrors?.filter(
    err => err.extensions?.code === 'BAD_USER_INPUT'
  );
  
  if (validationErrors?.length > 0) {
    // バリデーションエラーの表示
    const fieldErrors = validationErrors[0].extensions?.field_errors || {};
    return (
      <div className="error-container">
        <h4>入力エラー</h4>
        <ul>
          {Object.entries(fieldErrors).map(([field, message]) => (
            <li key={field}>{message}</li>
          ))}
        </ul>
      </div>
    );
  }
  
  // 認証エラーの表示
  if (error.graphQLErrors?.some(err => 
    err.extensions?.code === 'UNAUTHENTICATED' || 
    err.extensions?.code === 'FORBIDDEN'
  )) {
    return (
      <div className="error-container">
        <h4>アクセス権限エラー</h4>
        <p>この操作を実行する権限がありません。</p>
      </div>
    );
  }
  
  // その他のエラー
  return (
    <div className="error-container">
      <h4>エラーが発生しました</h4>
      <p>{error.message}</p>
    </div>
  );
};
```

## 7. 関連ファイルとリソース

### 7.1 主要ファイル一覧

#### GraphQLスキーマファイル
- `internal/interface/graphql/schema/schema.graphqls` - 基本スキーマ
- `internal/interface/graphql/schema/auth.graphqls` - 認証関連スキーマ
- `internal/interface/graphql/schema/user_management.graphqls` - ユーザー管理スキーマ
- `internal/interface/graphql/schema/role_test.graphqls` - ロールテスト用スキーマ

#### 認証関連ファイル
- `internal/interface/graphql/auth/middleware.go` - 認証ミドルウェア
- `internal/interface/graphql/auth/context.go` - 認証コンテキスト管理
- `internal/interface/graphql/directives/auth.go` - 認証ディレクティブ
- `internal/infrastructure/auth/jwt_service.go` - JWTサービス
- `internal/infrastructure/auth/password_service.go` - パスワードサービス

#### ユーザー管理関連ファイル
- `internal/domain/entity/user.go` - ユーザーエンティティ
- `internal/domain/repository/user_repository.go` - リポジトリインターフェース
- `internal/infrastructure/persistence/postgres/user_repository.go` - リポジトリ実装
- `internal/usecase/port/user_management_usecase.go` - ユースケースインターフェース
- `internal/usecase/interactor/user_management_interactor.go` - ユースケース実装

#### リゾルバーファイル
- `internal/interface/graphql/resolver/resolver.go` - リゾルバーインターフェース
- `internal/interface/graphql/resolver/user_management.resolvers.go` - ユーザー管理リゾルバー
- `internal/interface/graphql/resolver/auth.resolvers.go` - 認証リゾルバー

#### エラーハンドリング関連ファイル
- `support/customerrors/base_error.go` - 基本エラー型
- `support/customerrors/validation_error.go` - バリデーションエラー
- `support/customerrors/unauthorized_error.go` - 認証エラー
- `support/customerrors/not_found_error.go` - リソース未検出エラー
- `support/customerrors/conflict_error.go` - 競合エラー

#### デプロイ関連ファイル
- `makefiles/terraform.mk` - Terraformコマンド
- `makefiles/update-image-only.mk` - 高速デプロイコマンド
- `scripts/terraform/terraform-deploy.sh` - デプロイスクリプト
- `scripts/terraform/aws-migrate-ci.sh` - マイグレーションスクリプト
- `scripts/terraform/aws-seed-user.sh` - テストユーザー投入スクリプト
- `.gitlab-ci.yml` - GitLab CI/CD設定

### 7.2 参考リソース

- [gqlgen公式ドキュメント](https://gqlgen.com/)
- [JWT公式サイト](https://jwt.io/)
- [GraphQL公式ドキュメント](https://graphql.org/learn/)
- [Apollo Client公式ドキュメント](https://www.apollographql.com/docs/react/)
- [React認証パターン](https://reactrouter.com/en/main/guides/auth)
- [OWASP認証セキュリティチートシート](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html)

## 8. まとめ

本ガイドでは、GraphQL認証とユーザー管理機能の設計、実装、テスト結果について説明しました。JWTベースの認証機構とロールベースのアクセス制御により、セキュアなAPIアクセスを実現しています。

実装された主な機能は以下の通りです：

- **認証機能**: JWTトークン認証、認証ディレクティブ
- **ユーザー管理機能**: ユーザー作成、パスワード変更、ユーザー削除
- **バリデーション**: 入力値の検証（ユーザー名3文字以上、パスワード6文字以上）
- **権限制御**: ロールベースのアクセス制御（Admin/Manager/Tester）

AWS環境にデプロイされたこれらの機能は、GraphQL Playgroundを使用したテストで正常に動作することが確認されています。

今後の拡張としては、ユーザー一覧機能の追加、最後のAdmin削除防止、パスワード強度検証の強化などが推奨されます。フロントエンド開発では、Apollo ClientとGraphQL Code Generatorを活用することで、型安全なAPI連携を実現できます。

この認証システムをベースに、フロントエンドとの連携を進めることで、完全なユーザー管理機能を持つWebアプリケーションを構築することができます。
