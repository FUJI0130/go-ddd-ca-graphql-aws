package dto

import (
	"time"
)

// TestCaseResponseDTO はテストケースのレスポンスDTO
type TestCaseResponseDTO struct {
	ID            string    `json:"id"`
	GroupID       string    `json:"groupId"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	Status        string    `json:"status"`
	Priority      string    `json:"priority"`
	PlannedEffort float64   `json:"plannedEffort"`
	ActualEffort  float64   `json:"actualEffort"`
	IsDelayed     bool      `json:"isDelayed"`
	DelayDays     int       `json:"delayDays"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// TestCaseListResponseDTO はテストケース一覧のレスポンスDTO
type TestCaseListResponseDTO struct {
	TestCases []*TestCaseResponseDTO `json:"testCases"`
}

// TestCaseCreateDTO はテストケース作成用のDTO
type TestCaseCreateDTO struct {
	GroupID       string  `json:"groupId" validate:"required"`
	Title         string  `json:"title" validate:"required"`
	Description   string  `json:"description"`
	Priority      string  `json:"priority"`
	PlannedEffort float64 `json:"plannedEffort"`
}
