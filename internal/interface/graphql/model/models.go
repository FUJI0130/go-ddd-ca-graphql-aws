package model

import (
	"time"
)

// TestSuite はGraphQLモデルのテストスイート型
type TestSuite struct {
	ID                   string       `json:"id"`
	Name                 string       `json:"name"`
	Description          string       `json:"description"`
	Status               SuiteStatus  `json:"status"`
	EstimatedStartDate   time.Time    `json:"estimatedStartDate"`
	EstimatedEndDate     time.Time    `json:"estimatedEndDate"`
	RequireEffortComment bool         `json:"requireEffortComment"`
	Progress             float64      `json:"progress"`
	CreatedAt            time.Time    `json:"createdAt"`
	UpdatedAt            time.Time    `json:"updatedAt"`
	Groups               []*TestGroup `json:"groups,omitempty"`
}

// TestGroup はGraphQLモデルのテストグループ型
type TestGroup struct {
	ID           string      `json:"id"`
	Name         string      `json:"name"`
	Description  string      `json:"description"`
	DisplayOrder int         `json:"displayOrder"`
	SuiteID      string      `json:"suiteId"`
	Status       SuiteStatus `json:"status"`
	CreatedAt    time.Time   `json:"createdAt"`
	UpdatedAt    time.Time   `json:"updatedAt"`
	Cases        []*TestCase `json:"cases,omitempty"`
}

// TestCase はGraphQLモデルのテストケース型
type TestCase struct {
	ID            string     `json:"id"`
	Title         string     `json:"title"`
	Description   string     `json:"description"`
	Status        TestStatus `json:"status"`
	Priority      Priority   `json:"priority"`
	PlannedEffort *float64   `json:"plannedEffort,omitempty"`
	ActualEffort  *float64   `json:"actualEffort,omitempty"`
	IsDelayed     bool       `json:"isDelayed"`
	DelayDays     *int       `json:"delayDays,omitempty"`
	GroupID       string     `json:"groupId"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

// User は認証されたユーザーを表すGraphQLモデル
type User struct {
	ID          string     `json:"id"`
	Username    string     `json:"username"`
	Role        string     `json:"role"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	LastLoginAt *time.Time `json:"lastLoginAt,omitempty"`
}

// AuthPayload は認証レスポンスを表すGraphQLモデル
type AuthPayload struct {
	Token        string    `json:"token"`
	RefreshToken string    `json:"refreshToken"`
	User         *User     `json:"user"`
	ExpiresAt    time.Time `json:"expiresAt"`
}
