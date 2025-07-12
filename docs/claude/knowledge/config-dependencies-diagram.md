# 設定管理システムのファイル依存関係図

## 1. ファイル構造と依存関係

### 1.1 全体的な依存関係図

```
pkg/config/
├── provider.go (インターフェース定義)
│   └── ConfigProvider インターフェース
│       ├── 実装 → env_provider.go (EnvConfigProvider)
│       ├── 実装 → file_provider.go (FileConfigProvider, EmptyConfigProvider)
│       ├── 実装 → static_provider.go (StaticConfigProvider) - 新規作成
│       └── 実装 → chain_provider.go (ChainedConfigProvider)
│
├── config.go (メインロジック)
│   ├── 依存 → provider.go (ConfigProvider インターフェース)
│   ├── 依存 → env_provider.go (NewEnvConfigProvider 関数)
│   ├── 依存 → file_provider.go (NewFileConfigProvider 関数)
│   ├── 依存 → static_provider.go (NewStaticConfigProvider 関数) - 新規作成
│   ├── 依存 → chain_provider.go (NewChainedConfigProvider 関数)
│   ├── 依存 → database.go (DatabaseConfig 構造体、NewDatabaseConfigFromProvider 関数)
│   └── 公開 → LoadConfig 関数 (cmd/*/main.go で使用)
│
└── database.go (データベース設定)
    ├── 依存 → provider.go (ConfigProvider インターフェース)
    └── 公開 → NewDatabaseConnection 関数 (internal/infrastructure/persistence/postgres/db.go で使用)
```

### 1.2 インポート関係

```go
// provider.go
package config
// 外部インポートなし - インターフェース定義のみ

// env_provider.go
package config
import (
    "os"
    "strconv"
    "strings"
)

// file_provider.go
package config
import (
    "fmt"
    "log"
    "github.com/spf13/viper"
)

// chain_provider.go
package config
import (
    "strings"
)

// static_provider.go (新規作成)
package config
import (
    "fmt"
    "strconv"
    "strings"
)

// config.go
package config
import (
    "log"
    "os"
    "regexp"
    "strings"
    "time"
)

// database.go
package config
import (
    "database/sql"
    "fmt"
    "log"
    "time"
    "github.com/FUJI0130/go-ddd-ca/internal/infrastructure/persistence/postgres"
)
```

## 2. 詳細な依存関係

### 2.1 インターフェース実装の関係

**ConfigProviderインターフェース (provider.go)**:

```go
// ConfigProvider は設定値を提供するインターフェース
type ConfigProvider interface {
    // Get は指定したキーの設定値を取得
    Get(key string) (string, bool)
    
    // GetString は指定したキーの文字列値を取得（デフォルト値指定可能）
    GetString(key, defaultValue string) string
    
    // GetInt は指定したキーの整数値を取得（デフォルト値指定可能）
    GetInt(key string, defaultValue int) int
    
    // GetBool は指定したキーの真偽値を取得（デフォルト値指定可能）
    GetBool(key string, defaultValue bool) bool
    
    // Source は設定値のソース種別を取得
    Source() string
}
```

**実装クラス**:
1. `EnvConfigProvider` (env_provider.go)
2. `FileConfigProvider` (file_provider.go)
3. `EmptyConfigProvider` (file_provider.go)
4. `ChainedConfigProvider` (chain_provider.go)
5. `StaticConfigProvider` (static_provider.go) - 新規作成

### 2.2 主要な関数と呼び出し関係

**LoadConfig関数 (config.go)**:
- 呼び出し元: 各サービスのmain.go
- 呼び出す関数:
  - `NewEnvConfigProvider` (env_provider.go)
  - `NewFileConfigProvider` (file_provider.go)
  - `NewStaticConfigProvider` (static_provider.go)
  - `NewChainedConfigProvider` (chain_provider.go)
  - `NewDatabaseConfigFromProvider` (database.go)

**NewDatabaseConnection関数 (database.go)**:
- 呼び出し元: 設定値を使用するコード (internal/infrastructure/persistence/postgres/db.go)
- 内部で使用する機能:
  - postgres.NewDB (internal/infrastructure/persistence/postgres/db.go)

## 3. 修正のポイントと影響範囲

### 3.1 config.goの修正ポイント

1. **クラウド環境検出ロジック**:
   ```go
   isCloudEnv := os.Getenv("APP_ENVIRONMENT") == "production" ||
       os.Getenv("IS_CLOUD_ENV") == "true" ||
       os.Getenv("ECS_CONTAINER_METADATA_URI") != "" ||
       os.Getenv("KUBERNETES_SERVICE_HOST") != ""
   ```
   - 影響: 設定ファイルの読み込み有無に影響

2. **プロバイダー優先順位の明確化**:
   ```go
   // 環境変数プロバイダーを最優先で追加
   providers = append(providers, envProvider)
   
   // クラウド環境では設定ファイルを使用しない
   if !isCloudEnv {
       // 設定ファイルプロバイダーを追加（開発環境のみ）
   }
   
   // デフォルト値プロバイダーを最後に追加
   providers = append(providers, defaultProvider)
   ```
   - 影響: 設定値の取得優先順位に影響

### 3.2 static_provider.goの作成ポイント (新規)

1. **ConfigProviderインターフェースの実装**:
   ```go
   type StaticConfigProvider struct {
       config map[string]interface{}
   }
   
   func (p *StaticConfigProvider) Get(key string) (string, bool) {
       // 実装...
   }
   
   // 他のメソッド実装...
   ```
   - 影響: config.goでのStaticConfigProvider使用方法に影響

### 3.3 database.goの修正ポイント

1. **設定ソースの透明性向上**:
   ```go
   log.Printf("Database connection settings from %s: host=%s, port=%d, user=%s, dbname=%s, sslmode=%s",
       provider.Source(), config.Host, config.Port, config.User, config.DBName, config.SSLMode)
   ```
   - 影響: ログ出力のみ、機能的な影響なし

## 4. テスト時の依存関係

### 4.1 モック作成ポイント

1. **ConfigProviderインターフェースのモック**:
   - テスト時にStaticConfigProviderを使用して特定の設定値を提供
   
2. **環境変数のモック**:
   - os.Setenvとos.Unsetenvを使用して環境変数をモック
   
3. **チェーンプロバイダーのテスト**:
   - 複数のプロバイダーを連鎖させて優先順位を検証

### 4.2 テストファイルの依存関係

```
pkg/config/
├── config_test.go
│   ├── 依存 → provider.go (ConfigProvider インターフェース)
│   ├── 依存 → static_provider.go (StaticConfigProvider 実装)
│   ├── 依存 → env_provider.go (EnvConfigProvider 実装)
│   └── テスト対象 → config.go (LoadConfig 関数)
│
└── database_test.go
    ├── 依存 → provider.go (ConfigProvider インターフェース)
    ├── 依存 → static_provider.go (StaticConfigProvider 実装)
    ├── 依存 → env_provider.go (EnvConfigProvider 実装)
    └── テスト対象 → database.go (NewDatabaseConfigFromProvider 関数)
```

## 5. 修正時の注意点

1. **StaticConfigProviderの実装**:
   - ConfigProviderインターフェースの全メソッドを正確に実装
   - 型変換処理を適切に実装（文字列、整数、真偽値）
   
2. **config.goでの使用方法**:
   - 正しい関数名と引数でStaticConfigProviderを生成
   - チェーンプロバイダーへの追加順序を維持
   
3. **テストコードの更新**:
   - 環境変数をテスト前後で適切にクリア
   - クラウド環境検出のテストを追加
   - チェーンプロバイダーのテストを追加

## 6. Dockerfileとの関連

Dockerfileでの設定ファイル関連部分の修正は、ランタイム環境で以下の設定管理コードの動作に影響します：

```go
// config.goの条件分岐
if !isCloudEnv {
    // 設定ファイルが存在する場合、それを読み込む
    var fileProvider ConfigProvider
    fileConfigPaths := []string{configPath, "./configs", "."}
    for _, path := range fileConfigPaths {
        fp, err := NewFileConfigProvider(path, env, "yml")
        if err == nil {
            log.Printf("設定ファイル %s/%s.yml を読み込みました", path, env)
            fileProvider = fp
            break
        }
    }
    
    // ファイルが見つかった場合のみプロバイダーとして追加
    if fileProvider != nil {
        providers = append(providers, fileProvider)
    }
}
```

Dockerfileから設定ファイルのコピーを削除し、明示的に設定ファイルディレクトリを削除することで、コンテナ内に設定ファイルが存在しなくなり、環境変数からの設定読み込みが確実に優先されるようになります。