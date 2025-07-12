// config_integration_test.go
package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestConfigIntegration は複数の設定ソースを統合したテスト
func TestConfigIntegration(t *testing.T) {
	testCases := []struct {
		name              string
		envVars           map[string]string
		globalFileContent string
		localFileContent  string
		queryKeys         []string
		expectedValues    map[string]string
		success           bool // 期待するテスト結果（成功/失敗）
	}{
		{
			name: "すべての設定ソースからの取得",
			envVars: map[string]string{
				"ENV_ONLY_KEY":  "env_value",
				"OVERRIDE_KEY":  "env_value", // 環境変数が優先されるべき
			},
			globalFileContent: "GLOBAL_ONLY_KEY=global_value\nGLOBAL_LOCAL_KEY=global_value\n",
			localFileContent:  "LOCAL_ONLY_KEY=local_value\nGLOBAL_LOCAL_KEY=local_value\nOVERRIDE_KEY=file_value\n",
			queryKeys:         []string{"ENV_ONLY_KEY", "GLOBAL_ONLY_KEY", "LOCAL_ONLY_KEY", "GLOBAL_LOCAL_KEY", "OVERRIDE_KEY"},
			expectedValues: map[string]string{
				"ENV_ONLY_KEY":    "env_value",
				"GLOBAL_ONLY_KEY": "global_value",
				"LOCAL_ONLY_KEY":  "local_value",
				"GLOBAL_LOCAL_KEY": "local_value", // ローカルがグローバルより優先
				"OVERRIDE_KEY":    "env_value",   // 環境変数が最優先
			},
			success: true,
		},
		{
			name:    "環境変数のみ（ファイルなし）",
			envVars: map[string]string{
				"ENV_KEY1": "value1",
				"ENV_KEY2": "value2",
			},
			globalFileContent: "",
			localFileContent:  "",
			queryKeys:         []string{"ENV_KEY1", "ENV_KEY2", "NONEXISTENT_KEY"},
			expectedValues: map[string]string{
				"ENV_KEY1": "value1",
				"ENV_KEY2": "value2",
			},
			success: true,
		},
		{
			name:              "ローカルファイルのみ（環境変数なし）",
			envVars:           map[string]string{},
			globalFileContent: "",
			localFileContent:  "FILE_KEY1=value1\nFILE_KEY2=value2\n",
			queryKeys:         []string{"FILE_KEY1", "FILE_KEY2", "NONEXISTENT_KEY"},
			expectedValues: map[string]string{
				"FILE_KEY1": "value1",
				"FILE_KEY2": "value2",
			},
			success: true,
		},
		{
			name:              "グローバルファイルのみ（環境変数・ローカルファイルなし）",
			envVars:           map[string]string{},
			globalFileContent: "GLOBAL_KEY1=value1\nGLOBAL_KEY2=value2\n",
			localFileContent:  "",
			queryKeys:         []string{"GLOBAL_KEY1", "GLOBAL_KEY2", "NONEXISTENT_KEY"},
			expectedValues: map[string]string{
				"GLOBAL_KEY1": "value1",
				"GLOBAL_KEY2": "value2",
			},
			success: true,
		},
		{
			name: "フォーマットの異なる設定値",
			envVars: map[string]string{
				"QUOTED_KEY": "env_quoted",
			},
			globalFileContent: "",
			localFileContent:  "QUOTED_KEY=\"file_quoted\"\nSINGLE_QUOTED_KEY='single_quoted'\nNOQUOTE_KEY=no_quotes\n",
			queryKeys:         []string{"QUOTED_KEY", "SINGLE_QUOTED_KEY", "NOQUOTE_KEY"},
			expectedValues: map[string]string{
				"QUOTED_KEY":       "env_quoted", // 環境変数が優先
				"SINGLE_QUOTED_KEY": "single_quoted",
				"NOQUOTE_KEY":      "no_quotes",
			},
			success: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// テスト前のクリーンアップ（関連する環境変数をすべて削除）
			for key := range tc.envVars {
				os.Unsetenv(key)
			}
			for key := range tc.expectedValues {
				os.Unsetenv(key)
			}

			// テスト用の環境変数を設定
			for key, value := range tc.envVars {
				os.Setenv(key, value)
				defer os.Unsetenv(key) // テスト完了後にクリーンアップ
			}

			// テスト用の一時ファイル作成（グローバル設定）
			var globalConfigPath string
			if tc.globalFileContent != "" {
				globalFile, err := os.CreateTemp("", "test-global-env-*")
				if err != nil {
					t.Fatalf("グローバル設定ファイル作成に失敗: %v", err)
				}
				defer os.Remove(globalFile.Name())
				if _, err := globalFile.WriteString(tc.globalFileContent); err != nil {
					t.Fatalf("ファイル書き込みに失敗: %v", err)
				}
				if err := globalFile.Close(); err != nil {
					t.Fatalf("ファイルクローズに失敗: %v", err)
				}
				globalConfigPath = globalFile.Name()
			}

			// テスト用の一時ファイル作成（ローカル設定）
			var localConfigPath string
			if tc.localFileContent != "" {
				localFile, err := os.CreateTemp("", "test-local-env-*")
				if err != nil {
					t.Fatalf("ローカル設定ファイル作成に失敗: %v", err)
				}
				defer os.Remove(localFile.Name())
				if _, err := localFile.WriteString(tc.localFileContent); err != nil {
					t.Fatalf("ファイル書き込みに失敗: %v", err)
				}
				if err := localFile.Close(); err != nil {
					t.Fatalf("ファイルクローズに失敗: %v", err)
				}
				localConfigPath = localFile.Name()
			}

			// 元の関数を保存
			originalGetEnvFileName := getEnvFileName
			
			// テスト用の実装で上書き
			getEnvFileName = func() string {
				if tc.localFileContent != "" {
					return filepath.Base(localConfigPath)
				}
				return defaultEnvFileName
			}
			
			// テスト完了後に元の関数を復元
			defer func() {
				getEnvFileName = originalGetEnvFileName
			}()

			// 元のディレクトリを保存
			originalDir, err := os.Getwd()
			if err != nil {
				t.Fatalf("カレントディレクトリ取得に失敗: %v", err)
			}

			if tc.localFileContent != "" {
				// ローカル設定ファイルのあるディレクトリに移動
				tempDir := filepath.Dir(localConfigPath)
				if err := os.Chdir(tempDir); err != nil {
					t.Fatalf("カレントディレクトリ変更に失敗: %v", err)
				}
				defer os.Chdir(originalDir) // テスト完了後に元のディレクトリに戻る
			}

			// テスト用に環境変数HOMEを一時的に設定（グローバル設定ファイルのパス変更用）
			homeEnvKey := "HOME"
			if runtime.GOOS == "windows" {
				homeEnvKey = "USERPROFILE"
			}
			
			if tc.globalFileContent != "" {
				originalHome, homeExists := os.LookupEnv(homeEnvKey)
				os.Setenv(homeEnvKey, filepath.Dir(globalConfigPath))
				
				// テスト完了後に元のHOME環境変数を復元
				defer func() {
					if homeExists {
						os.Setenv(homeEnvKey, originalHome)
					} else {
						os.Unsetenv(homeEnvKey)
					}
				}()
			}

			// 設定マネージャーの初期化
			configManager, err := NewConfigManager()
			if err != nil {
				t.Fatalf("設定マネージャー作成に失敗: %v", err)
			}

			// 各キーの値を検証
			for _, key := range tc.queryKeys {
				value, exists := configManager.Get(key)
				expectedValue, shouldExist := tc.expectedValues[key]
				
				if shouldExist != exists {
					t.Errorf("キー %s の存在状態: 期待=%v, 実際=%v", key, shouldExist, exists)
				}
				
				if shouldExist && value != expectedValue {
					t.Errorf("キー %s の値: 期待=%s, 実際=%s", key, expectedValue, value)
				}
			}

			// 成功判定（すべてのチェックに合格したか）
			testSuccess := true
			for _, key := range tc.queryKeys {
				expectedValue, shouldExist := tc.expectedValues[key]
				value, exists := configManager.Get(key)
				
				if shouldExist != exists || (shouldExist && value != expectedValue) {
					testSuccess = false
					break
				}
			}
			
			if testSuccess != tc.success {
				t.Errorf("テスト成功判定: 期待=%v, 実際=%v", tc.success, testSuccess)
			}
		})
	}
}

// TestAWSCredentialsIntegration はAWS認証情報チェック機能の統合テスト
func TestAWSCredentialsIntegration(t *testing.T) {
	testCases := []struct {
		name        string
		envVars     map[string]string
		fileContent string
		expectError bool
		success     bool // 期待するテスト結果（成功/失敗）
	}{
		{
			name: "すべての認証情報あり（環境変数）",
			envVars: map[string]string{
				"AWS_ACCESS_KEY_ID":     "test_id",
				"AWS_SECRET_ACCESS_KEY": "test_secret",
				"AWS_REGION":            "us-west-2",
			},
			fileContent: "",
			expectError: false,
			success:     true,
		},
		{
			name: "すべての認証情報あり（ファイル）",
			envVars: map[string]string{},
			fileContent: "AWS_ACCESS_KEY_ID=test_id\nAWS_SECRET_ACCESS_KEY=test_secret\nAWS_REGION=us-west-2\n",
			expectError: false,
			success:     true,
		},
		{
			name: "一部の認証情報あり（環境変数+ファイル）",
			envVars: map[string]string{
				"AWS_ACCESS_KEY_ID": "test_id",
			},
			fileContent: "AWS_SECRET_ACCESS_KEY=test_secret\nAWS_REGION=us-west-2\n",
			expectError: false,
			success:     true,
		},
		{
			name: "一部の認証情報なし（どちらにもないケース）",
			envVars: map[string]string{
				"AWS_ACCESS_KEY_ID": "test_id",
			},
			fileContent: "AWS_REGION=us-west-2\n",
			expectError: true,
			success:     true,
		},
		{
			name:        "すべての認証情報なし",
			envVars:     map[string]string{},
			fileContent: "",
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

			// テスト用の一時ファイル作成
			var configPath string
			if tc.fileContent != "" {
				tempFile, err := os.CreateTemp("", "test-aws-cred-*")
				if err != nil {
					t.Fatalf("テストファイル作成に失敗: %v", err)
				}
				defer os.Remove(tempFile.Name())
				if _, err := tempFile.WriteString(tc.fileContent); err != nil {
					t.Fatalf("ファイル書き込みに失敗: %v", err)
				}
				if err := tempFile.Close(); err != nil {
					t.Fatalf("ファイルクローズに失敗: %v", err)
				}
				configPath = tempFile.Name()
			}

			// 元の関数を保存
			originalGetEnvFileName := getEnvFileName
			
			// テスト用の実装で上書き
			getEnvFileName = func() string {
				if tc.fileContent != "" {
					return filepath.Base(configPath)
				}
				return defaultEnvFileName
			}
			
			// テスト完了後に元の関数を復元
			defer func() {
				getEnvFileName = originalGetEnvFileName
			}()

			// 元のディレクトリを保存
			originalDir, err := os.Getwd()
			if err != nil {
				t.Fatalf("カレントディレクトリ取得に失敗: %v", err)
			}

			if tc.fileContent != "" {
				// 設定ファイルのあるディレクトリに移動
				tempDir := filepath.Dir(configPath)
				if err := os.Chdir(tempDir); err != nil {
					t.Fatalf("カレントディレクトリ変更に失敗: %v", err)
				}
				defer os.Chdir(originalDir) // テスト完了後に元のディレクトリに戻る
			}

			// 設定マネージャーの初期化と認証情報チェック
			configManager, err := NewConfigManager()
			if err != nil {
				t.Fatalf("設定マネージャー作成に失敗: %v", err)
			}
			
			err = CheckRequiredVariables(configManager)

			// エラー有無の検証
			hasError := err != nil
			if hasError != tc.expectError {
				t.Errorf("エラー発生: 期待=%v, 実際=%v", tc.expectError, hasError)
				if err != nil {
					t.Logf("エラーメッセージ: %v", err)
				}
			}

			// エラーメッセージの内容検証（エラーが期待される場合）
			if tc.expectError && err != nil {
				// 必須変数名がすべてエラーメッセージに含まれているか
				requiredVars := []string{
					"AWS_ACCESS_KEY_ID",
					"AWS_SECRET_ACCESS_KEY",
					"AWS_REGION",
				}
				
				for _, key := range requiredVars {
					if !hasVar(tc.envVars, key) && !strings.Contains(tc.fileContent, key) {
						if !strings.Contains(err.Error(), key) {
							t.Errorf("エラーメッセージに必須変数 %s が含まれていない: %s", key, err.Error())
						}
					}
				}
			}

			// 成功/失敗の検証
			testSuccess := hasError == tc.expectError
			if testSuccess != tc.success {
				t.Errorf("テスト成功判定: 期待=%v, 実際=%v", tc.success, testSuccess)
			}
		})
	}
}

// hasVar はマップにキーが存在するか確認するヘルパー関数
func hasVar(vars map[string]string, key string) bool {
	_, exists := vars[key]
	return exists
}

// TestRealWorldConfigScenarios は実際のユースケースに近いシナリオをテスト
func TestRealWorldConfigScenarios(t *testing.T) {
	testCases := []struct {
		name            string
		setupFunc       func(t *testing.T) (cleanup func())
		verifyFunc      func(t *testing.T, config ConfigProvider) bool
		success         bool // 期待するテスト結果（成功/失敗）
	}{
		{
			name: "環境変数優先の複合設定",
			setupFunc: func(t *testing.T) func() {
				// 複数の設定ソースをセットアップ（環境変数、グローバル、ローカル）
				os.Setenv("AWS_REGION", "env-region")
				
				globalFile, err := os.CreateTemp("", "global-env-*")
				if err != nil {
					t.Fatalf("グローバル設定ファイル作成に失敗: %v", err)
				}
				if _, err := globalFile.WriteString("AWS_ACCESS_KEY_ID=global-key\nAWS_SECRET_ACCESS_KEY=global-secret\nAWS_REGION=global-region\n"); err != nil {
					t.Fatalf("ファイル書き込みに失敗: %v", err)
				}
				globalFile.Close()
				
				localFile, err := os.CreateTemp("", "local-env-*")
				if err != nil {
					t.Fatalf("ローカル設定ファイル作成に失敗: %v", err)
				}
				if _, err := localFile.WriteString("AWS_SECRET_ACCESS_KEY=local-secret\n"); err != nil {
					t.Fatalf("ファイル書き込みに失敗: %v", err)
				}
				localFile.Close()
				
				// ディレクトリ操作の保存
				origDir, _ := os.Getwd()
				os.Chdir(filepath.Dir(localFile.Name()))
				
				// ホームディレクトリのモック設定
				homeEnvKey := "HOME"
				if runtime.GOOS == "windows" {
					homeEnvKey = "USERPROFILE"
				}
				origHome, homeExists := os.LookupEnv(homeEnvKey)
				os.Setenv(homeEnvKey, filepath.Dir(globalFile.Name()))
				
				// getEnvFileName関数のモック
				origGetEnvFileName := getEnvFileName
				getEnvFileName = func() string {
					return filepath.Base(localFile.Name())
				}
				
				// クリーンアップ関数を返す
				return func() {
					os.Unsetenv("AWS_REGION")
					os.Remove(globalFile.Name())
					os.Remove(localFile.Name())
					os.Chdir(origDir)
					
					if homeExists {
						os.Setenv(homeEnvKey, origHome)
					} else {
						os.Unsetenv(homeEnvKey)
					}
					
					getEnvFileName = origGetEnvFileName
				}
			},
			verifyFunc: func(t *testing.T, config ConfigProvider) bool {
				// 優先順位に従って正しい値が取得できるか検証
				keyID, _ := config.Get("AWS_ACCESS_KEY_ID")
				secret, _ := config.Get("AWS_SECRET_ACCESS_KEY")
				region, _ := config.Get("AWS_REGION")
				
				return keyID == "global-key" && // グローバル設定からのみ
					   secret == "local-secret" && // ローカル設定が優先
					   region == "env-region" // 環境変数が最優先
			},
			success: true,
		},
		{
			name: "空設定ファイルでのエラー処理",
			setupFunc: func(t *testing.T) func() {
				// 空の設定ファイルを作成
				emptyFile, err := os.CreateTemp("", "empty-env-*")
				if err != nil {
					t.Fatalf("空の設定ファイル作成に失敗: %v", err)
				}
				emptyFile.Close()
				
				// ディレクトリ操作の保存
				origDir, _ := os.Getwd()
				os.Chdir(filepath.Dir(emptyFile.Name()))
				
				// getEnvFileName関数のモック
				origGetEnvFileName := getEnvFileName
				getEnvFileName = func() string {
					return filepath.Base(emptyFile.Name())
				}
				
				// クリーンアップ関数を返す
				return func() {
					os.Remove(emptyFile.Name())
					os.Chdir(origDir)
					getEnvFileName = origGetEnvFileName
				}
			},
			verifyFunc: func(t *testing.T, config ConfigProvider) bool {
				// 空の設定ファイルでもエラーなく動作するか
				_, exists := config.Get("ANY_KEY")
				return !exists // 存在しないキーが正しく「存在しない」と報告されるか
			},
			success: true,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 環境のセットアップ
			cleanup := tc.setupFunc(t)
			defer cleanup()
			
			// 設定マネージャーの初期化
			configManager, err := NewConfigManager()
			if err != nil {
				t.Fatalf("設定マネージャー作成に失敗: %v", err)
			}
			
			// 検証関数の実行
			testResult := tc.verifyFunc(t, configManager)
			
			// 成功/失敗の検証
			if testResult != tc.success {
				t.Errorf("テスト成功判定: 期待=%v, 実際=%v", tc.success, testResult)
			}
		})
	}
}
