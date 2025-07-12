// internal/tools/terraform/logger/default_logger.go
package logger

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"sync"
	"time"
)

// DefaultLogger は標準的なロガー実装
type DefaultLogger struct {
	level      LogLevel
	output     io.Writer
	showTime   bool
	showLevel  bool
	showColors bool
	fields     map[string]interface{}
	mu         sync.Mutex
}

// NewDefaultLogger は新しいDefaultLoggerインスタンスを作成する
func NewDefaultLogger() *DefaultLogger {
	return &DefaultLogger{
		level:      InfoLevel,
		output:     os.Stdout,
		showTime:   true,
		showLevel:  true,
		showColors: true,
		fields:     make(map[string]interface{}),
	}
}

// Error はエラーレベルのログを出力する
func (l *DefaultLogger) Error(format string, args ...interface{}) {
	l.Log(ErrorLevel, format, args...)
}

// Warn は警告レベルのログを出力する
func (l *DefaultLogger) Warn(format string, args ...interface{}) {
	l.Log(WarnLevel, format, args...)
}

// Info は情報レベルのログを出力する
func (l *DefaultLogger) Info(format string, args ...interface{}) {
	l.Log(InfoLevel, format, args...)
}

// Debug はデバッグレベルのログを出力する
func (l *DefaultLogger) Debug(format string, args ...interface{}) {
	l.Log(DebugLevel, format, args...)
}

// Trace は詳細なトレースレベルのログを出力する
func (l *DefaultLogger) Trace(format string, args ...interface{}) {
	l.Log(TraceLevel, format, args...)
}

// Log は指定したレベルでログを出力する
func (l *DefaultLogger) Log(level LogLevel, format string, args ...interface{}) {
	if !l.IsLevelEnabled(level) {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	var message string

	// フォーマット文字列を処理
	if len(args) > 0 {
		message = fmt.Sprintf(format, args...)
	} else {
		message = format
	}

	// 機密情報のマスキング
	message = maskSensitiveInfo(message)

	// 前置文字列の構築
	var prefix string

	if l.showTime {
		prefix += fmt.Sprintf("[%s]", time.Now().Format("2006-01-02 15:04:05"))
	}

	if l.showLevel {
		levelStr := level.String()
		if l.showColors {
			// レベルに応じた色付け
			switch level {
			case ErrorLevel:
				levelStr = fmt.Sprintf("\033[31m%s\033[0m", levelStr) // 赤
			case WarnLevel:
				levelStr = fmt.Sprintf("\033[33m%s\033[0m", levelStr) // 黄
			case InfoLevel:
				levelStr = fmt.Sprintf("\033[32m%s\033[0m", levelStr) // 緑
			case DebugLevel:
				levelStr = fmt.Sprintf("\033[36m%s\033[0m", levelStr) // シアン
			case TraceLevel:
				levelStr = fmt.Sprintf("\033[90m%s\033[0m", levelStr) // グレー
			}
		}
		prefix += fmt.Sprintf("[%s]", levelStr)
	}

	// フィールド情報の追加
	if len(l.fields) > 0 {
		var fieldStr string
		for k, v := range l.fields {
			if fieldStr != "" {
				fieldStr += " "
			}
			fieldStr += fmt.Sprintf("%s=%v", k, v)
		}
		prefix += fmt.Sprintf("[%s]", fieldStr)
	}

	// 最終的なログメッセージ
	if prefix != "" {
		message = prefix + " " + message
	}

	fmt.Fprintln(l.output, message)
}

// IsLevelEnabled は指定したレベルのログが出力可能かを返す
func (l *DefaultLogger) IsLevelEnabled(level LogLevel) bool {
	return level <= l.level
}

// SetLevel はロガーのログレベルを設定する
func (l *DefaultLogger) SetLevel(level LogLevel) {
	l.level = level
}

// GetLevel は現在のログレベルを取得する
func (l *DefaultLogger) GetLevel() LogLevel {
	return l.level
}

// SetOutput はログの出力先を設定する
func (l *DefaultLogger) SetOutput(w io.Writer) {
	l.output = w
}

// WithField は指定したフィールドを持つ新しいロガーを返す
func (l *DefaultLogger) WithField(key string, value interface{}) Logger {
	newLogger := &DefaultLogger{
		level:      l.level,
		output:     l.output,
		showTime:   l.showTime,
		showLevel:  l.showLevel,
		showColors: l.showColors,
		fields:     make(map[string]interface{}),
	}

	// 既存のフィールドをコピー
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}

	// 新しいフィールドを追加
	newLogger.fields[key] = value

	return newLogger
}

// WithFields は指定したフィールド群を持つ新しいロガーを返す
func (l *DefaultLogger) WithFields(fields map[string]interface{}) Logger {
	newLogger := &DefaultLogger{
		level:      l.level,
		output:     l.output,
		showTime:   l.showTime,
		showLevel:  l.showLevel,
		showColors: l.showColors,
		fields:     make(map[string]interface{}),
	}

	// 既存のフィールドをコピー
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}

	// 新しいフィールドを追加
	for k, v := range fields {
		newLogger.fields[k] = v
	}

	return newLogger
}

// 機密情報をマスクする
func maskSensitiveInfo(message string) string {
	sensitivePatterns := []struct {
		pattern     *regexp.Regexp
		replacement string
	}{
		{
			// AWS_ACCESS_KEY_ID
			regexp.MustCompile(`(AWS_ACCESS_KEY_ID=)([A-Z0-9]{20})`),
			"$1********************",
		},
		{
			// AWS_SECRET_ACCESS_KEY
			regexp.MustCompile(`(AWS_SECRET_ACCESS_KEY=)([A-Za-z0-9/+]{40})`),
			"$1****************************************",
		},
		{
			// パスワード
			regexp.MustCompile(`(password=|PASSWORD=|Password=)([^,\s]+)`),
			"$1********",
		},
	}

	result := message
	for _, sp := range sensitivePatterns {
		result = sp.pattern.ReplaceAllString(result, sp.replacement)
	}

	return result
}
