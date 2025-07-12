// internal/tools/terraform/logger/logger.go
package logger

import (
	"io"
)

// Logger はログ出力機能を提供するインターフェース
type Logger interface {
	// Error はエラーレベルのログを出力する
	Error(format string, args ...interface{})

	// Warn は警告レベルのログを出力する
	Warn(format string, args ...interface{})

	// Info は情報レベルのログを出力する
	Info(format string, args ...interface{})

	// Debug はデバッグレベルのログを出力する
	Debug(format string, args ...interface{})

	// Trace は詳細なトレースレベルのログを出力する
	Trace(format string, args ...interface{})

	// Log は指定したレベルでログを出力する
	Log(level LogLevel, format string, args ...interface{})

	// IsLevelEnabled は指定したレベルのログが出力可能かを返す
	IsLevelEnabled(level LogLevel) bool

	// SetLevel はロガーのログレベルを設定する
	SetLevel(level LogLevel)

	// GetLevel は現在のログレベルを取得する
	GetLevel() LogLevel

	// SetOutput はログの出力先を設定する
	SetOutput(w io.Writer)

	// WithField は指定したフィールドを持つ新しいロガーを返す
	WithField(key string, value interface{}) Logger

	// WithFields は指定したフィールド群を持つ新しいロガーを返す
	WithFields(fields map[string]interface{}) Logger
}
