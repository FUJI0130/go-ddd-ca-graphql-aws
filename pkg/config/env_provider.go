package config

import (
	"os"
	"strconv"
	"strings"
)

// 設定キーから環境変数キーへのマッピング
var configToEnvMap = map[string]string{
	"database.host":     "DB_HOST",
	"database.port":     "DB_PORT",
	"database.user":     "DB_USER",
	"database.username": "DB_USERNAME", // 互換性のため両方サポート
	"database.password": "DB_PASSWORD",
	"database.dbname":   "DB_NAME",
	"database.sslmode":  "DB_SSLMODE",
}

// EnvConfigProvider は環境変数から設定を読み込むプロバイダー
type EnvConfigProvider struct {
	prefix string
}

// NewEnvConfigProvider は新しい環境変数設定プロバイダーを作成
func NewEnvConfigProvider(prefix string) *EnvConfigProvider {
	return &EnvConfigProvider{
		prefix: prefix,
	}
}

// Get は指定したキーの環境変数値を取得
func (p *EnvConfigProvider) Get(key string) (string, bool) {
	// 標準のキーフォーマットでチェック
	envKey := p.formatKey(key)
	val, exists := os.LookupEnv(envKey)
	if exists {
		return val, true
	}

	// マッピングされた環境変数キーでチェック
	if mappedKey, ok := configToEnvMap[key]; ok {
		val, exists := os.LookupEnv(mappedKey)
		if exists {
			return val, true
		}
	}

	return "", false
}

// GetString は指定したキーの文字列値を取得
func (p *EnvConfigProvider) GetString(key, defaultValue string) string {
	if val, exists := p.Get(key); exists {
		return val
	}
	return defaultValue
}

// GetInt は指定したキーの整数値を取得
func (p *EnvConfigProvider) GetInt(key string, defaultValue int) int {
	if val, exists := p.Get(key); exists {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultValue
}

// GetBool は指定したキーの真偽値を取得
func (p *EnvConfigProvider) GetBool(key string, defaultValue bool) bool {
	if val, exists := p.Get(key); exists {
		lowerVal := strings.ToLower(val)
		if lowerVal == "true" || lowerVal == "1" || lowerVal == "yes" || lowerVal == "y" {
			return true
		} else if lowerVal == "false" || lowerVal == "0" || lowerVal == "no" || lowerVal == "n" {
			return false
		}
	}
	return defaultValue
}

// Source は設定値のソース種別を取得
func (p *EnvConfigProvider) Source() string {
	return "environment"
}

// formatKey は環境変数キーを正規化（プレフィックス追加、ドットをアンダースコアに変換）
func (p *EnvConfigProvider) formatKey(key string) string {
	// プレフィックスが指定されている場合は追加
	formattedKey := key
	if p.prefix != "" {
		// キーがすでにプレフィックスで始まっていない場合のみ追加
		if !strings.HasPrefix(strings.ToUpper(key), strings.ToUpper(p.prefix)) {
			formattedKey = p.prefix + key
		}
	}

	// ドットをアンダースコアに変換
	formattedKey = strings.ReplaceAll(formattedKey, ".", "_")

	// 大文字に変換（環境変数の慣習）
	formattedKey = strings.ToUpper(formattedKey)

	return formattedKey
}
