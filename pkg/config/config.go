package config

import (
	"log"
	"os"
	"regexp"
	"time"
)

// Config はアプリケーション全体の設定を保持します
type Config struct {
	Environment string
	Server      ServerConfig
	Database    DatabaseConfig
	Auth        AuthConfig
}

// ServerConfig はWebサーバーの設定を保持します
type ServerConfig struct {
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// AuthConfig は認証関連の設定を保持します
type AuthConfig struct {
	JWTSecret     string
	TokenDuration time.Duration
}

// LoadConfig は設定ファイルと環境変数から設定を読み込みます
func LoadConfig(configPath string) (*Config, error) {
	// 環境変数プロバイダーを最優先に設定
	envProvider := NewEnvConfigProvider("")

	// 環境別の設定ファイル名
	env := os.Getenv("APP_ENVIRONMENT")
	if env == "" {
		env = "development"
	}

	// クラウド環境を検出
	isCloudEnv := os.Getenv("APP_ENVIRONMENT") == "production" ||
		os.Getenv("IS_CLOUD_ENV") == "true" ||
		os.Getenv("ECS_CONTAINER_METADATA_URI") != "" ||
		os.Getenv("KUBERNETES_SERVICE_HOST") != ""

	// プロバイダーリストを初期化
	var providers []ConfigProvider

	// 環境変数プロバイダーを最優先で追加
	providers = append(providers, envProvider)

	// クラウド環境では設定ファイルを使用しない
	if !isCloudEnv {
		// 設定ファイルプロバイダーの作成
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

		// 設定ファイルプロバイダーが作成できた場合のみリストに追加
		if fileProvider != nil {
			providers = append(providers, fileProvider)
		} else {
			log.Printf("警告: %s.yml 設定ファイルが見つかりません。環境変数またはデフォルト値を使用します", env)
		}
	} else {
		log.Printf("クラウド環境で実行中のため、設定ファイルは使用しません")
	}

	// デフォルト値プロバイダーの作成
	defaultConfig := map[string]interface{}{
		"server.port":         8080,
		"server.readTimeout":  "15s",
		"server.writeTimeout": "15s",
		"database.driver":     "postgres",
		"database.user":       "testuser",
		"database.password":   "testpass",
		"database.host":       "localhost",
		"database.port":       5432,
		"database.dbname":     "test_management",
		"database.sslmode":    "disable",
		"auth.tokenDuration":  "24h",
	}

	// StaticConfigProviderの作成
	defaultProvider := NewStaticConfigProvider(defaultConfig)

	// デフォルト値プロバイダーを最後に追加
	providers = append(providers, defaultProvider)

	// チェーン設定プロバイダー（優先順位: 環境変数 > 設定ファイル > デフォルト値）
	chainedProvider := NewChainedConfigProvider(providers...)

	// 設定値の読み込み
	config := &Config{
		Environment: env,
		Server: ServerConfig{
			Port:         chainedProvider.GetInt("server.port", 8080),
			ReadTimeout:  parseDuration(chainedProvider.GetString("server.readTimeout", "15s")),
			WriteTimeout: parseDuration(chainedProvider.GetString("server.writeTimeout", "15s")),
		},
		Database: *NewDatabaseConfigFromProvider(chainedProvider),
		Auth: AuthConfig{
			JWTSecret:     chainedProvider.GetString("auth.jwtSecret", ""),
			TokenDuration: parseDuration(chainedProvider.GetString("auth.tokenDuration", "24h")),
		},
	}

	// デバッグモードでログ出力
	if os.Getenv("DEBUG") == "true" || os.Getenv("APP_DEBUG") == "true" {
		log.Println("==================== 設定情報 ====================")
		log.Printf("Environment: %s", config.Environment)
		log.Printf("Server.Port: %d (source: %s)", config.Server.Port, getSettingSource(chainedProvider, "server.port"))
		log.Printf("Server.ReadTimeout: %s (source: %s)", config.Server.ReadTimeout, getSettingSource(chainedProvider, "server.readTimeout"))
		log.Printf("Server.WriteTimeout: %s (source: %s)", config.Server.WriteTimeout, getSettingSource(chainedProvider, "server.writeTimeout"))
		log.Printf("Database.Driver: %s (source: %s)", config.Database.Driver, getSettingSource(chainedProvider, "database.driver"))
		log.Printf("Database.Host: %s (source: %s)", config.Database.Host, getSettingSource(chainedProvider, "database.host", "DB_HOST"))
		log.Printf("Database.Port: %d (source: %s)", config.Database.Port, getSettingSource(chainedProvider, "database.port", "DB_PORT"))
		log.Printf("Database.User: %s (source: %s)", config.Database.User, getSettingSource(chainedProvider, "database.user", "DB_USER", "DB_USERNAME"))
		log.Printf("Database.DBName: %s (source: %s)", config.Database.DBName, getSettingSource(chainedProvider, "database.dbname", "DB_NAME"))
		log.Printf("Database.SSLMode: %s (source: %s)", config.Database.SSLMode, getSettingSource(chainedProvider, "database.sslmode", "DB_SSLMODE"))
		log.Printf("Auth.JWTSecret: %s (source: %s)", "***" /* セキュリティのため表示しない */, getSettingSource(chainedProvider, "auth.jwtSecret"))
		log.Printf("Auth.TokenDuration: %s (source: %s)", config.Auth.TokenDuration, getSettingSource(chainedProvider, "auth.tokenDuration"))
		log.Println("====================================================")
	}

	return config, nil
}

// getSettingSource は指定した設定がどのプロバイダーから取得されたかを返します
func getSettingSource(provider *ChainedConfigProvider, keys ...string) string {
	for _, p := range provider.providers {
		for _, key := range keys {
			if _, exists := p.Get(key); exists {
				return p.Source()
			}
		}
	}

	// どのプロバイダーでも見つからない場合はデフォルト値を使用
	return "default"
}

// parseDuration は文字列からDurationを解析します（エラー時はデフォルト値）
func parseDuration(durationStr string) time.Duration {
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		log.Printf("警告: 無効な期間形式 '%s': %v, デフォルト値の15秒を使用します", durationStr, err)
		return 15 * time.Second
	}
	return duration
}

// 以下はViperの環境変数展開のための既存のヘルパー関数

// 環境変数が未設定の場合はデフォルト値を試行するロジックを追加
func expandEnvVars(s string) string {
	// ${VAR:-default} の形式を捉える正規表現
	re := regexp.MustCompile(`\${([^:}]+)(?::-([^}]*))?}`)
	return re.ReplaceAllStringFunc(s, func(match string) string {
		parts := re.FindStringSubmatch(match)
		if len(parts) < 2 {
			return match
		}

		envName := parts[1]
		defaultValue := ""
		if len(parts) > 2 {
			defaultValue = parts[2]
		}

		envValue := os.Getenv(envName)
		if envValue == "" {
			if defaultValue != "" {
				return defaultValue
			}
			return match // 環境変数もデフォルト値も未設定の場合は元の文字列を返す
		}
		return envValue
	})
}
