package implementation

import (
	"os"

	"github.com/FUJI0130/go-ddd-ca/internal/tools/terraform/interfaces"
)

// DefaultFileSystem はFileSystemの標準実装
type DefaultFileSystem struct{}

// Getwd は現在の作業ディレクトリを取得します
func (fs *DefaultFileSystem) Getwd() (string, error) {
	return os.Getwd()
}

// Chdir はディレクトリを変更します
func (fs *DefaultFileSystem) Chdir(dir string) error {
	return os.Chdir(dir)
}

// NewDefaultFileSystem は新しいDefaultFileSystemインスタンスを作成する
func NewDefaultFileSystem() interfaces.FileSystem {
	return &DefaultFileSystem{}
}
