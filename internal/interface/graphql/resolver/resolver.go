package resolver

import (
	"context"

	"github.com/FUJI0130/go-ddd-ca/internal/usecase/dto"
	"github.com/FUJI0130/go-ddd-ca/internal/usecase/port"
)

// TestSuiteUseCase はGraphQLリゾルバーが使用するテストスイートのユースケースインターフェースです
// テストスイートの作成、取得、更新などの操作を提供します
type TestSuiteUseCase interface {
	// CreateTestSuite は新しいテストスイートを作成します
	CreateTestSuite(ctx context.Context, createDTO *dto.TestSuiteCreateDTO) (*dto.TestSuiteResponseDTO, error)

	// GetTestSuite はIDを指定してテストスイートを取得します
	GetTestSuite(ctx context.Context, id string) (*dto.TestSuiteResponseDTO, error)

	// ListTestSuites はフィルタリング条件に基づいてテストスイート一覧を取得します
	ListTestSuites(ctx context.Context, params *dto.TestSuiteQueryParamDTO) (*dto.TestSuiteListResponseDTO, error)

	// UpdateTestSuite はIDを指定してテストスイートを更新します
	UpdateTestSuite(ctx context.Context, id string, updateDTO *dto.TestSuiteUpdateDTO) (*dto.TestSuiteResponseDTO, error)

	// UpdateTestSuiteStatus はIDを指定してテストスイートのステータスを更新します
	UpdateTestSuiteStatus(ctx context.Context, id string, statusDTO *dto.TestSuiteStatusUpdateDTO) (*dto.TestSuiteResponseDTO, error)
}

// Resolver はGraphQLリゾルバーのルート構造体です
// すべてのリゾルバーはこの構造体のメソッドとして実装されています
type Resolver struct {
	// テストスイート関連のユースケース
	TestSuiteUseCase TestSuiteUseCase

	// テストグループ関連のユースケース
	TestGroupUseCase port.TestGroupUseCase

	// テストケース関連のユースケース
	TestCaseUseCase port.TestCaseUseCase

	// 認証用のユースケース
	AuthUseCase port.AuthUseCase

	UserManagementUseCase port.UserManagementUseCase
}

// NewResolver は新しいリゾルバーインスタンスを作成します
// 各種ユースケースを注入して初期化します
func NewResolver(
	testSuiteUseCase TestSuiteUseCase,
	testGroupUseCase port.TestGroupUseCase,
	testCaseUseCase port.TestCaseUseCase,
	authUseCase port.AuthUseCase,
	userManagementUseCase port.UserManagementUseCase, // ← 追加
) *Resolver {
	return &Resolver{
		TestSuiteUseCase:      testSuiteUseCase,
		TestGroupUseCase:      testGroupUseCase,
		TestCaseUseCase:       testCaseUseCase,
		AuthUseCase:           authUseCase,
		UserManagementUseCase: userManagementUseCase, // ← 追加
	}
}
