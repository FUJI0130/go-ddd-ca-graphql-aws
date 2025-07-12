// internal/tools/terraform/logger/global_test.go
package logger

import (
	"bytes"
	"strings"
	"testing"
)

func TestGlobalLogger(t *testing.T) {
	// 元のロガーを保存
	originalLogger := GetDefaultLogger()
	defer SetDefaultLogger(originalLogger)

	// テスト用バッファとロガー
	buf := &bytes.Buffer{}
	testLogger := NewDefaultLogger()
	testLogger.SetOutput(buf)
	testLogger.SetLevel(TraceLevel) // DebugLevel から TraceLevel に変更
	testLogger.showTime = false

	// グローバルロガーを置き換え
	SetDefaultLogger(testLogger)

	// 各レベルのログ出力をテスト
	testCases := []struct {
		name    string
		logFunc func(string, ...interface{})
		level   string
		message string
		success bool
	}{
		{
			name:    "グローバルErrorログ",
			logFunc: Error,
			level:   "ERROR",
			message: "エラーメッセージ",
			success: true,
		},
		{
			name:    "グローバルWarnログ",
			logFunc: Warn,
			level:   "WARN",
			message: "警告メッセージ",
			success: true,
		},
		{
			name:    "グローバルInfoログ",
			logFunc: Info,
			level:   "INFO",
			message: "情報メッセージ",
			success: true,
		},
		{
			name:    "グローバルDebugログ",
			logFunc: Debug,
			level:   "DEBUG",
			message: "デバッグメッセージ",
			success: true,
		},
		{
			name:    "グローバルTraceログ",
			logFunc: Trace,
			level:   "TRACE",
			message: "トレースメッセージ",
			success: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// バッファをクリア
			buf.Reset()

			// ログ出力
			tc.logFunc(tc.message)

			// 出力確認
			output := buf.String()
			if !strings.Contains(output, tc.level) || !strings.Contains(output, tc.message) {
				t.Errorf("期待する出力がありません。期待: レベル=%s, メッセージ=%s, 実際: %s",
					tc.level, tc.message, output)
			}

			// 成功/失敗の検証
			testSuccess := strings.Contains(output, tc.level) && strings.Contains(output, tc.message)
			if testSuccess != tc.success {
				t.Errorf("テスト成功判定: 期待=%v, 実際=%v", tc.success, testSuccess)
			}
		})
	}
}

// internal/tools/terraform/logger/global_test.go （続き）
func TestGlobalLoggerWithField(t *testing.T) {
	// 元のロガーを保存
	originalLogger := GetDefaultLogger()
	defer SetDefaultLogger(originalLogger)

	// テスト用バッファとロガー
	buf := &bytes.Buffer{}
	testLogger := NewDefaultLogger()
	testLogger.SetOutput(buf)
	testLogger.SetLevel(DebugLevel)
	testLogger.showTime = false

	// グローバルロガーを置き換え
	SetDefaultLogger(testLogger)

	// WithFieldのテスト
	WithField("component", "test").Info("フィールドテスト")

	output := buf.String()
	if !strings.Contains(output, "component=test") || !strings.Contains(output, "フィールドテスト") {
		t.Errorf("WithFieldが適切に機能していません。出力: %s", output)
	}

	// バッファをクリア
	buf.Reset()

	// WithFieldsのテスト
	fields := map[string]interface{}{
		"component": "test",
		"id":        123,
	}
	WithFields(fields).Info("複数フィールドテスト")

	output = buf.String()
	success := strings.Contains(output, "component=test") &&
		strings.Contains(output, "id=123") &&
		strings.Contains(output, "複数フィールドテスト")

	if !success {
		t.Errorf("WithFieldsが適切に機能していません。出力: %s", output)
	}
}

func TestSetLevelAndOutput(t *testing.T) {
	// 元のロガーを保存
	originalLogger := GetDefaultLogger()
	defer SetDefaultLogger(originalLogger)

	// テスト用バッファとロガー
	buf := &bytes.Buffer{}
	testLogger := NewDefaultLogger()
	testLogger.SetLevel(InfoLevel)

	// グローバルロガーを置き換え
	SetDefaultLogger(testLogger)

	// 出力先の設定
	SetOutput(buf)

	// デバッグログは出力されないはず（InfoレベルのLoggerのため）
	Debug("このメッセージは表示されないはず")
	if buf.Len() > 0 {
		t.Errorf("InfoレベルでDebugメッセージが出力されました: %s", buf.String())
	}

	// レベルをDebugに変更
	SetLevel(DebugLevel)

	// バッファをクリア
	buf.Reset()

	// デバッグログが出力されるはず
	Debug("このメッセージは表示されるはず")
	if buf.Len() == 0 {
		t.Errorf("DebugレベルでDebugメッセージが出力されませんでした")
	}

	success := buf.Len() > 0
	if !success {
		t.Errorf("SetLevelが適切に機能していません")
	}
}
