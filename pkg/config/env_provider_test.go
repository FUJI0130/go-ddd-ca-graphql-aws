// pkg/config/env_provider_test.go
package config

import (
	"os"
	"testing"
)

func TestEnvConfigProvider_Get(t *testing.T) {
	tests := []struct {
		name         string
		envKey       string
		envValue     string
		keyToGet     string
		prefix       string
		expectExists bool
		expectValue  string
	}{
		{
			name:         "環境変数から値を取得",
			envKey:       "TEST_KEY",
			envValue:     "test_value",
			keyToGet:     "TEST_KEY",
			prefix:       "",
			expectExists: true,
			expectValue:  "test_value",
		},
		{
			name:         "存在しない環境変数",
			envKey:       "",
			envValue:     "",
			keyToGet:     "NON_EXISTENT_KEY",
			prefix:       "",
			expectExists: false,
			expectValue:  "",
		},
		{
			name:         "プレフィックス付き環境変数を取得",
			envKey:       "APP_TEST_KEY",
			envValue:     "prefixed_value",
			keyToGet:     "TEST_KEY",
			prefix:       "APP_",
			expectExists: true,
			expectValue:  "prefixed_value",
		},
		{
			name:         "プレフィックスが既に含まれているキーで取得",
			envKey:       "APP_TEST_KEY",
			envValue:     "prefixed_value",
			keyToGet:     "APP_TEST_KEY",
			prefix:       "APP_",
			expectExists: true,
			expectValue:  "prefixed_value",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// 環境をクリーンにする
			if tc.envKey != "" {
				os.Setenv(tc.envKey, tc.envValue)
				defer os.Unsetenv(tc.envKey)
			}

			provider := NewEnvConfigProvider(tc.prefix)
			val, exists := provider.Get(tc.keyToGet)

			if exists != tc.expectExists {
				t.Errorf("期待: exists=%v, 実際: exists=%v", tc.expectExists, exists)
			}

			if val != tc.expectValue {
				t.Errorf("期待: value=%q, 実際: value=%q", tc.expectValue, val)
			}
		})
	}
}

func TestEnvConfigProvider_GetString(t *testing.T) {
	tests := []struct {
		name           string
		envKey         string
		envValue       string
		keyToGet       string
		defaultValue   string
		expectedResult string
	}{
		{
			name:           "環境変数から文字列を取得",
			envKey:         "TEST_STRING",
			envValue:       "test_string_value",
			keyToGet:       "TEST_STRING",
			defaultValue:   "default",
			expectedResult: "test_string_value",
		},
		{
			name:           "存在しない環境変数はデフォルト値を返す",
			envKey:         "",
			envValue:       "",
			keyToGet:       "NON_EXISTENT_KEY",
			defaultValue:   "default_value",
			expectedResult: "default_value",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// 環境をクリーンにする
			if tc.envKey != "" {
				os.Setenv(tc.envKey, tc.envValue)
				defer os.Unsetenv(tc.envKey)
			}

			provider := NewEnvConfigProvider("")
			result := provider.GetString(tc.keyToGet, tc.defaultValue)

			if result != tc.expectedResult {
				t.Errorf("期待: %q, 実際: %q", tc.expectedResult, result)
			}
		})
	}
}

func TestEnvConfigProvider_GetInt(t *testing.T) {
	tests := []struct {
		name           string
		envKey         string
		envValue       string
		keyToGet       string
		defaultValue   int
		expectedResult int
	}{
		{
			name:           "環境変数から整数を取得",
			envKey:         "TEST_INT",
			envValue:       "123",
			keyToGet:       "TEST_INT",
			defaultValue:   456,
			expectedResult: 123,
		},
		{
			name:           "存在しない環境変数はデフォルト値を返す",
			envKey:         "",
			envValue:       "",
			keyToGet:       "NON_EXISTENT_KEY",
			defaultValue:   456,
			expectedResult: 456,
		},
		{
			name:           "不正な整数値は変換に失敗してデフォルト値を返す",
			envKey:         "TEST_INT_INVALID",
			envValue:       "not_a_number",
			keyToGet:       "TEST_INT_INVALID",
			defaultValue:   789,
			expectedResult: 789,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// 環境をクリーンにする
			if tc.envKey != "" {
				os.Setenv(tc.envKey, tc.envValue)
				defer os.Unsetenv(tc.envKey)
			}

			provider := NewEnvConfigProvider("")
			result := provider.GetInt(tc.keyToGet, tc.defaultValue)

			if result != tc.expectedResult {
				t.Errorf("期待: %d, 実際: %d", tc.expectedResult, result)
			}
		})
	}
}

func TestEnvConfigProvider_GetBool(t *testing.T) {
	tests := []struct {
		name           string
		envKey         string
		envValue       string
		keyToGet       string
		defaultValue   bool
		expectedResult bool
	}{
		// trueとして解釈されるケース
		{
			name:           "trueの文字列はtrueを返す",
			envKey:         "TEST_BOOL",
			envValue:       "true",
			keyToGet:       "TEST_BOOL",
			defaultValue:   false,
			expectedResult: true,
		},
		{
			name:           "TRUEの文字列はtrueを返す",
			envKey:         "TEST_BOOL",
			envValue:       "TRUE",
			keyToGet:       "TEST_BOOL",
			defaultValue:   false,
			expectedResult: true,
		},
		{
			name:           "1の文字列はtrueを返す",
			envKey:         "TEST_BOOL",
			envValue:       "1",
			keyToGet:       "TEST_BOOL",
			defaultValue:   false,
			expectedResult: true,
		},
		{
			name:           "yesの文字列はtrueを返す",
			envKey:         "TEST_BOOL",
			envValue:       "yes",
			keyToGet:       "TEST_BOOL",
			defaultValue:   false,
			expectedResult: true,
		},

		// falseとして解釈されるケース
		{
			name:           "falseの文字列はfalseを返す",
			envKey:         "TEST_BOOL",
			envValue:       "false",
			keyToGet:       "TEST_BOOL",
			defaultValue:   true,
			expectedResult: false,
		},
		{
			name:           "FALSEの文字列はfalseを返す",
			envKey:         "TEST_BOOL",
			envValue:       "FALSE",
			keyToGet:       "TEST_BOOL",
			defaultValue:   true,
			expectedResult: false,
		},
		{
			name:           "0の文字列はfalseを返す",
			envKey:         "TEST_BOOL",
			envValue:       "0",
			keyToGet:       "TEST_BOOL",
			defaultValue:   true,
			expectedResult: false,
		},
		{
			name:           "noの文字列はfalseを返す",
			envKey:         "TEST_BOOL",
			envValue:       "no",
			keyToGet:       "TEST_BOOL",
			defaultValue:   true,
			expectedResult: false,
		},

		// デフォルト値のケース
		{
			name:           "存在しない環境変数はデフォルト値を返す",
			envKey:         "",
			envValue:       "",
			keyToGet:       "NON_EXISTENT_KEY",
			defaultValue:   true,
			expectedResult: true,
		},
		{
			name:           "不明な値はデフォルト値を返す",
			envKey:         "TEST_BOOL",
			envValue:       "unknown",
			keyToGet:       "TEST_BOOL",
			defaultValue:   true,
			expectedResult: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// 環境をクリーンにする
			if tc.envKey != "" {
				os.Setenv(tc.envKey, tc.envValue)
				defer os.Unsetenv(tc.envKey)
			}

			provider := NewEnvConfigProvider("")
			result := provider.GetBool(tc.keyToGet, tc.defaultValue)

			if result != tc.expectedResult {
				t.Errorf("期待: %v, 実際: %v", tc.expectedResult, result)
			}
		})
	}
}

func TestEnvConfigProvider_Source(t *testing.T) {
	provider := NewEnvConfigProvider("")
	source := provider.Source()

	expectedSource := "environment"
	if source != expectedSource {
		t.Errorf("期待: %q, 実際: %q", expectedSource, source)
	}
}
