package implementation

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/FUJI0130/go-ddd-ca/internal/tools/terraform/interfaces"
)

// DefaultAWSRunner はデフォルトのAWS CLI実行実装
type DefaultAWSRunner struct{}

// RunCommand はAWS CLIコマンドを実行し、結果を返す
func (r *DefaultAWSRunner) RunCommand(args ...string) (string, error) {
	cmd := exec.Command("aws", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("AWS CLI error: %v\nStderr: %s", err, stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}

// NewDefaultAWSRunner は新しいDefaultAWSRunnerインスタンスを作成する
func NewDefaultAWSRunner() interfaces.AWSCommandRunner {
	return &DefaultAWSRunner{}
}
