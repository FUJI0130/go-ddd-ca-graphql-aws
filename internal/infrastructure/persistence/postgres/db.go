package postgres

import (
	"database/sql"
	"time"

	_ "github.com/lib/pq"
)

// DBConfig はデータベース接続設定を保持します
type DBConfig struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// NewDB はPostgreSQLデータベース接続を作成します
func NewDB(config DBConfig) (*sql.DB, error) {
	db, err := sql.Open("postgres", config.DSN)
	if err != nil {
		return nil, err
	}

	// デフォルト値の設定
	maxOpenConns := 25
	if config.MaxOpenConns > 0 {
		maxOpenConns = config.MaxOpenConns
	}

	maxIdleConns := 25
	if config.MaxIdleConns > 0 {
		maxIdleConns = config.MaxIdleConns
	}

	connMaxLifetime := 5 * time.Minute
	if config.ConnMaxLifetime > 0 {
		connMaxLifetime = config.ConnMaxLifetime
	}

	// コネクションプールの設定
	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(connMaxLifetime)

	// 接続確認
	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

// NewPostgresDB はDSN文字列からPostgreSQLデータベース接続を作成します（互換性のため）
func NewPostgresDB(dsn string) (*sql.DB, error) {
	config := DBConfig{
		DSN: dsn,
	}
	return NewDB(config)
}
