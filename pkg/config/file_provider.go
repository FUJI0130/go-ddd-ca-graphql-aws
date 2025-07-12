// pkg/config/file_provider.go
package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

// FileConfigProvider は設定ファイルから設定を読み込むプロバイダー
type FileConfigProvider struct {
	viper  *viper.Viper
	source string
}

// NewFileConfigProvider は新しいファイル設定プロバイダーを作成
func NewFileConfigProvider(filePath, fileName, fileType string) (*FileConfigProvider, error) {
	v := viper.New()

	// 設定ファイルの設定
	v.SetConfigName(fileName)
	v.SetConfigType(fileType)
	v.AddConfigPath(filePath)

	// 設定ファイルの読み込み
	err := v.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// 設定ファイルが見つからない場合は警告を出力
			log.Printf("警告: %s.%s 設定ファイルが見つかりません。", fileName, fileType)
			return nil, fmt.Errorf("設定ファイルが見つかりません: %w", err)
		}
		return nil, fmt.Errorf("設定ファイルの読み込みに失敗しました: %w", err)
	}

	source := fmt.Sprintf("file(%s)", v.ConfigFileUsed())

	return &FileConfigProvider{
		viper:  v,
		source: source,
	}, nil
}

// Get は指定したキーの設定値を取得
func (p *FileConfigProvider) Get(key string) (string, bool) {
	if !p.viper.IsSet(key) {
		return "", false
	}

	value := p.viper.GetString(key)
	return value, true
}

// GetString は指定したキーの文字列値を取得
func (p *FileConfigProvider) GetString(key, defaultValue string) string {
	if !p.viper.IsSet(key) {
		return defaultValue
	}
	return p.viper.GetString(key)
}

// GetInt は指定したキーの整数値を取得
func (p *FileConfigProvider) GetInt(key string, defaultValue int) int {
	if !p.viper.IsSet(key) {
		return defaultValue
	}
	return p.viper.GetInt(key)
}

// GetBool は指定したキーの真偽値を取得
func (p *FileConfigProvider) GetBool(key string, defaultValue bool) bool {
	if !p.viper.IsSet(key) {
		return defaultValue
	}
	return p.viper.GetBool(key)
}

// Source は設定値のソース種別を取得
func (p *FileConfigProvider) Source() string {
	return p.source
}

// NewEmptyConfigProvider は空の設定プロバイダーを作成
func NewEmptyConfigProvider() *EmptyConfigProvider {
	return &EmptyConfigProvider{}
}

// EmptyConfigProvider は空の設定を提供するプロバイダー
type EmptyConfigProvider struct{}

// Get は常に値がないことを返す
func (p *EmptyConfigProvider) Get(key string) (string, bool) {
	return "", false
}

// GetString は常にデフォルト値を返す
func (p *EmptyConfigProvider) GetString(key, defaultValue string) string {
	return defaultValue
}

// GetInt は常にデフォルト値を返す
func (p *EmptyConfigProvider) GetInt(key string, defaultValue int) int {
	return defaultValue
}

// GetBool は常にデフォルト値を返す
func (p *EmptyConfigProvider) GetBool(key string, defaultValue bool) bool {
	return defaultValue
}

// Source は設定値のソース種別を取得
func (p *EmptyConfigProvider) Source() string {
	return "empty"
}
