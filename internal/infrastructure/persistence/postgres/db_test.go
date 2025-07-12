package postgres

import (
	"testing"
	"time"
)

func TestNewDB(t *testing.T) {
	// テスト用のDBConfig
	config := DBConfig{
		DSN:             "postgresql://test_user:test_pass@localhost:5433/test_db?sslmode=disable", // テスト用DBを使用
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 3 * time.Minute,
	}

	// 接続テスト
	db, err := NewDB(config)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// 接続確認
	err = db.Ping()
	if err != nil {
		t.Fatalf("Failed to ping database: %v", err)
	}

	// 設定値の確認
	if db.Stats().MaxOpenConnections != 10 {
		t.Errorf("Expected MaxOpenConnections to be 10, got %d", db.Stats().MaxOpenConnections)
	}
}

func TestNewPostgresDB(t *testing.T) {
	// DSNのみを指定
	dsn := "postgresql://test_user:test_pass@localhost:5433/test_db?sslmode=disable" // テスト用DBを使用

	// 接続テスト
	db, err := NewPostgresDB(dsn)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// 接続確認
	err = db.Ping()
	if err != nil {
		t.Fatalf("Failed to ping database: %v", err)
	}

	// デフォルト設定値の確認
	if db.Stats().MaxOpenConnections != 25 {
		t.Errorf("Expected default MaxOpenConnections to be 25, got %d", db.Stats().MaxOpenConnections)
	}
}
