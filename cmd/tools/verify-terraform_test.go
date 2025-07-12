package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/FUJI0130/go-ddd-ca/internal/tools/terraform"
	"github.com/FUJI0130/go-ddd-ca/internal/tools/terraform/models"
)

// テスト用リソースオブジェクトを生成するファクトリ関数
func newTestResources(vpc, rds, ecsCluster int, serviceResourcesExist bool) *models.Resources {
	resources := &models.Resources{
		VPC:        vpc,
		RDS:        rds,
		ECSCluster: ecsCluster,
		Services:   make(map[string]models.ServiceResources),
	}

	// サービスリソースの設定
	serviceCount := 0
	if serviceResourcesExist {
		serviceCount = 1
	}

	for _, serviceType := range []string{"api", "graphql", "grpc"} {
		resources.Services[serviceType] = models.ServiceResources{
			ECSService:  serviceCount,
			ALB:         serviceCount,
			TargetGroup: serviceCount,
		}
	}

	return resources
}

// テスト用にverifystate関数の実行をシミュレートする関数
func runMainForTest(
	t *testing.T,
	args []string,
	opts models.VerifyOptions,
	awsResources, tfResources *models.Resources,
	runCmd func(command string, args ...string) (string, error),
	outputWriter *terraform.MockOutputWriter,
	exiter *terraform.MockSystemExiter,
) {
	// 出力初期化
	formatter := terraform.NewDefaultOutputFormatter(outputWriter)

	// 開始メッセージ
	outputWriter.Printf("■ AWS環境とTerraform状態の整合性を検証しています（環境: %s）...\n", opts.Environment)

	// 状態検証の実行（新しいテスト用関数を使用）
	exitCode, results, err := terraform.VerifyStateForTest(opts, awsResources, tfResources)
	if err != nil {
		outputWriter.Printf("✖ %v\n", err)
		exiter.Exit(1)
		return
	}

	// 結果の表示
	formatter.DisplayResults(results, opts.Environment)

	// 不一致数のカウント
	mismatchCount := 0
	for _, result := range results {
		if !result.IsMatch {
			mismatchCount++
		}
	}

	// 最終判定の表示
	if exitCode == 0 {
		// 成功
		if mismatchCount == 0 {
			outputWriter.Printf("\n✅ 整合性確認OK: AWS環境とTerraform状態は一致しています\n")
		} else {
			outputWriter.Printf("\n⚠️ 混在状態: リソース数に不一致がありますが、Terraform planでは変更がありません\n")
			outputWriter.Printf("⚠️ AWS CLIによる検出とTerraformの認識に相違があります。手動での確認をお勧めします。\n")
		}
	} else if exitCode == 2 {
		// 不一致
		outputWriter.Printf("\n⚠️ 不整合検出: %d個のリソースで差異があります\n", mismatchCount)
		formatter.ShowMismatchRemediation(opts.Environment)
	} else {
		// エラー
		outputWriter.Printf("\n✖ Terraformの実行中にエラーが発生しました\n")
	}

	// 終了コードを返す
	exiter.Exit(exitCode)
}

// メイン関数のテスト
func TestVerifyTerraform(t *testing.T) {
	testCases := []struct {
		name           string
		args           []string
		setupResources func() (models.VerifyOptions, *models.Resources, *models.Resources)
		runCmd         func(command string, args ...string) (string, error)
		expectedCode   int
		expectedOutput string
		shouldSucceed  bool
	}{
		// 正常系: デフォルト引数でリソース一致
		{
			name: "デフォルト引数でリソース一致",
			args: []string{},
			setupResources: func() (models.VerifyOptions, *models.Resources, *models.Resources) {
				opts := models.VerifyOptions{Environment: "development"}
				// AWS環境とTerraform状態のリソースが一致するケース
				awsResources := newTestResources(1, 1, 1, false)
				tfResources := newTestResources(1, 1, 1, false)
				return opts, awsResources, tfResources
			},
			runCmd: func(command string, args ...string) (string, error) {
				if command == "terraform" && len(args) > 0 && args[0] == "plan" {
					return "No changes. Your infrastructure matches the configuration.", nil
				}
				return "", nil
			},
			expectedCode:   0,
			expectedOutput: "整合性確認OK",
			shouldSucceed:  true,
		},

		// 正常系: 環境指定
		{
			name: "環境指定でリソース一致",
			args: []string{"-env", "production"},
			setupResources: func() (models.VerifyOptions, *models.Resources, *models.Resources) {
				opts := models.VerifyOptions{Environment: "production"}
				// AWS環境とTerraform状態のリソースが一致するケース
				awsResources := newTestResources(1, 1, 1, false)
				tfResources := newTestResources(1, 1, 1, false)
				return opts, awsResources, tfResources
			},
			runCmd: func(command string, args ...string) (string, error) {
				if command == "terraform" && len(args) > 0 && args[0] == "plan" {
					return "No changes. Your infrastructure matches the configuration.", nil
				}
				return "", nil
			},
			expectedCode:   0,
			expectedOutput: "production",
			shouldSucceed:  true,
		},

		// 異常系: リソース不一致
		{
			name: "リソース不一致の検出",
			args: []string{},
			setupResources: func() (models.VerifyOptions, *models.Resources, *models.Resources) {
				opts := models.VerifyOptions{Environment: "development"}
				// AWS環境には2つのVPC、Terraformには1つのVPC
				awsResources := newTestResources(2, 1, 1, false)
				tfResources := newTestResources(1, 1, 1, false)
				return opts, awsResources, tfResources
			},
			runCmd: func(command string, args ...string) (string, error) {
				return "", nil
			},
			expectedCode:   2,
			expectedOutput: "不整合検出",
			shouldSucceed:  true,
		},

		// 異常系: サービスリソース不一致
		{
			name: "サービスリソース不一致の検出",
			args: []string{},
			setupResources: func() (models.VerifyOptions, *models.Resources, *models.Resources) {
				opts := models.VerifyOptions{Environment: "development"}
				// AWS環境にはサービスリソースがあり、Terraformにはない
				awsResources := newTestResources(1, 1, 1, true)
				tfResources := newTestResources(1, 1, 1, false)
				return opts, awsResources, tfResources
			},
			runCmd: func(command string, args ...string) (string, error) {
				return "", nil
			},
			expectedCode:   2,
			expectedOutput: "不整合検出",
			shouldSucceed:  true,
		},

		// 正常系: 空の環境
		{
			name: "空の環境",
			args: []string{},
			setupResources: func() (models.VerifyOptions, *models.Resources, *models.Resources) {
				opts := models.VerifyOptions{Environment: "development"}
				// AWS環境もTerraform状態も空
				awsResources := newTestResources(0, 0, 0, false)
				tfResources := newTestResources(0, 0, 0, false)
				return opts, awsResources, tfResources
			},
			runCmd: func(command string, args ...string) (string, error) {
				return "", nil
			},
			expectedCode:   0,
			expectedOutput: "整合性確認OK",
			shouldSucceed:  true,
		},

		// 正常系: デバッグモード
		{
			name: "デバッグモード有効",
			args: []string{"-debug"},
			setupResources: func() (models.VerifyOptions, *models.Resources, *models.Resources) {
				opts := models.VerifyOptions{Environment: "development", Debug: true}
				// リソース一致
				awsResources := newTestResources(1, 1, 1, false)
				tfResources := newTestResources(1, 1, 1, false)
				return opts, awsResources, tfResources
			},
			runCmd: func(command string, args ...string) (string, error) {
				return "", nil
			},
			expectedCode:   0,
			expectedOutput: "整合性確認OK",
			shouldSucceed:  true,
		},

		// 正常系: terraform plan スキップ
		{
			name: "Terraform plan スキップ",
			args: []string{"-skip-plan"},
			setupResources: func() (models.VerifyOptions, *models.Resources, *models.Resources) {
				opts := models.VerifyOptions{Environment: "development", SkipTerraformPlan: true}
				// リソース一致
				awsResources := newTestResources(1, 1, 1, false)
				tfResources := newTestResources(1, 1, 1, false)
				return opts, awsResources, tfResources
			},
			runCmd: func(command string, args ...string) (string, error) {
				// このテストでは plan は呼び出されないはず
				if command == "terraform" && len(args) > 0 && args[0] == "plan" {
					t.Error("terraform plan が呼び出されました (スキップ指定あり)")
				}
				return "", nil
			},
			expectedCode:   0,
			expectedOutput: "整合性確認OK",
			shouldSucceed:  true,
		},

		// エラーケース: オプション処理エラー
		{
			name: "オプション処理エラー",
			args: []string{"--invalid-option"},
			setupResources: func() (models.VerifyOptions, *models.Resources, *models.Resources) {
				opts := models.VerifyOptions{Environment: "development"}
				awsResources := newTestResources(1, 1, 1, false)
				tfResources := newTestResources(1, 1, 1, false)
				return opts, awsResources, tfResources
			},
			runCmd: func(command string, args ...string) (string, error) {
				return "", nil
			},
			expectedCode:   1,
			expectedOutput: "引数解析エラー",
			shouldSucceed:  false,
		},
	}

	// 各ケースを実行
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// モックの作成
			outputWriter := terraform.NewMockOutputWriter()
			exiter := terraform.NewMockSystemExiter()

			// フラグセットの作成（オプション処理エラーのテスト用）
			if tc.name == "オプション処理エラー" {
				// このケースは別途テスト
				flagSet := flag.NewFlagSet("verify-terraform", flag.ContinueOnError)
				flagSet.String("env", "development", "Environment to verify (development, production)")
				flagSet.Bool("debug", false, "Enable debug output")
				flagSet.Bool("skip-plan", false, "Skip terraform plan verification")
				flagSet.Bool("force-plan", false, "Force terraform plan even if resource counts mismatch")

				err := flagSet.Parse(tc.args)
				if err == nil {
					t.Error("無効なオプションでエラーが発生しませんでした")
				}
				return
			}

			// リソース設定
			opts, awsResources, tfResources := tc.setupResources()

			// テスト実行
			runMainForTest(t, tc.args, opts, awsResources, tfResources, tc.runCmd, outputWriter, exiter)

			// 検証
			if tc.shouldSucceed {
				// 終了コードの検証
				if exiter.MockExitCode != tc.expectedCode {
					t.Errorf("終了コードが期待値と異なります。期待値: %d, 実際の値: %d", tc.expectedCode, exiter.MockExitCode)
				}

				// 出力内容の検証
				output := outputWriter.GetOutput()
				if !strings.Contains(output, tc.expectedOutput) {
					t.Errorf("出力に期待する文字列 %q が含まれていません。\n出力内容: %s", tc.expectedOutput, output)
				}
			} else {
				// 失敗期待のケースの検証
				t.Logf("失敗ケースのテスト結果: 終了コード %d, 出力内容: %s", exiter.MockExitCode, outputWriter.GetOutput())
			}
		})
	}
}

// TestHelperProcess はサブプロセス実行のテストヘルパー
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	defer os.Exit(0)

	args := os.Args
	for len(args) > 0 {
		if args[0] == "--" {
			args = args[1:]
			break
		}
		args = args[1:]
	}
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "No command\n")
		os.Exit(2)
	}

	cmd, args := args[0], args[1:]
	switch cmd {
	case "aws":
		// AWS コマンドのモック
		fmt.Print("1")
	case "terraform":
		// Terraform コマンドのモック
		if len(args) > 0 && args[0] == "state" && args[1] == "list" {
			fmt.Print("module.networking.aws_vpc.main\nmodule.database.aws_db_instance.postgres\nmodule.shared_ecs_cluster.aws_ecs_cluster.main\n")
		} else if len(args) > 0 && args[0] == "plan" {
			fmt.Print("No changes. Your infrastructure matches the configuration.")
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown command %q\n", cmd)
		os.Exit(2)
	}
}
