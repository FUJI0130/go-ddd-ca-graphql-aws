// ファイル: internal/interface/graphql/dataloader/dataloader.go

package dataloader

import (
	"context"
	"fmt"
	"net/http"

	"github.com/FUJI0130/go-ddd-ca/internal/usecase/port"
)

type ctxKey string

const (
	testGroupLoaderKey ctxKey = "test_group_loader"
	testCaseLoaderKey  ctxKey = "test_case_loader"
)

// DataLoaders には、すべてのデータローダーのインスタンスが含まれます
type DataLoaders struct {
	TestGroupLoader *TestGroupLoader
	TestCaseLoader  *TestCaseLoader
}

// NewDataLoaders は新しいDataLoadersインスタンスを作成します
func NewDataLoaders(groupUseCase port.TestGroupUseCase, caseUseCase port.TestCaseUseCase) *DataLoaders {
	return &DataLoaders{
		TestGroupLoader: NewTestGroupLoader(groupUseCase),
		TestCaseLoader:  NewTestCaseLoader(caseUseCase),
	}
}

// Middleware はDataLoaderをコンテキストに追加するミドルウェアです
func Middleware(loaders *DataLoaders, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), testGroupLoaderKey, loaders.TestGroupLoader)
		ctx = context.WithValue(ctx, testCaseLoaderKey, loaders.TestCaseLoader)

		// リクエスト処理開始時にDataLoaderをリセット
		loaders.TestGroupLoader.Clear()
		loaders.TestCaseLoader.Clear()

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetTestGroupLoader はコンテキストからTestGroupLoaderを取得します
func GetTestGroupLoader(ctx context.Context) (*TestGroupLoader, error) {
	loader, ok := ctx.Value(testGroupLoaderKey).(*TestGroupLoader)
	if !ok {
		return nil, fmt.Errorf("TestGroupLoader not found in context")
	}
	return loader, nil
}

// GetTestCaseLoader はコンテキストからTestCaseLoaderを取得します
func GetTestCaseLoader(ctx context.Context) (*TestCaseLoader, error) {
	loader, ok := ctx.Value(testCaseLoaderKey).(*TestCaseLoader)
	if !ok {
		return nil, fmt.Errorf("TestCaseLoader not found in context")
	}
	return loader, nil
}
