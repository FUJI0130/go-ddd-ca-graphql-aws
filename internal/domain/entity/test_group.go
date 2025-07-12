// ファイル: internal/domain/entity/test_group.go
package entity

import (
	"time"

	"github.com/FUJI0130/go-ddd-ca/internal/domain/valueobject"
)

// TestGroup はテストグループを表すエンティティです
type TestGroup struct {
	ID           string
	SuiteID      string
	Name         string
	Description  string
	DisplayOrder int
	Status       valueobject.SuiteStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// NewTestGroup は新しいTestGroupを作成します
func NewTestGroup(id, suiteID, name, description string, displayOrder int, status valueobject.SuiteStatus) *TestGroup {
	currentTime := time.Now()
	return &TestGroup{
		ID:           id,
		SuiteID:      suiteID,
		Name:         name,
		Description:  description,
		DisplayOrder: displayOrder,
		Status:       status,
		CreatedAt:    currentTime,
		UpdatedAt:    currentTime,
	}
}

// ProgressSummary はテストグループの進捗状況を表します
type ProgressSummary struct {
	ProgressPercentage float64
	CompletedCount     int
	TotalCount         int
}

// CalculateProgress はテストグループの進捗率を計算します
// この実装ではダミーの値を返していますが、実際にはテストケースの状態に基づいて計算します
func (g *TestGroup) CalculateProgress() float64 {
	// 後ほど詳細実装
	return 0.0
}

// GetProgressSummary はテストグループの進捗サマリーを取得します
func (g *TestGroup) GetProgressSummary() *ProgressSummary {
	// 後ほど詳細実装
	return &ProgressSummary{
		ProgressPercentage: g.CalculateProgress(),
		CompletedCount:     0,
		TotalCount:         0,
	}
}

// UpdateDisplayOrder はテストグループの表示順序を更新します
func (g *TestGroup) UpdateDisplayOrder(newOrder int) {
	g.DisplayOrder = newOrder
	g.UpdatedAt = time.Now()
}

// UpdateStatus はテストグループのステータスを更新します
func (g *TestGroup) UpdateStatus(status valueobject.SuiteStatus) {
	g.Status = status
	g.UpdatedAt = time.Now()
}
