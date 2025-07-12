// file_provider.go
package config

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

// DotEnvConfigProvider は.envファイルから設定を提供するプロバイダー
type DotEnvConfigProvider struct {
	values map[string]string
}

// NewDotEnvConfigProvider は新しいDotEnvConfigProviderインスタンスを作成する
func NewDotEnvConfigProvider() *DotEnvConfigProvider {
	return &DotEnvConfigProvider{
		values: make(map[string]string),
	}
}

// Load は指定されたパスの設定ファイルを読み込む
func (p *DotEnvConfigProvider) Load(filePath string) error {
	// ファイルが存在しない場合は無視
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil
	}

	// ファイルパーミッションをチェック (Unixのみ)
	if runtime.GOOS != "windows" {
		info, err := os.Stat(filePath)
		if err != nil {
			return err
		}
		mode := info.Mode()
		if mode&0077 != 0 {
			return fmt.Errorf("設定ファイル %s のパーミッションが安全ではありません (現在: %o, 推奨: 600)", filePath, mode.Perm())
		}
	}

	// ファイル読み込み
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// 行ごとに処理
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// コメントや空行をスキップ
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// key=value 形式を解析
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// クォート除去
		if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
			(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
			value = value[1 : len(value)-1]
		}

		p.values[key] = value
	}

	return nil
}

// Get は指定されたキーの値を取得する。キーが存在しない場合は空文字列とfalseを返す
func (p *DotEnvConfigProvider) Get(key string) (string, bool) {
	value, exists := p.values[key]
	return value, exists
}

// GetWithDefault は指定されたキーの値を取得する。キーが存在しない場合はデフォルト値を返す
func (p *DotEnvConfigProvider) GetWithDefault(key, defaultValue string) string {
	value, exists := p.Get(key)
	if !exists {
		return defaultValue
	}
	return value
}

// GetRequired は指定されたキーの値を取得する。キーが存在しない場合はエラーを返す
func (p *DotEnvConfigProvider) GetRequired(key string) (string, error) {
	value, exists := p.Get(key)
	if !exists {
		return "", fmt.Errorf("必須設定 %s が設定ファイルにありません", key)
	}
	return value, nil
}
