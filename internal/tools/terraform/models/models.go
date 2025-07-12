package models

import (
	"time"

	"github.com/FUJI0130/go-ddd-ca/internal/tools/terraform/config"
	"github.com/FUJI0130/go-ddd-ca/internal/tools/terraform/logger"
)

// リソース構造体
type Resources struct {
	VPC        int
	RDS        int
	ECSCluster int
	Services   map[string]ServiceResources
}

// サービスリソース構造体
type ServiceResources struct {
	ECSService  int
	ALB         int
	TargetGroup int
}

// 比較結果
type ComparisonResult struct {
	ResourceName   string
	AWSCount       int
	TerraformCount int
	IsMatch        bool
}

// VerifyOptions は検証オプションを保持する構造体
type VerifyOptions struct {
	Environment          string
	Debug                bool
	SkipTerraformPlan    bool
	ForceCleanup         bool
	ConfigProvider       config.ConfigProvider
	LogLevel             logger.LogLevel
	Timeout              time.Duration // 追加: タイムアウト時間
	IgnoreResourceErrors bool          // 追加: リソース不在エラーを無視するオプション
	ServiceSuffix        string        // サービス名のサフィックス（例: "-new"）
}

type ResourceType struct {
	Name      string   // リソース名
	DependsOn []string // 依存するリソース名
}
