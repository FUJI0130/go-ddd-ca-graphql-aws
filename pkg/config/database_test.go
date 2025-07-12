// pkg/config/database_test.go
package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDatabaseConfigFromProvider(t *testing.T) {
	// 静的プロバイダーのテスト
	t.Run("静的プロバイダーからのデータベース設定", func(t *testing.T) {
		staticProvider := NewStaticConfigProvider(map[string]interface{}{
			"database.driver":   "postgres",
			"database.host":     "static-host",
			"database.port":     5678,
			"database.user":     "static-user",
			"database.password": "static-pass",
			"database.dbname":   "static-db",
			"database.sslmode":  "require",
		})

		dbConfig := NewDatabaseConfigFromProvider(staticProvider)
		assert.Equal(t, "postgres", dbConfig.Driver)
		assert.Equal(t, "static-host", dbConfig.Host)
		assert.Equal(t, 5678, dbConfig.Port)
		assert.Equal(t, "static-user", dbConfig.User)
		assert.Equal(t, "static-pass", dbConfig.Password)
		assert.Equal(t, "static-db", dbConfig.DBName)
		assert.Equal(t, "require", dbConfig.SSLMode)
	})

	// 環境変数プロバイダーのテスト
	t.Run("環境変数プロバイダーからのデータベース設定", func(t *testing.T) {
		os.Setenv("DB_HOST", "env-host")
		os.Setenv("DB_PORT", "9012")
		os.Setenv("DB_USER", "env-user")
		os.Setenv("DB_PASSWORD", "env-pass")
		os.Setenv("DB_NAME", "env-db")
		os.Setenv("DB_SSLMODE", "disable")

		envProvider := NewEnvConfigProvider("")
		dbConfig := NewDatabaseConfigFromProvider(envProvider)
		assert.Equal(t, "postgres", dbConfig.Driver) // デフォルト値
		assert.Equal(t, "env-host", dbConfig.Host)
		assert.Equal(t, 9012, dbConfig.Port)
		assert.Equal(t, "env-user", dbConfig.User)
		assert.Equal(t, "env-pass", dbConfig.Password)
		assert.Equal(t, "env-db", dbConfig.DBName)
		assert.Equal(t, "disable", dbConfig.SSLMode)

		// 環境変数をクリア
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("DB_NAME")
		os.Unsetenv("DB_SSLMODE")
	})

	// チェーンプロバイダーでの環境変数優先のテスト
	t.Run("チェーンプロバイダーでの環境変数優先テスト", func(t *testing.T) {
		// 環境変数を設定
		os.Setenv("DB_HOST", "env-host")

		// 静的プロバイダーを設定
		staticProvider := NewStaticConfigProvider(map[string]interface{}{
			"database.host": "static-host",
		})

		// 環境変数プロバイダーを設定
		envProvider := NewEnvConfigProvider("")

		// チェーンプロバイダーを作成（環境変数を優先）
		chainedProvider := NewChainedConfigProvider(envProvider, staticProvider)

		dbConfig := NewDatabaseConfigFromProvider(chainedProvider)
		assert.Equal(t, "env-host", dbConfig.Host)

		// 環境変数をクリア
		os.Unsetenv("DB_HOST")
	})

	// デフォルト値のテスト
	t.Run("デフォルト値のテスト", func(t *testing.T) {
		emptyProvider := NewEmptyConfigProvider()
		dbConfig := NewDatabaseConfigFromProvider(emptyProvider)
		assert.Equal(t, "postgres", dbConfig.Driver)
		assert.Equal(t, "localhost", dbConfig.Host)
		assert.Equal(t, 5432, dbConfig.Port)
		assert.Equal(t, "postgres", dbConfig.User)
		assert.Equal(t, "", dbConfig.Password)
		assert.Equal(t, "postgres", dbConfig.DBName)
		assert.Equal(t, "disable", dbConfig.SSLMode)
	})
}

func TestDSN(t *testing.T) {
	// DSN生成のテスト
	t.Run("DSN生成", func(t *testing.T) {
		dbConfig := &DatabaseConfig{
			Driver:   "postgres",
			Host:     "test-host",
			Port:     5432,
			User:     "test-user",
			Password: "test-pass",
			DBName:   "test-db",
			SSLMode:  "disable",
		}

		expected := "host=test-host port=5432 user=test-user password=test-pass dbname=test-db sslmode=disable"
		assert.Equal(t, expected, dbConfig.DSN())
	})
}
