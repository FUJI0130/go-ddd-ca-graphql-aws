// chain_provider_test.go
package config

import (
	"fmt"
	"strings"
	"testing"
)

// テスト用のモックプロバイダー
type MockConfigProvider struct {
	values map[string]string
	name   string // プロバイダーの名前を追加（デバッグ用）
}

func NewMockConfigProvider(values map[string]string, name string) *MockConfigProvider {
	return &MockConfigProvider{
		values: values,
		name:   name,
	}
}

func (p *MockConfigProvider) Get(key string) (string, bool) {
	value, exists := p.values[key]
	return value, exists
}

func (p *MockConfigProvider) GetWithDefault(key, defaultValue string) string {
	value, exists := p.Get(key)
	if !exists {
		return defaultValue
	}
	return value
}

func (p *MockConfigProvider) GetRequired(key string) (string, error) {
	value, exists := p.Get(key)
	if !exists {
		return "", fmt.Errorf("必須設定 %s が設定されていません", key)
	}
	return value, nil
}

// TestProviderPriority は追加順による優先順位を明示的にテストする
func TestProviderPriority(t *testing.T) {
	// チェーンプロバイダー作成
	chain := NewChainConfigProvider()

	// 最初のプロバイダーを追加
	provider1 := NewMockConfigProvider(map[string]string{"PRIORITY_KEY": "value1"}, "provider1")
	chain.AddProvider(provider1)

	// 2番目のプロバイダーを追加
	provider2 := NewMockConfigProvider(map[string]string{"PRIORITY_KEY": "value2"}, "provider2")
	chain.AddProvider(provider2)

	// 現在の実装では、先に追加したプロバイダー（provider1）が優先されるはず
	value, exists := chain.Get("PRIORITY_KEY")
	if !exists {
		t.Error("キーが見つかりませんでした")
	}
	if value != "value1" {
		t.Errorf("優先順位エラー：期待=value1, 実際=%s", value)
	}

	// 順序を逆にしたテスト
	chain2 := NewChainConfigProvider()

	// 順序を逆にして追加
	chain2.AddProvider(provider2) // 先に provider2 を追加
	chain2.AddProvider(provider1) // 後に provider1 を追加

	// 現在の実装では、先に追加したプロバイダー（provider2）が優先されるはず
	value, exists = chain2.Get("PRIORITY_KEY")
	if !exists {
		t.Error("キーが見つかりませんでした")
	}
	if value != "value2" {
		t.Errorf("優先順位エラー：期待=value2, 実際=%s", value)
	}
}

func TestChainConfigProvider(t *testing.T) {
	testCases := []struct {
		name           string
		providers      []map[string]string // 各プロバイダーの値マップ
		key            string
		expectedValue  string
		expectedExists bool
		success        bool // 期待するテスト結果（成功/失敗）
	}{
		{
			name: "配列の先頭のプロバイダーから取得",
			providers: []map[string]string{
				{"TEST_KEY": "value1"},
				{"TEST_KEY": "value2"},
			},
			key:            "TEST_KEY",
			expectedValue:  "value1",
			expectedExists: true,
			success:        true,
		},
		{
			name: "2番目のプロバイダーから取得",
			providers: []map[string]string{
				{},
				{"TEST_KEY": "value2"},
			},
			key:            "TEST_KEY",
			expectedValue:  "value2",
			expectedExists: true,
			success:        true,
		},
		{
			name: "存在しないキー",
			providers: []map[string]string{
				{"KEY1": "value1"},
				{"KEY2": "value2"},
			},
			key:            "TEST_KEY",
			expectedValue:  "",
			expectedExists: false,
			success:        true,
		},
		{
			name:           "プロバイダーなし",
			providers:      []map[string]string{},
			key:            "TEST_KEY",
			expectedValue:  "",
			expectedExists: false,
			success:        true,
		},
		{
			name: "複数プロバイダーの優先順位",
			providers: []map[string]string{
				{"SHARED_KEY": "first"},
				{"SHARED_KEY": "second"},
				{"SHARED_KEY": "third"},
			},
			key:            "SHARED_KEY",
			expectedValue:  "first", // 配列の最初の要素が優先される
			expectedExists: true,
			success:        true,
		},
		// 異常系テストケース追加
		{
			name: "すべてのプロバイダーが空",
			providers: []map[string]string{
				{},
				{},
				{},
			},
			key:            "ANY_KEY",
			expectedValue:  "",
			expectedExists: false,
			success:        true,
		},
		{
			name: "一部のプロバイダーだけが値を持つ",
			providers: []map[string]string{
				{},
				{"PARTIAL_KEY": "only_here"},
				{},
			},
			key:            "PARTIAL_KEY",
			expectedValue:  "only_here",
			expectedExists: true,
			success:        true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// チェーンプロバイダー作成
			chain := NewChainConfigProvider()

			// モックプロバイダーを追加
			for i, providerValues := range tc.providers {
				providerName := fmt.Sprintf("provider%d", i+1)
				chain.AddProvider(NewMockConfigProvider(providerValues, providerName))
			}

			// Get メソッドのテスト
			value, exists := chain.Get(tc.key)

			// 結果検証
			if exists != tc.expectedExists {
				t.Errorf("Get(): exists 期待=%v, 実際=%v", tc.expectedExists, exists)
			}

			if exists && value != tc.expectedValue {
				t.Errorf("Get(): 値 期待=%v, 実際=%v", tc.expectedValue, value)
			}

			// GetWithDefault メソッドのテスト
			defaultValue := "default_value"
			valueWithDefault := chain.GetWithDefault(tc.key, defaultValue)

			expectedValueWithDefault := tc.expectedValue
			if !tc.expectedExists {
				expectedValueWithDefault = defaultValue
			}

			if valueWithDefault != expectedValueWithDefault {
				t.Errorf("GetWithDefault(): 値 期待=%v, 実際=%v", expectedValueWithDefault, valueWithDefault)
			}

			// GetRequired メソッドのテスト（存在する場合のみ）
			if tc.expectedExists {
				valueRequired, err := chain.GetRequired(tc.key)
				if err != nil {
					t.Errorf("GetRequired(): 予期しないエラー: %v", err)
				}
				if valueRequired != tc.expectedValue {
					t.Errorf("GetRequired(): 値 期待=%v, 実際=%v", tc.expectedValue, valueRequired)
				}
			} else {
				_, err := chain.GetRequired(tc.key)
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

// TestAddProviderOrder は追加順序を明示的にテストする
func TestAddProviderOrder(t *testing.T) {
	testCases := []struct {
		name          string
		addOrder      []string // プロバイダーを追加する順序
		expectedOrder []string // 期待される優先順位
		success       bool     // 期待するテスト結果
	}{
		{
			name:          "2つのプロバイダー - 先に追加した方が優先される",
			addOrder:      []string{"first", "second"},
			expectedOrder: []string{"first", "second"},
			success:       true,
		},
		{
			name:          "3つのプロバイダー - 追加順が優先順位",
			addOrder:      []string{"A", "B", "C"},
			expectedOrder: []string{"A", "B", "C"},
			success:       true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// チェーンプロバイダー作成
			chain := NewChainConfigProvider()

			// テスト用のキー
			testKey := "ORDER_TEST_KEY"

			// 順番にプロバイダーを追加
			for _, name := range tc.addOrder {
				provider := NewMockConfigProvider(map[string]string{testKey: name}, name)
				chain.AddProvider(provider)
			}

			// キーの値を取得して優先順位をテスト
			value, exists := chain.Get(testKey)
			if !exists {
				t.Error("キーが見つかりませんでした")
			} else {
				// 最も優先度の高いプロバイダーの値と比較
				expectedValue := tc.expectedOrder[0]
				if value != expectedValue {
					t.Errorf("優先順位エラー: 期待=%s, 実際=%s", expectedValue, value)
				}
			}

			// 成功/失敗の検証
			testSuccess := exists && value == tc.expectedOrder[0]
			if testSuccess != tc.success {
				t.Errorf("テスト成功判定: 期待=%v, 実際=%v", tc.success, testSuccess)
			}
		})
	}
}

// 異常系テストを追加
func TestChainConfigProviderErrorCases(t *testing.T) {
	// GetRequiredのエラーメッセージテスト
	t.Run("GetRequiredのエラーメッセージ", func(t *testing.T) {
		chain := NewChainConfigProvider()

		// プロバイダーを追加せず空の状態でテスト
		key := "MISSING_KEY"
		_, err := chain.GetRequired(key)

		// エラーが発生することを確認
		if err == nil {
			t.Error("GetRequired(): プロバイダーなしでもエラーが発生しませんでした")
			return
		}

		// エラーメッセージに期待する文字列が含まれているか確認
		expectedErrorMsgPart := "必須設定"
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

	// 同じプロバイダーを複数回追加するテスト
	t.Run("同じプロバイダーの複数回追加", func(t *testing.T) {
		chain := NewChainConfigProvider()

		// 同じプロバイダーインスタンスを作成
		provider := NewMockConfigProvider(map[string]string{"DUPLICATE_KEY": "original_value"}, "duplicate_provider")

		// 同じプロバイダーを2回追加
		chain.AddProvider(provider)
		chain.AddProvider(provider)

		// 値を取得して1回だけ追加した場合と同じ結果になることを確認
		value, exists := chain.Get("DUPLICATE_KEY")
		if !exists {
			t.Error("キーが見つかりませんでした")
			return
		}

		if value != "original_value" {
			t.Errorf("値の不一致: 期待=original_value, 実際=%s", value)
		}

		// プロバイダーの値を変更して、両方の追加に影響するか確認
		provider.values["DUPLICATE_KEY"] = "changed_value"

		value, exists = chain.Get("DUPLICATE_KEY")
		if !exists {
			t.Error("値変更後、キーが見つかりませんでした")
			return
		}

		if value != "changed_value" {
			t.Errorf("値変更後の不一致: 期待=changed_value, 実際=%s", value)
		}
	})
}

// 多層構造での優先順位テスト
func TestComplexProviderPriority(t *testing.T) {
	t.Run("多層構造での優先順位", func(t *testing.T) {
		// 最上位チェーン
		topChain := NewChainConfigProvider()

		// 第2レベルチェーン1
		subChain1 := NewChainConfigProvider()
		subChain1.AddProvider(NewMockConfigProvider(map[string]string{
			"KEY1": "subChain1_value1",
			"KEY2": "subChain1_value2",
		}, "subProvider1"))

		// 第2レベルチェーン2
		subChain2 := NewChainConfigProvider()
		subChain2.AddProvider(NewMockConfigProvider(map[string]string{
			"KEY1": "subChain2_value1",
			"KEY3": "subChain2_value3",
		}, "subProvider2"))

		// 最上位プロバイダー
		topProvider := NewMockConfigProvider(map[string]string{
			"KEY1": "top_value1",
			"KEY4": "top_value4",
		}, "topProvider")

		// チェーンを構築（追加順が優先順位）
		topChain.AddProvider(topProvider) // 最優先
		topChain.AddProvider(subChain1)   // 2番目
		topChain.AddProvider(subChain2)   // 3番目

		// 各キーをテスト
		testCases := []struct {
			key           string
			expectedValue string
			expectedFrom  string
		}{
			{"KEY1", "top_value1", "topProvider"},     // 最上位から取得
			{"KEY2", "subChain1_value2", "subChain1"}, // 2番目のチェーンから取得
			{"KEY3", "subChain2_value3", "subChain2"}, // 3番目のチェーンから取得
			{"KEY4", "top_value4", "topProvider"},     // 最上位のみにある値
			{"KEY5", "", "存在しない"},                     // 存在しないキー
		}

		for _, tc := range testCases {
			value, exists := topChain.Get(tc.key)

			if tc.expectedValue == "" {
				if exists {
					t.Errorf("キー %s: 存在しないはずが存在している（値=%s）", tc.key, value)
				}
			} else {
				if !exists {
					t.Errorf("キー %s: 存在するはずが存在していない", tc.key)
				} else if value != tc.expectedValue {
					t.Errorf("キー %s: 値の不一致 期待=%s, 実際=%s", tc.key, tc.expectedValue, value)
				}
			}
		}
	})
}

// 空のプロバイダーのエッジケーステスト
func TestEmptyProviderEdgeCases(t *testing.T) {
	t.Run("完全に空のプロバイダーチェーン", func(t *testing.T) {
		chain := NewChainConfigProvider()

		// 空のチェーンでGetを呼び出す
		value, exists := chain.Get("ANY_KEY")
		if exists {
			t.Errorf("空のチェーンでキーが存在するという結果: 値=%s", value)
		}
		if value != "" {
			t.Errorf("空のチェーンで空文字列以外の値が返された: 値=%s", value)
		}

		// 空のチェーンでGetWithDefaultを呼び出す
		defaultValue := "default_for_empty"
		valueWithDefault := chain.GetWithDefault("ANY_KEY", defaultValue)
		if valueWithDefault != defaultValue {
			t.Errorf("GetWithDefault(): デフォルト値が返されなかった: 期待=%s, 実際=%s",
				defaultValue, valueWithDefault)
		}
	})

	t.Run("空のプロバイダーのみのチェーン", func(t *testing.T) {
		chain := NewChainConfigProvider()

		// 空の値マップを持つプロバイダーを追加
		chain.AddProvider(NewMockConfigProvider(map[string]string{}, "empty1"))
		chain.AddProvider(NewMockConfigProvider(map[string]string{}, "empty2"))

		// 空のプロバイダーでGetを呼び出す
		value, exists := chain.Get("ANY_KEY")
		if exists {
			t.Errorf("空のプロバイダーのみでキーが存在するという結果: 値=%s", value)
		}
		if value != "" {
			t.Errorf("空のプロバイダーのみで空文字列以外の値が返された: 値=%s", value)
		}
	})
}
