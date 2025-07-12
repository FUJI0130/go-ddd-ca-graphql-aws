// pkg/config/chain_provider_test.go
package config

import (
	"os"
	"testing"
)

// 静的な設定値を返すテスト用プロバイダー
type TestConfigProvider struct {
	values map[string]string
	source string
}

func NewTestConfigProvider(values map[string]string, source string) *TestConfigProvider {
	return &TestConfigProvider{
		values: values,
		source: source,
	}
}

func (p *TestConfigProvider) Get(key string) (string, bool) {
	val, exists := p.values[key]
	return val, exists
}

func (p *TestConfigProvider) GetString(key, defaultValue string) string {
	if val, exists := p.Get(key); exists {
		return val
	}
	return defaultValue
}

func (p *TestConfigProvider) GetInt(key string, defaultValue int) int {
	return defaultValue // 簡略化のため常にデフォルト値を返す
}

func (p *TestConfigProvider) GetBool(key string, defaultValue bool) bool {
	return defaultValue // 簡略化のため常にデフォルト値を返す
}

func (p *TestConfigProvider) Source() string {
	return p.source
}

func TestChainedConfigProvider_Get(t *testing.T) {
	// テスト用プロバイダーの作成
	provider1 := NewTestConfigProvider(map[string]string{
		"KEY1": "value1_from_provider1",
		"KEY2": "value2_from_provider1",
	}, "provider1")

	provider2 := NewTestConfigProvider(map[string]string{
		"KEY2": "value2_from_provider2",
		"KEY3": "value3_from_provider2",
	}, "provider2")

	// チェーンプロバイダーの作成
	chainProvider := NewChainedConfigProvider(provider1, provider2)

	tests := []struct {
		name         string
		key          string
		expectExists bool
		expectValue  string
	}{
		{
			name:         "provider1のみに存在する値",
			key:          "KEY1",
			expectExists: true,
			expectValue:  "value1_from_provider1",
		},
		{
			name:         "両方に存在する値（provider1が優先）",
			key:          "KEY2",
			expectExists: true,
			expectValue:  "value2_from_provider1",
		},
		{
			name:         "provider2のみに存在する値",
			key:          "KEY3",
			expectExists: true,
			expectValue:  "value3_from_provider2",
		},
		{
			name:         "存在しない値",
			key:          "NON_EXISTENT_KEY",
			expectExists: false,
			expectValue:  "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			val, exists := chainProvider.Get(tc.key)

			if exists != tc.expectExists {
				t.Errorf("期待: exists=%v, 実際: exists=%v", tc.expectExists, exists)
			}

			if val != tc.expectValue {
				t.Errorf("期待: value=%q, 実際: value=%q", tc.expectValue, val)
			}
		})
	}
}

func TestChainedConfigProvider_GetString(t *testing.T) {
	// テスト用プロバイダーの作成
	provider1 := NewTestConfigProvider(map[string]string{
		"KEY1": "value1_from_provider1",
	}, "provider1")

	provider2 := NewTestConfigProvider(map[string]string{
		"KEY2": "value2_from_provider2",
	}, "provider2")

	// チェーンプロバイダーの作成
	chainProvider := NewChainedConfigProvider(provider1, provider2)

	tests := []struct {
		name           string
		key            string
		defaultValue   string
		expectedResult string
	}{
		{
			name:           "provider1の値を取得",
			key:            "KEY1",
			defaultValue:   "default",
			expectedResult: "value1_from_provider1",
		},
		{
			name:           "provider2の値を取得",
			key:            "KEY2",
			defaultValue:   "default",
			expectedResult: "value2_from_provider2",
		},
		{
			name:           "存在しない値はデフォルト値",
			key:            "NON_EXISTENT_KEY",
			defaultValue:   "default_value",
			expectedResult: "default_value",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := chainProvider.GetString(tc.key, tc.defaultValue)

			if result != tc.expectedResult {
				t.Errorf("期待: %q, 実際: %q", tc.expectedResult, result)
			}
		})
	}
}

func TestChainedConfigProvider_Source(t *testing.T) {
	tests := []struct {
		name           string
		providers      []ConfigProvider
		expectedSource string
	}{
		{
			name: "2つのプロバイダー",
			providers: []ConfigProvider{
				NewTestConfigProvider(nil, "provider1"),
				NewTestConfigProvider(nil, "provider2"),
			},
			expectedSource: "chain(provider1 > provider2)",
		},
		{
			name: "3つのプロバイダー",
			providers: []ConfigProvider{
				NewTestConfigProvider(nil, "provider1"),
				NewTestConfigProvider(nil, "provider2"),
				NewTestConfigProvider(nil, "provider3"),
			},
			expectedSource: "chain(provider1 > provider2 > provider3)",
		},
		{
			name:           "プロバイダーなし",
			providers:      []ConfigProvider{},
			expectedSource: "chain()",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			chainProvider := NewChainedConfigProvider(tc.providers...)
			source := chainProvider.Source()

			if source != tc.expectedSource {
				t.Errorf("期待: %q, 実際: %q", tc.expectedSource, source)
			}
		})
	}
}

func TestChainedConfigProvider_WithEnvProvider(t *testing.T) {
	tests := []struct {
		name           string
		envKey         string
		envValue       string
		keyToGet       string
		defaultValue   string
		expectedResult string
		setupFunc      func()
		cleanupFunc    func()
	}{
		{
			name:           "環境変数が優先される",
			envKey:         "TEST_CHAIN_KEY",
			envValue:       "env_value",
			keyToGet:       "TEST_CHAIN_KEY",
			defaultValue:   "default",
			expectedResult: "env_value",
			setupFunc: func() {
				os.Setenv("TEST_CHAIN_KEY", "env_value")
			},
			cleanupFunc: func() {
				os.Unsetenv("TEST_CHAIN_KEY")
			},
		},
		{
			name:           "環境変数がない場合は静的プロバイダーの値",
			envKey:         "TEST_CHAIN_KEY",
			envValue:       "",
			keyToGet:       "TEST_CHAIN_KEY",
			defaultValue:   "default",
			expectedResult: "static_value",
			setupFunc: func() {
				os.Unsetenv("TEST_CHAIN_KEY")
			},
			cleanupFunc: func() {},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// テスト環境のセットアップ
			tc.setupFunc()
			defer tc.cleanupFunc()

			// 環境変数プロバイダー
			envProvider := NewEnvConfigProvider("")

			// 静的プロバイダー
			staticProvider := NewTestConfigProvider(map[string]string{
				"TEST_CHAIN_KEY": "static_value",
			}, "static")

			// チェーンプロバイダー作成
			chainProvider := NewChainedConfigProvider(envProvider, staticProvider)

			// テスト実行
			result := chainProvider.GetString(tc.keyToGet, tc.defaultValue)

			if result != tc.expectedResult {
				t.Errorf("期待: %q, 実際: %q", tc.expectedResult, result)
			}
		})
	}
}
