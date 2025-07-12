// pkg/config/config_test.go
package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	// 環境変数をクリア
	os.Clearenv()

	// テスト用の基本設定
	t.Run("基本的な設定読み込み", func(t *testing.T) {
		config, err := LoadConfig("../../configs")
		assert.NoError(t, err)
		assert.NotNil(t, config)
		assert.Equal(t, "development", config.Environment)
		assert.Equal(t, 8080, config.Server.Port)
	})

	// 環境変数が設定ファイルよりも優先されることをテスト
	t.Run("環境変数が設定ファイルよりも優先される", func(t *testing.T) {
		// テスト用の環境変数を設定
		os.Setenv("DB_HOST", "env-db-host")
		os.Setenv("DB_PORT", "1234")

		config, err := LoadConfig("../../configs")
		assert.NoError(t, err)
		assert.Equal(t, "env-db-host", config.Database.Host)
		assert.Equal(t, 1234, config.Database.Port)

		// テスト用の環境変数をクリア
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
	})

	// クラウド環境検出のテスト
	t.Run("クラウド環境での設定ファイル不使用", func(t *testing.T) {
		// クラウド環境フラグを設定
		os.Setenv("IS_CLOUD_ENV", "true")
		os.Setenv("DB_HOST", "cloud-db-host")

		config, err := LoadConfig("../../configs")
		assert.NoError(t, err)
		assert.Equal(t, "cloud-db-host", config.Database.Host)

		// テスト用の環境変数をクリア
		os.Unsetenv("IS_CLOUD_ENV")
		os.Unsetenv("DB_HOST")
	})

	// ECS環境検出のテスト
	t.Run("ECS環境での設定ファイル不使用", func(t *testing.T) {
		// ECS環境フラグを設定
		os.Setenv("ECS_CONTAINER_METADATA_URI", "http://169.254.170.2/v3")
		os.Setenv("DB_HOST", "ecs-db-host")

		config, err := LoadConfig("../../configs")
		assert.NoError(t, err)
		assert.Equal(t, "ecs-db-host", config.Database.Host)

		// テスト用の環境変数をクリア
		os.Unsetenv("ECS_CONTAINER_METADATA_URI")
		os.Unsetenv("DB_HOST")
	})

	// 期間のパース
	t.Run("期間のパース", func(t *testing.T) {
		assert.Equal(t, 15*time.Second, parseDuration("15s"))
		assert.Equal(t, 24*time.Hour, parseDuration("24h"))
		// 無効な期間形式の場合は15秒（デフォルト値）
		assert.Equal(t, 15*time.Second, parseDuration("invalid"))
	})
}

func TestGetSettingSource(t *testing.T) {
	// テスト用のプロバイダー
	envProvider := NewEnvConfigProvider("")
	staticProvider := NewStaticConfigProvider(map[string]interface{}{
		"test.key": "static-value",
	})
	chainedProvider := NewChainedConfigProvider(envProvider, staticProvider)

	// 静的プロバイダーから取得
	t.Run("静的プロバイダーからの取得", func(t *testing.T) {
		source := getSettingSource(chainedProvider, "test.key")
		assert.Equal(t, "static", source)
	})

	// 環境変数から取得
	t.Run("環境変数からの取得", func(t *testing.T) {
		os.Setenv("TEST_KEY", "env-value")
		source := getSettingSource(chainedProvider, "test.key")
		assert.Equal(t, "environment", source)
		os.Unsetenv("TEST_KEY")
	})

	// 複数のキーから取得
	t.Run("複数のキーから取得", func(t *testing.T) {
		os.Setenv("DB_HOST", "env-db-host")
		source := getSettingSource(chainedProvider, "database.host", "DB_HOST")
		assert.Equal(t, "environment", source)
		os.Unsetenv("DB_HOST")
	})
}

func TestExpandEnvVars(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		envVars  map[string]string
		expected string
	}{
		// 基本ケース
		{
			name:     "環境変数が存在する場合",
			input:    "${TEST_VAR}",
			envVars:  map[string]string{"TEST_VAR": "test-value"},
			expected: "test-value",
		},
		{
			name:     "環境変数が存在し、前後に文字列がある場合",
			input:    "prefix-${TEST_VAR}-suffix",
			envVars:  map[string]string{"TEST_VAR": "test-value"},
			expected: "prefix-test-value-suffix",
		},

		// デフォルト値のケース
		{
			name:     "環境変数が存在せず、デフォルト値がある場合",
			input:    "${NON_EXISTENT_VAR:-default-value}",
			envVars:  map[string]string{},
			expected: "default-value",
		},
		{
			name:     "環境変数が存在せず、デフォルト値もない場合",
			input:    "${NON_EXISTENT_VAR}",
			envVars:  map[string]string{},
			expected: "${NON_EXISTENT_VAR}",
		},

		// 境界ケース
		{
			name:     "環境変数の値が空文字列でデフォルト値がある場合",
			input:    "${EMPTY_VAR:-default-value}",
			envVars:  map[string]string{"EMPTY_VAR": ""},
			expected: "default-value", // 空文字列はデフォルト値を使用する
		},
		{
			name:     "デフォルト値に特殊文字が含まれる場合",
			input:    "${NON_EXISTENT_VAR:-special $@#%^&*()}",
			envVars:  map[string]string{},
			expected: "special $@#%^&*()",
		},

		// 複合ケース
		{
			name:     "複数の環境変数が存在する場合",
			input:    "${FIRST_VAR} and ${SECOND_VAR}",
			envVars:  map[string]string{"FIRST_VAR": "first", "SECOND_VAR": "second"},
			expected: "first and second",
		},
		{
			name:     "複数の環境変数で一部が存在しデフォルト値を使用する場合",
			input:    "${FIRST_VAR} and ${NON_EXISTENT_VAR:-default}",
			envVars:  map[string]string{"FIRST_VAR": "first"},
			expected: "first and default",
		},
		{
			name:     "複数の展開パターンが同一文字列内にある場合",
			input:    "${FIRST_VAR:-default1} and ${SECOND_VAR:-default2}",
			envVars:  map[string]string{"FIRST_VAR": "first"},
			expected: "first and default2",
		},

		// エッジケース
		{
			name:     "変数名に数字とアンダースコアを含む場合",
			input:    "${TEST_VAR_123}",
			envVars:  map[string]string{"TEST_VAR_123": "numbered-value"},
			expected: "numbered-value",
		},
		{
			name:     "不完全な構文 - 閉じ括弧がない",
			input:    "${UNCLOSED_VAR",
			envVars:  map[string]string{"UNCLOSED_VAR": "value"},
			expected: "${UNCLOSED_VAR", // 正規表現が一致しないので変更なし
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テスト前に環境をクリア
			os.Clearenv()

			// テストケースの環境変数を設定
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			// テスト対象関数を実行して結果を検証
			result := expandEnvVars(tt.input)
			assert.Equal(t, tt.expected, result, "入力: %s", tt.input)

			// テスト後に環境をクリア
			os.Clearenv()
		})
	}
}
