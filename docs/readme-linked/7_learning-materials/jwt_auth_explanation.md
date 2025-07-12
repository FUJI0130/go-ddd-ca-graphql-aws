# JWT認証システム解説資料
*プロジェクト内でのJWT認証の仕組みと役割*

## 🎯 この資料の目的

JWT認証がどのような仕組みで動いているか、そしてあなたのプロジェクトのどこで使われているかを分かりやすく解説します。

---

## 1. JWT（JSON Web Token）とは何か

### 1.1 JWTの基本構造

JWTは3つの部分から構成される文字列です：

```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMTIzIiwicm9sZSI6ImFkbWluIiwiZXhwIjoxNjg5ODQwMDAwfQ.signature_part
      ↑ Header             ↑ Payload                                    ↑ Signature
```

#### 🔹 Header（ヘッダー）
```json
{
  "alg": "HS256",    // 署名アルゴリズム
  "typ": "JWT"       // トークンタイプ
}
```

#### 🔹 Payload（ペイロード）- ユーザー情報
```json
{
  "user_id": "123",           // ユーザーID
  "role": "admin",            // ユーザーの権限
  "exp": 1689840000,          // 有効期限（Unixタイムスタンプ）
  "iat": 1689836400           // 発行時刻
}
```

#### 🔹 Signature（署名）- 改ざん防止
```
HMACSHA256(
  base64UrlEncode(header) + "." + base64UrlEncode(payload),
  secret_key  // サーバーだけが知っている秘密鍵
)
```

### 1.2 JWTがなぜ安全なのか

**🔐 改ざん検出の仕組み**：
1. **秘密鍵による署名**: サーバーだけが知っている秘密鍵で署名
2. **検証プロセス**: 受信時に署名を検証し、改ざんされていないか確認
3. **有効期限チェック**: トークンの有効期限を自動確認

```mermaid
flowchart LR
    A[ユーザー情報] --> B[JSON形式に変換]
    B --> C[秘密鍵で署名]
    C --> D[Base64エンコード]
    D --> E[JWTトークン完成]
    
    E --> F[クライアント送信]
    F --> G[サーバー受信]
    G --> H[署名検証]
    H --> I{改ざんされていない?}
    I -->|Yes| J[認証成功]
    I -->|No| K[認証失敗]
```

---

## 2. あなたのプロジェクトでのJWT認証の配置

### 2.1 システム全体での位置づけ

```mermaid
graph TB
    subgraph "フロントエンド（React）"
        LOGIN[LoginPage.tsx<br/>ログインフォーム]
        AUTH_CTX[AuthContext.tsx<br/>認証状態管理]
        COOKIE[HttpOnly Cookie<br/>トークン安全保存]
    end
    
    subgraph "バックエンド（Go）"
        MIDDLEWARE[認証ミドルウェア<br/>JWT検証]
        JWT_SERVICE[JWT Service<br/>トークン生成・検証]
        PWD_SERVICE[Password Service<br/>BCryptハッシュ化]
    end
    
    subgraph "GraphQL"
        DIRECTIVE["@auth ディレクティブ<br/>宣言的認証"]
        RESOLVER[リゾルバー<br/>認証済み処理]
    end
    
    subgraph "データベース"
        USER_TABLE[Userテーブル<br/>認証情報保存]
        REFRESH_TOKEN[RefreshTokenテーブル<br/>長期認証]
    end
    
    LOGIN --> AUTH_CTX
    AUTH_CTX --> COOKIE
    COOKIE -.->|HTTP Request| MIDDLEWARE
    MIDDLEWARE --> JWT_SERVICE
    JWT_SERVICE --> DIRECTIVE
    DIRECTIVE --> RESOLVER
    RESOLVER --> USER_TABLE
    JWT_SERVICE --> PWD_SERVICE
    PWD_SERVICE --> USER_TABLE
    JWT_SERVICE --> REFRESH_TOKEN
```

### 2.2 各コンポーネントの役割

#### 🎨 **フロントエンド部分**
- **LoginPage.tsx**: ユーザーがIDとパスワードを入力する画面
- **AuthContext.tsx**: ログイン状態を管理（ログイン中か、ログアウト中か）
- **HttpOnly Cookie**: JWTトークンを安全に保存する場所

#### ⚙️ **バックエンド部分**
- **認証ミドルウェア**: すべてのリクエストでJWTトークンを確認
- **JWT Service**: トークンの作成と検証を担当
- **Password Service**: パスワードを安全にハッシュ化

#### 🔗 **GraphQL部分**
- **@auth ディレクティブ**: 「この機能は認証が必要」と宣言
- **リゾルバー**: 実際のビジネスロジック処理

---

## 3. JWT認証の実際の流れ

### 3.1 ログイン処理の詳細

```mermaid
sequenceDiagram
    participant U as ユーザー
    participant F as フロントエンド
    participant M as 認証ミドルウェア
    participant J as JWT Service
    participant P as Password Service
    participant D as データベース
    
    U->>F: ID・パスワード入力
    F->>M: ログインリクエスト
    M->>P: パスワード検証
    P->>D: ユーザー情報取得
    D-->>P: ハッシュ化パスワード
    P->>P: BCrypt検証
    P-->>M: 検証結果
    
    alt パスワード正しい
        M->>J: JWTトークン生成依頼
        J->>J: ユーザー情報をJWTに変換
        J-->>M: JWTトークン
        M-->>F: HttpOnly Cookieでトークン送信
        F->>F: 認証状態をログイン中に変更
        F-->>U: ダッシュボード画面に移動
    else パスワード間違い
        M-->>F: エラーメッセージ
        F-->>U: 「IDまたはパスワードが間違っています」
    end
```

### 3.2 認証が必要な機能にアクセスする流れ

```mermaid
sequenceDiagram
    participant U as ユーザー
    participant F as フロントエンド
    participant M as 認証ミドルウェア
    participant J as JWT Service
    participant G as GraphQL
    participant R as リゾルバー
    
    U->>F: テストスイート作成ボタンクリック
    F->>M: GraphQLリクエスト（Cookie付き）
    M->>J: トークン検証
    J->>J: 署名確認・有効期限確認
    
    alt トークン有効
        J-->>M: ユーザー情報
        M->>G: 認証済みリクエスト
        G->>G: @authディレクティブ確認
        G->>R: ビジネスロジック実行
        R-->>F: 作成成功レスポンス
        F-->>U: 「テストスイートが作成されました」
    else トークン無効
        J-->>M: 認証エラー
        M-->>F: 401 Unauthorized
        F->>F: ログアウト状態に変更
        F-->>U: ログイン画面に自動移動
    end
```

---

## 4. JWTの有効期限とセキュリティ

### 4.1 トークンの寿命管理

```mermaid
timeline
    title JWTトークンのライフサイクル
    
    section ログイン時
        トークン生成    : 有効期限15分に設定
        リフレッシュトークン生成 : 有効期限7日に設定
    
    section 通常利用
        API呼び出し     : 毎回JWTトークンを送信
        自動検証       : 有効期限・署名を確認
    
    section 期限切れ前
        自動更新       : リフレッシュトークンで新しいJWTを取得
        継続利用       : ユーザーは何も気づかず利用継続
    
    section 完全期限切れ
        認証失敗       : 7日後にリフレッシュトークンも期限切れ
        再ログイン     : ユーザーに再度ログインを要求
```

### 4.2 セキュリティ対策の重複防御

**🛡️ 第1の防御: HttpOnly Cookie**
- JavaScriptからアクセス不可
- XSS攻撃でもトークンを盗めない

**🛡️ 第2の防御: 署名検証**
- 改ざんされたトークンを即座に検出
- 偽造トークンでのアクセスを防止

**🛡️ 第3の防御: 有効期限**
- 短い有効期限（15分）でリスク最小化
- 盗まれても短期間で無効化

**🛡️ 第4の防御: SameSite設定**
- CSRF攻撃を防止
- 他のサイトからのリクエストでCookieを送信しない

---

## 5. プロジェクト内での具体的な実装コード

### 5.1 フロントエンド: 認証状態管理

```typescript
// AuthContext.tsx - 認証状態を管理
interface AuthContextType {
  isAuthenticated: boolean;      // ログインしているか
  isLoading: boolean;           // 認証状態確認中か
  user: AuthUser | null;        // ログインユーザー情報
  login: (credentials) => Promise<void>;  // ログイン実行
  logout: () => Promise<void>;             // ログアウト実行
}

// LoginPage.tsx - ログイン処理
const handleSubmit = async (credentials) => {
  try {
    await login(credentials);  // JWTトークンを取得・Cookie保存
    navigate('/dashboard');    // ダッシュボードに移動
  } catch (error) {
    setError('ログインに失敗しました');
  }
};
```

### 5.2 バックエンド: JWT生成・検証

```go
// JWT Service - トークンの生成
func (s *JWTService) GenerateToken(user *User) (*TokenResult, error) {
    claims := JWTClaims{
        UserID: user.ID,
        Role:   string(user.Role),
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := token.SignedString(s.secretKey)
    return &TokenResult{Token: tokenString, ExpiresAt: expiresAt}, err
}

// 認証ミドルウェア - トークンの検証
func (m *AuthMiddleware) ValidateRequest(r *http.Request) (*User, error) {
    cookie, err := r.Cookie("auth_token")
    if err != nil {
        return nil, errors.New("認証トークンがありません")
    }
    
    user, err := m.jwtService.ValidateToken(cookie.Value)
    if err != nil {
        return nil, errors.New("無効な認証トークンです")
    }
    
    return user, nil
}
```

### 5.3 GraphQL: 宣言的認証

```graphql
# GraphQLスキーマ - 認証が必要な操作を宣言
directive @auth on FIELD_DEFINITION
directive @hasRole(role: String!) on FIELD_DEFINITION

type Mutation {
  # 認証必須: ログインしていないとアクセスできない
  createTestSuite(input: CreateTestSuiteInput!): TestSuite! @auth
  
  # 管理者限定: Adminロールのみアクセス可能
  deleteUser(userId: ID!): Boolean! @hasRole(role: "Admin")
  
  # 権限制限: AdminまたはManagerのみアクセス可能
  updateTestSuite(id: ID!, input: UpdateTestSuiteInput!): TestSuite! @hasRole(role: "Manager")
}
```

---

## 6. 実際の使用場面での理解確認

### 6.1 あなたのプロジェクトでのJWT認証確認

**🌐 実際に確認できる動作**:
1. **ログイン**: https://example-frontend.cloudfront.net/ にアクセス
2. **認証情報**: demo_user / password でログイン
3. **認証後の画面**: ダッシュボードでユーザー情報表示
4. **認証が必要な機能**: テストスイート作成・管理

### 6.2 開発者ツールでの確認方法

**ブラウザでJWTトークンを確認する手順**:

1. **ログイン後にF12でDevToolsを開く**
2. **Application → Cookies → あなたのサイトURL**
3. **`auth_token`という名前のCookieを確認**
4. **値をコピーして jwt.io にペースト**
5. **デコードされたユーザー情報を確認**

```json
// デコード結果例
{
  "user_id": "demo_user",
  "role": "admin", 
  "exp": 1689840000,  // 有効期限
  "iat": 1689836400   // 発行時刻
}
```

---

## 7. よくある質問と回答

### Q1. なぜJWTではなくセッションCookieを使わないのか？

**A**: JWTの方が以下の利点があります：
- **ステートレス**: サーバーでセッション情報を保存不要
- **スケーラブル**: 複数サーバーでセッション共有が不要
- **マイクロサービス対応**: 他のサービスでも同じトークンで認証可能

### Q2. 15分という有効期限は短すぎないか？

**A**: セキュリティのための設計です：
- **盗難リスク軽減**: 盗まれても15分で無効化
- **リフレッシュトークン**: 自動更新でユーザーは気づかない
- **バランス**: セキュリティと利便性のバランス

### Q3. HttpOnly Cookieとは何が違うのか？

**A**: 通常のCookieとの違い：
- **通常のCookie**: JavaScriptで `document.cookie` でアクセス可能
- **HttpOnly Cookie**: JavaScriptからアクセス不可、HTTP通信のみ
- **XSS対策**: 悪意あるJavaScriptでもトークンを盗めない

### Q4. リフレッシュトークンはどう動くのか？

**A**: 以下の仕組みで自動更新：
1. **JWT期限切れ**: 15分後にアクセストークンが無効
2. **リフレッシュ実行**: 7日有効のリフレッシュトークンで新しいJWTを取得
3. **透明な更新**: ユーザーは何も気づかずに利用継続
4. **完全期限切れ**: 7日後に再ログインが必要

---

## 8. まとめ: JWTがあなたのプロジェクトに与える価値

### 8.1 技術的価値
- ✅ **現代的認証**: 業界標準の認証方式を実装
- ✅ **セキュリティ**: 多層防御による堅牢なセキュリティ
- ✅ **スケーラビリティ**: サーバー負荷軽減・マイクロサービス対応

### 8.2 実用的価値
- ✅ **ユーザビリティ**: シームレスなログイン体験
- ✅ **保守性**: 宣言的認証による分かりやすいコード
- ✅ **拡張性**: 新機能追加時の認証制御が容易

### 8.3 学習価値
- ✅ **JWT理解**: 現代的認証技術の実践的習得
- ✅ **セキュリティ**: Webセキュリティの包括的理解
- ✅ **フルスタック**: フロントエンド・バックエンド統合の実践

---

**💡 重要なポイント**: JWTは単なる認証技術ではなく、現代的Webアプリケーションの基盤技術です。あなたのプロジェクトではセキュリティ・利便性・拡張性すべてを満たす形で実装されています。