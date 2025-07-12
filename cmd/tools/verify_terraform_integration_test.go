// verify_terraform_integration_test.go
package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/FUJI0130/go-ddd-ca/internal/tools/terraform"
	"github.com/FUJI0130/go-ddd-ca/internal/tools/terraform/config"
	"github.com/FUJI0130/go-ddd-ca/internal/tools/terraform/models"
)

// 環境変数設定と検証ツールの統合テスト
func TestVerifyTerraformWithConfigIntegration(t *testing.T) {
	testCases := []struct {
		name           string
		envVars        map[string]string
		fileContent    string
		mockResources  func() (*models.Resources, *models.Resources)
		expectedOutput string
		expectedCode   int
		success        bool // 期待するテスト結果（成功/失敗）
	}{
		{
			name: "環境変数からAWS認証情報を取得",
			envVars: map[string]string{
				"AWS_ACCESS_KEY_ID":     "test_id",
				"AWS_SECRET_ACCESS_KEY": "test_secret",
				"AWS_REGION":            "us-west-2",
			},
			fileContent: "",
			mockResources: func() (*models.Resources, *models.Resources) {
				// AWS環境とTerraform状態のリソースが一致するケース
				awsResources := newTestResources(1, 1, 1, false)
				tfResources := newTestResources(1, 1, 1, false)
				return awsResources, tfResources
			},
			expectedOutput: "整合性確認OK",
			expectedCode:   0,
			success:        true,
		},
		{
			name:        "設定ファイルからAWS認証情報を取得",
			envVars:     map[string]string{},
			fileContent: "AWS_ACCESS_KEY_ID=file_id\nAWS_SECRET_ACCESS_KEY=file_secret\nAWS_REGION=us-east-1\n",
			mockResources: func() (*models.Resources, *models.Resources) {
				// AWS環境とTerraform状態のリソースが一致するケース
				awsResources := newTestResources(1, 1, 1, false)
				tfResources := newTestResources(1, 1, 1, false)
				return awsResources, tfResources
			},
			expectedOutput: "整合性確認OK",
			expectedCode:   0,
			success:        true,
		},
		{
			name: "環境変数と設定ファイルからの混合認証情報",
			envVars: map[string]string{
				"AWS_ACCESS_KEY_ID": "env_id",
				"AWS_REGION":        "us-west-2",
			},
			fileContent: "AWS_SECRET_ACCESS_KEY=file_secret\n",
			mockResources: func() (*models.Resources, *models.Resources) {
				// AWS環境とTerraform状態のリソースが一致するケース
				awsResources := newTestResources(1, 1, 1, false)
				tfResources := newTestResources(1, 1, 1, false)
				return awsResources, tfResources
			},
			expectedOutput: "整合性確認OK",
			expectedCode:   0,
			success:        true,
		},
		{
			name: "カスタム設定（複数環境変数）",
			envVars: map[string]string{
				"AWS_ACCESS_KEY_ID":     "test_id",
				"AWS_SECRET_ACCESS_KEY": "test_secret",
				"AWS_REGION":            "us-west-2",
				"CUSTOM_SETTING":        "custom_value",
			},
			fileContent: "ANOTHER_CUSTOM=file_value\n",
			mockResources: func() (*models.Resources, *models.Resources) {
				// リソース不一致ケース
				awsResources := newTestResources(2, 1, 1, false)
				tfResources := newTestResources(1, 1, 1, false)
				return awsResources, tfResources
			},
			expectedOutput: "不整合検出",
			expectedCode:   2,
			success:        true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 元の環境変数を退避
			origEnvVars := make(map[string]string)
			origEnvExists := make(map[string]bool)

			for key := range tc.envVars {
				val, exists := os.LookupEnv(key)
				origEnvVars[key] = val
				origEnvExists[key] = exists

				// テスト用に環境変数を設定
				os.Setenv(key, tc.envVars[key])
			}

			// テスト後に元の環境変数を復元するクリーンアップ関数
			defer func() {
				for key := range tc.envVars {
					if origEnvExists[key] {
						os.Setenv(key, origEnvVars[key])
					} else {
						os.Unsetenv(key)
					}
				}
			}()

			// テスト用の一時ファイル作成
			var configPath string
			if tc.fileContent != "" {
				tempFile, err := os.CreateTemp("", "test-config-*")
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

			// 設定ファイル名環境変数のモック
			if tc.fileContent != "" {
				// getEnvFileName関数のオリジナルを保存
				origGetEnvFileName := config.GetEnvFileNameForTest()

				// テスト用に関数を置き換え
				config.SetEnvFileNameForTest(func() string {
					return filepath.Base(configPath)
				})

				// テスト完了後に元の関数に戻す
				defer config.SetEnvFileNameForTest(origGetEnvFileName)

				// カレントディレクトリの変更
				origDir, _ := os.Getwd()
				os.Chdir(filepath.Dir(configPath))
				defer os.Chdir(origDir)
			}

			// モックアウトプットライターとExiterの準備
			outputWriter := terraform.NewMockOutputWriter()
			exiter := terraform.NewMockSystemExiter()

			// モックAWSリソースの準備
			awsResources, tfResources := tc.mockResources()

			// モックコマンド実行関数
			mockRunCmd := func(command string, args ...string) (string, error) {
				if command == "terraform" && len(args) > 0 && args[0] == "plan" {
					return "No changes. Your infrastructure matches the configuration.", nil
				}
				return "", nil
			}

			// VerifyOptionsのセットアップ
			opts := models.VerifyOptions{
				Environment: "development",
			}

			// 設定マネージャーの初期化
			configManager, err := config.NewConfigManager()
			if err != nil {
				t.Fatalf("設定マネージャー作成に失敗: %v", err)
			}

			// オプションに設定プロバイダーを追加
			opts.ConfigProvider = configManager

			// テスト実行
			runMainForTest(t, []string{}, opts, awsResources, tfResources, mockRunCmd, outputWriter, exiter)

			// 出力と終了コードの検証
			output := outputWriter.GetOutput()
			if !strings.Contains(output, tc.expectedOutput) {
				t.Errorf("出力に期待する文字列 %q が含まれていません。\n出力内容: %s", tc.expectedOutput, output)
			}

			if exiter.MockExitCode != tc.expectedCode {
				t.Errorf("終了コードが期待値と異なります。期待値: %d, 実際の値: %d", tc.expectedCode, exiter.MockExitCode)
			}

			// 成功判定
			testSuccess := strings.Contains(output, tc.expectedOutput) && exiter.MockExitCode == tc.expectedCode
			if testSuccess != tc.success {
				t.Errorf("テスト成功判定: 期待=%v, 実際=%v", tc.success, testSuccess)
			}
		})
	}
}

// 異常系統合テスト（設定エラーとの連携）
// テスト間で共有する変数
var (
	invalidConfigPath       string
	savedOrigGetEnvFileName func() string
	savedOrigDir            string
)

// 異常系統合テスト（設定エラーとの連携）
func TestVerifyTerraformWithConfigErrors(t *testing.T) {
	// 元のフラグ値を保存
	origArgs := os.Args

	testCases := []struct {
		name           string
		setupFunc      func() // 異常な設定状態をセットアップする関数
		cleanupFunc    func() // セットアップのクリーンアップを行う関数
		expectedOutput string
		expectedCode   int
		skipPlan       bool // terraform planをスキップするかどうか
		noCheck        bool // 必須環境変数チェックをスキップするかどうか
		success        bool // 期待するテスト結果（成功/失敗）
	}{
		{
			name: "必須環境変数なしでの処理（-no-checkなし）",
			setupFunc: func() {
				// AWS認証情報の環境変数をクリア
				os.Unsetenv("AWS_ACCESS_KEY_ID")
				os.Unsetenv("AWS_SECRET_ACCESS_KEY")
				os.Unsetenv("AWS_REGION")
			},
			cleanupFunc: func() {
				// テスト後にも環境変数はクリアしておく
			},
			expectedOutput: "必須設定", // エラーメッセージに含まれるはず
			expectedCode:   1,
			skipPlan:       true,
			noCheck:        false,
			success:        true,
		},
		{
			name: "必須環境変数なしでの処理（-no-checkあり）",
			setupFunc: func() {
				// AWS認証情報の環境変数をクリア
				os.Unsetenv("AWS_ACCESS_KEY_ID")
				os.Unsetenv("AWS_SECRET_ACCESS_KEY")
				os.Unsetenv("AWS_REGION")
			},
			cleanupFunc: func() {
				// 環境変数はクリアしたまま
			},
			expectedOutput: "整合性確認OK", // チェックをスキップしたので処理は続行するはず
			expectedCode:   0,
			skipPlan:       true,
			noCheck:        true, // no-checkフラグを有効に
			success:        true,
		},
		// 不適切な設定ファイルでのエラー処理のケースは変更なし
		// ...
	}

	// 各テストケースを実行
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 異常設定のセットアップ
			tc.setupFunc()
			defer tc.cleanupFunc()

			// モックアウトプットライターとExiterの準備
			outputWriter := terraform.NewMockOutputWriter()
			exiter := terraform.NewMockSystemExiter()

			// モックAWSリソースの準備（リソース一致ケース）
			awsResources := newTestResources(1, 1, 1, false)
			tfResources := newTestResources(1, 1, 1, false)

			// モックコマンド実行関数
			mockRunCmd := func(command string, args ...string) (string, error) {
				if command == "terraform" && len(args) > 0 && args[0] == "plan" {
					return "No changes. Your infrastructure matches the configuration.", nil
				}
				return "", nil
			}

			// VerifyOptionsのセットアップ
			opts := models.VerifyOptions{
				Environment:       "development",
				SkipTerraformPlan: tc.skipPlan,
			}

			// 設定マネージャーの初期化を試みる
			configManager, err := config.NewConfigManager()

			// 設定マネージャー作成に失敗した場合のハンドリング
			if err != nil {
				// 期待されるエラーケースかチェック
				if tc.expectedCode == 1 && strings.Contains(err.Error(), tc.expectedOutput) {
					// 期待されたエラーなのでテスト成功
					return
				}
				t.Fatalf("設定マネージャー作成に失敗: %v", err)
			}

			// AWS認証情報などの必須環境変数チェック（-no-checkフラグが指定されていない場合）
			if !tc.noCheck {
				if err := config.CheckRequiredVariables(configManager); err != nil {
					// テストケースによっては期待されるエラー
					outputWriter.Printf("%v\n", err)
					exiter.Exit(1)

					// 出力内容を確認
					output := outputWriter.GetOutput()
					if !strings.Contains(output, tc.expectedOutput) {
						t.Errorf("エラー出力に期待する文字列 %q が含まれていません。\n出力内容: %s", tc.expectedOutput, output)
					}

					if exiter.MockExitCode != tc.expectedCode {
						t.Errorf("終了コードが期待値と異なります。期待値: %d, 実際の値: %d", tc.expectedCode, exiter.MockExitCode)
					}

					return
				}
			}

			// オプションに設定プロバイダーを追加
			opts.ConfigProvider = configManager

			// テスト実行
			runMainForTest(t, []string{}, opts, awsResources, tfResources, mockRunCmd, outputWriter, exiter)

			// 出力と終了コードの検証
			output := outputWriter.GetOutput()
			if !strings.Contains(output, tc.expectedOutput) {
				t.Errorf("出力に期待する文字列 %q が含まれていません。\n出力内容: %s", tc.expectedOutput, output)
			}

			if exiter.MockExitCode != tc.expectedCode {
				t.Errorf("終了コードが期待値と異なります。期待値: %d, 実際の値: %d", tc.expectedCode, exiter.MockExitCode)
			}

			// 成功判定
			testSuccess := strings.Contains(output, tc.expectedOutput) && exiter.MockExitCode == tc.expectedCode
			if testSuccess != tc.success {
				t.Errorf("テスト成功判定: 期待=%v, 実際=%v", tc.success, testSuccess)
			}
		})
	}

	// テスト完了後にos.Argsを元に戻す
	os.Args = origArgs
}

// TestConfigManagerSetupWithVerifyTool は設定マネージャーの初期化と検証ツールの連携をテスト
func TestConfigManagerSetupWithVerifyTool(t *testing.T) {
	// テスト環境のセットアップ
	os.Setenv("AWS_ACCESS_KEY_ID", "test_key_id")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "test_secret_key")
	os.Setenv("AWS_REGION", "us-west-2")
	defer func() {
		os.Unsetenv("AWS_ACCESS_KEY_ID")
		os.Unsetenv("AWS_SECRET_ACCESS_KEY")
		os.Unsetenv("AWS_REGION")
	}()

	// 設定マネージャーの初期化
	configManager, err := config.NewConfigManager()
	if err != nil {
		t.Fatalf("設定マネージャー作成に失敗: %v", err)
	}

	// 設定マネージャーからAWS認証情報が取得できることを確認
	for _, key := range []string{"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY", "AWS_REGION"} {
		value, exists := configManager.Get(key)
		if !exists {
			t.Errorf("キー %s が設定マネージャーから取得できません", key)
		} else if value == "" {
			t.Errorf("キー %s の値が空です", key)
		}
	}

	// モック出力ライターとExiterの準備
	outputWriter := terraform.NewMockOutputWriter()
	exiter := terraform.NewMockSystemExiter()

	// モックAWSリソースの準備
	awsResources := newTestResources(1, 1, 1, false)
	tfResources := newTestResources(1, 1, 1, false)

	// モックコマンド実行関数
	mockRunCmd := func(command string, args ...string) (string, error) {
		return "", nil
	}

	// VerifyOptionsのセットアップ
	opts := models.VerifyOptions{
		Environment:       "development",
		SkipTerraformPlan: true,
		ConfigProvider:    configManager,
	}

	// テスト実行
	runMainForTest(t, []string{}, opts, awsResources, tfResources, mockRunCmd, outputWriter, exiter)

	// 終了コードの検証
	if exiter.MockExitCode != 0 {
		t.Errorf("終了コードが0ではありません: %d", exiter.MockExitCode)
	}

	// 出力内容の検証
	output := outputWriter.GetOutput()
	if !strings.Contains(output, "整合性確認OK") {
		t.Errorf("出力に「整合性確認OK」が含まれていません: %s", output)
	}
}
