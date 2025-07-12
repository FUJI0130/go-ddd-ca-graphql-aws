package valueobject

import "fmt"

// SuiteStatus はテストスイートの状態を表す値オブジェクト
type SuiteStatus string

const (
	SuiteStatusPreparation SuiteStatus = "準備中"
	SuiteStatusInProgress  SuiteStatus = "実行中"
	SuiteStatusCompleted   SuiteStatus = "完了"
	SuiteStatusSuspended   SuiteStatus = "中断"
)

// IsValid は有効なステータス値かどうかを検証する
func (s SuiteStatus) IsValid() bool {
	switch s {
	case SuiteStatusPreparation,
		SuiteStatusInProgress,
		SuiteStatusCompleted,
		SuiteStatusSuspended:
		return true
	default:
		return false
	}
}

// String はステータスを文字列として返す
func (s SuiteStatus) String() string {
	return string(s)
}

// NewSuiteStatus は文字列からSuiteStatusを生成します
func NewSuiteStatus(status string) (SuiteStatus, error) {
	s := SuiteStatus(status)
	if !s.IsValid() {
		return "", fmt.Errorf("invalid status: %s", status)
	}
	return s, nil
}
