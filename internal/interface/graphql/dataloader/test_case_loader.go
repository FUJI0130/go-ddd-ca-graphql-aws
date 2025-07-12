// ファイル: internal/interface/graphql/dataloader/test_case_loader.go

package dataloader

import (
	"context"
	"sync"
	"time"

	"github.com/FUJI0130/go-ddd-ca/internal/usecase/dto"
	"github.com/FUJI0130/go-ddd-ca/internal/usecase/port"
)

// TestCaseLoader はテストケースをバッチ処理するためのデータローダーです
type TestCaseLoader struct {
	// 実際のデータを取得するためのユースケース
	useCase port.TestCaseUseCase

	// キャッシュとロックオブジェクト
	cache map[string][]*dto.TestCaseResponseDTO
	mutex sync.Mutex

	// バッチ処理のためのパラメータ
	maxBatchSize int
	wait         time.Duration
}

// NewTestCaseLoader は新しいTestCaseLoaderインスタンスを作成します
func NewTestCaseLoader(useCase port.TestCaseUseCase) *TestCaseLoader {
	return &TestCaseLoader{
		useCase:      useCase,
		cache:        make(map[string][]*dto.TestCaseResponseDTO),
		maxBatchSize: 100,
		wait:         1 * time.Millisecond,
	}
}

// Clear はキャッシュをクリアします
func (l *TestCaseLoader) Clear() {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.cache = make(map[string][]*dto.TestCaseResponseDTO)
}

// GetCasesByGroupID はグループIDに基づいてテストケースを取得します
// キャッシュが存在する場合はキャッシュから返し、なければDBから取得してキャッシュに保存します
func (l *TestCaseLoader) GetCasesByGroupID(ctx context.Context, groupID string) ([]*dto.TestCaseResponseDTO, error) {
	l.mutex.Lock()

	// キャッシュにあるか確認
	if cases, exists := l.cache[groupID]; exists {
		l.mutex.Unlock()
		return cases, nil
	}

	// キャッシュになければDBから取得
	l.mutex.Unlock()
	cases, err := l.useCase.GetCasesByGroupID(ctx, groupID)
	if err != nil {
		return nil, err
	}

	// 取得したデータをキャッシュに保存
	l.mutex.Lock()
	l.cache[groupID] = cases
	l.mutex.Unlock()

	return cases, nil
}

// LoadMany は複数のグループIDに対応するケースを一括で取得します
// 現在は単純に並列処理していますが、将来的にはバッチクエリに最適化可能です
func (l *TestCaseLoader) LoadMany(ctx context.Context, groupIDs []string) ([][]*dto.TestCaseResponseDTO, []error) {
	results := make([][]*dto.TestCaseResponseDTO, len(groupIDs))
	errors := make([]error, len(groupIDs))

	var wg sync.WaitGroup
	for i, id := range groupIDs {
		wg.Add(1)
		go func(idx int, groupID string) {
			defer wg.Done()
			cases, err := l.GetCasesByGroupID(ctx, groupID)
			results[idx] = cases
			errors[idx] = err
		}(i, id)
	}
	wg.Wait()

	return results, errors
}
