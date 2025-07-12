package terraform

import (
	"context"
	"fmt"
	"strings"
)

// MockAWSRunner はテスト用のAWSCommandRunnerモック
type MockAWSRunner struct {
	MockOutput  string
	MockError   error
	CommandFunc func(args ...string) (string, error)
}

// RunCommand はモックされたAWS CLIコマンド実行
func (m *MockAWSRunner) RunCommand(args ...string) (string, error) {
	if m.CommandFunc != nil {
		return m.CommandFunc(args...)
	}
	return m.MockOutput, m.MockError
}

// NewMockAWSRunner は新しいMockAWSRunnerインスタンスを作成する
func NewMockAWSRunner() *MockAWSRunner {
	return &MockAWSRunner{}
}

// WithMockOutput はモック出力を設定する
func (m *MockAWSRunner) WithMockOutput(output string) *MockAWSRunner {
	m.MockOutput = output
	return m
}

// WithMockError はモックエラーを設定する
func (m *MockAWSRunner) WithMockError(err error) *MockAWSRunner {
	m.MockError = err
	return m
}

// WithCommandFunc はカスタムコマンド関数を設定する
func (m *MockAWSRunner) WithCommandFunc(fn func(args ...string) (string, error)) *MockAWSRunner {
	m.CommandFunc = fn
	return m
}

// MockCommandExecutor はテスト用のCommandExecutorモック
type MockCommandExecutor struct {
	MockOutput  string
	MockError   error
	ExecuteFunc func(command string, args ...string) (string, error)
}

// Execute はモックされたコマンド実行
func (m *MockCommandExecutor) Execute(command string, args ...string) (string, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(command, args...)
	}
	return m.MockOutput, m.MockError
}

// NewMockCommandExecutor は新しいMockCommandExecutorインスタンスを作成する
func NewMockCommandExecutor() *MockCommandExecutor {
	return &MockCommandExecutor{}
}

// WithMockOutput はモック出力を設定する
func (m *MockCommandExecutor) WithMockOutput(output string) *MockCommandExecutor {
	m.MockOutput = output
	return m
}

// WithMockError はモックエラーを設定する
func (m *MockCommandExecutor) WithMockError(err error) *MockCommandExecutor {
	m.MockError = err
	return m
}

// WithExecuteFunc はカスタム実行関数を設定する
func (m *MockCommandExecutor) WithExecuteFunc(fn func(command string, args ...string) (string, error)) *MockCommandExecutor {
	m.ExecuteFunc = fn
	return m
}

// MockContextCommandExecutor はContextAwareCommandExecutorのモック
type MockContextCommandExecutor struct {
	// ポインタではなく直接埋め込み
	MockCommandExecutor
	executeWithContextFunc func(ctx context.Context, command string, args ...string) (string, error)
}

// NewMockContextCommandExecutor は新しいMockContextCommandExecutorインスタンスを作成する
func NewMockContextCommandExecutor() *MockContextCommandExecutor {
	return &MockContextCommandExecutor{
		MockCommandExecutor: *NewMockCommandExecutor(),
	}
}

// WithExecuteFunc のオーバーライド
func (m *MockContextCommandExecutor) WithExecuteFunc(fn func(command string, args ...string) (string, error)) *MockContextCommandExecutor {
	m.ExecuteFunc = fn
	return m
}

// WithExecuteWithContextFunc の実装
func (m *MockContextCommandExecutor) WithExecuteWithContextFunc(fn func(ctx context.Context, command string, args ...string) (string, error)) *MockContextCommandExecutor {
	m.executeWithContextFunc = fn
	return m
}

// ExecuteWithContext の実装
func (m *MockContextCommandExecutor) ExecuteWithContext(ctx context.Context, command string, args ...string) (string, error) {
	if m.executeWithContextFunc != nil {
		return m.executeWithContextFunc(ctx, command, args...)
	}
	return m.Execute(command, args...)
}

// Execute メソッドをオーバーライド
func (m *MockContextCommandExecutor) Execute(command string, args ...string) (string, error) {
	return m.MockCommandExecutor.Execute(command, args...)
}

// MockFileSystem はテスト用のFileSystemモック
type MockFileSystem struct {
	MockGetwd    string
	MockGetwdErr error
	MockChdirErr error
	GetwdFunc    func() (string, error)
	ChdirFunc    func(dir string) error
}

// Getwd はモックされた作業ディレクトリ取得
func (m *MockFileSystem) Getwd() (string, error) {
	if m.GetwdFunc != nil {
		return m.GetwdFunc()
	}
	return m.MockGetwd, m.MockGetwdErr
}

// Chdir はモックされたディレクトリ変更
func (m *MockFileSystem) Chdir(dir string) error {
	if m.ChdirFunc != nil {
		return m.ChdirFunc(dir)
	}
	return m.MockChdirErr
}

// NewMockFileSystem は新しいMockFileSystemインスタンスを作成する
func NewMockFileSystem() *MockFileSystem {
	return &MockFileSystem{}
}

// WithMockGetwd はモックGetwd結果を設定する
func (m *MockFileSystem) WithMockGetwd(getwd string) *MockFileSystem {
	m.MockGetwd = getwd
	return m
}

// WithMockGetwdErr はモックGetwdエラーを設定する
func (m *MockFileSystem) WithMockGetwdErr(err error) *MockFileSystem {
	m.MockGetwdErr = err
	return m
}

// WithMockChdirErr はモックChdirエラーを設定する
func (m *MockFileSystem) WithMockChdirErr(err error) *MockFileSystem {
	m.MockChdirErr = err
	return m
}

// WithGetwdFunc はカスタムGetwd関数を設定する
func (m *MockFileSystem) WithGetwdFunc(fn func() (string, error)) *MockFileSystem {
	m.GetwdFunc = fn
	return m
}

// WithChdirFunc はカスタムChdir関数を設定する
func (m *MockFileSystem) WithChdirFunc(fn func(dir string) error) *MockFileSystem {
	m.ChdirFunc = fn
	return m
}

// MockSystemExiter はテスト用のSystemExiterモック
type MockSystemExiter struct {
	MockExitCode int
	ExitCalled   bool
	ExitFunc     func(code int)
}

// Exit はモックされたシステム終了
func (m *MockSystemExiter) Exit(code int) {
	if m.ExitFunc != nil {
		m.ExitFunc(code)
		return
	}
	m.MockExitCode = code
	m.ExitCalled = true
}

// NewMockSystemExiter は新しいMockSystemExiterインスタンスを作成する
func NewMockSystemExiter() *MockSystemExiter {
	return &MockSystemExiter{}
}

// WithExitFunc はカスタムExit関数を設定する
func (m *MockSystemExiter) WithExitFunc(fn func(code int)) *MockSystemExiter {
	m.ExitFunc = fn
	return m
}

// MockOutputWriter はテスト用のOutputWriterモック
type MockOutputWriter struct {
	Outputs     []string
	PrintfFunc  func(format string, args ...interface{})
	PrintlnFunc func(args ...interface{})
}

// Printf はフォーマット付きで出力する
func (m *MockOutputWriter) Printf(format string, args ...interface{}) {
	if m.PrintfFunc != nil {
		m.PrintfFunc(format, args...)
		return
	}
	output := fmt.Sprintf(format, args...)
	m.Outputs = append(m.Outputs, output)
}

// Println は改行付きで出力する
func (m *MockOutputWriter) Println(args ...interface{}) {
	if m.PrintlnFunc != nil {
		m.PrintlnFunc(args...)
		return
	}
	output := fmt.Sprint(args...)
	m.Outputs = append(m.Outputs, output+"\n")
}

// NewMockOutputWriter は新しいMockOutputWriterを作成する
func NewMockOutputWriter() *MockOutputWriter {
	return &MockOutputWriter{
		Outputs: make([]string, 0),
	}
}

// WithPrintfFunc はカスタムPrintf関数を設定する
func (m *MockOutputWriter) WithPrintfFunc(fn func(format string, args ...interface{})) *MockOutputWriter {
	m.PrintfFunc = fn
	return m
}

// WithPrintlnFunc はカスタムPrintln関数を設定する
func (m *MockOutputWriter) WithPrintlnFunc(fn func(args ...interface{})) *MockOutputWriter {
	m.PrintlnFunc = fn
	return m
}

// GetOutput は全ての出力を結合して返す
func (m *MockOutputWriter) GetOutput() string {
	return strings.Join(m.Outputs, "")
}

// GetOutputs は全ての出力を配列で返す
func (m *MockOutputWriter) GetOutputs() []string {
	return m.Outputs
}

// ClearOutputs は出力履歴をクリアする
func (m *MockOutputWriter) ClearOutputs() {
	m.Outputs = make([]string, 0)
}
