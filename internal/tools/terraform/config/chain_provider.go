// chain_provider.go
package config

import (
	"fmt"
)

// DefaultChainConfigProvider は複数のConfigProviderを優先順に検索するプロバイダー
type DefaultChainConfigProvider struct {
	providers []ConfigProvider
}

// NewChainConfigProvider は新しいDefaultChainConfigProviderインスタンスを作成する
func NewChainConfigProvider() *DefaultChainConfigProvider {
	return &DefaultChainConfigProvider{
		providers: make([]ConfigProvider, 0),
	}
}

// AddProvider は新しいプロバイダーを追加する
// 複数のプロバイダーが同じキーを提供する場合、先に追加されたプロバイダーが優先される
func (p *DefaultChainConfigProvider) AddProvider(provider ConfigProvider) {
	p.providers = append(p.providers, provider)
}

// Get は指定されたキーの値を取得する。キーが存在しない場合は空文字列とfalseを返す
func (p *DefaultChainConfigProvider) Get(key string) (string, bool) {
	for _, provider := range p.providers {
		if value, exists := provider.Get(key); exists {
			return value, true
		}
	}
	return "", false
}

// GetWithDefault は指定されたキーの値を取得する。キーが存在しない場合はデフォルト値を返す
func (p *DefaultChainConfigProvider) GetWithDefault(key, defaultValue string) string {
	value, exists := p.Get(key)
	if !exists {
		return defaultValue
	}
	return value
}

// GetRequired は指定されたキーの値を取得する。キーが存在しない場合はエラーを返す
func (p *DefaultChainConfigProvider) GetRequired(key string) (string, error) {
	value, exists := p.Get(key)
	if !exists {
		return "", fmt.Errorf("必須設定 %s が設定されていません", key)
	}
	return value, nil
}
