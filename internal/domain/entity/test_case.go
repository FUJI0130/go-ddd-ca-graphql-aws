// ファイル: internal/domain/entity/test_case.go
package entity

import (
	"time"
)

// TestStatus はテストケースの状態を表す列挙型です
type TestStatus string

const (
	TestStatusCreated       TestStatus = "作成"
	TestStatusTesting       TestStatus = "テスト"
	TestStatusFixing        TestStatus = "修正"
	TestStatusReviewWaiting TestStatus = "レビュー待ち"
	TestStatusReviewing     TestStatus = "レビュー中"
	TestStatusCompleted     TestStatus = "完了"
	TestStatusRetesting     TestStatus = "再テスト"
)

// Priority はテストケースの優先度を表す列挙型です
type Priority string

const (
	PriorityCritical Priority = "Critical"
	PriorityHigh     Priority = "High"
	PriorityMedium   Priority = "Medium"
	PriorityLow      Priority = "Low"
)

// PriorityWeight は優先度の重みを返します
func (p Priority) Weight() float64 {
	switch p {
	case PriorityCritical:
		return 4.0
	case PriorityHigh:
		return 3.0
	case PriorityMedium:
		return 2.0
	case PriorityLow:
		return 1.0
	default:
		return 1.0
	}
}

// BaseProgressRate は状態に基づく基本進捗率を返します
func (s TestStatus) BaseProgressRate() float64 {
	switch s {
	case TestStatusCreated:
		return 0.0
	case TestStatusTesting:
		return 0.25
	case TestStatusFixing:
		return 0.25
	case TestStatusReviewWaiting:
		return 0.5
	case TestStatusReviewing:
		return 0.75
	case TestStatusCompleted:
		return 1.0
	case TestStatusRetesting:
		return 0.5
	default:
		return 0.0
	}
}

// TestCase はテストケースを表すエンティティです
type TestCase struct {
	ID            string
	GroupID       string
	Title         string
	Description   string
	Status        TestStatus
	Priority      Priority
	PlannedEffort float64
	ActualEffort  float64
	IsDelayed     bool
	DelayDays     int
	CurrentEditor string
	IsLocked      bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// NewTestCase は新しいTestCaseを作成します
func NewTestCase(id, groupID, title, description string, status TestStatus, priority Priority, plannedEffort float64) *TestCase {
	currentTime := time.Now()
	return &TestCase{
		ID:            id,
		GroupID:       groupID,
		Title:         title,
		Description:   description,
		Status:        status,
		Priority:      priority,
		PlannedEffort: plannedEffort,
		ActualEffort:  0,
		IsDelayed:     false,
		DelayDays:     0,
		IsLocked:      false,
		CreatedAt:     currentTime,
		UpdatedAt:     currentTime,
	}
}

// CalculateProgress はテストケースの進捗率を計算します
func (c *TestCase) CalculateProgress() float64 {
	// 状態に基づく基本進捗率 × 優先度の重み
	return c.Status.BaseProgressRate() * 100
}

// UpdateStatus はテストケースのステータスを更新します
func (c *TestCase) UpdateStatus(status TestStatus) {
	c.Status = status
	c.UpdatedAt = time.Now()
}

// AddEffort はテストケースに工数を追加します
func (c *TestCase) AddEffort(effort float64) {
	c.ActualEffort += effort
	c.UpdatedAt = time.Now()
}

// Lock はテストケースをロックします
func (c *TestCase) Lock(editor string) {
	c.IsLocked = true
	c.CurrentEditor = editor
	c.UpdatedAt = time.Now()
}

// Unlock はテストケースのロックを解除します
func (c *TestCase) Unlock() {
	c.IsLocked = false
	c.CurrentEditor = ""
	c.UpdatedAt = time.Now()
}

// MarkDelayed はテストケースを遅延としてマークします
func (c *TestCase) MarkDelayed(delayDays int) {
	c.IsDelayed = true
	c.DelayDays = delayDays
	c.UpdatedAt = time.Now()
}

// ClearDelay はテストケースの遅延マークを解除します
func (c *TestCase) ClearDelay() {
	c.IsDelayed = false
	c.DelayDays = 0
	c.UpdatedAt = time.Now()
}

// MoveToGroup はテストケースを別のグループに移動します
func (c *TestCase) MoveToGroup(targetGroupID string) {
	c.GroupID = targetGroupID
	c.UpdatedAt = time.Now()
}
