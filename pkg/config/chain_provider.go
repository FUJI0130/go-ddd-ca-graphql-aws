// pkg/config/chain_provider.go
package config

import (
	"strings"
)

// ChainedConfigProvider は複数のプロバイダーをチェーンするプロバイダー
type ChainedConfigProvider struct {
	providers []ConfigProvider
	sources   []string
}

// NewChainedConfigProvider は指定した優先順位のプロバイダーをチェーンした設定プロバイダーを作成
func NewChainedConfigProvider(providers ...ConfigProvider) *ChainedConfigProvider {
	sources := make([]string, len(providers))
	for i, provider := range providers {
		sources[i] = provider.Source()
	}

	return &ChainedConfigProvider{
		providers: providers,
		sources:   sources,
	}
}

// Get は最初に値が見つかったプロバイダーから設定値を取得
func (p *ChainedConfigProvider) Get(key string) (string, bool) {
	for _, provider := range p.providers {
		if val, exists := provider.Get(key); exists {
			return val, true
		}
	}
	return "", false
}

// GetString は最初に値が見つかったプロバイダーから文字列値を取得
func (p *ChainedConfigProvider) GetString(key, defaultValue string) string {
	for _, provider := range p.providers {
		if val, exists := provider.Get(key); exists {
			return val
		}
	}
	return defaultValue
}

// GetInt は最初に値が見つかったプロバイダーから整数値を取得
func (p *ChainedConfigProvider) GetInt(key string, defaultValue int) int {
	for _, provider := range p.providers {
		if intVal := provider.GetInt(key, defaultValue); intVal != defaultValue {
			return intVal
		}
	}
	return defaultValue
}

// GetBool は最初に値が見つかったプロバイダーから真偽値を取得
// GetBool は最初に値が見つかったプロバイダーから真偽値を取得
func (p *ChainedConfigProvider) GetBool(key string, defaultValue bool) bool {
	for _, provider := range p.providers {
		// GetのBool値を確認
		_, exists := provider.Get(key)
		if exists {
			return provider.GetBool(key, defaultValue)
		}
	}
	return defaultValue
}

// Source は全てのプロバイダーのソースを結合して返す
func (p *ChainedConfigProvider) Source() string {
	return "chain(" + strings.Join(p.sources, " > ") + ")"
}
