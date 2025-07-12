// env_provider.go
package config

import (
	"fmt"
	"os"
)

// EnvConfigProvider は環境変数から設定を提供するプロバイダー
type EnvConfigProvider struct{}

// NewEnvConfigProvider は新しいEnvConfigProviderインスタンスを作成する
func NewEnvConfigProvider() *EnvConfigProvider {
	return &EnvConfigProvider{}
}

// Get は指定されたキーの環境変数値を取得する。キーが存在しない場合は空文字列とfalseを返す
func (p *EnvConfigProvider) Get(key string) (string, bool) {
	value, exists := os.LookupEnv(key)
	return value, exists
}

// GetWithDefault は指定されたキーの環境変数値を取得する。キーが存在しない場合はデフォルト値を返す
func (p *EnvConfigProvider) GetWithDefault(key, defaultValue string) string {
	value, exists := p.Get(key)
	if !exists {
		return defaultValue
	}
	return value
}

// GetRequired は指定されたキーの環境変数値を取得する。キーが存在しない場合はエラーを返す
func (p *EnvConfigProvider) GetRequired(key string) (string, error) {
	value, exists := p.Get(key)
	if !exists {
		return "", fmt.Errorf("必須環境変数 %s が設定されていません", key)
	}
	return value, nil
}
