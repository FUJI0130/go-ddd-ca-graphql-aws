// env_provider_test.go
package config

import (
	"os"
	"strings"
	"testing"
)

func TestEnvConfigProvider(t *testing.T) {
	testCases := []struct {
		name           string
		envVars        map[string]string
		key            string
		expectedValue  string
		expectedExists bool
		success        bool // 期待するテスト結果（成功/失敗）
	}{
		{
			name:           "環境変数あり",
			envVars:        map[string]string{"TEST_KEY": "test_value"},
			key:            "TEST_KEY",
			expectedValue:  "test_value",
			expectedExists: true,
			success:        true,
		},
		{
			name:           "環境変数なし",
			envVars:        map[string]string{},
			key:            "TEST_KEY",
			expectedValue:  "",
			expectedExists: false,
			success:        true,
		},
		{
			name:           "GetWithDefaultで存在する環境変数",
			envVars:        map[string]string{"TEST_KEY": "test_value"},
			key:            "TEST_KEY",
			expectedValue:  "test_value",
			expectedExists: true,
			success:        true,
		},
		{
			name:           "GetWithDefaultでデフォルト値使用",
			envVars:        map[string]string{},
			key:            "TEST_KEY",
			expectedValue:  "default_value",
			expectedExists: false,
			success:        true,
		},
		// 異常系テストケース追加
		{
			name:           "空の環境変数値",
			envVars:        map[string]string{"TEST_KEY": ""},
			key:            "TEST_KEY",
			expectedValue:  "",
			expectedExists: true, // 空文字列でも存在はする
			success:        true,
		},
		{
			name:           "環境変数の明示的な削除",
			envVars:        map[string]string{"TEST_KEY": "delete_me"},
			key:            "TEST_KEY",
			expectedValue:  "",
			expectedExists: false,
			success:        true,
		},
		{
			name:           "存在しないキーのGetRequired",
			envVars:        map[string]string{},
			key:            "NONEXISTENT_KEY",
			expectedValue:  "",
			expectedExists: false,
			success:        true,
		},
		{
			name:           "複数環境変数での対象キー",
			envVars:        map[string]string{"TEST_KEY": "correct_value", "OTHER_KEY": "other_value"},
			key:            "TEST_KEY",
			expectedValue:  "correct_value",
			expectedExists: true,
			success:        true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 特殊なケース：環境変数を明示的に削除するテスト
			if tc.name == "環境変数の明示的な削除" {
				// 一旦環境変数を設定
				os.Setenv(tc.key, tc.envVars[tc.key])
				// 次にすぐに削除
				os.Unsetenv(tc.key)
			} else {
				// 通常のケース：テストケースの環境変数を設定
				for k, v := range tc.envVars {
					os.Setenv(k, v)
					defer os.Unsetenv(k)
				}
			}

			// テスト対象のプロバイダー作成
			provider := NewEnvConfigProvider()

			// Get メソッドのテスト
			value, exists := provider.Get(tc.key)

			// 結果検証
			if exists != tc.expectedExists {
				t.Errorf("Get(): exists 期待=%v, 実際=%v", tc.expectedExists, exists)
			}

			if exists && value != tc.expectedValue {
				t.Errorf("Get(): 値 期待=%v, 実際=%v", tc.expectedValue, value)
			}

			// GetWithDefault メソッドのテスト
			defaultValue := "default_value"
			valueWithDefault := provider.GetWithDefault(tc.key, defaultValue)

			expectedValueWithDefault := tc.expectedValue
			if !tc.expectedExists {
				expectedValueWithDefault = defaultValue
			}

			if valueWithDefault != expectedValueWithDefault {
				t.Errorf("GetWithDefault(): 値 期待=%v, 実際=%v", expectedValueWithDefault, valueWithDefault)
			}

			// GetRequired メソッドのテスト（存在する場合のみ）
			if tc.expectedExists {
				valueRequired, err := provider.GetRequired(tc.key)
				if err != nil {
					t.Errorf("GetRequired(): 予期しないエラー: %v", err)
				}
				if valueRequired != tc.expectedValue {
					t.Errorf("GetRequired(): 値 期待=%v, 実際=%v", tc.expectedValue, valueRequired)
				}
			} else {
				_, err := provider.GetRequired(tc.key)
				if err == nil {
					t.Errorf("GetRequired(): エラーが発生しませんでした")
				}
			}

			// 成功/失敗の検証
			testSuccess := true
			if tc.expectedExists {
				testSuccess = (value == tc.expectedValue)
			} else {
				testSuccess = !exists
			}

			if testSuccess != tc.success {
				t.Errorf("テスト成功判定: 期待=%v, 実際=%v", tc.success, testSuccess)
			}
		})
	}
}

// 異常系テストを別途追加
func TestEnvConfigProviderErrorCases(t *testing.T) {
	// GetRequired関数の詳細なエラーメッセージテスト
	t.Run("GetRequiredのエラーメッセージ", func(t *testing.T) {
		// 確実に存在しない環境変数キー
		key := "DEFINITELY_NON_EXISTENT_ENV_VAR_FOR_TEST"

		// 環境変数が存在しないことを確保
		os.Unsetenv(key)

		provider := NewEnvConfigProvider()
		_, err := provider.GetRequired(key)

		// エラーが発生することを確認
		if err == nil {
			t.Error("GetRequired(): 存在しない環境変数でもエラーが発生しませんでした")
			return
		}

		// エラーメッセージに期待する文字列が含まれているか確認
		expectedErrorMsgPart := "必須環境変数"
		if !strings.Contains(err.Error(), expectedErrorMsgPart) {
			t.Errorf("GetRequired(): エラーメッセージに '%s' が含まれていません: %v",
				expectedErrorMsgPart, err)
		}

		expectedKeyInMsg := key
		if !strings.Contains(err.Error(), expectedKeyInMsg) {
			t.Errorf("GetRequired(): エラーメッセージにキー '%s' が含まれていません: %v",
				expectedKeyInMsg, err)
		}
	})

	// 複数のGet呼び出しの独立性テスト
	t.Run("複数のGet呼び出しの独立性", func(t *testing.T) {
		// 2つの異なる環境変数を設定
		key1 := "TEST_ENV_VAR_1"
		value1 := "value1"
		key2 := "TEST_ENV_VAR_2"
		value2 := "value2"

		os.Setenv(key1, value1)
		os.Setenv(key2, value2)
		defer os.Unsetenv(key1)
		defer os.Unsetenv(key2)

		provider := NewEnvConfigProvider()

		// 1つ目の変数の値を取得
		val1, exists1 := provider.Get(key1)
		if !exists1 || val1 != value1 {
			t.Errorf("Get(%s): 期待=%s, 実際=%s, exists=%v", key1, value1, val1, exists1)
		}

		// 2つ目の変数の値を取得
		val2, exists2 := provider.Get(key2)
		if !exists2 || val2 != value2 {
			t.Errorf("Get(%s): 期待=%s, 実際=%s, exists=%v", key2, value2, val2, exists2)
		}

		// 存在しない変数の値を取得
		val3, exists3 := provider.Get("NON_EXISTENT_KEY")
		if exists3 || val3 != "" {
			t.Errorf("Get(NON_EXISTENT_KEY): 期待=\"\", 実際=%s, exists=%v", val3, exists3)
		}
	})
}

// 追加の複雑なテストケース
func TestEnvConfigProviderEdgeCases(t *testing.T) {
	// 特殊文字を含む環境変数
	t.Run("特殊文字を含む環境変数", func(t *testing.T) {
		key := "TEST_SPECIAL_CHARS"
		value := "!@#$%^&*()_+{}[]|\\:;\"'<>,.?/"

		os.Setenv(key, value)
		defer os.Unsetenv(key)

		provider := NewEnvConfigProvider()

		val, exists := provider.Get(key)
		if !exists {
			t.Errorf("特殊文字を含む環境変数が存在しませんでした")
		}

		if val != value {
			t.Errorf("特殊文字の値が一致しません: 期待=%s, 実際=%s", value, val)
		}
	})

	// 長い値を持つ環境変数
	t.Run("長い値を持つ環境変数", func(t *testing.T) {
		key := "TEST_LONG_VALUE"
		// 1000文字の長さの値
		value := strings.Repeat("abcdefghij", 100)

		os.Setenv(key, value)
		defer os.Unsetenv(key)

		provider := NewEnvConfigProvider()

		val, exists := provider.Get(key)
		if !exists {
			t.Errorf("長い値を持つ環境変数が存在しませんでした")
		}

		if val != value {
			t.Errorf("長い値が一致しません: 長さ(期待)=%d, 長さ(実際)=%d",
				len(value), len(val))
		}
	})
}
