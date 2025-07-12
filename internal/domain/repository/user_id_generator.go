package repository

import (
	"context"
)

// UserIDGenerator はユーザーIDを生成するインターフェース
type UserIDGenerator interface {
	// Generate は新しいユーザーIDを生成する
	Generate(ctx context.Context) (string, error)
}
