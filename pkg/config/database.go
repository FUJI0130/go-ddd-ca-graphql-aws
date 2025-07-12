package config

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/FUJI0130/go-ddd-ca/internal/infrastructure/persistence/postgres"
)

// DatabaseConfig はデータベース接続の設定を保持します
type DatabaseConfig struct {
	Driver   string
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// NewDatabaseConfigFromProvider は設定プロバイダーからデータベース設定を作成します
func NewDatabaseConfigFromProvider(provider ConfigProvider) *DatabaseConfig {
	// データベース設定をプロバイダーから一貫して取得
	// 環境変数のマッピングはプロバイダー内部で処理される
	config := &DatabaseConfig{
		Driver:   provider.GetString("database.driver", "postgres"),
		Host:     provider.GetString("database.host", "localhost"),
		Port:     provider.GetInt("database.port", 5432),
		User:     provider.GetString("database.user", "postgres"),
		Password: provider.GetString("database.password", ""),
		DBName:   provider.GetString("database.dbname", "postgres"),
		SSLMode:  provider.GetString("database.sslmode", "disable"),
	}

	// ソース情報のログ出力（パスワードは除く）
	log.Printf("Database connection settings from %s: host=%s, port=%d, user=%s, dbname=%s, sslmode=%s",
		provider.Source(), config.Host, config.Port, config.User, config.DBName, config.SSLMode)

	return config
}

// NewDatabaseConnection はデータベース接続を作成します
func (c *DatabaseConfig) NewDatabaseConnection() (*sql.DB, error) {
	// PostgreSQL接続文字列の生成
	dsn := fmt.Sprintf(
		"postgresql://%s:%s@%s:%d/%s?sslmode=%s",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.DBName,
		c.SSLMode,
	)

	// パスワードを隠した接続情報のログ出力
	log.Printf("Connecting to: postgresql://%s:******@%s:%d/%s?sslmode=%s",
		c.User, c.Host, c.Port, c.DBName, c.SSLMode)

	dbConfig := postgres.DBConfig{
		DSN:             dsn,
		MaxOpenConns:    25,
		MaxIdleConns:    25,
		ConnMaxLifetime: 5 * time.Minute,
	}

	return postgres.NewDB(dbConfig)
}

// 既存のConfig構造体からデータベース接続を作成するメソッド（後方互換性のため維持）
func (c *Config) NewDatabaseConnection() (*sql.DB, error) {
	return c.Database.NewDatabaseConnection()
}

// DSN はデータベース接続文字列を返します（後方互換性のために維持）
func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}
