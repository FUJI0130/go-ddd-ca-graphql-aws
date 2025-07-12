package terraform

import (
	"errors"
	"strings"
	"testing"

	"github.com/FUJI0130/go-ddd-ca/internal/tools/terraform/models"
)

// TestGetVPCCount はGetVPCCount関数のテスト
func TestGetVPCCount(t *testing.T) {
	testCases := []struct {
		name          string
		mockOutput    string
		mockError     error
		env           string
		expectedCount int
		shouldSucceed bool
	}{
		{
			name:          "正常系：VPCが1つ存在する",
			mockOutput:    "1",
			mockError:     nil,
			env:           "development",
			expectedCount: 1,
			shouldSucceed: true,
		},
		{
			name:          "正常系：VPCが存在しない",
			mockOutput:    "0",
			mockError:     nil,
			env:           "development",
			expectedCount: 0,
			shouldSucceed: true,
		},
		{
			name:          "異常系：AWS CLIエラー",
			mockOutput:    "",
			mockError:     errors.New("AWS CLI error"),
			env:           "development",
			expectedCount: 0,
			shouldSucceed: false,
		},
		{
			name:          "異常系：数値変換エラー",
			mockOutput:    "invalid",
			mockError:     nil,
			env:           "development",
			expectedCount: 0,
			shouldSucceed: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// モックランナーのセットアップ
			mockRunner := NewMockAWSRunner().
				WithMockOutput(tc.mockOutput).
				WithMockError(tc.mockError)

			// テスト対象関数の実行
			count, err := GetVPCCount(mockRunner, tc.env)

			// 結果の検証
			if tc.shouldSucceed {
				if err != nil {
					t.Errorf("エラーは期待されていませんでした。得られたエラー: %v", err)
				}
				if count != tc.expectedCount {
					t.Errorf("期待されたカウント %d, 実際の値: %d", tc.expectedCount, count)
				}
			} else {
				if err == nil {
					t.Error("エラーが期待されましたが、エラーはありませんでした")
				}
			}
		})
	}
}

// TestGetRDSCount はGetRDSCount関数のテスト
func TestGetRDSCount(t *testing.T) {
	testCases := []struct {
		name          string
		mockOutput    string
		mockError     error
		env           string
		expectedCount int
		shouldSucceed bool
	}{
		{
			name:          "正常系：RDSインスタンスが1つ存在する",
			mockOutput:    "1",
			mockError:     nil,
			env:           "development",
			expectedCount: 1,
			shouldSucceed: true,
		},
		{
			name:          "正常系：RDSインスタンスが存在しない",
			mockOutput:    "0",
			mockError:     nil,
			env:           "development",
			expectedCount: 0,
			shouldSucceed: true,
		},
		{
			name:          "異常系：AWS CLIエラー",
			mockOutput:    "",
			mockError:     errors.New("AWS CLI error"),
			env:           "development",
			expectedCount: 0,
			shouldSucceed: false,
		},
		{
			name:          "異常系：数値変換エラー",
			mockOutput:    "invalid",
			mockError:     nil,
			env:           "development",
			expectedCount: 0,
			shouldSucceed: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// モックランナーのセットアップ
			// モックランナーのセットアップ - ファクトリメソッドを使用
			mockRunner := NewMockAWSRunner().
				WithMockOutput(tc.mockOutput).
				WithMockError(tc.mockError)

			// テスト対象関数の実行
			count, err := GetRDSCount(mockRunner, tc.env)

			// 結果の検証
			if tc.shouldSucceed {
				if err != nil {
					t.Errorf("エラーは期待されていませんでした。得られたエラー: %v", err)
				}
				if count != tc.expectedCount {
					t.Errorf("期待されたカウント %d, 実際の値: %d", tc.expectedCount, count)
				}
			} else {
				if err == nil {
					t.Error("エラーが期待されましたが、エラーはありませんでした")
				}
			}
		})
	}
}

// TestGetECSClusterCount はGetECSClusterCount関数のテスト
func TestGetECSClusterCount(t *testing.T) {
	testCases := []struct {
		name          string
		mockOutput    string
		mockError     error
		env           string
		expectedCount int
		shouldSucceed bool
	}{
		{
			name:          "正常系：ECSクラスターが1つ存在する",
			mockOutput:    "1",
			mockError:     nil,
			env:           "development",
			expectedCount: 1,
			shouldSucceed: true,
		},
		{
			name:          "正常系：ECSクラスターが存在しない",
			mockOutput:    "0",
			mockError:     nil,
			env:           "development",
			expectedCount: 0,
			shouldSucceed: true,
		},
		{
			name:          "異常系：AWS CLIエラー",
			mockOutput:    "",
			mockError:     errors.New("AWS CLI error"),
			env:           "development",
			expectedCount: 0,
			shouldSucceed: false,
		},
		{
			name:          "異常系：数値変換エラー",
			mockOutput:    "invalid",
			mockError:     nil,
			env:           "development",
			expectedCount: 0,
			shouldSucceed: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// モックランナーのセットアップ
			mockRunner := NewMockAWSRunner().
				WithMockOutput(tc.mockOutput).
				WithMockError(tc.mockError)

			// テスト対象関数の実行
			count, err := GetECSClusterCount(mockRunner, tc.env)

			// 結果の検証
			if tc.shouldSucceed {
				if err != nil {
					t.Errorf("エラーは期待されていませんでした。得られたエラー: %v", err)
				}
				if count != tc.expectedCount {
					t.Errorf("期待されたカウント %d, 実際の値: %d", tc.expectedCount, count)
				}
			} else {
				if err == nil {
					t.Error("エラーが期待されましたが、エラーはありませんでした")
				}
			}
		})
	}
}

// TestGetServiceResources はGetServiceResources関数のテスト
func TestGetServiceResources(t *testing.T) {
	testCases := []struct {
		name              string
		mockOutputs       []string
		mockErrors        []error
		env               string
		serviceType       string
		expectedResources models.ServiceResources
		shouldSucceed     bool
	}{
		{
			name: "正常系：すべてのサービスリソースが存在する",
			mockOutputs: []string{
				"1", // ECSサービス
				"1", // ALB
				"1", // ターゲットグループ
			},
			mockErrors:  []error{nil, nil, nil},
			env:         "development",
			serviceType: "api",
			expectedResources: models.ServiceResources{
				ECSService:  1,
				ALB:         1,
				TargetGroup: 1,
			},
			shouldSucceed: true,
		},
		{
			name: "正常系：サービスリソースが存在しない",
			mockOutputs: []string{
				"0", // ECSサービス
				"0", // ALB
				"0", // ターゲットグループ
			},
			mockErrors:        []error{nil, nil, nil},
			env:               "development",
			serviceType:       "api",
			expectedResources: models.ServiceResources{},
			shouldSucceed:     true,
		},
		{
			name: "異常系：ECSサービス取得時のエラー",
			mockOutputs: []string{
				"",  // ECSサービス
				"1", // ALB
				"1", // ターゲットグループ
			},
			mockErrors: []error{
				errors.New("AWS CLI error"), // ECSサービス
				nil,                         // ALB
				nil,                         // ターゲットグループ
			},
			env:               "development",
			serviceType:       "api",
			expectedResources: models.ServiceResources{},
			shouldSucceed:     false,
		},
		{
			name: "異常系：ALB取得時のエラー",
			mockOutputs: []string{
				"1", // ECSサービス
				"",  // ALB
				"1", // ターゲットグループ
			},
			mockErrors: []error{
				nil,                         // ECSサービス
				errors.New("AWS CLI error"), // ALB
				nil,                         // ターゲットグループ
			},
			env:               "development",
			serviceType:       "api",
			expectedResources: models.ServiceResources{ECSService: 1},
			shouldSucceed:     false,
		},
		{
			name: "異常系：ターゲットグループ取得時のエラー",
			mockOutputs: []string{
				"1", // ECSサービス
				"1", // ALB
				"",  // ターゲットグループ
			},
			mockErrors: []error{
				nil,                         // ECSサービス
				nil,                         // ALB
				errors.New("AWS CLI error"), // ターゲットグループ
			},
			env:               "development",
			serviceType:       "api",
			expectedResources: models.ServiceResources{ECSService: 1, ALB: 1},
			shouldSucceed:     false,
		},
		{
			name: "異常系：数値変換エラー（ECSサービス）",
			mockOutputs: []string{
				"invalid", // ECSサービス
				"1",       // ALB
				"1",       // ターゲットグループ
			},
			mockErrors:        []error{nil, nil, nil},
			env:               "development",
			serviceType:       "api",
			expectedResources: models.ServiceResources{},
			shouldSucceed:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// このテストケースのためのランナーを準備
			callCount := 0
			mockRunner := NewMockAWSRunner().WithCommandFunc(func(args ...string) (string, error) {
				// 呼ばれる順番に応じて異なる結果を返す
				result := tc.mockOutputs[callCount]
				err := tc.mockErrors[callCount]
				callCount++
				return result, err
			})

			// テスト対象関数の実行
			resources, err := GetServiceResources(mockRunner, tc.env, tc.serviceType)

			// 結果の検証
			if tc.shouldSucceed {
				if err != nil {
					t.Errorf("エラーは期待されていませんでした。得られたエラー: %v", err)
				}
				if resources.ECSService != tc.expectedResources.ECSService {
					t.Errorf("ECSService: 期待値 %d, 実際の値 %d",
						tc.expectedResources.ECSService, resources.ECSService)
				}
				if resources.ALB != tc.expectedResources.ALB {
					t.Errorf("ALB: 期待値 %d, 実際の値 %d",
						tc.expectedResources.ALB, resources.ALB)
				}
				if resources.TargetGroup != tc.expectedResources.TargetGroup {
					t.Errorf("TargetGroup: 期待値 %d, 実際の値 %d",
						tc.expectedResources.TargetGroup, resources.TargetGroup)
				}
			} else {
				if err == nil {
					t.Error("エラーが期待されましたが、エラーはありませんでした")
				}
			}
		})
	}
}

// getCommandType はAWS CLIコマンドの引数からコマンドタイプを判定する
func getCommandType(args []string) string {
	if len(args) == 0 {
		return ""
	}

	switch args[0] {
	case "ec2":
		if len(args) >= 4 && args[1] == "describe-vpcs" {
			return "vpc"
		}
	case "rds":
		if len(args) >= 2 && args[1] == "describe-db-instances" {
			return "rds"
		}
	case "ecs":
		if len(args) >= 2 {
			if args[1] == "list-clusters" {
				return "ecsCluster"
			} else if args[1] == "list-services" {
				// サービスタイプを抽出
				for i, arg := range args {
					if arg == "--query" && i+1 < len(args) {
						queryArg := args[i+1]
						if strings.Contains(queryArg, "api") {
							return "api_service"
						} else if strings.Contains(queryArg, "graphql") {
							return "graphql_service"
						} else if strings.Contains(queryArg, "grpc") {
							return "grpc_service"
						}
					}
				}
			}
		}
	case "elbv2":
		if len(args) >= 2 {
			if args[1] == "describe-load-balancers" {
				// ALBの種類を判断するために--queryパラメータを探す
				for i, arg := range args {
					if arg == "--query" && i+1 < len(args) {
						queryArg := args[i+1]
						if strings.Contains(queryArg, "api") {
							return "api_alb"
						} else if strings.Contains(queryArg, "graphql") {
							return "graphql_alb"
						} else if strings.Contains(queryArg, "grpc") {
							return "grpc_alb"
						}
					}
				}
			} else if args[1] == "describe-target-groups" {
				// ターゲットグループの種類を判断するために--queryパラメータを探す
				for i, arg := range args {
					if arg == "--query" && i+1 < len(args) {
						queryArg := args[i+1]
						if strings.Contains(queryArg, "api") {
							return "api_tg"
						} else if strings.Contains(queryArg, "graphql") {
							return "graphql_tg"
						} else if strings.Contains(queryArg, "grpc") {
							return "grpc_tg"
						}
					}
				}
			}
		}
	}

	return ""
}

func TestGetAWSResources(t *testing.T) {
	testCases := []struct {
		name           string
		mockResponses  map[string]string
		mockErrors     map[string]error
		env            string
		expectedResult *models.Resources
		shouldSucceed  bool
	}{
		{
			name: "正常系：すべてのリソースが存在する",
			mockResponses: map[string]string{
				"vpc":             "1",
				"rds":             "1",
				"ecsCluster":      "1",
				"api_service":     "1",
				"api_alb":         "1",
				"api_tg":          "1",
				"graphql_service": "1",
				"graphql_alb":     "1",
				"graphql_tg":      "1",
				"grpc_service":    "1",
				"grpc_alb":        "1",
				"grpc_tg":         "1",
			},
			mockErrors: map[string]error{},
			env:        "development",
			expectedResult: &models.Resources{
				VPC:        1,
				RDS:        1,
				ECSCluster: 1,
				Services: map[string]models.ServiceResources{
					"api": {
						ECSService:  1,
						ALB:         1,
						TargetGroup: 1,
					},
					"graphql": {
						ECSService:  1,
						ALB:         1,
						TargetGroup: 1,
					},
					"grpc": {
						ECSService:  1,
						ALB:         1,
						TargetGroup: 1,
					},
				},
			},
			shouldSucceed: true,
		},
		{
			name: "正常系：リソースが存在しない",
			mockResponses: map[string]string{
				"vpc":             "0",
				"rds":             "0",
				"ecsCluster":      "0",
				"api_service":     "0",
				"api_alb":         "0",
				"api_tg":          "0",
				"graphql_service": "0",
				"graphql_alb":     "0",
				"graphql_tg":      "0",
				"grpc_service":    "0",
				"grpc_alb":        "0",
				"grpc_tg":         "0",
			},
			mockErrors: map[string]error{},
			env:        "development",
			expectedResult: &models.Resources{
				VPC:        0,
				RDS:        0,
				ECSCluster: 0,
				Services:   map[string]models.ServiceResources{}, // 修正: ECSクラスターが0の場合はサービスマップは空
			},
			shouldSucceed: true,
		},
		{
			name: "異常系：VPC取得エラー",
			mockResponses: map[string]string{
				"vpc": "",
			},
			mockErrors: map[string]error{
				"vpc": errors.New("AWS CLI error"),
			},
			env:            "development",
			expectedResult: nil,
			shouldSucceed:  false,
		},
		{
			name: "異常系：RDS取得エラー",
			mockResponses: map[string]string{
				"vpc": "1",
				"rds": "",
			},
			mockErrors: map[string]error{
				"rds": errors.New("AWS CLI error"),
			},
			env:            "development",
			expectedResult: nil,
			shouldSucceed:  false,
		},
		{
			name: "異常系：ECSクラスター取得エラー",
			mockResponses: map[string]string{
				"vpc":        "1",
				"rds":        "1",
				"ecsCluster": "",
			},
			mockErrors: map[string]error{
				"ecsCluster": errors.New("AWS CLI error"),
			},
			env:            "development",
			expectedResult: nil,
			shouldSucceed:  false,
		},
		{
			name: "異常系：api_service取得エラー",
			mockResponses: map[string]string{
				"vpc":         "1",
				"rds":         "1",
				"ecsCluster":  "1",
				"api_service": "",
			},
			mockErrors: map[string]error{
				"api_service": errors.New("AWS CLI error"),
			},
			env:            "development",
			expectedResult: nil,
			shouldSucceed:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// このテストケースのためのランナーを準備
			mockRunner := NewMockAWSRunner().WithCommandFunc(func(args ...string) (string, error) {
				// コマンドの種類を判断（汎用ヘルパー関数を使用）
				cmdType := getCommandType(args)

				// 該当するモックレスポンスとエラーを返す
				if response, ok := tc.mockResponses[cmdType]; ok {
					var err error
					if errVal, errOk := tc.mockErrors[cmdType]; errOk {
						err = errVal
					}
					return response, err
				}

				// デフォルトのレスポンス
				return "0", nil
			})

			// テスト対象関数の実行
			var opts models.VerifyOptions
			result, err := GetAWSResources(mockRunner, tc.env, opts)

			// 結果の検証
			if tc.shouldSucceed {
				if err != nil {
					t.Errorf("エラーは期待されていませんでした。得られたエラー: %v", err)
				}

				// リソース数の検証
				if result.VPC != tc.expectedResult.VPC {
					t.Errorf("VPC: 期待値 %d, 実際の値 %d", tc.expectedResult.VPC, result.VPC)
				}
				if result.RDS != tc.expectedResult.RDS {
					t.Errorf("RDS: 期待値 %d, 実際の値 %d", tc.expectedResult.RDS, result.RDS)
				}
				if result.ECSCluster != tc.expectedResult.ECSCluster {
					t.Errorf("ECSCluster: 期待値 %d, 実際の値 %d", tc.expectedResult.ECSCluster, result.ECSCluster)
				}

				// サービスリソースの検証
				for svcType, expected := range tc.expectedResult.Services {
					actual, ok := result.Services[svcType]
					if !ok {
						// 期待値がある場合のみエラーとする
						if tc.expectedResult.ECSCluster > 0 {
							t.Errorf("サービス %s が結果に含まれていません", svcType)
						}
						continue
					}

					if actual.ECSService != expected.ECSService {
						t.Errorf("%s ECSService: 期待値 %d, 実際の値 %d",
							svcType, expected.ECSService, actual.ECSService)
					}
					if actual.ALB != expected.ALB {
						t.Errorf("%s ALB: 期待値 %d, 実際の値 %d",
							svcType, expected.ALB, actual.ALB)
					}
					if actual.TargetGroup != expected.TargetGroup {
						t.Errorf("%s TargetGroup: 期待値 %d, 実際の値 %d",
							svcType, expected.TargetGroup, actual.TargetGroup)
					}
				}
			} else {
				if err == nil {
					t.Error("エラーが期待されましたが、エラーはありませんでした")
				}
			}
		})
	}
}

func TestGetAWSResourcesWithIgnoreOption(t *testing.T) {
	testCases := []struct {
		name               string
		mockResponses      map[string]string
		mockErrors         map[string]error
		env                string
		ignoreResErrors    bool
		expectedResultNil  bool
		expectErrorMessage string
	}{
		{
			name: "VPC取得エラー（無視オプションあり）",
			mockResponses: map[string]string{
				"vpc": "",
			},
			mockErrors: map[string]error{
				"vpc": errors.New("VpcNotFound"),
			},
			env:                "development",
			ignoreResErrors:    true,
			expectedResultNil:  false, // リソースオブジェクトは返される
			expectErrorMessage: "",    // エラーなし
		},
		{
			name: "VPC取得エラー（無視オプションなし）",
			mockResponses: map[string]string{
				"vpc": "",
			},
			mockErrors: map[string]error{
				"vpc": errors.New("VpcNotFound"),
			},
			env:                "development",
			ignoreResErrors:    false,
			expectedResultNil:  true,          // リソースオブジェクトはnil
			expectErrorMessage: "VpcNotFound", // エラーメッセージの一部
		},
		{
			name: "ECSクラスター取得エラー（無視オプションあり）",
			mockResponses: map[string]string{
				"vpc":        "1",
				"rds":        "1",
				"ecsCluster": "",
			},
			mockErrors: map[string]error{
				"ecsCluster": errors.New("ClusterNotFoundException"),
			},
			env:                "development",
			ignoreResErrors:    true,
			expectedResultNil:  false,
			expectErrorMessage: "",
		},
		{
			name: "サービス取得エラー（無視オプションあり）",
			mockResponses: map[string]string{
				"vpc":         "1",
				"rds":         "1",
				"ecsCluster":  "1",
				"api_service": "",
			},
			mockErrors: map[string]error{
				"api_service": errors.New("ServiceNotFoundException"),
			},
			env:                "development",
			ignoreResErrors:    true,
			expectedResultNil:  false,
			expectErrorMessage: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// このテストケースのためのランナーを準備
			mockRunner := NewMockAWSRunner().WithCommandFunc(func(args ...string) (string, error) {
				// コマンドの種類を判断（汎用ヘルパー関数を使用）
				cmdType := getCommandType(args)

				// 該当するモックレスポンスとエラーを返す
				if response, ok := tc.mockResponses[cmdType]; ok {
					var err error
					if errVal, errOk := tc.mockErrors[cmdType]; errOk {
						err = errVal
					}
					return response, err
				}

				// デフォルトのレスポンス
				return "0", nil
			})

			// テスト対象関数の実行（オプション追加）
			opts := models.VerifyOptions{
				IgnoreResourceErrors: tc.ignoreResErrors,
			}
			result, err := GetAWSResources(mockRunner, tc.env, opts)

			// 結果の検証
			if tc.expectedResultNil && result != nil {
				t.Errorf("nilの結果が期待されていましたが、実際には非nilの結果が返されました")
			} else if !tc.expectedResultNil && result == nil {
				t.Errorf("非nilの結果が期待されていましたが、実際にはnilの結果が返されました")
			}

			// エラーメッセージの検証
			if tc.expectErrorMessage == "" && err != nil {
				t.Errorf("エラーは期待されていませんでした。得られたエラー: %v", err)
			} else if tc.expectErrorMessage != "" && err == nil {
				t.Errorf("エラーが期待されていましたが、エラーはありませんでした")
			} else if tc.expectErrorMessage != "" && err != nil && !strings.Contains(err.Error(), tc.expectErrorMessage) {
				t.Errorf("期待されたエラーメッセージ '%s' が実際のエラー '%v' に含まれていません", tc.expectErrorMessage, err)
			}
		})
	}
}

func TestResourceDependencies(t *testing.T) {
	testCases := []struct {
		name          string
		mockResponses map[string]string
		mockErrors    map[string]error
		env           string
		expectedCalls []string // どのリソース取得が呼ばれるべきか
	}{
		{
			name: "すべてのリソースが正常",
			mockResponses: map[string]string{
				"vpc":        "1",
				"rds":        "1",
				"ecsCluster": "1",
			},
			mockErrors: map[string]error{},
			env:        "development",
			expectedCalls: []string{
				"vpc", "rds", "ecsCluster",
				"api_service", "api_alb", "api_tg",
				"graphql_service", "graphql_alb", "graphql_tg",
				"grpc_service", "grpc_alb", "grpc_tg",
			},
		},
		{
			name: "VPCが存在しない場合",
			mockResponses: map[string]string{
				"vpc":        "0",
				"rds":        "0", // 追加: rdsも呼ばれるので応答を定義
				"ecsCluster": "0", // 追加: ecsClusterも呼ばれるので応答を定義
			},
			mockErrors: map[string]error{},
			env:        "development",
			expectedCalls: []string{
				"vpc", "rds", "ecsCluster", // 修正: 現在の実装ではすべて呼ばれる
			},
		},
		{
			name: "ECSクラスターが存在しない場合",
			mockResponses: map[string]string{
				"vpc":        "1",
				"rds":        "1",
				"ecsCluster": "0",
			},
			mockErrors: map[string]error{},
			env:        "development",
			expectedCalls: []string{
				"vpc", "rds", "ecsCluster", // サービス系は呼ばれない
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 呼び出されたコマンドを記録
			calledCommands := []string{}

			// このテストケースのためのランナーを準備
			mockRunner := NewMockAWSRunner().WithCommandFunc(func(args ...string) (string, error) {
				// コマンドの種類を判断
				cmdType := getCommandType(args)
				calledCommands = append(calledCommands, cmdType)

				// 該当するモックレスポンスとエラーを返す
				if response, ok := tc.mockResponses[cmdType]; ok {
					var err error
					if errVal, errOk := tc.mockErrors[cmdType]; errOk {
						err = errVal
					}
					return response, err
				}

				// デフォルトのレスポンス
				return "0", nil
			})

			// テスト対象関数の実行
			opts := models.VerifyOptions{
				IgnoreResourceErrors: true,
			}
			_, _ = GetAWSResources(mockRunner, tc.env, opts)

			// 呼び出されたコマンドの検証
			for _, expected := range tc.expectedCalls {
				found := false
				for _, called := range calledCommands {
					if called == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("期待されたコマンド '%s' が呼び出されませんでした", expected)
				}
			}

			// 予期しないコマンド呼び出しがないことを確認
			for _, called := range calledCommands {
				found := false
				for _, expected := range tc.expectedCalls {
					if called == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("予期しないコマンド '%s' が呼び出されました", called)
				}
			}
		})
	}
}

func TestIsResourceNotFoundError(t *testing.T) {
	testCases := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "VPC not found error",
			err:      errors.New("Error getting VPC: VpcNotFound: The vpc ID 'vpc-123' does not exist"),
			expected: true,
		},
		{
			name:     "Cluster not found error",
			err:      errors.New("Error getting ECS cluster: ClusterNotFoundException: Cluster not found."),
			expected: true,
		},
		{
			name:     "DB instance not found error",
			err:      errors.New("Error getting RDS: DBInstanceNotFound: DB instance not found"),
			expected: true,
		},
		{
			name:     "Service not found error",
			err:      errors.New("Error getting service: ServiceNotFoundException: Service not found."),
			expected: true,
		},
		{
			name:     "Other error",
			err:      errors.New("Some other error"),
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isResourceNotFoundError(tc.err)
			if result != tc.expected {
				t.Errorf("isResourceNotFoundError(%v) = %v, expected %v", tc.err, result, tc.expected)
			}
		})
	}
}
