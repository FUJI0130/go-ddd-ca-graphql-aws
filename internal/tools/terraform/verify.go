package terraform

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/FUJI0130/go-ddd-ca/internal/tools/terraform/implementation"
	"github.com/FUJI0130/go-ddd-ca/internal/tools/terraform/interfaces"
	"github.com/FUJI0130/go-ddd-ca/internal/tools/terraform/logger"
	"github.com/FUJI0130/go-ddd-ca/internal/tools/terraform/models"
)

// GetTerraformResources はTerraform状態からリソース情報を取得する
func GetTerraformResources(env string, fs interfaces.FileSystem, cmdExecutor interfaces.CommandExecutor, opts models.VerifyOptions) (*models.Resources, error) {
	logger.Debug("Terraform状態からリソース情報取得を開始: 環境=%s, サフィックス=%s", env, opts.ServiceSuffix)

	resources := &models.Resources{
		Services: make(map[string]models.ServiceResources),
	}

	// 現在のディレクトリを保存
	currentDir, err := fs.Getwd()
	if err != nil {
		logger.Error("現在のディレクトリ取得に失敗: %v", err)
		return nil, err
	}
	logger.Debug("現在のディレクトリ: %s", currentDir)

	// 環境ディレクトリに移動
	envDir := filepath.Join("deployments", "terraform", "environments", env)
	logger.Debug("環境ディレクトリに移動: %s", envDir)

	err = fs.Chdir(envDir)
	if err != nil {
		logger.Error("環境ディレクトリへの移動に失敗: %s", envDir)
		return nil, fmt.Errorf("環境ディレクトリが見つかりません: %s", envDir)
	}

	// 処理完了後に元のディレクトリに戻ることを保証
	defer fs.Chdir(currentDir)

	// Terraformの状態をJSON形式で取得
	logger.Debug("terraform show -json コマンドを実行")
	output, err := cmdExecutor.Execute("terraform", "show", "-json")
	if err != nil {
		logger.Warn("terraform show -json コマンドでエラーが発生: %v", err)
		logger.Info("空のリソース情報を返します")
		return resources, nil
	}

	// JSONデータを解析
	var tfState map[string]interface{}
	if err := json.Unmarshal([]byte(output), &tfState); err != nil {
		logger.Error("JSON解析エラー: %v", err)
		logger.Debug("JSON解析エラー時のリソース状態: %+v", resources) // 追加
		return resources, nil                             // エラーがあっても空のリソース情報を返す
	}

	// リソースタイプごとのIDマップを初期化（重複カウント防止用）
	resourceIDs := make(map[string]map[string]bool)

	// リソース抽出関数（opts引数を追加）
	extractResources(tfState, resourceIDs, resources, opts)

	logger.Info("Terraform状態からのリソース情報取得が完了")
	logger.Trace("Terraformリソース詳細: %+v", resources)
	return resources, nil
}

// 再帰的にリソースを抽出する関数
// extractResources関数にデバッグログを追加
func extractResources(tfState map[string]interface{}, resourceIDs map[string]map[string]bool, resources *models.Resources, opts models.VerifyOptions) {
	logger.Debug("Terraform状態のJSONパース開始")

	// values構造の確認
	values, ok := tfState["values"].(map[string]interface{})
	if !ok {
		logger.Error("JSONパース: 'values'キーが存在しないか、予期せぬ型です")
		return
	}

	// root_module構造の確認
	rootModule, ok := values["root_module"].(map[string]interface{})
	if !ok {
		logger.Error("JSONパース: 'values.root_module'キーが存在しないか、予期せぬ型です")
		return
	}

	// モジュールからリソースを抽出（opts引数を追加）
	logger.Debug("モジュール処理を開始")
	processModule(rootModule, "", resourceIDs, resources, opts)
}

// processModule関数にデバッグログを追加し、モジュール処理を行う
func processModule(module map[string]interface{}, path string, resourceIDs map[string]map[string]bool, resources *models.Resources, opts models.VerifyOptions) {
	if path != "" {
		logger.Debug("モジュール処理: %s", path)
	} else {
		logger.Debug("ルートモジュール処理")
	}

	// リソース処理
	if resourceList, ok := module["resources"].([]interface{}); ok {
		logger.Debug("モジュールのリソース数: %d", len(resourceList))
		for i, res := range resourceList {
			resource, ok := res.(map[string]interface{})
			if !ok {
				logger.Warn("リソース[%d]がマップ型でない", i)
				continue
			}

			// リソースタイプとアドレスを取得
			resType, typeOk := resource["type"].(string)
			address, addrOk := resource["address"].(string)
			if !typeOk || !addrOk {
				logger.Warn("リソース[%d]にtype/addressがない: typeOk=%v, addrOk=%v", i, typeOk, addrOk)
				continue
			}

			logger.Debug("リソース処理: type=%s, address=%s", resType, address)

			// リソースのIDを抽出
			var resourceID string
			if values, ok := resource["values"].(map[string]interface{}); ok {
				if id, ok := values["id"].(string); ok {
					resourceID = id
					logger.Debug("リソースID: %s", resourceID)
				} else {
					logger.Debug("リソースにIDがない")
				}
			} else {
				logger.Debug("リソースにvaluesがない")
			}

			// リソースIDが空の場合はアドレスを使用
			if resourceID == "" {
				resourceID = address
				logger.Debug("IDが空なのでアドレスを使用: %s", resourceID)
			}

			// リソースタイプに基づいて処理（opts引数を追加）
			processResourceByType(resType, resourceID, address, resourceIDs, resources, opts)
		}
	} else {
		logger.Debug("モジュールにリソースがない")
	}

	// 子モジュールの処理
	if childModules, ok := module["child_modules"].([]interface{}); ok {
		logger.Debug("子モジュール数: %d", len(childModules))
		for i, child := range childModules {
			childModule, ok := child.(map[string]interface{})
			if !ok {
				logger.Warn("子モジュール[%d]がマップ型でない", i)
				continue
			}

			childAddress, ok := childModule["address"].(string)
			if !ok {
				childAddress = path + ".child"
				logger.Debug("子モジュール[%d]にアドレスがない、デフォルト使用: %s", i, childAddress)
			} else {
				logger.Debug("子モジュール[%d]アドレス: %s", i, childAddress)
			}

			// 子モジュールの再帰的処理（opts引数を追加）
			processModule(childModule, childAddress, resourceIDs, resources, opts)
		}
	} else {
		logger.Debug("モジュールに子モジュールがない")
	}
}

// processResourceByType関数にデバッグログを追加
// processResourceByType はリソースタイプに基づいてリソースを処理する
func processResourceByType(resType, resourceID, address string, resourceIDs map[string]map[string]bool, resources *models.Resources, opts models.VerifyOptions) {
	logger.Debug("リソースタイプ処理: type=%s, id=%s, address=%s", resType, resourceID, address)

	// リソースタイプごとのIDマップがない場合は初期化
	if _, ok := resourceIDs[resType]; !ok {
		resourceIDs[resType] = make(map[string]bool)
		logger.Debug("リソースタイプのIDマップを初期化: %s", resType)
	}

	// すでに処理済みのリソースIDであればスキップ
	if resourceIDs[resType][resourceID] {
		logger.Debug("すでに処理済みのリソースIDなのでスキップ: %s", resourceID)
		return
	}

	// 処理済みとしてマーク
	resourceIDs[resType][resourceID] = true
	logger.Debug("リソースを処理済みとしてマーク: type=%s, id=%s", resType, resourceID)

	// リソースタイプに応じてカウント
	switch resType {
	case "aws_vpc":
		resources.VPC++
		logger.Debug("VPCカウント増加: 現在=%d", resources.VPC)
	case "aws_db_instance":
		resources.RDS++
		logger.Debug("RDSカウント増加: 現在=%d", resources.RDS)
	case "aws_ecs_cluster":
		resources.ECSCluster++
		logger.Debug("ECSクラスターカウント増加: 現在=%d", resources.ECSCluster)
	case "aws_ecs_service":
		// サービスタイプを判定（サフィックス対応）
		found := false
		for _, svcType := range []string{"api", "graphql", "grpc"} {
			if strings.Contains(address, svcType) {
				// サフィックスあり/なしの両方に対応
				suffix := opts.ServiceSuffix
				// サフィックスが設定されている場合、そのサフィックスを含むアドレスのみを処理
				// サフィックスが空の場合は、通常の（サフィックスなし）処理
				if (suffix != "" && strings.Contains(address, suffix)) ||
					(suffix == "" && !containsKnownSuffix(address)) {

					svc, exists := resources.Services[svcType]
					if !exists {
						svc = models.ServiceResources{}
						resources.Services[svcType] = svc
						logger.Debug("新しいサービスタイプを追加: %s", svcType)
					}
					svc.ECSService++
					resources.Services[svcType] = svc
					logger.Debug("%s ECSサービスカウント増加: 現在=%d, サフィックス=%s",
						svcType, svc.ECSService, suffix)
					found = true
					break
				}
			}
		}
		if !found {
			logger.Debug("ECSサービスのサービスタイプが不明またはサフィックス不一致: %s", address)
		}
	case "aws_lb":
		// ALBのサービスタイプを判定（サフィックス対応）
		found := false
		for _, svcType := range []string{"api", "graphql", "grpc"} {
			if strings.Contains(address, svcType) {
				// サフィックスあり/なしの両方に対応
				suffix := opts.ServiceSuffix
				if (suffix != "" && strings.Contains(address, suffix)) ||
					(suffix == "" && !containsKnownSuffix(address)) {

					svc, exists := resources.Services[svcType]
					if !exists {
						svc = models.ServiceResources{}
						resources.Services[svcType] = svc
						logger.Debug("新しいサービスタイプを追加: %s", svcType)
					}
					svc.ALB++
					resources.Services[svcType] = svc
					logger.Debug("%s ALBカウント増加: 現在=%d, サフィックス=%s",
						svcType, svc.ALB, suffix)
					found = true
					break
				}
			}
		}
		if !found {
			logger.Debug("ALBのサービスタイプが不明またはサフィックス不一致: %s", address)
		}
	case "aws_lb_target_group":
		// ターゲットグループのサービスタイプを判定（サフィックス対応）
		found := false
		for _, svcType := range []string{"api", "graphql", "grpc"} {
			if strings.Contains(address, svcType) {
				// サフィックスあり/なしの両方に対応
				suffix := opts.ServiceSuffix
				if (suffix != "" && strings.Contains(address, suffix)) ||
					(suffix == "" && !containsKnownSuffix(address)) {

					svc, exists := resources.Services[svcType]
					if !exists {
						svc = models.ServiceResources{}
						resources.Services[svcType] = svc
						logger.Debug("新しいサービスタイプを追加: %s", svcType)
					}
					svc.TargetGroup++
					resources.Services[svcType] = svc
					logger.Debug("%s ターゲットグループカウント増加: 現在=%d, サフィックス=%s",
						svcType, svc.TargetGroup, suffix)
					found = true
					break
				}
			}
		}
		if !found {
			logger.Debug("ターゲットグループのサービスタイプが不明またはサフィックス不一致: %s", address)
		}
	}
}

// containsKnownSuffix は既知のサフィックスがアドレスに含まれているかをチェックする
func containsKnownSuffix(address string) bool {
	knownSuffixes := []string{"-new"}
	for _, suffix := range knownSuffixes {
		if strings.Contains(address, suffix) {
			return true
		}
	}
	return false
}

// CompareResources はAWS環境とTerraform状態のリソース数を比較する
func CompareResources(awsResources, tfResources *models.Resources) []models.ComparisonResult {
	logger.Debug("リソース比較を開始")
	logger.Trace("AWS環境リソース: %+v", awsResources)
	logger.Trace("Terraform状態リソース: %+v", tfResources)

	var results []models.ComparisonResult

	// コアリソースの比較
	results = append(results, models.ComparisonResult{
		ResourceName:   "VPC",
		AWSCount:       awsResources.VPC,
		TerraformCount: tfResources.VPC,
		IsMatch:        awsResources.VPC == tfResources.VPC,
	})

	if awsResources.VPC != tfResources.VPC {
		logger.Warn("VPC数の不一致: AWS=%d, Terraform=%d", awsResources.VPC, tfResources.VPC)
	} else {
		logger.Debug("VPC数が一致しています: %d", awsResources.VPC)
	}

	results = append(results, models.ComparisonResult{
		ResourceName:   "RDS",
		AWSCount:       awsResources.RDS,
		TerraformCount: tfResources.RDS,
		IsMatch:        awsResources.RDS == tfResources.RDS,
	})

	if awsResources.RDS != tfResources.RDS {
		logger.Warn("RDS数の不一致: AWS=%d, Terraform=%d", awsResources.RDS, tfResources.RDS)
	} else {
		logger.Debug("RDS数が一致しています: %d", awsResources.RDS)
	}

	results = append(results, models.ComparisonResult{
		ResourceName:   "ECSクラスター",
		AWSCount:       awsResources.ECSCluster,
		TerraformCount: tfResources.ECSCluster,
		IsMatch:        awsResources.ECSCluster == tfResources.ECSCluster,
	})

	if awsResources.ECSCluster != tfResources.ECSCluster {
		logger.Warn("ECSクラスター数の不一致: AWS=%d, Terraform=%d", awsResources.ECSCluster, tfResources.ECSCluster)
	} else {
		logger.Debug("ECSクラスター数が一致しています: %d", awsResources.ECSCluster)
	}

	// サービスリソースの比較
	for _, serviceType := range []string{"api", "graphql", "grpc"} {
		awsSvc, awsExists := awsResources.Services[serviceType]
		tfSvc, tfExists := tfResources.Services[serviceType]

		if !awsExists {
			awsSvc = models.ServiceResources{}
		}

		if !tfExists {
			tfSvc = models.ServiceResources{}
		}

		// ECSサービス
		results = append(results, models.ComparisonResult{
			ResourceName:   fmt.Sprintf("%s-ECSサービス", serviceType),
			AWSCount:       awsSvc.ECSService,
			TerraformCount: tfSvc.ECSService,
			IsMatch:        awsSvc.ECSService == tfSvc.ECSService,
		})

		if awsSvc.ECSService != tfSvc.ECSService {
			logger.Warn("%s-ECSサービス数の不一致: AWS=%d, Terraform=%d", serviceType, awsSvc.ECSService, tfSvc.ECSService)
		} else {
			logger.Debug("%s-ECSサービス数が一致しています: %d", serviceType, awsSvc.ECSService)
		}

		// ALB
		results = append(results, models.ComparisonResult{
			ResourceName:   fmt.Sprintf("%s-ALB", serviceType),
			AWSCount:       awsSvc.ALB,
			TerraformCount: tfSvc.ALB,
			IsMatch:        awsSvc.ALB == tfSvc.ALB,
		})

		if awsSvc.ALB != tfSvc.ALB {
			logger.Warn("%s-ALB数の不一致: AWS=%d, Terraform=%d", serviceType, awsSvc.ALB, tfSvc.ALB)
		} else {
			logger.Debug("%s-ALB数が一致しています: %d", serviceType, awsSvc.ALB)
		}

		// ターゲットグループ
		results = append(results, models.ComparisonResult{
			ResourceName:   fmt.Sprintf("%s-ターゲットグループ", serviceType),
			AWSCount:       awsSvc.TargetGroup,
			TerraformCount: tfSvc.TargetGroup,
			IsMatch:        awsSvc.TargetGroup == tfSvc.TargetGroup,
		})

		if awsSvc.TargetGroup != tfSvc.TargetGroup {
			logger.Warn("%s-ターゲットグループ数の不一致: AWS=%d, Terraform=%d", serviceType, awsSvc.TargetGroup, tfSvc.TargetGroup)
		} else {
			logger.Debug("%s-ターゲットグループ数が一致しています: %d", serviceType, awsSvc.TargetGroup)
		}
	}

	// 不一致数のカウント
	mismatchCount := 0
	for _, result := range results {
		if !result.IsMatch {
			mismatchCount++
		}
	}

	logger.Info("リソース比較完了: 全%d項目中%d項目の不一致", len(results), mismatchCount)
	return results
}

// RunTerraformPlanWithContext はコンテキスト付きでterraform planを実行して差分を検証する
func RunTerraformPlanWithContext(ctx context.Context, env string, fs interfaces.FileSystem,
	cmdExecutor interfaces.ContextAwareCommandExecutor) (int, string, error) {

	logger.Debug("コンテキスト付きterraform planを実行: 環境=%s", env)

	// 現在のディレクトリを保存
	currentDir, err := fs.Getwd()
	if err != nil {
		logger.Error("現在のディレクトリ取得に失敗: %v", err)
		return 1, "", err
	}

	// 環境ディレクトリに移動
	envDir := filepath.Join("deployments", "terraform", "environments", env)
	logger.Debug("環境ディレクトリに移動: %s", envDir)

	err = fs.Chdir(envDir)
	if err != nil {
		logger.Error("環境ディレクトリへの移動に失敗: %s", envDir)
		return 1, "", fmt.Errorf("環境ディレクトリが見つかりません: %s", envDir)
	}

	// 処理完了後に元のディレクトリに戻ることを保証
	defer fs.Chdir(currentDir)

	// terraform planをコンテキスト付きで実行
	logger.Debug("terraform plan コマンドをコンテキスト付きで実行: -lock=false -input=false -detailed-exitcode")
	output, err := cmdExecutor.ExecuteWithContext(ctx, "terraform", "plan", "-lock=false", "-input=false", "-detailed-exitcode")

	// 終了コードの取得
	exitCode := 0
	if err != nil {
		// コンテキストエラーの場合
		if ctx.Err() != nil {
			return 1, output, err
		}

		// exec.ExitError型としての取得を試みる
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
			logger.Debug("terraform plan終了コード: %d", exitCode)
		} else {
			// テスト用のカスタムエラータイプに対応（ExitCode()メソッドを持つインターフェース）
			type ExitCoder interface {
				ExitCode() int
			}
			if exitCoder, ok := err.(ExitCoder); ok {
				exitCode = exitCoder.ExitCode()
				logger.Debug("カスタムエラーからの終了コード: %d", exitCode)
			} else {
				logger.Error("terraform plan実行中にエラー発生: %v", err)
				return 1, output, err
			}
		}
	}

	logger.Debug("terraform plan完了: 出力長=%d, 終了コード=%d", len(output), exitCode)
	logger.Trace("terraform plan出力: %s", output)
	return exitCode, output, nil
}

// RunTerraformPlan はterraform planを実行して差分を検証する
func RunTerraformPlan(env string, fs interfaces.FileSystem, cmdExecutor interfaces.CommandExecutor) (int, string, error) {
	// CommandExecutorをContextAwareCommandExecutorに変換
	var ctxExecutor interfaces.ContextAwareCommandExecutor
	if contextAware, ok := cmdExecutor.(interfaces.ContextAwareCommandExecutor); ok {
		ctxExecutor = contextAware
	} else {
		// 直接キャストできない場合はラップする
		ctxExecutor = implementation.NewCommandExecutorWrapper(cmdExecutor)
	}

	// 無期限コンテキストで実行
	return RunTerraformPlanWithContext(context.Background(), env, fs, ctxExecutor)

}

// VerifyStateWithContext はコンテキスト付きでAWS環境とTerraform状態の整合性を検証する
func VerifyStateWithContext(ctx context.Context, opts models.VerifyOptions,
	awsRunner interfaces.AWSCommandRunner, fs interfaces.FileSystem,
	cmdExecutor interfaces.ContextAwareCommandExecutor) (int, []models.ComparisonResult, error) {

	logger.Info("コンテキスト付きでAWS環境とTerraform状態の整合性検証を開始: 環境=%s", opts.Environment)
	logger.Debug("検証オプション: %+v", opts)

	// キャンセルされているかチェック
	if ctx.Err() != nil {
		logger.Error("コンテキストが既にキャンセルされています: %v", ctx.Err())
		return 1, nil, ctx.Err()
	}

	// AWS リソース取得
	logger.Debug("AWS環境のリソース情報を取得...")
	awsResources, err := GetAWSResources(awsRunner, opts.Environment, opts)
	if err != nil {
		logger.Error("AWS情報取得エラー: %v", err)
		return 1, nil, fmt.Errorf("AWS情報取得エラー: %v", err)
	}

	// キャンセルされているかチェック
	if ctx.Err() != nil {
		logger.Error("AWS情報取得後にコンテキストがキャンセルされました: %v", ctx.Err())
		return 1, nil, ctx.Err()
	}

	// Terraform状態取得
	logger.Debug("Terraform状態からリソース情報を取得...")
	tfResources, err := GetTerraformResources(opts.Environment, fs, cmdExecutor, opts) // opts引数を追加
	if err != nil {
		logger.Error("Terraform情報取得エラー: %v", err)
		return 1, nil, fmt.Errorf("Terraform情報取得エラー: %v", err)
	}

	// キャンセルされているかチェック
	if ctx.Err() != nil {
		logger.Error("Terraform情報取得後にコンテキストがキャンセルされました: %v", ctx.Err())
		return 1, nil, ctx.Err()
	}

	// リソース比較
	logger.Debug("AWS環境とTerraform状態のリソース比較を実行...")
	results := CompareResources(awsResources, tfResources)

	// 不一致数のカウント
	mismatchCount := 0
	for _, result := range results {
		if !result.IsMatch {
			mismatchCount++
		}
	}

	logger.Debug("不一致リソース数: %d", mismatchCount)

	// 早期判定
	if mismatchCount == 0 {
		logger.Debug("不一致なし: AWS=%+v, Terraform=%+v",
			map[string]int{"VPC": awsResources.VPC, "RDS": awsResources.RDS, "ECSCluster": awsResources.ECSCluster},
			map[string]int{"VPC": tfResources.VPC, "RDS": tfResources.RDS, "ECSCluster": tfResources.ECSCluster})

		// すべてのリソース数が一致
		if awsResources.VPC == 0 && tfResources.VPC == 0 &&
			awsResources.RDS == 0 && tfResources.RDS == 0 &&
			awsResources.ECSCluster == 0 && tfResources.ECSCluster == 0 {
			// 環境が空（リソースなし）でTerraform状態も空の場合
			logger.Info("環境が空でTerraform状態も空の状態です")
			// ユーザー向けガイダンスの追加
			logger.Info("環境をセットアップするには: make start-api-dev TF_ENV=%s", opts.Environment)
			return 0, results, nil
		}
		// リソースがあってすべて一致する場合は terraform plan で最終確認
		if !opts.SkipTerraformPlan {
			logger.Debug("リソース数が一致したため、terraform planで詳細確認を実行...")
			exitCode, _, err := RunTerraformPlanWithContext(ctx, opts.Environment, fs, cmdExecutor)
			if err != nil {
				// コンテキストエラーの場合はそのままエラーを返す
				if ctx.Err() != nil {
					logger.Error("terraform plan実行中にコンテキストがキャンセルされました: %v", ctx.Err())
					return 1, results, err
				}

				logger.Error("Terraformプラン実行エラー: %v", err)
				return 1, results, fmt.Errorf("Terraformプラン実行エラー: %v", err)
			}

			if exitCode == 0 {
				logger.Info("terraform planの結果: 変更なし (exitCode=0)")
			} else if exitCode == 2 {
				logger.Warn("terraform planの結果: 変更あり (exitCode=2)")
			} else {
				logger.Warn("terraform planの結果: 予期しない終了コード (exitCode=%d)", exitCode)
			}

			return exitCode, results, nil
		}

		logger.Info("リソース数が一致し、terraform planはスキップされました")
		return 0, results, nil
	}

	// 不一致がある場合
	logger.Warn("リソース数に不一致があります: %d個のリソースで差異検出", mismatchCount)
	return 2, results, nil
}

// VerifyState はAWS環境とTerraform状態の整合性を検証する
func VerifyState(opts models.VerifyOptions, awsRunner interfaces.AWSCommandRunner, fs interfaces.FileSystem, cmdExecutor interfaces.CommandExecutor) (int, []models.ComparisonResult, error) {
	// デフォルトタイムアウト（60秒）
	if opts.Timeout == 0 {
		opts.Timeout = 60 * time.Second
	}

	// タイムアウト付きコンテキスト作成
	ctx, cancel := context.WithTimeout(context.Background(), opts.Timeout)
	defer cancel()

	// CommandExecutorをContextAwareCommandExecutorに変換
	var ctxExecutor interfaces.ContextAwareCommandExecutor
	if contextAware, ok := cmdExecutor.(interfaces.ContextAwareCommandExecutor); ok {
		ctxExecutor = contextAware
	} else {
		// 直接キャストできない場合はラップする
		ctxExecutor = implementation.NewCommandExecutorWrapper(cmdExecutor)
	}

	return VerifyStateWithContext(ctx, opts, awsRunner, fs, ctxExecutor)
}

// VerifyStateForTest はテスト用のVerifyState関数で、リソースオブジェクトを直接受け取る
func VerifyStateForTest(opts models.VerifyOptions,
	awsResources, tfResources *models.Resources) (int, []models.ComparisonResult, error) {
	logger.Debug("テスト用の状態検証を実行: 環境=%s", opts.Environment)
	logger.Trace("テスト用AWS環境リソース: %+v", awsResources)
	logger.Trace("テスト用Terraform状態リソース: %+v", tfResources)

	// リソース比較
	results := CompareResources(awsResources, tfResources)

	// 不一致数のカウント
	mismatchCount := 0
	for _, result := range results {
		if !result.IsMatch {
			mismatchCount++
			logger.Debug("不一致項目: %s (AWS=%d, TF=%d)", // 追加
				result.ResourceName, result.AWSCount, result.TerraformCount)
		}
	}
	// 三項演算子ではなく通常のif文を使用
	exitCodeValue := 0
	if mismatchCount > 0 {
		exitCodeValue = 2
	}
	logger.Debug("不一致数=%d, 終了コード判定=%d", mismatchCount, exitCodeValue)
	logger.Debug("テスト用比較: 不一致リソース数=%d", mismatchCount)

	// 早期判定
	if mismatchCount == 0 {
		// すべてのリソース数が一致
		if awsResources.VPC == 0 && tfResources.VPC == 0 &&
			awsResources.RDS == 0 && tfResources.RDS == 0 &&
			awsResources.ECSCluster == 0 && tfResources.ECSCluster == 0 {
			// 環境が空（リソースなし）でTerraform状態も空の場合
			logger.Info("テスト用比較: 環境が空でTerraform状態も空")
			return 0, results, nil
		}

		// リソースがあってすべて一致する場合は成功
		logger.Info("テスト用比較: リソース数が一致")
		return 0, results, nil
	}

	// 不一致がある場合
	logger.Warn("テスト用比較: リソース数に不一致があります: %d個のリソースで差異検出", mismatchCount)
	return 2, results, nil
}
