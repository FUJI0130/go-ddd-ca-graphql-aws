// config.go
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// デフォルトの環境設定ファイル名（定数から変数に変更）
var defaultEnvFileName = ".env.terraform"

// getEnvFileName は環境設定ファイル名を取得する関数
// テスト時に置き換え可能
var getEnvFileName = func() string {
	return defaultEnvFileName
}

// GetEnvFileNameForTest は getEnvFileName 関数を取得するテスト用関数
func GetEnvFileNameForTest() func() string {
	return getEnvFileName
}

// SetEnvFileNameForTest は getEnvFileName 関数を設定するテスト用関数
func SetEnvFileNameForTest(fn func() string) func() string {
	orig := getEnvFileName
	getEnvFileName = fn
	return orig
}

// NewConfigManager は新しい設定マネージャーを作成する
func NewConfigManager() (ChainConfigProvider, error) {
	chain := NewChainConfigProvider()

	// 環境変数プロバイダー（最優先）
	chain.AddProvider(NewEnvConfigProvider())

	// .env ファイルプロバイダー
	dotenvProvider := NewDotEnvConfigProvider()

	// グローバル設定ファイル
	homeDir, err := os.UserHomeDir()
	if err == nil {
		globalConfig := filepath.Join(homeDir, getEnvFileName())
		if err := dotenvProvider.Load(globalConfig); err != nil {
			fmt.Fprintf(os.Stderr, "警告: グローバル設定ファイルの読み込みに失敗しました: %v\n", err)
		}
	}

	// ローカル設定ファイル（カレントディレクトリ）
	if err := dotenvProvider.Load(getEnvFileName()); err != nil {
		fmt.Fprintf(os.Stderr, "警告: ローカル設定ファイルの読み込みに失敗しました: %v\n", err)
	}

	chain.AddProvider(dotenvProvider)

	return chain, nil
}

// CheckRequiredVariables は必須環境変数が設定されているかチェックする
func CheckRequiredVariables(config ConfigProvider) error {
	missingKeys := []string{}

	requiredKeys := []string{
		"AWS_ACCESS_KEY_ID",
		"AWS_SECRET_ACCESS_KEY",
		"AWS_REGION",
	}

	for _, key := range requiredKeys {
		if _, exists := config.Get(key); !exists {
			missingKeys = append(missingKeys, key)
		}
	}

	if len(missingKeys) > 0 {
		return fmt.Errorf("以下の必須設定が不足しています:\n"+
			"  %s\n\n"+
			"設定方法:\n"+
			"  1. 環境変数を設定: export %s=値\n"+
			"  2. ~/.env.terraform ファイルに追加: %s=値\n"+
			"  3. AWS設定ファイル（~/.aws/credentials）を使用\n",
			strings.Join(missingKeys, ", "),
			missingKeys[0],
			missingKeys[0])
	}

	return nil
}
