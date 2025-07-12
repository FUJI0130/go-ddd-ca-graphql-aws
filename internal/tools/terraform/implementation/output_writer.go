package implementation

import (
	"fmt"

	"github.com/FUJI0130/go-ddd-ca/internal/tools/terraform/interfaces"
)

// DefaultOutputWriter は標準の出力ライター
type DefaultOutputWriter struct{}

// Printf はフォーマット付きで出力する
func (w *DefaultOutputWriter) Printf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

// Println は改行付きで出力する
func (w *DefaultOutputWriter) Println(args ...interface{}) {
	fmt.Println(args...)
}

// NewDefaultOutputWriter は新しいDefaultOutputWriterを作成する
func NewDefaultOutputWriter() interfaces.OutputWriter {
	return &DefaultOutputWriter{}
}

// ColoredOutputWriter はカラー対応の出力ライター
type ColoredOutputWriter struct{}

// Printf はカラー対応でフォーマット付きで出力する
func (w *ColoredOutputWriter) Printf(format string, args ...interface{}) {
	// カラー処理のロジックを実装
	fmt.Printf(format, args...)
}

// Println はカラー対応で改行付きで出力する
func (w *ColoredOutputWriter) Println(args ...interface{}) {
	// カラー処理のロジックを実装
	fmt.Println(args...)
}

// NewColoredOutputWriter は新しいColoredOutputWriterを作成する
func NewColoredOutputWriter() interfaces.OutputWriter {
	return &ColoredOutputWriter{}
}
