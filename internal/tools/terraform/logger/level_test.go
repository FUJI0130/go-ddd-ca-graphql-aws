// internal/tools/terraform/logger/level_test.go
package logger

import (
	"testing"
)

func TestLogLevelString(t *testing.T) {
	testCases := []struct {
		name     string
		level    LogLevel
		expected string
		success  bool
	}{
		{
			name:     "ErrorLevel",
			level:    ErrorLevel,
			expected: "ERROR",
			success:  true,
		},
		{
			name:     "WarnLevel",
			level:    WarnLevel,
			expected: "WARN",
			success:  true,
		},
		{
			name:     "InfoLevel",
			level:    InfoLevel,
			expected: "INFO",
			success:  true,
		},
		{
			name:     "DebugLevel",
			level:    DebugLevel,
			expected: "DEBUG",
			success:  true,
		},
		{
			name:     "TraceLevel",
			level:    TraceLevel,
			expected: "TRACE",
			success:  true,
		},
		{
			name:     "未定義のレベル",
			level:    LogLevel(99),
			expected: "UNKNOWN",
			success:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.level.String()
			if result != tc.expected {
				t.Errorf("期待値: %s, 実際の値: %s", tc.expected, result)
			}
		})
	}
}

func TestStringToLogLevel(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected LogLevel
		success  bool
	}{
		{
			name:     "ERROR文字列",
			input:    "ERROR",
			expected: ErrorLevel,
			success:  true,
		},
		{
			name:     "小文字error",
			input:    "error",
			expected: ErrorLevel,
			success:  true,
		},
		{
			name:     "WARN文字列",
			input:    "WARN",
			expected: WarnLevel,
			success:  true,
		},
		{
			name:     "INFO文字列",
			input:    "INFO",
			expected: InfoLevel,
			success:  true,
		},
		{
			name:     "DEBUG文字列",
			input:    "DEBUG",
			expected: DebugLevel,
			success:  true,
		},
		{
			name:     "TRACE文字列",
			input:    "TRACE",
			expected: TraceLevel,
			success:  true,
		},
		{
			name:     "未定義の文字列",
			input:    "UNDEFINED",
			expected: InfoLevel, // デフォルトはINFO
			success:  true,
		},
		{
			name:     "空文字列",
			input:    "",
			expected: InfoLevel, // デフォルトはINFO
			success:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := StringToLogLevel(tc.input)
			if result != tc.expected {
				t.Errorf("期待値: %v, 実際の値: %v", tc.expected, result)
			}
		})
	}
}
