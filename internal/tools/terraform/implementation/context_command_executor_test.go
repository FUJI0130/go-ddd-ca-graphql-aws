package implementation

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/FUJI0130/go-ddd-ca/internal/tools/terraform/interfaces"
)

func TestExecuteWithContext(t *testing.T) {
	testCases := []struct {
		name          string
		command       string
		args          []string
		timeout       time.Duration
		sleepDuration time.Duration
		expectTimeout bool
		success       bool
	}{
		{
			name:          "正常終了（タイムアウトなし）",
			command:       "echo",
			args:          []string{"test"},
			timeout:       5 * time.Second,
			sleepDuration: 0,
			expectTimeout: false,
			success:       true,
		},
		{
			name:          "タイムアウト発生",
			command:       "sleep",
			args:          []string{"3"},
			timeout:       1 * time.Second,
			sleepDuration: 0,
			expectTimeout: true,
			success:       true, // タイムアウトが期待通りのため成功
		},
		{
			name:          "コンテキストキャンセル",
			command:       "sleep",
			args:          []string{"5"},
			timeout:       10 * time.Second,
			sleepDuration: 0,
			expectTimeout: false, // タイムアウトではなくキャンセル
			success:       true,  // キャンセルが期待通りのため成功
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			executor := NewContextAwareCommandExecutor()

			// タイムアウト付きコンテキスト作成
			ctx, cancel := context.WithTimeout(context.Background(), tc.timeout)
			defer cancel()

			// コンテキストキャンセルのテストの場合
			if tc.name == "コンテキストキャンセル" {
				go func() {
					time.Sleep(1 * time.Second) // 1秒後にキャンセル
					cancel()
				}()
			}

			output, err := executor.(interfaces.ContextAwareCommandExecutor).ExecuteWithContext(ctx, tc.command, tc.args...)

			// タイムアウト確認
			isTimeout := err != nil && ctx.Err() == context.DeadlineExceeded

			if isTimeout != tc.expectTimeout {
				t.Errorf("タイムアウト状態: 期待=%v, 実際=%v", tc.expectTimeout, isTimeout)
			}

			// 成功/失敗の検証
			var testSuccess bool
			if tc.name == "コンテキストキャンセル" {
				isCanceled := err != nil && ctx.Err() == context.Canceled
				testSuccess = isCanceled
			} else {
				testSuccess = isTimeout == tc.expectTimeout
				if tc.name == "正常終了（タイムアウトなし）" {
					testSuccess = testSuccess && strings.Contains(output, "test")
				}
			}

			if testSuccess != tc.success {
				t.Errorf("テスト成功判定: 期待=%v, 実際=%v", tc.success, testSuccess)
				if err != nil {
					t.Logf("エラー: %v", err)
				}
				t.Logf("出力: %s", output)
			}
		})
	}
}

// TestCommandExecutorWrapper はcommandExecutorWrapperのテスト
func TestCommandExecutorWrapper(t *testing.T) {
	testCases := []struct {
		name      string
		mockCmd   *MockCommandExecutor
		ctx       context.Context
		expectErr bool
		success   bool
	}{
		{
			name: "正常系: コンテキスト有効",
			mockCmd: NewMockCommandExecutor().
				WithExecuteFunc(func(command string, args ...string) (string, error) {
					return "executed", nil
				}),
			ctx:       context.Background(),
			expectErr: false,
			success:   true,
		},
		{
			name: "異常系: コンテキストキャンセル済み",
			mockCmd: NewMockCommandExecutor().
				WithExecuteFunc(func(command string, args ...string) (string, error) {
					return "executed", nil
				}),
			ctx:       canceledContext(),
			expectErr: true,
			success:   true,
		},
		{
			name: "異常系: コマンド実行エラー",
			mockCmd: NewMockCommandExecutor().
				WithExecuteFunc(func(command string, args ...string) (string, error) {
					return "", fmt.Errorf("execution error")
				}),
			ctx:       context.Background(),
			expectErr: true,
			success:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			wrapper := &CommandExecutorWrapper{tc.mockCmd}

			output, err := wrapper.ExecuteWithContext(tc.ctx, "test", "arg1", "arg2")

			hasError := err != nil
			if hasError != tc.expectErr {
				t.Errorf("エラー発生: 期待=%v, 実際=%v", tc.expectErr, hasError)
				if err != nil {
					t.Logf("エラー内容: %v", err)
				}
			}

			// 基本的なラッパーの動作検証
			if !tc.expectErr && output != "executed" {
				t.Errorf("出力: 期待='executed', 実際='%s'", output)
			}

			// 成功/失敗の検証
			testSuccess := hasError == tc.expectErr

			if testSuccess != tc.success {
				t.Errorf("テスト成功判定: 期待=%v, 実際=%v", tc.success, testSuccess)
			}
		})
	}
}

// canceledContext はキャンセル済みのコンテキストを返す
func canceledContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

// MockCommandExecutor は CommandExecutor のモック
type MockCommandExecutor struct {
	executeFunc func(command string, args ...string) (string, error)
}

// NewMockCommandExecutor は新しい MockCommandExecutor を作成する
func NewMockCommandExecutor() *MockCommandExecutor {
	return &MockCommandExecutor{}
}

// WithExecuteFunc は Execute 関数の挙動を設定する
func (m *MockCommandExecutor) WithExecuteFunc(fn func(command string, args ...string) (string, error)) *MockCommandExecutor {
	m.executeFunc = fn
	return m
}

// Execute はインターフェースを満たす
func (m *MockCommandExecutor) Execute(command string, args ...string) (string, error) {
	if m.executeFunc != nil {
		return m.executeFunc(command, args...)
	}
	return "", nil
}
