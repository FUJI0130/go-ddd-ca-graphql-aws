package valueobject_test

import (
    "testing"
    "github.com/FUJI0130/go-ddd-ca/internal/domain/valueobject"
)

func TestSuiteStatus_IsValid(t *testing.T) {
    tests := []struct {
        name   string
        status valueobject.SuiteStatus
        want   bool
    }{
        {
            name:   "準備中は有効",
            status: valueobject.SuiteStatusPreparation,
            want:   true,
        },
        {
            name:   "実行中は有効",
            status: valueobject.SuiteStatusInProgress,
            want:   true,
        },
        {
            name:   "完了は有効",
            status: valueobject.SuiteStatusCompleted,
            want:   true,
        },
        {
            name:   "中断は有効",
            status: valueobject.SuiteStatusSuspended,
            want:   true,
        },
        {
            name:   "未定義の値は無効",
            status: "不正な値",
            want:   false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if got := tt.status.IsValid(); got != tt.want {
                t.Errorf("SuiteStatus.IsValid() = %v, want %v", got, tt.want)
            }
        })
    }
}