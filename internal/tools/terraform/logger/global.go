// internal/tools/terraform/logger/global.go
package logger

import (
	"io"
	"sync"
)

var (
	// defaultLogger はグローバルロガーのデフォルトインスタンス
	defaultLogger Logger = NewDefaultLogger()

	// loggerMu はグローバルロガーへのアクセスを同期するためのミューテックス
	loggerMu sync.RWMutex
)

// SetDefaultLogger はグローバルロガーを設定する
func SetDefaultLogger(logger Logger) {
	loggerMu.Lock()
	defaultLogger = logger
	loggerMu.Unlock()
}

// GetDefaultLogger はグローバルロガーを取得する
func GetDefaultLogger() Logger {
	loggerMu.RLock()
	defer loggerMu.RUnlock()
	return defaultLogger
}

// SetLevel はグローバルロガーのログレベルを設定する
func SetLevel(level LogLevel) {
	loggerMu.RLock()
	defer loggerMu.RUnlock()
	defaultLogger.SetLevel(level)
}

// SetOutput はグローバルロガーの出力先を設定する
func SetOutput(w io.Writer) {
	loggerMu.RLock()
	defer loggerMu.RUnlock()
	defaultLogger.SetOutput(w)
}

// Error はグローバルロガーを使用してエラーレベルのログを出力する
func Error(format string, args ...interface{}) {
	loggerMu.RLock()
	defer loggerMu.RUnlock()
	defaultLogger.Error(format, args...)
}

// Warn はグローバルロガーを使用して警告レベルのログを出力する
func Warn(format string, args ...interface{}) {
	loggerMu.RLock()
	defer loggerMu.RUnlock()
	defaultLogger.Warn(format, args...)
}

// Info はグローバルロガーを使用して情報レベルのログを出力する
func Info(format string, args ...interface{}) {
	loggerMu.RLock()
	defer loggerMu.RUnlock()
	defaultLogger.Info(format, args...)
}

// Debug はグローバルロガーを使用してデバッグレベルのログを出力する
func Debug(format string, args ...interface{}) {
	loggerMu.RLock()
	defer loggerMu.RUnlock()
	defaultLogger.Debug(format, args...)
}

// Trace はグローバルロガーを使用して詳細なトレースレベルのログを出力する
func Trace(format string, args ...interface{}) {
	loggerMu.RLock()
	defer loggerMu.RUnlock()
	defaultLogger.Trace(format, args...)
}

// WithField はグローバルロガーにフィールドを追加した新しいロガーを返す
func WithField(key string, value interface{}) Logger {
	loggerMu.RLock()
	defer loggerMu.RUnlock()
	return defaultLogger.WithField(key, value)
}

// WithFields はグローバルロガーに複数フィールドを追加した新しいロガーを返す
func WithFields(fields map[string]interface{}) Logger {
	loggerMu.RLock()
	defer loggerMu.RUnlock()
	return defaultLogger.WithFields(fields)
}
