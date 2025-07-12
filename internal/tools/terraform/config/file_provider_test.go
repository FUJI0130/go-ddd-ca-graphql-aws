// file_provider_test.go
package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestDotEnvConfigProvider(t *testing.T) {
	testCases := []struct {
		name           string
		fileContent    string
		fileExists     bool
		key            string
		expectedValue  string
		expectedExists bool
		success        bool // 期待するテスト結果（成功/失敗）
	}{
		{
			name:           "設定ファイルあり・キーあり",
			fileContent:    "TEST_KEY=file_value",
			fileExists:     true,
			key:            "TEST_KEY",
			expectedValue:  "file_value",
			expectedExists: true,
			success:        true,
		},
		{
			name:           "設定ファイルあり・キーなし",
			fileContent:    "OTHER_KEY=value",
			fileExists:     true,
			key:            "TEST_KEY",
			expectedValue:  "",
			expectedExists: false,
			success:        true,
		},
		{
			name:           "設定ファイルなし",
			fileContent:    "",
			fileExists:     false,
			key:            "TEST_KEY",
			expectedValue:  "",
			expectedExists: false,
			success:        true,
		},
		{
			name:           "コメント行スキップ",
			fileContent:    "# This is a comment\nTEST_KEY=file_value",
			fileExists:     true,
			key:            "TEST_KEY",
			expectedValue:  "file_value",
			expectedExists: true,
			success:        true,
		},
		{
			name:           "引用符付き値",
			fileContent:    "TEST_KEY=\"quoted value\"",
			fileExists:     true,
			key:            "TEST_KEY",
			expectedValue:  "quoted value",
			expectedExists: true,
			success:        true,
		},
		// 異常系テストケース追加
		{
			name:           "空のファイル",
			fileContent:    "",
			fileExists:     true,
			key:            "TEST_KEY",
			expectedValue:  "",
			expectedExists: false,
			success:        true,
		},
		{
			name:           "不正なフォーマット行を含むファイル",
			fileContent:    "INVALID_LINE\nTEST_KEY=valid_value",
			fileExists:     true,
			key:            "TEST_KEY",
			expectedValue:  "valid_value",
			expectedExists: true,
			success:        true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// テスト用一時ファイルの設定
			var tempFilePath string
			if tc.fileExists {
				// 一時ファイル作成
				tempFile, err := os.CreateTemp("", "test-env-*")
				if err != nil {
					t.Fatalf("一時ファイル作成に失敗: %v", err)
				}
				defer os.Remove(tempFile.Name()) // テスト終了後にファイル削除

				// ファイル内容の書き込み
				if _, err := tempFile.WriteString(tc.fileContent); err != nil {
					t.Fatalf("ファイル書き込みに失敗: %v", err)
				}
				if err := tempFile.Close(); err != nil {
					t.Fatalf("ファイルクローズに失敗: %v", err)
				}

				tempFilePath = tempFile.Name()
			} else {
				// 存在しないファイルパス
				tempFilePath = filepath.Join(os.TempDir(), "non-existent-file")
			}

			// テスト対象のプロバイダー作成
			provider := NewDotEnvConfigProvider()

			// ファイル読み込み
			err := provider.Load(tempFilePath)
			if tc.fileExists && err != nil {
				t.Fatalf("Load(): 予期しないエラー: %v", err)
			}

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
func TestDotEnvConfigProviderErrorCases(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("このテストはUnix系OSでのみ実行されます")
	}

	// パーミッションエラーテスト
	t.Run("ファイルパーミッションエラー", func(t *testing.T) {
		// 一時ファイル作成
		tempFile, err := os.CreateTemp("", "test-env-perm-*")
		if err != nil {
			t.Fatalf("一時ファイル作成に失敗: %v", err)
		}
		defer os.Remove(tempFile.Name())

		// ファイル内容の書き込み
		if _, err := tempFile.WriteString("TEST_KEY=value"); err != nil {
			t.Fatalf("ファイル書き込みに失敗: %v", err)
		}
		if err := tempFile.Close(); err != nil {
			t.Fatalf("ファイルクローズに失敗: %v", err)
		}

		// 危険なパーミッションに設定 (666 = rw-rw-rw-)
		if err := os.Chmod(tempFile.Name(), 0666); err != nil {
			t.Fatalf("ファイルパーミッション変更に失敗: %v", err)
		}

		// テスト対象のプロバイダー作成
		provider := NewDotEnvConfigProvider()

		// ファイル読み込み - パーミッションエラーが発生するはず
		err = provider.Load(tempFile.Name())
		if err == nil {
			t.Error("Load(): 危険なパーミッションでもエラーが発生しませんでした")
		} else if !strings.Contains(err.Error(), "パーミッションが安全ではありません") {
			t.Errorf("Load(): 予期しないエラー: %v", err)
		}
	})

	// 読み込み権限なしテスト
	t.Run("読み込み権限なしファイル", func(t *testing.T) {
		// 一時ファイル作成
		tempFile, err := os.CreateTemp("", "test-env-noperm-*")
		if err != nil {
			t.Fatalf("一時ファイル作成に失敗: %v", err)
		}
		defer os.Remove(tempFile.Name())

		// ファイル内容の書き込み
		if _, err := tempFile.WriteString("TEST_KEY=value"); err != nil {
			t.Fatalf("ファイル書き込みに失敗: %v", err)
		}
		if err := tempFile.Close(); err != nil {
			t.Fatalf("ファイルクローズに失敗: %v", err)
		}

		// 読み込み権限なしに設定 (200 = -w-------)
		if err := os.Chmod(tempFile.Name(), 0200); err != nil {
			t.Fatalf("ファイルパーミッション変更に失敗: %v", err)
		}

		// テスト対象のプロバイダー作成
		provider := NewDotEnvConfigProvider()

		// ファイル読み込み - 権限エラーが発生するはず
		err = provider.Load(tempFile.Name())
		if err == nil {
			t.Error("Load(): 読み込み権限なしでもエラーが発生しませんでした")
		}
	})
}

// 非存在ファイルの扱いに特化したテスト
func TestDotEnvConfigProviderNonExistentFile(t *testing.T) {
	// 確実に存在しないファイルパス
	nonExistentPath := filepath.Join(os.TempDir(), "definitely-non-existent-file-for-test")

	// 既存ファイルがある場合に備えて削除（エラーは無視）
	os.Remove(nonExistentPath)

	provider := NewDotEnvConfigProvider()

	// 非存在ファイルの読み込みはエラーにならないことを確認
	err := provider.Load(nonExistentPath)
	if err != nil {
		t.Errorf("非存在ファイルの読み込みでエラーが発生: %v", err)
	}

	// 読み込み後に値が空であることを確認
	value, exists := provider.Get("ANY_KEY")
	if exists {
		t.Error("非存在ファイル読み込み後にキーが存在していました")
	}
	if value != "" {
		t.Errorf("非存在ファイル読み込み後に空でない値が返されました: %s", value)
	}
}
