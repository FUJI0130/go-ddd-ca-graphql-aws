package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/FUJI0130/go-ddd-ca/internal/tools/terraform"
	"github.com/FUJI0130/go-ddd-ca/internal/tools/terraform/config"
	"github.com/FUJI0130/go-ddd-ca/internal/tools/terraform/implementation"
	"github.com/FUJI0130/go-ddd-ca/internal/tools/terraform/interfaces"
	"github.com/FUJI0130/go-ddd-ca/internal/tools/terraform/logger"
	"github.com/FUJI0130/go-ddd-ca/internal/tools/terraform/models"
	"github.com/fatih/color"
)

// バージョン情報（ビルド時に上書きされる）
var (
	version   = "dev"     // ビルド時に -ldflags で上書き
	buildTime = "unknown" // ビルド時に自動設定
	gitCommit = "unknown" // ビルド時に自動設定
)

func main() {
	// コマンドライン引数の解析
	envFlag := flag.String("env", "development", "Environment to verify (development, production)")
	debugFlag := flag.Bool("debug", false, "Enable debug output (deprecated, use --log-level=debug instead)")
	logLevelFlag := flag.String("log-level", "", "Log level (error, warn, info, debug, trace)")
	skipPlanFlag := flag.Bool("skip-plan", false, "Skip terraform plan verification")
	forcePlanFlag := flag.Bool("force-plan", false, "Force terraform plan even if resource counts mismatch")
	noCheckFlag := flag.Bool("no-check", false, "Skip required environment variables check")
	logFileFlag := flag.String("log-file", "", "Log file path (defaults to stdout)")
	versionFlag := flag.Bool("version", false, "Show version information")
	ignoreResourceErrorsFlag := flag.Bool("ignore-resource-errors", false, "リソースが存在しない場合のエラーを無視")

	// 追加: タイムアウトオプション
	timeoutFlag := flag.Int("timeout", 60, "Timeout for terraform plan execution in seconds")

	flag.Parse()

	// バージョン表示
	if *versionFlag {
		fmt.Printf("AWS Terraform検証ツール v%s\n", version)
		fmt.Printf("ビルド時間: %s\n", buildTime)
		fmt.Printf("GitCommit: %s\n", gitCommit)
		os.Exit(0)
	}

	// ログレベルの初期化
	logLevel := logger.InfoLevel // デフォルトはINFO

	// --debug フラグの後方互換性
	if *debugFlag {
		logLevel = logger.DebugLevel
	}

	// --log-level フラグが指定されていれば優先
	if *logLevelFlag != "" {
		logLevel = logger.StringToLogLevel(*logLevelFlag)
	}

	// ロガーレベルの設定
	logger.SetLevel(logLevel)

	// ログファイルの設定
	if *logFileFlag != "" {
		logFile, err := os.OpenFile(*logFileFlag, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ログファイルを開けませんでした: %v\n", err)
			os.Exit(1)
		}
		defer logFile.Close()
		logger.SetOutput(logFile)
		logger.Info("ログ出力をファイルに設定: %s", *logFileFlag)
	}

	logger.Info("AWS Terraform検証ツールを起動します")
	logger.Debug("コマンドライン引数: env=%s, debug=%v, log-level=%s, skip-plan=%v, force-plan=%v, no-check=%v, timeout=%ds",
		*envFlag, *debugFlag, *logLevelFlag, *skipPlanFlag, *forcePlanFlag, *noCheckFlag, *timeoutFlag)

	// 設定マネージャーの初期化
	configManager, err := config.NewConfigManager()
	if err != nil {
		logger.Error("設定の初期化に失敗: %v", err)
		fmt.Fprintf(os.Stderr, "設定の初期化に失敗しました: %v\n", err)
		os.Exit(1)
	}
	logger.Debug("設定マネージャーの初期化に成功")

	// AWS認証情報などの必須環境変数チェック（-no-checkフラグが指定されていない場合）
	if !*noCheckFlag {
		logger.Debug("必須環境変数チェックを実行")
		if err := config.CheckRequiredVariables(configManager); err != nil {
			logger.Error("必須環境変数チェックに失敗: %v", err)
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		logger.Debug("必須環境変数チェックに成功")
	} else {
		logger.Warn("必須環境変数チェックはスキップされました")
	}

	suffixFlag := flag.String("suffix", "", "Service name suffix (e.g. -new)")

	// 検証オプションの設定
	opts := models.VerifyOptions{
		Environment:          *envFlag,
		Debug:                *debugFlag,
		SkipTerraformPlan:    *skipPlanFlag,
		ForceCleanup:         *forcePlanFlag,
		ConfigProvider:       configManager,
		LogLevel:             logLevel,
		Timeout:              time.Duration(*timeoutFlag) * time.Second, // 追加: タイムアウト設定
		IgnoreResourceErrors: *ignoreResourceErrorsFlag,                 // 追加
		ServiceSuffix:        *suffixFlag,
	}
	logger.Debug("検証オプション: %+v", opts)

	// 実装コンポーネントの初期化
	awsRunner := implementation.NewDefaultAWSRunner()
	fileSystem := implementation.NewDefaultFileSystem()
	cmdExecutor := implementation.NewContextAwareCommandExecutor() // 変更: コンテキスト対応版を使用
	outputWriter := implementation.NewDefaultOutputWriter()
	formatter := terraform.NewDefaultOutputFormatter(outputWriter)
	exiter := implementation.NewDefaultSystemExiter()
	logger.Debug("実装コンポーネントの初期化に成功")

	// 追加: タイムアウト付きコンテキストの作成
	ctx, cancel := context.WithTimeout(context.Background(), opts.Timeout)
	defer cancel()

	// 色付き出力の初期化
	color.NoColor = false

	fmt.Printf("%s AWS環境とTerraform状態の整合性を検証しています（環境: %s）...\n",
		color.BlueString("■"),
		color.BlueString(*envFlag))

	// 状態検証の実行
	logger.Info("状態検証を実行: 環境=%s, タイムアウト=%v", *envFlag, opts.Timeout)
	exitCode, results, err := terraform.VerifyStateWithContext(ctx, opts, awsRunner, fileSystem, cmdExecutor.(interfaces.ContextAwareCommandExecutor))

	// タイムアウトエラーの特別処理
	if ctx.Err() == context.DeadlineExceeded {
		logger.Error("検証処理がタイムアウトしました: タイムアウト=%v", opts.Timeout)
		fmt.Fprintf(os.Stderr, "%s タイムアウトが発生しました（%d秒）\n",
			color.RedString("✖"),
			*timeoutFlag)
		fmt.Fprintf(os.Stderr, "推奨アクション:\n")
		fmt.Fprintf(os.Stderr, "1. タイムアウト時間を増やす: --timeout %d\n", *timeoutFlag*2)
		fmt.Fprintf(os.Stderr, "2. terraform planをスキップする: --skip-plan\n")
		fmt.Fprintf(os.Stderr, "3. 手動でコマンドを実行: cd deployments/terraform/environments/%s && terraform plan\n", *envFlag)
		exiter.Exit(1)
	}

	if err != nil {
		logger.Error("検証エラー: %v", err)
		fmt.Fprintf(os.Stderr, "%s %v\n", color.RedString("✖"), err)
		exiter.Exit(1)
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

	logger.Debug("不一致リソース数: %d", mismatchCount)

	// 最終判定の表示
	if exitCode == 0 {
		// 成功
		if mismatchCount == 0 {
			logger.Info("検証結果: 整合性確認OK - AWS環境とTerraform状態は一致")
			fmt.Printf("\n%s 整合性確認OK: AWS環境とTerraform状態は一致しています\n", color.GreenString("✅"))
		} else {
			logger.Warn("検証結果: 混在状態 - リソース数に不一致がありますが、Terraform planでは変更がありません")
			fmt.Printf("\n%s 混在状態: リソース数に不一致がありますが、Terraform planでは変更がありません\n", color.YellowString("⚠️"))
			fmt.Printf("%s AWS CLIによる検出とTerraformの認識に相違があります。手動での確認をお勧めします。\n", color.YellowString("⚠️"))
		}
	} else if exitCode == 2 {
		// 不一致
		logger.Warn("検証結果: 不整合検出 - %d個のリソースで差異があります", mismatchCount)
		fmt.Printf("\n%s 不整合検出: %d個のリソースで差異があります\n", color.YellowString("⚠️"), mismatchCount)
		formatter.ShowMismatchRemediation(opts.Environment)
	} else {
		// エラー
		logger.Error("検証結果: Terraformの実行中にエラーが発生")
		fmt.Printf("\n%s Terraformの実行中にエラーが発生しました\n", color.RedString("✖"))
	}

	// 終了コードを返す
	logger.Info("AWS Terraform検証ツールを終了: 終了コード=%d", exitCode)
	exiter.Exit(exitCode)
}
