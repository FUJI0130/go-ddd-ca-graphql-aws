package implementation

import (
	"os"

	"github.com/FUJI0130/go-ddd-ca/internal/tools/terraform/interfaces"
)

// DefaultSystemExiter はSystemExiterの標準実装
type DefaultSystemExiter struct{}

// Exit はプログラムを指定のコードで終了します
func (e *DefaultSystemExiter) Exit(code int) {
	os.Exit(code)
}

// NewDefaultSystemExiter は新しいDefaultSystemExiterインスタンスを作成する
func NewDefaultSystemExiter() interfaces.SystemExiter {
	return &DefaultSystemExiter{}
}
