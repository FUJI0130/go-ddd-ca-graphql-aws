// verify_terraform_json_integration_test.go
package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/FUJI0130/go-ddd-ca/internal/tools/terraform"
	"github.com/FUJI0130/go-ddd-ca/internal/tools/terraform/models"
)

// TestVerifyTerraformWithJSONProcessing はJSONパース処理を含む統合テスト
func TestVerifyTerraformWithJSONProcessing(t *testing.T) {
	testCases := []struct {
		name         string                   // テスト名
		jsonResponse string                   // モックJSON応答
		awsResources func() *models.Resources // AWSリソース生成関数
		expectedCode int                      // 期待終了コード
		expectError  bool                     // 関数がエラーを返すことを期待するか
	}{
		{
			name: "正常系: 空のリソース",
			jsonResponse: `{
				"values": {
					"root_module": {
						"resources": []
					}
				}
			}`,
			awsResources: func() *models.Resources {
				return &models.Resources{
					VPC:        0,
					RDS:        0,
					ECSCluster: 0,
					Services:   make(map[string]models.ServiceResources),
				}
			},
			expectedCode: 0, // リソースが一致するので0
			expectError:  false,
		},
		{
			name: "正常系: 単一VPC",
			jsonResponse: `{
				"values": {
					"root_module": {
						"resources": [
							{
								"type": "aws_vpc",
								"address": "module.networking.aws_vpc.main",
								"values": { "id": "vpc-12345" }
							}
						]
					}
				}
			}`,
			awsResources: func() *models.Resources {
				return &models.Resources{
					VPC:        1,
					RDS:        0,
					ECSCluster: 0,
					Services:   make(map[string]models.ServiceResources),
				}
			},
			expectedCode: 0, // リソースが一致するので0
			expectError:  false,
		},
		{
			name: "正常系: 複数リソース",
			jsonResponse: `{
				"values": {
					"root_module": {
						"resources": [
							{
								"type": "aws_vpc",
								"address": "module.networking.aws_vpc.main",
								"values": { "id": "vpc-12345" }
							},
							{
								"type": "aws_db_instance",
								"address": "module.database.aws_db_instance.main",
								"values": { "id": "db-12345" }
							},
							{
								"type": "aws_ecs_cluster",
								"address": "module.shared.aws_ecs_cluster.main",
								"values": { "id": "cluster-12345" }
							}
						]
					}
				}
			}`,
			awsResources: func() *models.Resources {
				return &models.Resources{
					VPC:        1,
					RDS:        1,
					ECSCluster: 1,
					Services:   make(map[string]models.ServiceResources),
				}
			},
			expectedCode: 0, // リソースが一致するので0
			expectError:  false,
		},
		{
			name: "正常系: サービスリソース",
			jsonResponse: `{
				"values": {
					"root_module": {
						"resources": [
							{
								"type": "aws_vpc",
								"address": "module.networking.aws_vpc.main",
								"values": { "id": "vpc-12345" }
							},
							{
								"type": "aws_ecs_service",
								"address": "module.service_api.aws_ecs_service.main",
								"values": { "id": "service-api-12345" }
							},
							{
								"type": "aws_lb",
								"address": "module.loadbalancer_api.aws_lb.main",
								"values": { "id": "alb-api-12345" }
							},
							{
								"type": "aws_lb_target_group",
								"address": "module.target_group_api.aws_lb_target_group.main",
								"values": { "id": "tg-api-12345" }
							}
						]
					}
				}
			}`,
			awsResources: func() *models.Resources {
				resources := &models.Resources{
					VPC:        1,
					RDS:        0,
					ECSCluster: 0,
					Services:   make(map[string]models.ServiceResources),
				}
				resources.Services["api"] = models.ServiceResources{
					ECSService:  1,
					ALB:         1,
					TargetGroup: 1,
				}
				return resources
			},
			expectedCode: 0, // リソースが一致するので0
			expectError:  false,
		},
		{
			name: "正常系: 子モジュール",
			jsonResponse: `{
				"values": {
					"root_module": {
						"resources": [
							{
								"type": "aws_vpc",
								"address": "module.networking.aws_vpc.main",
								"values": { "id": "vpc-12345" }
							}
						],
						"child_modules": [
							{
								"address": "module.service_api",
								"resources": [
									{
										"type": "aws_ecs_service",
										"address": "module.service_api.aws_ecs_service.main",
										"values": { "id": "service-api-12345" }
									},
									{
										"type": "aws_lb",
										"address": "module.service_api.aws_lb.main",
										"values": { "id": "alb-api-12345" }
									},
									{
										"type": "aws_lb_target_group",
										"address": "module.service_api.aws_lb_target_group.main",
										"values": { "id": "tg-api-12345" }
									}
								]
							}
						]
					}
				}
			}`,
			awsResources: func() *models.Resources {
				resources := &models.Resources{
					VPC:        1,
					RDS:        0,
					ECSCluster: 0,
					Services:   make(map[string]models.ServiceResources),
				}
				resources.Services["api"] = models.ServiceResources{
					ECSService:  1,
					ALB:         1,
					TargetGroup: 1,
				}
				return resources
			},
			expectedCode: 0, // リソースが一致するので0
			expectError:  false,
		},
		{
			name: "異常系: リソース不一致",
			jsonResponse: `{
				"values": {
					"root_module": {
						"resources": [
							{
								"type": "aws_vpc",
								"address": "module.networking.aws_vpc.main",
								"values": { "id": "vpc-12345" }
							}
						]
					}
				}
			}`,
			awsResources: func() *models.Resources {
				return &models.Resources{
					VPC:        2, // AWS側は2つのVPC（不一致）
					RDS:        0,
					ECSCluster: 0,
					Services:   make(map[string]models.ServiceResources),
				}
			},
			expectedCode: 2, // 不一致なので2を期待
			expectError:  false,
		},
		{
			name: "異常系: 無効なJSON構文",
			jsonResponse: `{
				"values": {
					"root_module": {
						"resources": [
							{
								"type": "aws_vpc",
								"address": "module.networking.aws_vpc.main",
								"values": { "id": "vpc-12345" }
							}
						]
					}
				`, // 閉じ括弧が不足
			awsResources: func() *models.Resources {
				return &models.Resources{
					VPC:        1, // AWS側は1つのVPCがある
					RDS:        0,
					ECSCluster: 0,
					Services:   make(map[string]models.ServiceResources),
				}
			},
			expectedCode: 2,     // JSON解析エラーで空のリソース情報返却→不一致→終了コード2
			expectError:  false, // エラーは返さず空のリソース情報を返す
		},
		{
			name: "異常系: 階層不足",
			jsonResponse: `{
				"values": {
					"resources": [
						{
							"type": "aws_vpc",
							"address": "module.networking.aws_vpc.main",
							"values": { "id": "vpc-12345" }
						}
					]
				}
			}`, // root_moduleキーがない
			awsResources: func() *models.Resources {
				return &models.Resources{
					VPC:        1, // AWS側は1つのVPCがある
					RDS:        0,
					ECSCluster: 0,
					Services:   make(map[string]models.ServiceResources),
				}
			},
			expectedCode: 2,     // 解析問題で空のリソース情報→不一致→終了コード2
			expectError:  false, // エラーは返さず空のリソース情報を返す
		},
		{
			name: "異常系: キー不足",
			jsonResponse: `{
				"values": {
					"root_module": {
						"resources": [
							{
								"address": "module.networking.aws_vpc.main",
								"values": { "id": "vpc-12345" }
							}
						]
					}
				}
			}`, // typeキーが不足
			awsResources: func() *models.Resources {
				return &models.Resources{
					VPC:        1, // AWS側は1つのVPCがある
					RDS:        0,
					ECSCluster: 0,
					Services:   make(map[string]models.ServiceResources),
				}
			},
			expectedCode: 2, // リソースのスキップにより不一致となり終了コード2
			expectError:  false,
		},
		{
			name: "異常系: 不正な値型",
			jsonResponse: `{
				"values": {
					"root_module": {
						"resources": 123
					}
				}
			}`, // resourcesが配列ではなく数値
			awsResources: func() *models.Resources {
				return &models.Resources{
					VPC:        1, // AWS側は1つのVPCがある
					RDS:        0,
					ECSCluster: 0,
					Services:   make(map[string]models.ServiceResources),
				}
			},
			expectedCode: 2, // 解析問題→リソース不一致で終了コード2
			expectError:  false,
		},
	}

	// 各テストケースを実行
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// テスト環境のセットアップ
			mockFS := terraform.NewMockFileSystem().
				WithGetwdFunc(func() (string, error) {
					return "/test/path", nil
				}).
				WithChdirFunc(func(dir string) error {
					return nil
				})

			mockCmd := terraform.NewMockContextCommandExecutor().
				WithExecuteFunc(func(command string, args ...string) (string, error) {
					if command == "terraform" && args[0] == "show" && args[1] == "-json" {
						return tc.jsonResponse, nil
					}
					return "", fmt.Errorf("unexpected command: %s %v", command, args)
				}).
				WithExecuteWithContextFunc(func(ctx context.Context, command string, args ...string) (string, error) {
					if command == "terraform" && args[0] == "show" && args[1] == "-json" {
						return tc.jsonResponse, nil
					}
					if command == "terraform" && args[0] == "plan" {
						return "No changes. Your infrastructure matches the configuration.", nil
					}
					return "", fmt.Errorf("unexpected command: %s %v", command, args)
				})

			// AWS側のリソース情報を生成
			awsResources := tc.awsResources()

			// verify.GetTerraformResources を使用してリソース情報を取得
			// これにより実際のJSONパース処理が実行される
			tfResources, err := terraform.GetTerraformResources("development", mockFS, mockCmd)

			// エラー検証
			if tc.expectError && err == nil {
				t.Errorf("エラーが期待されていましたが、エラーはありませんでした")
			} else if !tc.expectError && err != nil {
				t.Errorf("エラーは期待されていませんでした。得られたエラー: %v", err)
			}

			// tfResourcesがnilでないことを確認
			if tfResources == nil {
				t.Errorf("Terraformリソースがnilです")
				return
			}

			// 取得したリソース情報を使用して検証
			opts := models.VerifyOptions{
				Environment:       "development",
				SkipTerraformPlan: true,
			}

			// 検証処理の実行
			exitCode, results, err := terraform.VerifyStateForTest(opts, awsResources, tfResources)

			// リソース比較結果のログ出力（デバッグ用）
			t.Logf("リソース比較結果: %+v", results)
			t.Logf("Terraformリソース情報: VPC=%d, RDS=%d, ECSCluster=%d, サービス=%v",
				tfResources.VPC, tfResources.RDS, tfResources.ECSCluster, tfResources.Services)
			t.Logf("デバッグ情報: テスト=%s, exitCode=%d, expectedCode=%d",
				tc.name, exitCode, tc.expectedCode)

			// 終了コードの検証（テストの主要な判定基準）
			if exitCode != tc.expectedCode {
				t.Errorf("終了コードの不一致: 期待=%d, 実際=%d", tc.expectedCode, exitCode)
			}
		})
	}
}

// TestVerifyTerraformWithMixedResources はリソース種別を混在させたケースの統合テスト
func TestVerifyTerraformWithMixedResources(t *testing.T) {
	// AWSサービスリソースを含む複雑なJSON
	jsonResponse := `{
		"values": {
			"root_module": {
				"resources": [
					{
						"type": "aws_vpc",
						"address": "module.networking.aws_vpc.main",
						"values": { "id": "vpc-12345" }
					},
					{
						"type": "aws_db_instance",
						"address": "module.database.aws_db_instance.main",
						"values": { "id": "db-12345" }
					},
					{
						"type": "aws_ecs_cluster",
						"address": "module.shared.aws_ecs_cluster.main",
						"values": { "id": "cluster-12345" }
					},
					{
						"type": "aws_ecs_service",
						"address": "module.service_api.aws_ecs_service.main",
						"values": { "id": "service-api-12345" }
					},
					{
						"type": "aws_lb",
						"address": "module.loadbalancer_api.aws_lb.main",
						"values": { "id": "alb-api-12345" }
					},
					{
						"type": "aws_lb_target_group",
						"address": "module.target_group_api.aws_lb_target_group.main",
						"values": { "id": "tg-api-12345" }
					},
					{
						"type": "aws_ecs_service",
						"address": "module.service_graphql.aws_ecs_service.main",
						"values": { "id": "service-graphql-12345" }
					},
					{
						"type": "aws_lb",
						"address": "module.loadbalancer_graphql.aws_lb.main",
						"values": { "id": "alb-graphql-12345" }
					},
					{
						"type": "aws_lb_target_group",
						"address": "module.target_group_graphql.aws_lb_target_group.main",
						"values": { "id": "tg-graphql-12345" }
					}
				]
			}
		}
	}`

	// テストケースのセットアップ
	mockFS := terraform.NewMockFileSystem().
		WithGetwdFunc(func() (string, error) {
			return "/test/path", nil
		}).
		WithChdirFunc(func(dir string) error {
			return nil
		})

	mockCmd := terraform.NewMockContextCommandExecutor().
		WithExecuteFunc(func(command string, args ...string) (string, error) {
			if command == "terraform" && args[0] == "show" && args[1] == "-json" {
				return jsonResponse, nil
			}
			return "", fmt.Errorf("unexpected command: %s %v", command, args)
		}).
		WithExecuteWithContextFunc(func(ctx context.Context, command string, args ...string) (string, error) {
			if command == "terraform" && args[0] == "show" && args[1] == "-json" {
				return jsonResponse, nil
			}
			if command == "terraform" && args[0] == "plan" {
				return "No changes. Your infrastructure matches the configuration.", nil
			}
			return "", fmt.Errorf("unexpected command: %s %v", command, args)
		})

	// AWS側のリソース情報を生成（すべて一致するケース）
	awsResources := &models.Resources{
		VPC:        1,
		RDS:        1,
		ECSCluster: 1,
		Services:   make(map[string]models.ServiceResources),
	}
	awsResources.Services["api"] = models.ServiceResources{
		ECSService:  1,
		ALB:         1,
		TargetGroup: 1,
	}
	awsResources.Services["graphql"] = models.ServiceResources{
		ECSService:  1,
		ALB:         1,
		TargetGroup: 1,
	}

	// GetTerraformResources を使用してリソース情報を取得（実際のJSONパース処理を実行）
	tfResources, err := terraform.GetTerraformResources("development", mockFS, mockCmd)
	if err != nil {
		t.Fatalf("Terraformリソース取得エラー: %v", err)
	}

	// JSON応答から解析されたリソース情報を検証
	if tfResources.VPC != 1 {
		t.Errorf("VPC数が期待値と異なります: 期待=1, 実際=%d", tfResources.VPC)
	}
	if tfResources.RDS != 1 {
		t.Errorf("RDS数が期待値と異なります: 期待=1, 実際=%d", tfResources.RDS)
	}
	if tfResources.ECSCluster != 1 {
		t.Errorf("ECSクラスター数が期待値と異なります: 期待=1, 実際=%d", tfResources.ECSCluster)
	}

	// サービスリソースの検証
	for svcType, expected := range map[string]models.ServiceResources{
		"api":     {ECSService: 1, ALB: 1, TargetGroup: 1},
		"graphql": {ECSService: 1, ALB: 1, TargetGroup: 1},
	} {
		actual, exists := tfResources.Services[svcType]
		if !exists {
			t.Errorf("サービスタイプ %s のリソースが存在しません", svcType)
			continue
		}
		if actual.ECSService != expected.ECSService {
			t.Errorf("%s ECSサービス数が期待値と異なります: 期待=%d, 実際=%d",
				svcType, expected.ECSService, actual.ECSService)
		}
		if actual.ALB != expected.ALB {
			t.Errorf("%s ALB数が期待値と異なります: 期待=%d, 実際=%d",
				svcType, expected.ALB, actual.ALB)
		}
		if actual.TargetGroup != expected.TargetGroup {
			t.Errorf("%s ターゲットグループ数が期待値と異なります: 期待=%d, 実際=%d",
				svcType, expected.TargetGroup, actual.TargetGroup)
		}
	}

	// リソースが一致する場合の検証
	opts := models.VerifyOptions{
		Environment:       "development",
		SkipTerraformPlan: true,
	}

	exitCode, results, err := terraform.VerifyStateForTest(opts, awsResources, tfResources)
	if err != nil {
		t.Fatalf("検証エラー: %v", err)
	}

	// すべてのリソースが一致するので終了コードは0
	if exitCode != 0 {
		t.Errorf("終了コードが期待値と異なります: 期待=0, 実際=%d", exitCode)
	}

	// 詳細なリソース比較結果の検証
	for _, result := range results {
		if !result.IsMatch {
			t.Errorf("リソース %s の不一致: AWS=%d, Terraform=%d",
				result.ResourceName, result.AWSCount, result.TerraformCount)
		}
	}
}

// TestVerifyTerraformWithContextTimeout はタイムアウト処理の統合テスト
func TestVerifyTerraformWithContextTimeout(t *testing.T) {
	t.Logf("[DEBUG] タイムアウトテスト開始")

	// 非常に短いタイムアウトを設定（20msに短縮）
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	t.Logf("[DEBUG] コンテキスト作成: タイムアウト=20ms")

	// 簡素化したJSON応答
	jsonResponse := `{
        "values": {
            "root_module": {
                "resources": [
                    {
                        "type": "aws_vpc",
                        "address": "module.networking.aws_vpc.main",
                        "values": { "id": "vpc-12345" }
                    }
                ]
            }
        }
    }`

	// モックファイルシステム
	mockFS := terraform.NewMockFileSystem().
		WithGetwdFunc(func() (string, error) {
			t.Logf("[DEBUG] GetWd呼び出し")
			return "/test/path", nil
		}).
		WithChdirFunc(func(dir string) error {
			t.Logf("[DEBUG] Chdir呼び出し: dir=%s", dir)
			return nil
		})

	// モックコマンド実行
	mockCmd := terraform.NewMockContextCommandExecutor().
		WithExecuteFunc(func(command string, args ...string) (string, error) {
			t.Logf("[DEBUG] Execute呼び出し: command=%s args=%v", command, args)
			if command == "terraform" && args[0] == "show" && args[1] == "-json" {
				return jsonResponse, nil
			}
			return "", fmt.Errorf("unexpected command: %s %v", command, args)
		}).
		WithExecuteWithContextFunc(func(ctx context.Context, command string, args ...string) (string, error) {
			t.Logf("[DEBUG] ExecuteWithContext呼び出し: command=%s args=%v", command, args)

			// タイムアウト確認（即時）
			if ctx.Err() != nil {
				t.Logf("[DEBUG] 呼び出し時点でコンテキストエラー: %v", ctx.Err())
				return "", ctx.Err()
			}

			if command == "terraform" && args[0] == "show" && args[1] == "-json" {
				// terraform show は短時間で返す
				t.Logf("[DEBUG] terraform show -json 応答返却")
				return jsonResponse, nil
			}

			if command == "terraform" && args[0] == "plan" {
				// 確実にタイムアウトするよう長めの遅延
				t.Logf("[DEBUG] terraform plan 実行開始 - 遅延前 ctx.Err()=%v", ctx.Err())

				// 複数ステップでタイムアウト発生を確認
				for i := 1; i <= 5; i++ {
					time.Sleep(10 * time.Millisecond)
					t.Logf("[DEBUG] sleep %dms後: ctx.Err()=%v", i*10, ctx.Err())

					if ctx.Err() != nil {
						t.Logf("[DEBUG] コンテキストエラー検出: %v", ctx.Err())
						return "", ctx.Err()
					}
				}

				t.Logf("[DEBUG] 警告: 50ms待機後もタイムアウトなし")
				return "No changes.", nil
			}

			return "", fmt.Errorf("unexpected command: %s %v", command, args)
		})

	// AWS Runnerのモック（シンプル化）
	mockAWSRunner := terraform.NewMockAWSRunner().
		WithCommandFunc(func(args ...string) (string, error) {
			// 単純に1つのVPCだけを返すようにする
			if len(args) > 0 && args[0] == "ec2" {
				t.Logf("[DEBUG] AWS Runner VPC返却: 1")
				return "1", nil
			}
			// その他は全て0を返す
			t.Logf("[DEBUG] AWS Runner 0返却")
			return "0", nil
		})

	// オプション設定
	opts := models.VerifyOptions{
		Environment:       "development",
		SkipTerraformPlan: false,                 // planを確実に実行
		Timeout:           20 * time.Millisecond, // 短いタイムアウト
	}

	t.Logf("[DEBUG] VerifyStateWithContext 実行開始")

	// タイムアウト付きで実行
	exitCode, _, err := terraform.VerifyStateWithContext(ctx, opts,
		mockAWSRunner, mockFS, mockCmd)

	// 実行後の詳細なログ
	t.Logf("[DEBUG] 実行結果: exitCode=%d, err=%v, ctx.Err()=%v", exitCode, err, ctx.Err())

	// タイムアウトの確認
	if ctx.Err() == nil {
		t.Errorf("タイムアウトが発生しませんでした: ctx.Err()=nil")
	} else {
		t.Logf("[DEBUG] 成功：コンテキストエラー種別: %v", ctx.Err())
	}

	// エラー確認
	if err == nil {
		t.Errorf("エラーが返されませんでした: err=nil")
	} else {
		t.Logf("[DEBUG] 返却されたエラー種別: %v", err)
	}

	// タイムアウト時は終了コード1になるはず
	if exitCode != 1 {
		t.Errorf("タイムアウト時の終了コードが期待値と異なります: 期待=1, 実際=%d", exitCode)
	}
}
