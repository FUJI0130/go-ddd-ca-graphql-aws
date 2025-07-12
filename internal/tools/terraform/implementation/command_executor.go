package implementation

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"

	"github.com/FUJI0130/go-ddd-ca/internal/tools/terraform/interfaces"
	"github.com/FUJI0130/go-ddd-ca/internal/tools/terraform/logger"
)

// DefaultCommandExecutor はCommandExecutorの標準実装
type DefaultCommandExecutor struct{}

// Execute はコマンドを実行します
func (e *DefaultCommandExecutor) Execute(command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// ExecuteWithContext はコンテキスト付きでコマンドを実行します
func (e *DefaultCommandExecutor) ExecuteWithContext(ctx context.Context, command string, args ...string) (string, error) {
	logger.Debug("コンテキスト付きコマンド実行: %s %v", command, args)

	// context付きのコマンド作成
	cmd := exec.CommandContext(ctx, command, args...)

	// 出力バッファの準備
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// コマンド実行
	err := cmd.Run()

	// コンテキストがキャンセルされた場合の特別なエラーメッセージ
	if ctx.Err() == context.DeadlineExceeded {
		logger.Error("コマンド実行がタイムアウトしました: %s %v", command, args)
		return stdout.String(), fmt.Errorf("コマンド実行がタイムアウトしました: %s %v\n部分的な出力:\n%s\n\nエラー出力:\n%s",
			command, args, stdout.String(), stderr.String())
	}

	if ctx.Err() == context.Canceled {
		logger.Error("コマンド実行がキャンセルされました: %s %v", command, args)
		return stdout.String(), fmt.Errorf("コマンド実行がキャンセルされました: %s %v", command, args)
	}

	// 通常のエラー処理
	if err != nil {
		logger.Error("コマンド実行エラー: %v\nstderr: %s", err, stderr.String())
		return stdout.String(), fmt.Errorf("コマンド実行エラー: %v\n標準エラー出力:\n%s", err, stderr.String())
	}

	logger.Debug("コマンド実行完了: 出力サイズ=%d バイト", stdout.Len())
	logger.Trace("コマンド実行出力: %s", stdout.String())

	return stdout.String(), nil
}

// NewDefaultCommandExecutor は新しいDefaultCommandExecutorインスタンスを作成する
func NewDefaultCommandExecutor() interfaces.CommandExecutor {
	return &DefaultCommandExecutor{}
}

// NewContextAwareCommandExecutor は新しいContextAwareCommandExecutorインスタンスを作成する
func NewContextAwareCommandExecutor() interfaces.ContextAwareCommandExecutor {
	return &DefaultCommandExecutor{}
}

// AsContextAware は既存のCommandExecutorをContextAwareCommandExecutorとして返す
func (e *DefaultCommandExecutor) AsContextAware() interfaces.ContextAwareCommandExecutor {
	return e
}

// CommandExecutorWrapper はコンテキスト非対応のCommandExecutorをラップするヘルパー
type CommandExecutorWrapper struct {
	executor interfaces.CommandExecutor
}

// NewCommandExecutorWrapper は新しいCommandExecutorWrapperを作成します
func NewCommandExecutorWrapper(executor interfaces.CommandExecutor) interfaces.ContextAwareCommandExecutor {
	return &CommandExecutorWrapper{executor}
}

// Execute はインターフェースを満たす
func (w *CommandExecutorWrapper) Execute(command string, args ...string) (string, error) {
	return w.executor.Execute(command, args...)
}

// ExecuteWithContext はコンテキスト付きでコマンドを実行します
func (w *CommandExecutorWrapper) ExecuteWithContext(ctx context.Context, command string, args ...string) (string, error) {
	// コンテキストのキャンセルチェック
	if ctx.Err() != nil {
		return "", ctx.Err()
	}

	// 通常の実行（コンテキストは無視）
	// 注: 理想的には goroutine とチャネルを使ってキャンセルを処理すべきだが、
	// 互換性のための簡易実装として提供
	return w.executor.Execute(command, args...)
}
