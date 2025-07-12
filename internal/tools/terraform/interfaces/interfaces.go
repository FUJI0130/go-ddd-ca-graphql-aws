package interfaces

import (
	"context"

	"github.com/FUJI0130/go-ddd-ca/internal/tools/terraform/models"
)

// AWSCommandRunner はAWS CLIコマンド実行のインターフェース
type AWSCommandRunner interface {
	// RunCommand はAWS CLIコマンドを実行し、標準出力を文字列として返します
	RunCommand(args ...string) (string, error)
}

// CommandExecutor はコマンドを実行するインターフェース
type CommandExecutor interface {
	// Execute は指定されたコマンドを実行し、標準出力を文字列として返します
	Execute(command string, args ...string) (string, error)
}

// ContextAwareCommandExecutor はコンテキスト対応のコマンド実行インターフェース
type ContextAwareCommandExecutor interface {
	CommandExecutor // 既存インターフェースを埋め込む

	// ExecuteWithContext はコンテキスト付きでコマンドを実行します
	// コンテキストがキャンセルされた場合、プロセスは中断されます
	ExecuteWithContext(ctx context.Context, command string, args ...string) (string, error)
}

// FileSystem はファイルシステム操作を抽象化するインターフェース
type FileSystem interface {
	// Getwd は現在の作業ディレクトリを取得します
	Getwd() (string, error)
	// Chdir はディレクトリを変更します
	Chdir(dir string) error
}

// OutputFormatter は出力整形を担当するインターフェース
type OutputFormatter interface {
	// FormatComparisonTable は比較結果をテーブル形式でフォーマットします
	FormatComparisonTable(results []models.ComparisonResult) string
	// DisplayResults は比較結果を表示します
	DisplayResults(results []models.ComparisonResult, env string)
	// ShowMismatchRemediation はリソース不一致時の修復オプションを表示します
	ShowMismatchRemediation(env string)
	// PrintDebugInfo はデバッグ情報を表示します
	PrintDebugInfo(message string, args ...interface{})
	// PrintError はエラーメッセージを表示します
	PrintError(message string, args ...interface{})
	// PrintWarning は警告メッセージを表示します
	PrintWarning(message string, args ...interface{})
	// PrintSuccess は成功メッセージを表示します
	PrintSuccess(message string, args ...interface{})
	// PrintInfo は情報メッセージを表示します
	PrintInfo(message string, args ...interface{})
}

// ResourceVerifier はリソース検証を担当するインターフェース
type ResourceVerifier interface {
	// VerifyState はAWS環境とTerraform状態の整合性を検証します
	VerifyState(opts models.VerifyOptions) (int, []models.ComparisonResult, error)
	// GetAWSResources はAWS環境からリソース情報を取得します
	GetAWSResources(env string) (*models.Resources, error)
	// GetTerraformResources はTerraform状態からリソース情報を取得します
	GetTerraformResources(env string) (*models.Resources, error)
	// CompareResources はAWS環境とTerraform状態のリソース数を比較します
	CompareResources(awsResources, tfResources *models.Resources) []models.ComparisonResult
	// RunTerraformPlan はterraform planを実行して差分を検証します
	RunTerraformPlan(env string) (int, string, error)
}

// ResultsDisplayer は結果表示を担当するインターフェース
type ResultsDisplayer interface {
	// DisplayResults は比較結果を表示します
	DisplayResults(results []models.ComparisonResult, env string)
	// ShowMismatchRemediation はリソース不一致時の修復オプションを表示します
	ShowMismatchRemediation(env string)
}

// SystemExiter はシステム終了を担当するインターフェース
type SystemExiter interface {
	// Exit はプログラムを指定のコードで終了します
	Exit(code int)
}

// OutputWriter は出力操作を抽象化するインターフェース
type OutputWriter interface {
	Printf(format string, args ...interface{})
	Println(args ...interface{})
}
