// provider.go
package config

// ConfigProvider は設定値を提供するインターフェース
type ConfigProvider interface {
	// Get は指定されたキーの値を取得する。キーが存在しない場合は空文字列とfalseを返す
	Get(key string) (string, bool)

	// GetWithDefault は指定されたキーの値を取得する。キーが存在しない場合はデフォルト値を返す
	GetWithDefault(key, defaultValue string) string

	// GetRequired は指定されたキーの値を取得する。キーが存在しない場合はエラーを返す
	GetRequired(key string) (string, error)
}

// FileConfigProvider は設定ファイルから値を提供するインターフェース
type FileConfigProvider interface {
	ConfigProvider

	// Load は指定されたパスの設定ファイルを読み込む
	Load(filePath string) error
}

// ChainConfigProvider は複数のConfigProviderを優先順に検索するプロバイダー
type ChainConfigProvider interface {
	ConfigProvider

	// AddProvider は新しいプロバイダーを追加する（先に追加されたものが優先される）
	AddProvider(provider ConfigProvider)
}
