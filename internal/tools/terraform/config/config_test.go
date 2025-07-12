// config_test.go
package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestNewConfigManager(t *testing.T) {
	// テスト環境のセットアップ
	// 環境変数設定
	os.Setenv("ENV_TEST_KEY", "env_value")
	defer os.Unsetenv("ENV_TEST_KEY")

	// テスト用の一時ファイル作成（ローカル設定）
	localContent := "FILE_TEST_KEY=local_value\nBOTH_TEST_KEY=local_value"
	localFile, err := os.CreateTemp("", "local-env-*")
	if err != nil {
		t.Fatalf("ローカル設定ファイル作成に失敗: %v", err)
	}
	defer os.Remove(localFile.Name())
	if _, err := localFile.WriteString(localContent); err != nil {
		t.Fatalf("ファイル書き込みに失敗: %v", err)
	}
	if err := localFile.Close(); err != nil {
		t.Fatalf("ファイルクローズに失敗: %v", err)
	}

	// 元の関数を保存
	originalGetEnvFileName := getEnvFileName
	// テスト用に上書き
	getEnvFileName = func() string {
		return filepath.Base(localFile.Name())
	}
	// テスト後に復元
	defer func() {
		getEnvFileName = originalGetEnvFileName
	}()

	// 環境変数と設定ファイルの両方にある値を設定
	os.Setenv("BOTH_TEST_KEY", "env_value")
	defer os.Unsetenv("BOTH_TEST_KEY")

	// カレントディレクトリを一時的に変更
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("カレントディレクトリ取得に失敗: %v", err)
	}
	tempDir := filepath.Dir(localFile.Name())
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("カレントディレクトリ変更に失敗: %v", err)
	}
	defer os.Chdir(originalDir)

	// 設定マネージャーのテスト
	configManager, err := NewConfigManager()
	if err != nil {
		t.Fatalf("設定マネージャー作成に失敗: %v", err)
	}

	// 環境変数からの取得テスト
	envValue, envExists := configManager.Get("ENV_TEST_KEY")
	if !envExists {
		t.Error("環境変数からの取得に失敗")
	}
	if envValue != "env_value" {
		t.Errorf("環境変数値: 期待=%s, 実際=%s", "env_value", envValue)
	}

	// ファイルからの取得テスト
	fileValue, fileExists := configManager.Get("FILE_TEST_KEY")
	if !fileExists {
		t.Error("ファイルからの取得に失敗")
	}
	if fileValue != "local_value" {
		t.Errorf("ファイル値: 期待=%s, 実際=%s", "local_value", fileValue)
	}

	// 優先順位テスト（環境変数が優先されるべき）
	bothValue, bothExists := configManager.Get("BOTH_TEST_KEY")
	if !bothExists {
		t.Error("環境変数とファイルの両方にある値の取得に失敗")
	}
	if bothValue != "env_value" {
		t.Errorf("優先順位: 期待=%s, 実際=%s", "env_value", bothValue)
	}
}

func TestCheckRequiredVariables(t *testing.T) {
	testCases := []struct {
		name        string
		envVars     map[string]string
		expectError bool
		success     bool // 期待するテスト結果（成功/失敗）
	}{
		{
			name: "すべての必須変数あり",
			envVars: map[string]string{
				"AWS_ACCESS_KEY_ID":     "test_id",
				"AWS_SECRET_ACCESS_KEY": "test_secret",
				"AWS_REGION":            "us-west-2",
			},
			expectError: false,
			success:     true,
		},
		{
			name: "一部の必須変数なし",
			envVars: map[string]string{
				"AWS_ACCESS_KEY_ID": "test_id",
				"AWS_REGION":        "us-west-2",
			},
			expectError: true,
			success:     true,
		},
		{
			name:        "すべての必須変数なし",
			envVars:     map[string]string{},
			expectError: true,
			success:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 環境変数のクリーンアップ
			os.Unsetenv("AWS_ACCESS_KEY_ID")
			os.Unsetenv("AWS_SECRET_ACCESS_KEY")
			os.Unsetenv("AWS_REGION")

			// テスト用環境変数の設定
			for k, v := range tc.envVars {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			// モックプロバイダーの作成
			provider := NewChainConfigProvider()
			provider.AddProvider(NewEnvConfigProvider())

			// 検証実行
			err := CheckRequiredVariables(provider)

			// エラー検証
			if (err != nil) != tc.expectError {
				t.Errorf("エラー発生: 期待=%v, 実際=%v", tc.expectError, (err != nil))
			}

			// 成功/失敗の検証
			testSuccess := (err != nil) == tc.expectError
			if testSuccess != tc.success {
				t.Errorf("テスト成功判定: 期待=%v, 実際=%v", tc.success, testSuccess)
			}
		})
	}
}

// 異常系テストの追加
// 異常系テストの追加
func TestConfigManagerErrorCases(t *testing.T) {
	// ※ UserHomeDirをモックするために、パッケージ内の一時変数を導入するようコードを修正する必要があります
	// 実際のコードをテストできるよう、各テストケースを分割しています

	// グローバル設定とローカル設定の競合テスト
	t.Run("グローバル設定とローカル設定の競合", func(t *testing.T) {
		// テスト用の一時ファイル作成（グローバル設定）
		globalContent := "CONFLICT_KEY=global_value\nGLOBAL_ONLY=global_value"
		globalFile, err := os.CreateTemp("", "global-env-*")
		if err != nil {
			t.Fatalf("グローバル設定ファイル作成に失敗: %v", err)
		}
		defer os.Remove(globalFile.Name())
		if _, err := globalFile.WriteString(globalContent); err != nil {
			t.Fatalf("ファイル書き込みに失敗: %v", err)
		}
		if err := globalFile.Close(); err != nil {
			t.Fatalf("ファイルクローズに失敗: %v", err)
		}

		// テスト用の一時ファイル作成（ローカル設定）
		localContent := "CONFLICT_KEY=local_value\nLOCAL_ONLY=local_value"
		localFile, err := os.CreateTemp("", "local-env-*")
		if err != nil {
			t.Fatalf("ローカル設定ファイル作成に失敗: %v", err)
		}
		defer os.Remove(localFile.Name())
		if _, err := localFile.WriteString(localContent); err != nil {
			t.Fatalf("ファイル書き込みに失敗: %v", err)
		}
		if err := localFile.Close(); err != nil {
			t.Fatalf("ファイルクローズに失敗: %v", err)
		}

		// 環境変数をクリア
		os.Unsetenv("CONFLICT_KEY")
		os.Unsetenv("GLOBAL_ONLY")
		os.Unsetenv("LOCAL_ONLY")

		// ホームディレクトリモック（代替アプローチ）
		// 元の関数を保存し、テスト用モック関数で置換
		originalGetEnvFileName := getEnvFileName
		getEnvFileName = func() string {
			return filepath.Base(localFile.Name())
		}
		defer func() {
			getEnvFileName = originalGetEnvFileName
		}()

		// NewConfigManagerを修正してテスト用のグローバルパスを使用できるようにする必要があります
		// 以下は修正方法の例として、実際の実装に合わせる必要があります
		globalConfigPath := globalFile.Name()
		localConfigPath := localFile.Name()

		// 手動で実装（NewConfigManagerを直接使わず）
		chain := NewChainConfigProvider()
		chain.AddProvider(NewEnvConfigProvider())
		dotenvProvider := NewDotEnvConfigProvider()

		// グローバル設定を読み込み
		if err := dotenvProvider.Load(globalConfigPath); err != nil {
			t.Fatalf("グローバル設定の読み込みに失敗: %v", err)
		}

		// ローカル設定を読み込み
		if err := dotenvProvider.Load(localConfigPath); err != nil {
			t.Fatalf("ローカル設定の読み込みに失敗: %v", err)
		}

		chain.AddProvider(dotenvProvider)
		configManager := chain

		// 競合するキーのテスト - ローカル設定が優先されるべき
		conflictValue, exists := configManager.Get("CONFLICT_KEY")
		if !exists {
			t.Error("競合キーが見つかりませんでした")
		} else if conflictValue != "local_value" {
			t.Errorf("競合キーの値: 期待=local_value, 実際=%s", conflictValue)
		}

		// グローバル設定のみのキー
		globalValue, exists := configManager.Get("GLOBAL_ONLY")
		if !exists {
			t.Error("グローバル設定のみのキーが見つかりませんでした")
		} else if globalValue != "global_value" {
			t.Errorf("グローバル設定のみのキーの値: 期待=global_value, 実際=%s", globalValue)
		}

		// ローカル設定のみのキー
		localValue, exists := configManager.Get("LOCAL_ONLY")
		if !exists {
			t.Error("ローカル設定のみのキーが見つかりませんでした")
		} else if localValue != "local_value" {
			t.Errorf("ローカル設定のみのキーの値: 期待=local_value, 実際=%s", localValue)
		}
	})
}

// CheckRequiredVariablesの詳細テスト
// CheckRequiredVariablesの詳細テスト
func TestCheckRequiredVariablesErrorMessage(t *testing.T) {
	// 環境変数のクリーンアップ
	os.Unsetenv("AWS_ACCESS_KEY_ID")
	os.Unsetenv("AWS_SECRET_ACCESS_KEY")
	os.Unsetenv("AWS_REGION")

	// モックプロバイダーの作成
	provider := NewChainConfigProvider()
	provider.AddProvider(NewEnvConfigProvider())

	// 検証実行
	err := CheckRequiredVariables(provider)

	// エラーが発生することを確認
	if err == nil {
		t.Fatal("必須変数なしでもエラーが発生しない")
	}

	// エラーメッセージの検証
	errorMsg := err.Error()

	// "必須設定"という文字列が含まれているか
	if !strings.Contains(errorMsg, "必須設定") {
		t.Errorf("エラーメッセージに「必須設定」が含まれていない: %s", errorMsg)
	}

	// すべての必須変数名が含まれているか
	requiredKeys := []string{
		"AWS_ACCESS_KEY_ID",
		"AWS_SECRET_ACCESS_KEY",
		"AWS_REGION",
	}

	for _, key := range requiredKeys {
		if !strings.Contains(errorMsg, key) {
			t.Errorf("エラーメッセージに必須変数 %s が含まれていない: %s", key, errorMsg)
		}
	}

	// 設定方法の説明が含まれているか
	if !strings.Contains(errorMsg, "設定方法") {
		t.Errorf("エラーメッセージに設定方法の説明が含まれていない: %s", errorMsg)
	}

	// 環境変数設定の例が含まれているか
	if !strings.Contains(errorMsg, "export") {
		t.Errorf("エラーメッセージに環境変数設定の例が含まれていない: %s", errorMsg)
	}

	// ファイル設定の例が含まれているか
	if !strings.Contains(errorMsg, "~/.env.terraform") {
		t.Errorf("エラーメッセージに設定ファイルの例が含まれていない: %s", errorMsg)
	}
}

// 空または無効な設定ファイル名テスト
// 空または無効な設定ファイル名テスト
func TestInvalidConfigFileName(t *testing.T) {
	t.Run("空の設定ファイル名", func(t *testing.T) {
		originalGetEnvFileName := getEnvFileName
		getEnvFileName = func() string {
			return ""
		}
		defer func() {
			getEnvFileName = originalGetEnvFileName
		}()

		// 環境変数を設定
		testKey := "EMPTY_FILENAME_TEST"
		testValue := "test_value"
		os.Setenv(testKey, testValue)
		defer os.Unsetenv(testKey)

		// 設定マネージャーの代わりにチェーンを手動で構築
		chain := NewChainConfigProvider()
		chain.AddProvider(NewEnvConfigProvider())
		dotenvProvider := NewDotEnvConfigProvider()

		// 空のファイル名で読み込みを試みる
		err := dotenvProvider.Load("")
		if err != nil {
			t.Logf("空のファイル名での読み込み結果: %v", err)
		}

		chain.AddProvider(dotenvProvider)
		configManager := chain

		// 環境変数が正常に機能することを確認
		value, exists := configManager.Get(testKey)
		if !exists {
			t.Error("空のファイル名の状態で環境変数が取得できない")
		} else if value != testValue {
			t.Errorf("空のファイル名状態の環境変数値: 期待=%s, 実際=%s", testValue, value)
		}
	})

	t.Run("存在しないディレクトリのファイル名", func(t *testing.T) {
		// 環境変数を設定
		testKey := "NONEXIST_PATH_TEST"
		testValue := "test_value"
		os.Setenv(testKey, testValue)
		defer os.Unsetenv(testKey)

		// 設定マネージャーの代わりにチェーンを手動で構築
		chain := NewChainConfigProvider()
		chain.AddProvider(NewEnvConfigProvider())
		dotenvProvider := NewDotEnvConfigProvider()

		// 明らかに存在しないパスで読み込みを試みる
		nonExistPath := "/definitely/not/existing/path/to/config/file.env"
		err := dotenvProvider.Load(nonExistPath)
		if err != nil {
			t.Logf("存在しないパスでの読み込み結果: %v", err)
		}

		chain.AddProvider(dotenvProvider)
		configManager := chain

		// 環境変数が正常に機能することを確認
		value, exists := configManager.Get(testKey)
		if !exists {
			t.Error("存在しないパスの状態で環境変数が取得できない")
		} else if value != testValue {
			t.Errorf("存在しないパス状態の環境変数値: 期待=%s, 実際=%s", testValue, value)
		}
	})
}

// パーミッション問題テスト
func TestConfigFilePermissionIssue(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("このテストはUnix系OSでのみ実行されます")
		return
	}

	// テスト用の一時ファイル作成
	tempFile, err := os.CreateTemp("", "perm-test-*")
	if err != nil {
		t.Fatalf("テストファイル作成に失敗: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// 内容書き込み
	if _, err := tempFile.WriteString("TEST_KEY=value"); err != nil {
		t.Fatalf("ファイル書き込みに失敗: %v", err)
	}
	if err := tempFile.Close(); err != nil {
		t.Fatalf("ファイルクローズに失敗: %v", err)
	}

	// 危険なパーミッションに設定
	if err := os.Chmod(tempFile.Name(), 0666); err != nil {
		t.Fatalf("パーミッション変更に失敗: %v", err)
	}

	// ドットエンブプロバイダー直接テスト
	provider := NewDotEnvConfigProvider()
	err = provider.Load(tempFile.Name())

	// エラーが発生することを確認
	if err == nil {
		t.Error("危険なパーミッションのファイルでもエラーが発生しない")
	} else if !strings.Contains(err.Error(), "パーミッション") {
		t.Errorf("期待したパーミッションエラーではない: %v", err)
	}

	// キーが存在しないことを確認
	value, exists := provider.Get("TEST_KEY")
	if exists {
		t.Errorf("パーミッションエラー後にもキーが存在する: %s", value)
	}
}

// ホームディレクトリが利用できない場合のテスト（直接的なアプローチ）
func TestConfigWithNoHomeDir(t *testing.T) {
	// 環境変数HOMEを一時的に削除してテスト
	// これにより一部の環境ではUserHomeDirが失敗するようになる

	homeEnvKey := "HOME"
	if runtime.GOOS == "windows" {
		homeEnvKey = "USERPROFILE"
	}

	// 元の値を保存
	originalHome, homeExists := os.LookupEnv(homeEnvKey)

	// 環境変数を削除
	os.Unsetenv(homeEnvKey)

	// テスト完了後に復元
	defer func() {
		if homeExists {
			os.Setenv(homeEnvKey, originalHome)
		} else {
			os.Unsetenv(homeEnvKey)
		}
	}()

	// 環境変数を設定
	testKey := "TEST_NO_HOME_KEY"
	testValue := "test_value"
	os.Setenv(testKey, testValue)
	defer os.Unsetenv(testKey)

	// config.NewConfigManagerの代わりに必要な処理を手動で行う
	chain := NewChainConfigProvider()
	chain.AddProvider(NewEnvConfigProvider())
	dotenvProvider := NewDotEnvConfigProvider()

	// グローバル設定ファイルの処理は省略（ホームディレクトリがないため）

	// ローカル設定ファイルの処理だけ行う
	if err := dotenvProvider.Load(getEnvFileName()); err != nil {
		// エラーは無視（通常の処理と同じ）
	}

	chain.AddProvider(dotenvProvider)
	configManager := chain

	// 環境変数が正常に取得できることを確認
	value, exists := configManager.Get(testKey)
	if !exists {
		t.Error("ホームディレクトリなしの状態で環境変数が取得できない")
	} else if value != testValue {
		t.Errorf("ホームディレクトリなしの状態の環境変数値: 期待=%s, 実際=%s", testValue, value)
	}
}
