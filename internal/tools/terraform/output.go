package terraform

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/FUJI0130/go-ddd-ca/internal/tools/terraform/implementation"
	"github.com/FUJI0130/go-ddd-ca/internal/tools/terraform/interfaces"
	"github.com/FUJI0130/go-ddd-ca/internal/tools/terraform/models"
)

// テキストメッセージの定数
const (
	// ヘッダーと区切り
	TableHeader        = "リソース\t\tAWS\tTerraform\t状態"
	HeaderSeparator    = "--------------------------------"
	ResultHeaderFormat = "\n■ 検証結果: %s\n"

	// 状態メッセージ
	StatusMatch    = "一致"
	StatusMismatch = "不一致"

	// 修復オプション
	RemediationHeader   = "\n■ 修復オプション:\n"
	RemediationImport   = "1. terraform importで不足リソースをインポート: make terraform-import TF_ENV=%s\n"
	RemediationStateRm  = "2. terraform state rmで余分なリソースを削除: terraform state rm <リソースパス>\n"
	RemediationTagClean = "3. タグベース削除を使用: make tag-cleanup TF_ENV=%s\n"

	// メッセージプレフィックス
	DebugPrefix   = "DEBUG: %s\n"
	ErrorPrefix   = "✖ %s\n"
	WarningPrefix = "⚠️ %s\n"
	SuccessPrefix = "✅ %s\n"
	InfoPrefix    = "■ %s\n"
)

// DefaultOutputFormatter は標準の出力フォーマッター
// OutputFormatterインターフェースを実装し、比較結果の表示と修復オプションの提示を担当する
type DefaultOutputFormatter struct {
	writer interfaces.OutputWriter // 出力先のwriter
}

// NewDefaultOutputFormatter は新しいDefaultOutputFormatterを作成する
// writerがnilの場合はデフォルトのOutputWriterを使用する
func NewDefaultOutputFormatter(writer interfaces.OutputWriter) interfaces.OutputFormatter {
	if writer == nil {
		writer = implementation.NewDefaultOutputWriter()
	}
	return &DefaultOutputFormatter{
		writer: writer,
	}
}

// formatResultLine は1行の比較結果を文字列としてフォーマットする
// 共通のフォーマットロジックを提供する
func formatResultLine(result models.ComparisonResult) string {
	status := StatusMatch
	if !result.IsMatch {
		status = StatusMismatch
	}

	return fmt.Sprintf("%s\t\t%d\t%d\t%s\n",
		result.ResourceName,
		result.AWSCount,
		result.TerraformCount,
		status)
}

// FormatComparisonTable は比較結果をテーブル形式でフォーマットする
// 結果を文字列として返すため、出力先に依存せずフォーマット結果を取得できる
func (f *DefaultOutputFormatter) FormatComparisonTable(results []models.ComparisonResult) string {
	var builder strings.Builder

	builder.WriteString(TableHeader + "\n")
	builder.WriteString(HeaderSeparator + "\n")

	// nilチェック
	if results == nil {
		return builder.String()
	}

	for _, result := range results {
		builder.WriteString(formatResultLine(result))
	}

	return builder.String()
}

// DisplayResults は比較結果を標準出力に表示する
// 環境名と共に結果テーブルを出力する
func (f *DefaultOutputFormatter) DisplayResults(results []models.ComparisonResult, env string) {
	f.writer.Printf(ResultHeaderFormat, env)
	f.writer.Println(HeaderSeparator)

	// テーブル形式で結果を表示
	f.writer.Println(TableHeader)
	f.writer.Println(HeaderSeparator)

	// nilチェック
	if results == nil {
		return
	}

	for _, result := range results {
		status := StatusMatch
		if !result.IsMatch {
			status = StatusMismatch
		}

		f.writer.Printf("%s\t\t%s\t%s\t%s\n",
			result.ResourceName,
			strconv.Itoa(result.AWSCount),
			strconv.Itoa(result.TerraformCount),
			status)
	}
}

// ShowMismatchRemediation はリソース不一致時の修復オプションを表示する
// 環境名を指定して環境固有の修復コマンドを表示する
func (f *DefaultOutputFormatter) ShowMismatchRemediation(env string) {
	f.writer.Printf(RemediationHeader)
	f.writer.Printf(RemediationImport, env)
	f.writer.Printf(RemediationStateRm)
	f.writer.Printf(RemediationTagClean, env)
}

// PrintDebugInfo はデバッグ情報を表示する
// デバッグモード時のみ表示される情報
func (f *DefaultOutputFormatter) PrintDebugInfo(message string, args ...interface{}) {
	f.writer.Printf(DebugPrefix, fmt.Sprintf(message, args...))
}

// PrintError はエラーメッセージを表示する
// エラーアイコン(✖)付きでメッセージを表示
func (f *DefaultOutputFormatter) PrintError(message string, args ...interface{}) {
	f.writer.Printf(ErrorPrefix, fmt.Sprintf(message, args...))
}

// PrintWarning は警告メッセージを表示する
// 警告アイコン(⚠️)付きでメッセージを表示
func (f *DefaultOutputFormatter) PrintWarning(message string, args ...interface{}) {
	f.writer.Printf(WarningPrefix, fmt.Sprintf(message, args...))
}

// PrintSuccess は成功メッセージを表示する
// 成功アイコン(✅)付きでメッセージを表示
func (f *DefaultOutputFormatter) PrintSuccess(message string, args ...interface{}) {
	f.writer.Printf(SuccessPrefix, fmt.Sprintf(message, args...))
}

// PrintInfo は情報メッセージを表示する
// 情報アイコン(■)付きでメッセージを表示
func (f *DefaultOutputFormatter) PrintInfo(message string, args ...interface{}) {
	f.writer.Printf(InfoPrefix, fmt.Sprintf(message, args...))
}
