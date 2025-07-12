// internal/tools/terraform/logger/level.go
package logger

import (
	"strings"
)

// LogLevel はログの重要度レベルを表す型
type LogLevel int

const (
	// ErrorLevel は重大なエラーを表すログレベル
	ErrorLevel LogLevel = iota
	// WarnLevel は警告を表すログレベル
	WarnLevel
	// InfoLevel は一般的な情報を表すログレベル
	InfoLevel
	// DebugLevel はデバッグ情報を表すログレベル
	DebugLevel
	// TraceLevel は詳細なトレース情報を表すログレベル
	TraceLevel
)

// LogLevelStrings はログレベルに対応する文字列表現
var LogLevelStrings = map[LogLevel]string{
	ErrorLevel: "ERROR",
	WarnLevel:  "WARN",
	InfoLevel:  "INFO",
	DebugLevel: "DEBUG",
	TraceLevel: "TRACE",
}

// String はログレベルの文字列表現を返す
func (l LogLevel) String() string {
	if str, ok := LogLevelStrings[l]; ok {
		return str
	}
	return "UNKNOWN"
}

// StringToLogLevel は文字列からログレベルに変換する
func StringToLogLevel(level string) LogLevel {
	switch strings.ToUpper(level) {
	case "ERROR":
		return ErrorLevel
	case "WARN":
		return WarnLevel
	case "INFO":
		return InfoLevel
	case "DEBUG":
		return DebugLevel
	case "TRACE":
		return TraceLevel
	default:
		return InfoLevel // デフォルトはINFO
	}
}
