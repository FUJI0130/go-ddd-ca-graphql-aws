// internal/infrastructure/persistence/common/db.go
package common

import (
	"context"
	"database/sql"
)

// SQLExecutor はデータベース操作を抽象化するインターフェース
// *sql.DB と *sql.Tx の両方で実装される共通のメソッドを定義
type SQLExecutor interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Query(query string, args ...interface{}) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}
