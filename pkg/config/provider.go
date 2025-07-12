// pkg/config/provider.go
package config

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
