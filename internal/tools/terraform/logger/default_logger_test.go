// internal/tools/terraform/logger/default_logger_test.go
package logger

import (
	"bytes"
	"strings"
	"testing"
)

func TestDefaultLogger(t *testing.T) {
	testCases := []struct {
		name      string
		level     LogLevel
		msgLevel  LogLevel
		message   string
		expected  string
		shouldLog bool
		success   bool
	}{
		{
			name:      "INFO出力（INFOレベル設定）",
			level:     InfoLevel,
			msgLevel:  InfoLevel,
			message:   "テストメッセージ",
			expected:  "テストメッセージ",
			shouldLog: true,
			success:   true,
		},
		{
			name:      "DEBUG出力（INFOレベル設定）",
			level:     InfoLevel,
			msgLevel:  DebugLevel,
			message:   "テストメッセージ",
			expected:  "",
			shouldLog: false,
			success:   true,
		},
		{
			name:      "ERROR出力（DEBUGレベル設定）",
			level:     DebugLevel,
			msgLevel:  ErrorLevel,
			message:   "テストメッセージ",
			expected:  "テストメッセージ",
			shouldLog: true,
			success:   true,
		},
		{
			name:      "DEBUG出力（DEBUGレベル設定）",
			level:     DebugLevel,
			msgLevel:  DebugLevel,
			message:   "テストメッセージ",
			expected:  "テストメッセージ",
			shouldLog: true,
			success:   true,
		},
		{
			name:      "機密情報マスキング",
			level:     InfoLevel,
			msgLevel:  InfoLevel,
			message:   "AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE",
			expected:  "AWS_ACCESS_KEY_ID=********************",
			shouldLog: true,
			success:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// バッファに出力をキャプチャ
			buf := &bytes.Buffer{}
			logger := NewDefaultLogger()
			logger.SetOutput(buf)
			logger.SetLevel(tc.level)

			// 時刻表示を無効化（テスト安定化のため）
			logger.showTime = false

			// ログ出力
			switch tc.msgLevel {
			case ErrorLevel:
				logger.Error(tc.message)
			case WarnLevel:
				logger.Warn(tc.message)
			case InfoLevel:
				logger.Info(tc.message)
			case DebugLevel:
				logger.Debug(tc.message)
			case TraceLevel:
				logger.Trace(tc.message)
			}

			// 結果確認
			output := buf.String()

			// 出力があるかチェック
			hasOutput := output != ""
			if hasOutput != tc.shouldLog {
				t.Errorf("ログ出力の有無: 期待=%v, 実際=%v", tc.shouldLog, hasOutput)
			}

			// 内容チェック（出力がある場合のみ）
			if tc.shouldLog && !strings.Contains(output, tc.expected) {
				t.Errorf("ログ内容: 期待文字列「%s」が出力「%s」に含まれていません", tc.expected, output)
			}

			// 成功/失敗の検証
			testSuccess := true
			if tc.shouldLog {
				testSuccess = strings.Contains(output, tc.expected)
			} else {
				testSuccess = output == ""
			}

			if testSuccess != tc.success {
				t.Errorf("テスト成功判定: 期待=%v, 実際=%v", tc.success, testSuccess)
			}
		})
	}
}

func TestDefaultLoggerWithFields(t *testing.T) {
	testCases := []struct {
		name     string
		fieldKey string
		fieldVal interface{}
		message  string
		expected string
		success  bool
	}{
		{
			name:     "単一フィールド",
			fieldKey: "component",
			fieldVal: "test",
			message:  "テストメッセージ",
			expected: "component=test",
			success:  true,
		},
		{
			name:     "数値フィールド",
			fieldKey: "id",
			fieldVal: 123,
			message:  "IDテスト",
			expected: "id=123",
			success:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			logger := NewDefaultLogger()
			logger.SetOutput(buf)
			logger.SetLevel(DebugLevel)
			logger.showTime = false

			// フィールド付きでログ出力
			logger.WithField(tc.fieldKey, tc.fieldVal).Info(tc.message)

			output := buf.String()
			if !strings.Contains(output, tc.expected) {
				t.Errorf("フィールドが出力に含まれていません: 期待=%s, 実際=%s", tc.expected, output)
			}

			// 成功/失敗の検証
			testSuccess := strings.Contains(output, tc.expected)
			if testSuccess != tc.success {
				t.Errorf("テスト成功判定: 期待=%v, 実際=%v", tc.success, testSuccess)
			}
		})
	}
}

func TestDefaultLoggerWithMultipleFields(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewDefaultLogger()
	logger.SetOutput(buf)
	logger.SetLevel(DebugLevel)
	logger.showTime = false

	// 複数フィールド
	fields := map[string]interface{}{
		"component": "test",
		"id":        123,
	}

	logger.WithFields(fields).Info("複数フィールドテスト")

	output := buf.String()

	// 両方のフィールドが含まれていることを確認
	success := strings.Contains(output, "component=test") &&
		strings.Contains(output, "id=123")

	if !success {
		t.Errorf("複数フィールドが出力に含まれていません: %s", output)
	}
}
