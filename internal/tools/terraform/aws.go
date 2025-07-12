package terraform

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/FUJI0130/go-ddd-ca/internal/tools/terraform/implementation"
	"github.com/FUJI0130/go-ddd-ca/internal/tools/terraform/interfaces"
	"github.com/FUJI0130/go-ddd-ca/internal/tools/terraform/logger"
	"github.com/FUJI0130/go-ddd-ca/internal/tools/terraform/models"
)

// handleResourceError はリソースエラーを処理する
func handleResourceError(err error, resourceType string, opts models.VerifyOptions) (int, error) {
	if err == nil {
		return 0, nil
	}

	if isResourceNotFoundError(err) {
		if opts.IgnoreResourceErrors {
			logger.WithField("resource", resourceType).Warn("リソースが存在しません: %v", err)
			logger.WithField("resource", resourceType).Info("環境をセットアップするには: make start-api-dev TF_ENV=%s", opts.Environment)
			return 0, nil // リソース数0で処理継続
		}
		// 無視しない場合はエラーを返す（ただしWarningレベルでログ出力）
		logger.WithField("resource", resourceType).Warn("リソースが存在しません: %v", err)
		return 0, fmt.Errorf("%sが見つかりません: %w", resourceType, err)
	}

	// その他のエラーはそのまま返す（Errorレベルでログ出力）
	logger.WithField("resource", resourceType).Error("検索に失敗: %v", err)
	return 0, err
}

// GetAWSResources はAWS環境からリソース情報を取得する
// 指定された環境のすべてのAWSリソース（VPC、RDS、ECSクラスター、サービス）を取得します
// エラーが発生した場合は、nilとエラーを返します
func GetAWSResources(runner interfaces.AWSCommandRunner, env string, opts models.VerifyOptions) (*models.Resources, error) {
	logger.Debug("AWS環境のリソース情報取得を開始: 環境=%s", env)

	resources := &models.Resources{
		Services: make(map[string]models.ServiceResources),
	}

	// リソース依存関係の定義
	resourceOrder := []struct {
		name      string
		fetchFunc func() (int, error)
		dependsOn []string
	}{
		{
			name: "VPC",
			fetchFunc: func() (int, error) {
				return GetVPCCount(runner, env)
			},
		},
		{
			name: "RDS",
			fetchFunc: func() (int, error) {
				return GetRDSCount(runner, env)
			},
			dependsOn: []string{"VPC"},
		},
		{
			name: "ECSクラスター",
			fetchFunc: func() (int, error) {
				return GetECSClusterCount(runner, env)
			},
			dependsOn: []string{"VPC"},
		},
	}

	// リソースの依存関係チェックとフェッチ
	skipMap := make(map[string]bool)

	for _, res := range resourceOrder {
		// 依存関係チェック
		shouldSkip := false
		for _, dep := range res.dependsOn {
			if skipMap[dep] {
				logger.WithFields(map[string]interface{}{
					"resource":   res.name,
					"depends_on": dep,
				}).Info("依存リソースがないためスキップします")
				skipMap[res.name] = true
				shouldSkip = true
				break
			}
		}

		if shouldSkip {
			continue
		}

		// リソース取得
		count, err := res.fetchFunc()
		if err != nil {
			// エラーハンドリング
			if isResourceNotFoundError(err) {
				if opts.IgnoreResourceErrors {
					logger.WithField("resource", res.name).Warn("リソースが存在しません: %v", err)
					logger.WithField("resource", res.name).Info("環境をセットアップするには: make start-api-dev TF_ENV=%s", env)

					// リソース数を0にセット
					count = 0

					// 後続のリソース検証をスキップするフラグを設定
					skipMap[res.name] = true
				} else {
					// IgnoreResourceErrors=falseなら従来通りエラーを返す
					logger.Error("%s検索コマンドの実行に失敗: %v", res.name, err)
					return nil, fmt.Errorf("failed to get %s: %v", res.name, err)
				}
			} else {
				// リソース不在エラー以外は常にエラーを返す
				logger.Error("%s検索コマンドの実行に失敗: %v", res.name, err)
				return nil, fmt.Errorf("failed to get %s: %v", res.name, err)
			}
		}

		// リソース数を設定
		switch res.name {
		case "VPC":
			resources.VPC = count
		case "RDS":
			resources.RDS = count
		case "ECSクラスター":
			resources.ECSCluster = count
		}

		logger.Debug("%s数: %d", res.name, count)
	}

	// サービス関連リソースはECSクラスターがあるときのみ検証
	if !skipMap["ECSクラスター"] && resources.ECSCluster > 0 {
		for _, serviceType := range []string{"api", "graphql", "grpc"} {
			logger.Debug("サービスタイプ '%s' のリソース情報を取得中...", serviceType)
			// serviceResources, err := GetServiceResources(runner, env, serviceType)
			serviceResources, err := GetServiceResources(runner, env, serviceType, opts)
			if err != nil {
				if opts.IgnoreResourceErrors {
					logger.WithField("service", serviceType).Warn("サービスリソース取得エラー: %v", err)
					resources.Services[serviceType] = models.ServiceResources{}
				} else {
					logger.Error("サービス '%s' のリソース取得に失敗: %v", serviceType, err)
					return nil, fmt.Errorf("failed to get service resources for %s: %v", serviceType, err)
				}
			} else {
				resources.Services[serviceType] = serviceResources
				logger.Debug("サービス '%s' のリソース: ECSサービス=%d, ALB=%d, ターゲットグループ=%d",
					serviceType, serviceResources.ECSService, serviceResources.ALB, serviceResources.TargetGroup)
			}
		}
	} else {
		logger.Info("ECSクラスターが存在しないため、サービスリソースの検証をスキップします")
	}

	logger.Info("AWS環境のリソース情報取得が完了しました")
	logger.Trace("AWS環境のリソース詳細: %+v", resources)
	return resources, nil
}

// isResourceNotFoundError はリソースが存在しないエラーかどうかを判定する
func isResourceNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	return strings.Contains(errMsg, "ClusterNotFoundException") ||
		strings.Contains(errMsg, "VpcNotFound") ||
		strings.Contains(errMsg, "DBInstanceNotFound") ||
		strings.Contains(errMsg, "ServiceNotFoundException")
}

// GetVPCCount はVPC数を取得する
// 指定された環境のVPC数を取得します
// エラーが発生した場合は、0とエラーを返します
func GetVPCCount(runner interfaces.AWSCommandRunner, env string) (int, error) {
	filters := fmt.Sprintf("Name=tag:Environment,Values=%s", env)
	query := "length(Vpcs)"
	output := "text"

	logger.Debug("VPCの検索クエリ: --filters %s --query %s --output %s", filters, query, output)

	output, err := runner.RunCommand("ec2", "describe-vpcs",
		"--filters", filters,
		"--query", query,
		"--output", output)
	if err != nil {
		logger.Error("VPC検索コマンドの実行に失敗: %v", err)
		return 0, err
	}

	logger.Trace("AWS CLI出力 (VPC): %s", output)

	count, err := strconv.Atoi(output)
	if err != nil {
		logger.Error("VPC数のパースに失敗: %v", err)
		return 0, fmt.Errorf("failed to parse VPC count: %v", err)
	}

	logger.Debug("検出されたVPC数: %d", count)
	return count, nil
}

// GetRDSCount はRDSインスタンス数を取得する
// 指定された環境のRDSインスタンス数を取得します
// エラーが発生した場合は、0とエラーを返します
func GetRDSCount(runner interfaces.AWSCommandRunner, env string) (int, error) {
	query := fmt.Sprintf("length(DBInstances[?DBInstanceIdentifier=='%s-postgres'])", env)

	logger.Debug("RDSの検索クエリ: --query %s --output text", query)

	output, err := runner.RunCommand("rds", "describe-db-instances",
		"--query", query,
		"--output", "text")
	if err != nil {
		logger.Error("RDS検索コマンドの実行に失敗: %v", err)
		return 0, err
	}

	logger.Trace("AWS CLI出力 (RDS): %s", output)

	count, err := strconv.Atoi(output)
	if err != nil {
		logger.Error("RDS数のパースに失敗: %v", err)
		return 0, fmt.Errorf("failed to parse RDS count: %v", err)
	}

	logger.Debug("検出されたRDS数: %d", count)
	return count, nil
}

// GetECSClusterCount はECSクラスター数を取得する
// 指定された環境のECSクラスター数を取得します
// エラーが発生した場合は、0とエラーを返します
func GetECSClusterCount(runner interfaces.AWSCommandRunner, env string) (int, error) {
	query := fmt.Sprintf("length(clusterArns[?contains(@,'%s-shared-cluster')])", env)

	logger.Debug("ECSクラスターの検索クエリ: --query %s --output text", query)

	output, err := runner.RunCommand("ecs", "list-clusters",
		"--query", query,
		"--output", "text")
	if err != nil {
		logger.Error("ECSクラスター検索コマンドの実行に失敗: %v", err)
		return 0, err
	}

	logger.Trace("AWS CLI出力 (ECSクラスター): %s", output)

	count, err := strconv.Atoi(output)
	if err != nil {
		logger.Error("ECSクラスター数のパースに失敗: %v", err)
		return 0, fmt.Errorf("failed to parse ECS cluster count: %v", err)
	}

	logger.Debug("検出されたECSクラスター数: %d", count)
	return count, nil
}

// GetServiceResources はサービス関連リソース数を取得する
// 指定された環境とサービスタイプに対するリソース数（ECSサービス、ALB、ターゲットグループ）を取得します
// エラーが発生した場合は、部分的に入力されたServiceResourcesとエラーを返します
func GetServiceResources(runner interfaces.AWSCommandRunner, env, serviceType string, opts models.VerifyOptions) (models.ServiceResources, error) {
	logger.Debug("サービスリソースの取得: 環境=%s, サービスタイプ=%s, サフィックス=%s", env, serviceType, opts.ServiceSuffix)

	serviceResources := models.ServiceResources{}

	// サフィックスを含んだサービス名のパターンを作成
	serviceName := fmt.Sprintf("%s-%s%s", env, serviceType, opts.ServiceSuffix)
	logger.Debug("検索用サービス名パターン: %s", serviceName)

	// ECSサービス数の取得
	serviceQuery := fmt.Sprintf("length(serviceArns[?contains(@,'%s')])", serviceName)
	logger.Debug("ECSサービス検索クエリ: --cluster %s-shared-cluster --query %s", env, serviceQuery)

	serviceOutput, err := runner.RunCommand("ecs", "list-services",
		"--cluster", fmt.Sprintf("%s-shared-cluster", env),
		"--query", serviceQuery,
		"--output", "text")
	if err != nil {
		logger.Error("ECSサービス検索コマンドの実行に失敗: %v", err)
		return serviceResources, err
	}

	logger.Trace("AWS CLI出力 (ECSサービス): %s", serviceOutput)

	serviceCount, err := strconv.Atoi(serviceOutput)
	if err != nil {
		logger.Error("ECSサービス数のパースに失敗: %v", err)
		return serviceResources, fmt.Errorf("failed to parse ECS service count: %v", err)
	}
	serviceResources.ECSService = serviceCount
	logger.Debug("検出されたECSサービス数: %d", serviceCount)

	// ALB名を生成（サフィックス付き）
	albName := fmt.Sprintf("%s-alb", serviceName)

	// ALB数の取得
	albQuery := fmt.Sprintf("length(LoadBalancers[?LoadBalancerName=='%s'])", albName)
	logger.Debug("ALB検索クエリ: --query %s", albQuery)

	albOutput, err := runner.RunCommand("elbv2", "describe-load-balancers",
		"--query", albQuery,
		"--output", "text")
	if err != nil {
		logger.Error("ALB検索コマンドの実行に失敗: %v", err)
		return serviceResources, err
	}

	logger.Trace("AWS CLI出力 (ALB): %s", albOutput)

	albCount, err := strconv.Atoi(albOutput)
	if err != nil {
		logger.Error("ALB数のパースに失敗: %v", err)
		return serviceResources, fmt.Errorf("failed to parse ALB count: %v", err)
	}
	serviceResources.ALB = albCount
	logger.Debug("検出されたALB数: %d", albCount)

	// ターゲットグループ名を生成（サフィックス付き）
	tgName := fmt.Sprintf("%s-tg", serviceName)

	// ターゲットグループ数の取得
	tgQuery := fmt.Sprintf("length(TargetGroups[?TargetGroupName=='%s'])", tgName)
	logger.Debug("ターゲットグループ検索クエリ: --query %s", tgQuery)

	tgOutput, err := runner.RunCommand("elbv2", "describe-target-groups",
		"--query", tgQuery,
		"--output", "text")
	if err != nil {
		logger.Error("ターゲットグループ検索コマンドの実行に失敗: %v", err)
		return serviceResources, err
	}

	logger.Trace("AWS CLI出力 (ターゲットグループ): %s", tgOutput)

	tgCount, err := strconv.Atoi(tgOutput)
	if err != nil {
		logger.Error("ターゲットグループ数のパースに失敗: %v", err)
		return serviceResources, fmt.Errorf("failed to parse target group count: %v", err)
	}
	serviceResources.TargetGroup = tgCount
	logger.Debug("検出されたターゲットグループ数: %d", tgCount)

	return serviceResources, nil
}

// RunAWSCommand は後方互換性のためのヘルパー関数
// 非推奨: 代わりに AWSCommandRunner インターフェースと DefaultAWSRunner を使用してください
// このメソッドは後方互換性のためにのみ存在し、新しいコードでは使用すべきではありません
func RunAWSCommand(args ...string) (string, error) {
	logger.Warn("非推奨の RunAWSCommand メソッドが呼び出されました。代わりに AWSCommandRunner インターフェースを使用してください")
	runner := implementation.NewDefaultAWSRunner()
	return runner.RunCommand(args...)
}
