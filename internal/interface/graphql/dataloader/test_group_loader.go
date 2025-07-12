// ファイル: internal/interface/graphql/dataloader/test_group_loader.go

package dataloader

import (
	"context"
	"sync"
	"time"

	"github.com/FUJI0130/go-ddd-ca/internal/usecase/dto"
	"github.com/FUJI0130/go-ddd-ca/internal/usecase/port"
)

// TestGroupLoader はテストグループをバッチ処理するためのデータローダーです
type TestGroupLoader struct {
	// 実際のデータを取得するためのユースケース
	useCase port.TestGroupUseCase

	// キャッシュとロックオブジェクト
	cache map[string][]*dto.TestGroupResponseDTO
	mutex sync.Mutex

	// バッチ処理のためのパラメータ
	maxBatchSize int
	wait         time.Duration
}

// NewTestGroupLoader は新しいTestGroupLoaderインスタンスを作成します
func NewTestGroupLoader(useCase port.TestGroupUseCase) *TestGroupLoader {
	return &TestGroupLoader{
		useCase:      useCase,
		cache:        make(map[string][]*dto.TestGroupResponseDTO),
		maxBatchSize: 100,
		wait:         1 * time.Millisecond,
	}
}

// Clear はキャッシュをクリアします
func (l *TestGroupLoader) Clear() {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.cache = make(map[string][]*dto.TestGroupResponseDTO)
}

// GetGroupsBySuiteID はテストスイートIDに基づいてテストグループを取得します
// キャッシュが存在する場合はキャッシュから返し、なければDBから取得してキャッシュに保存します
func (l *TestGroupLoader) GetGroupsBySuiteID(ctx context.Context, suiteID string) ([]*dto.TestGroupResponseDTO, error) {
	l.mutex.Lock()

	// キャッシュにあるか確認
	if groups, exists := l.cache[suiteID]; exists {
		l.mutex.Unlock()
		return groups, nil
	}

	// キャッシュになければDBから取得
	l.mutex.Unlock()
	groups, err := l.useCase.GetGroupsBySuiteID(ctx, suiteID)
	if err != nil {
		return nil, err
	}

	// 取得したデータをキャッシュに保存
	l.mutex.Lock()
	l.cache[suiteID] = groups
	l.mutex.Unlock()

	return groups, nil
}

// LoadMany は複数のテストスイートIDに対応するグループを一括で取得します
// 現在は単純に並列処理していますが、将来的にはバッチクエリに最適化可能です
func (l *TestGroupLoader) LoadMany(ctx context.Context, suiteIDs []string) ([][]*dto.TestGroupResponseDTO, []error) {
	results := make([][]*dto.TestGroupResponseDTO, len(suiteIDs))
	errors := make([]error, len(suiteIDs))

	var wg sync.WaitGroup
	for i, id := range suiteIDs {
		wg.Add(1)
		go func(idx int, suiteID string) {
			defer wg.Done()
			groups, err := l.GetGroupsBySuiteID(ctx, suiteID)
			results[idx] = groups
			errors[idx] = err
		}(i, id)
	}
	wg.Wait()

	return results, errors
}
