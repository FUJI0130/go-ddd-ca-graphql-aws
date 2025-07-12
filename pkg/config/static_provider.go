package config

import (
	"fmt"
	"strconv"
	"strings"
)

// StaticConfigProvider は静的なマップから設定を読み込むプロバイダー
type StaticConfigProvider struct {
	config map[string]interface{}
}

// NewStaticConfigProvider は新しい静的設定プロバイダーを作成
func NewStaticConfigProvider(config map[string]interface{}) *StaticConfigProvider {
	return &StaticConfigProvider{
		config: config,
	}
}

// Get は指定したキーの設定値を取得
func (p *StaticConfigProvider) Get(key string) (string, bool) {
	if value, exists := p.config[key]; exists {
		// 型アサーションで文字列に変換
		if strValue, ok := value.(string); ok {
			return strValue, true
		}
		// 数値やboolの場合は文字列に変換
		return fmt.Sprintf("%v", value), true
	}
	return "", false
}

// GetString は指定したキーの文字列値を取得
func (p *StaticConfigProvider) GetString(key, defaultValue string) string {
	if value, exists := p.config[key]; exists {
		if strValue, ok := value.(string); ok {
			return strValue
		}
		return fmt.Sprintf("%v", value)
	}
	return defaultValue
}

// GetInt は指定したキーの整数値を取得
func (p *StaticConfigProvider) GetInt(key string, defaultValue int) int {
	if value, exists := p.config[key]; exists {
		switch v := value.(type) {
		case int:
			return v
		case float64:
			return int(v)
		case string:
			if intVal, err := strconv.Atoi(v); err == nil {
				return intVal
			}
		}
	}
	return defaultValue
}

// GetBool は指定したキーの真偽値を取得
func (p *StaticConfigProvider) GetBool(key string, defaultValue bool) bool {
	if value, exists := p.config[key]; exists {
		switch v := value.(type) {
		case bool:
			return v
		case string:
			lowerStr := strings.ToLower(v)
			if lowerStr == "true" || lowerStr == "1" || lowerStr == "yes" || lowerStr == "y" {
				return true
			} else if lowerStr == "false" || lowerStr == "0" || lowerStr == "no" || lowerStr == "n" {
				return false
			}
		}
	}
	return defaultValue
}

// Source は設定値のソース種別を取得
func (p *StaticConfigProvider) Source() string {
	return "static"
}
