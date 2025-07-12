package dto

import (
	"time"
)

// TestGroupResponseDTO はテストグループのレスポンスDTO
type TestGroupResponseDTO struct {
	ID           string    `json:"id"`
	SuiteID      string    `json:"suiteId"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	DisplayOrder int       `json:"displayOrder"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// TestGroupListResponseDTO はテストグループ一覧のレスポンスDTO
type TestGroupListResponseDTO struct {
	TestGroups []*TestGroupResponseDTO `json:"testGroups"`
}

// TestGroupCreateDTO はテストグループ作成用のDTO
type TestGroupCreateDTO struct {
	SuiteID      string `json:"suiteId" validate:"required"`
	Name         string `json:"name" validate:"required"`
	Description  string `json:"description"`
	DisplayOrder int    `json:"displayOrder"`
}
