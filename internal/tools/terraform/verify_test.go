package terraform

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/FUJI0130/go-ddd-ca/internal/tools/terraform/models"
)

func debugJson(t *testing.T, jsonData string) {
	var parsed map[string]interface{}
	err := json.Unmarshal([]byte(jsonData), &parsed)
	if err != nil {
		t.Logf("JSONパースエラー: %v", err)
		return
	}

	// 構造確認
	t.Logf("JSONルートキー: %v", getKeys(parsed))

	if values, ok := parsed["values"].(map[string]interface{}); ok {
		t.Logf("values.keys: %v", getKeys(values))

		if rootModule, ok := values["root_module"].(map[string]interface{}); ok {
			t.Logf("root_module.keys: %v", getKeys(rootModule))

			// リソースの確認
			if resources, ok := rootModule["resources"].([]interface{}); ok {
				t.Logf("root_module.resources数: %d", len(resources))
				for i, res := range resources {
					if resMap, ok := res.(map[string]interface{}); ok {
						t.Logf("  resource[%d]: %v", i, getKeys(resMap))
					}
				}
			}

			// 子モジュールの確認
			if childModules, ok := rootModule["child_modules"].([]interface{}); ok {
				t.Logf("root_module.child_modules数: %d", len(childModules))
				for i, module := range childModules {
					if moduleMap, ok := module.(map[string]interface{}); ok {
						t.Logf("  module[%d].address: %v", i, moduleMap["address"])
						t.Logf("  module[%d].keys: %v", i, getKeys(moduleMap))
					}
				}
			}
		}
	}
}

// マップのキーを取得するヘルパー関数
func getKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// TestGetTerraformResources はGetTerraformResources関数のテストです
// TestGetTerraformResources はGetTerraformResources関数のテストです
func TestGetTerraformResources(t *testing.T) {
	testCases := []struct {
		name          string
		env           string
		mockFS        *MockFileSystem
		mockCmd       *MockCommandExecutor
		expected      *models.Resources
		expectedError string
		shouldSucceed bool
	}{
		{
			name: "正常系: 全リソースが存在する",
			env:  "development",
			mockFS: NewMockFileSystem().
				WithGetwdFunc(func() (string, error) {
					return "/original/path", nil
				}).
				WithChdirFunc(func(dir string) error {
					return nil
				}),
			mockCmd: NewMockCommandExecutor().
				WithExecuteFunc(func(command string, args ...string) (string, error) {
					if command == "terraform" && args[0] == "show" && args[1] == "-json" {
						// terraform show -json の出力をモック
						return `{
							"values": {
								"root_module": {
									"resources": [],
									"child_modules": [
										{
											"address": "module.networking",
											"resources": [
												{
													"address": "module.networking.aws_vpc.main",
													"type": "aws_vpc",
													"values": {
														"id": "vpc-12345"
													}
												}
											]
										},
										{
											"address": "module.database",
											"resources": [
												{
													"address": "module.database.aws_db_instance.main",
													"type": "aws_db_instance",
													"values": {
														"id": "db-12345"
													}
												}
											]
										},
										{
											"address": "module.shared_ecs_cluster",
											"resources": [
												{
													"address": "module.shared_ecs_cluster.aws_ecs_cluster.main",
													"type": "aws_ecs_cluster",
													"values": {
														"id": "cluster-12345"
													}
												}
											]
										},
										{
											"address": "module.service_api",
											"resources": [
												{
													"address": "module.service_api.aws_ecs_service.main",
													"type": "aws_ecs_service",
													"values": {
														"id": "ecs-svc-api-12345"
													}
												}
											]
										},
										{
											"address": "module.loadbalancer_api",
											"resources": [
												{
													"address": "module.loadbalancer_api.aws_lb.main",
													"type": "aws_lb",
													"values": {
														"id": "lb-api-12345"
													}
												}
											]
										},
										{
											"address": "module.target_group_api",
											"resources": [
												{
													"address": "module.target_group_api.aws_lb_target_group.main",
													"type": "aws_lb_target_group",
													"values": {
														"id": "tg-api-12345"
													}
												}
											]
										},
										{
											"address": "module.service_graphql",
											"resources": [
												{
													"address": "module.service_graphql.aws_ecs_service.main",
													"type": "aws_ecs_service",
													"values": {
														"id": "ecs-svc-graphql-12345"
													}
												}
											]
										},
										{
											"address": "module.loadbalancer_graphql",
											"resources": [
												{
													"address": "module.loadbalancer_graphql.aws_lb.main",
													"type": "aws_lb",
													"values": {
														"id": "lb-graphql-12345"
													}
												}
											]
										},
										{
											"address": "module.target_group_graphql",
											"resources": [
												{
													"address": "module.target_group_graphql.aws_lb_target_group.main",
													"type": "aws_lb_target_group",
													"values": {
														"id": "tg-graphql-12345"
													}
												}
											]
										},
										{
											"address": "module.service_grpc",
											"resources": [
												{
													"address": "module.service_grpc.aws_ecs_service.main",
													"type": "aws_ecs_service",
													"values": {
														"id": "ecs-svc-grpc-12345"
													}
												}
											]
										},
										{
											"address": "module.loadbalancer_grpc",
											"resources": [
												{
													"address": "module.loadbalancer_grpc.aws_lb.main",
													"type": "aws_lb",
													"values": {
														"id": "lb-grpc-12345"
													}
												}
											]
										},
										{
											"address": "module.target_group_grpc",
											"resources": [
												{
													"address": "module.target_group_grpc.aws_lb_target_group.main",
													"type": "aws_lb_target_group",
													"values": {
														"id": "tg-grpc-12345"
													}
												}
											]
										}
									]
								}
							}
						}`, nil
					}
					return "", errors.New("unexpected command")
				}),
			expected: &models.Resources{
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
			expectedError: "",
			shouldSucceed: true,
		},
		{
			name: "正常系: リソースが存在しない",
			env:  "development",
			mockFS: NewMockFileSystem().
				WithGetwdFunc(func() (string, error) {
					return "/original/path", nil
				}).
				WithChdirFunc(func(dir string) error {
					return nil
				}),
			mockCmd: NewMockCommandExecutor().
				WithExecuteFunc(func(command string, args ...string) (string, error) {
					if command == "terraform" && args[0] == "show" && args[1] == "-json" {
						// 空のJSONを返す
						return `{
							"values": {
								"root_module": {
									"resources": [],
									"child_modules": []
								}
							}
						}`, nil
					}
					return "", errors.New("unexpected command")
				}),
			expected: &models.Resources{
				VPC:        0,
				RDS:        0,
				ECSCluster: 0,
				Services:   map[string]models.ServiceResources{},
			},
			expectedError: "",
			shouldSucceed: true,
		},
		{
			name: "異常系: 環境ディレクトリが見つからない",
			env:  "invalid",
			mockFS: NewMockFileSystem().
				WithGetwdFunc(func() (string, error) {
					return "/original/path", nil
				}).
				WithChdirFunc(func(dir string) error {
					return errors.New("no such directory")
				}),
			mockCmd:       NewMockCommandExecutor(),
			expected:      nil,
			expectedError: "環境ディレクトリが見つかりません",
			shouldSucceed: false,
		},
		{
			name: "異常系: カレントディレクトリ取得エラー",
			env:  "development",
			mockFS: NewMockFileSystem().
				WithGetwdFunc(func() (string, error) {
					return "", errors.New("getwd error")
				}),
			mockCmd:       NewMockCommandExecutor(),
			expected:      nil,
			expectedError: "getwd error",
			shouldSucceed: false,
		},
		{
			name: "異常系: terraform show -jsonコマンドエラー",
			env:  "development",
			mockFS: NewMockFileSystem().
				WithGetwdFunc(func() (string, error) {
					return "/original/path", nil
				}).
				WithChdirFunc(func(dir string) error {
					return nil
				}),
			mockCmd: NewMockCommandExecutor().
				WithExecuteFunc(func(command string, args ...string) (string, error) {
					if command == "terraform" && args[0] == "show" && args[1] == "-json" {
						return "", errors.New("terraform show error")
					}
					return "", errors.New("unexpected command")
				}),
			expected: &models.Resources{
				VPC:        0,
				RDS:        0,
				ECSCluster: 0,
				Services:   map[string]models.ServiceResources{},
			},
			expectedError: "",
			shouldSucceed: true, // エラー時でも空のリソース情報を返すため
		},
		{
			name: "異常系: 不正なJSON形式",
			env:  "development",
			mockFS: NewMockFileSystem().
				WithGetwdFunc(func() (string, error) {
					return "/original/path", nil
				}).
				WithChdirFunc(func(dir string) error {
					return nil
				}),
			mockCmd: NewMockCommandExecutor().
				WithExecuteFunc(func(command string, args ...string) (string, error) {
					if command == "terraform" && args[0] == "show" && args[1] == "-json" {
						return `{invalid json`, nil
					}
					return "", errors.New("unexpected command")
				}),
			expected: &models.Resources{
				VPC:        0,
				RDS:        0,
				ECSCluster: 0,
				Services:   map[string]models.ServiceResources{},
			},
			expectedError: "",
			shouldSucceed: true, // エラー時でも空のリソース情報を返すため
		},
		{
			name: "正常系: 一部のサービスのみ存在",
			env:  "development",
			mockFS: NewMockFileSystem().
				WithGetwdFunc(func() (string, error) {
					return "/original/path", nil
				}).
				WithChdirFunc(func(dir string) error {
					return nil
				}),
			mockCmd: NewMockCommandExecutor().
				WithExecuteFunc(func(command string, args ...string) (string, error) {
					if command == "terraform" && args[0] == "show" && args[1] == "-json" {
						return `{
							"values": {
								"root_module": {
									"resources": [],
									"child_modules": [
										{
											"address": "module.networking",
											"resources": [
												{
													"address": "module.networking.aws_vpc.main",
													"type": "aws_vpc",
													"values": {
														"id": "vpc-12345"
													}
												}
											]
										},
										{
											"address": "module.database",
											"resources": [
												{
													"address": "module.database.aws_db_instance.main",
													"type": "aws_db_instance",
													"values": {
														"id": "db-12345"
													}
												}
											]
										},
										{
											"address": "module.shared_ecs_cluster",
											"resources": [
												{
													"address": "module.shared_ecs_cluster.aws_ecs_cluster.main",
													"type": "aws_ecs_cluster",
													"values": {
														"id": "cluster-12345"
													}
												}
											]
										},
										{
											"address": "module.service_api",
											"resources": [
												{
													"address": "module.service_api.aws_ecs_service.main",
													"type": "aws_ecs_service",
													"values": {
														"id": "ecs-svc-api-12345"
													}
												}
											]
										},
										{
											"address": "module.loadbalancer_api",
											"resources": [
												{
													"address": "module.loadbalancer_api.aws_lb.main",
													"type": "aws_lb",
													"values": {
														"id": "lb-api-12345"
													}
												}
											]
										},
										{
											"address": "module.target_group_api",
											"resources": [
												{
													"address": "module.target_group_api.aws_lb_target_group.main",
													"type": "aws_lb_target_group",
													"values": {
														"id": "tg-api-12345"
													}
												}
											]
										}
									]
								}
							}
						}`, nil
					}
					return "", errors.New("unexpected command")
				}),
			expected: &models.Resources{
				VPC:        1,
				RDS:        1,
				ECSCluster: 1,
				Services: map[string]models.ServiceResources{
					"api": {
						ECSService:  1,
						ALB:         1,
						TargetGroup: 1,
					},
				},
			},
			expectedError: "",
			shouldSucceed: true,
		},
		{
			name: "正常系: 重複リソース参照でも正しくカウント",
			env:  "development",
			mockFS: NewMockFileSystem().
				WithGetwdFunc(func() (string, error) {
					return "/original/path", nil
				}).
				WithChdirFunc(func(dir string) error {
					return nil
				}),
			mockCmd: NewMockCommandExecutor().
				WithExecuteFunc(func(command string, args ...string) (string, error) {
					if command == "terraform" && args[0] == "show" && args[1] == "-json" {
						// 重複リソース参照のテストケース（同じVPCを異なるモジュールパスで参照）
						return `{
							"values": {
								"root_module": {
									"resources": [
										{
											"address": "aws_vpc.main",
											"type": "aws_vpc",
											"values": {
												"id": "vpc-12345"
											}
										}
									],
									"child_modules": [
										{
											"address": "module.networking",
											"resources": [
												{
													"address": "module.networking.aws_vpc.main",
													"type": "aws_vpc",
													"values": {
														"id": "vpc-12345"
													}
												}
											]
										},
										{
											"address": "module.service_api",
											"resources": [
												{
													"address": "module.service_api.aws_ecs_service.main",
													"type": "aws_ecs_service",
													"values": {
														"id": "ecs-svc-api-12345"
													}
												}
											],
											"child_modules": [
												{
													"address": "module.service_api.module.nested",
													"resources": [
														{
															"address": "module.service_api.module.nested.aws_vpc.reference",
															"type": "aws_vpc",
															"values": {
																"id": "vpc-12345"
															}
														}
													]
												}
											]
										}
									]
								}
							}
						}`, nil
					}
					return "", errors.New("unexpected command")
				}),
			expected: &models.Resources{
				VPC:        1, // 3つのVPC参照があるが、同じIDなので1つとしてカウント
				RDS:        0,
				ECSCluster: 0,
				Services: map[string]models.ServiceResources{
					"api": {
						ECSService:  1,
						ALB:         0,
						TargetGroup: 0,
					},
				},
			},
			expectedError: "",
			shouldSucceed: true,
		},
		{
			name: "正常系: 深いネストのモジュール階層",
			env:  "development",
			mockFS: NewMockFileSystem().
				WithGetwdFunc(func() (string, error) {
					return "/original/path", nil
				}).
				WithChdirFunc(func(dir string) error {
					return nil
				}),
			mockCmd: NewMockCommandExecutor().
				WithExecuteFunc(func(command string, args ...string) (string, error) {
					if command == "terraform" && args[0] == "show" && args[1] == "-json" {
						// 深いネストのモジュール階層
						return `{
							"values": {
								"root_module": {
									"resources": [],
									"child_modules": [
										{
											"address": "module.networking",
											"resources": [],
											"child_modules": [
												{
													"address": "module.networking.module.vpc",
													"resources": [
														{
															"address": "module.networking.module.vpc.aws_vpc.main",
															"type": "aws_vpc",
															"values": {
																"id": "vpc-12345"
															}
														}
													]
												}
											]
										},
										{
											"address": "module.services",
											"resources": [],
											"child_modules": [
												{
													"address": "module.services.module.api",
													"resources": [],
													"child_modules": [
														{
															"address": "module.services.module.api.module.ecs",
															"resources": [
																{
																	"address": "module.services.module.api.module.ecs.aws_ecs_service.main",
																	"type": "aws_ecs_service",
																	"values": {
																		"id": "ecs-svc-api-12345"
																	}
																}
															]
														},
														{
															"address": "module.services.module.api.module.lb",
															"resources": [
																{
																	"address": "module.services.module.api.module.lb.aws_lb.main",
																	"type": "aws_lb",
																	"values": {
																		"id": "lb-api-12345"
																	}
																}
															]
														}
													]
												}
											]
										}
									]
								}
							}
						}`, nil
					}
					return "", errors.New("unexpected command")
				}),
			expected: &models.Resources{
				VPC:        1,
				RDS:        0,
				ECSCluster: 0,
				Services: map[string]models.ServiceResources{
					"api": {
						ECSService:  1,
						ALB:         1,
						TargetGroup: 0,
					},
				},
			},
			expectedError: "",
			shouldSucceed: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// JSONデータを解析してデバッグログを出力（正常系テストのみ）
			if strings.HasPrefix(tc.name, "正常系") {
				mockCmd := tc.mockCmd
				mockCmdOutput := ""
				if mockCmd.ExecuteFunc != nil {
					// terraform show -json コマンドをシミュレート
					output, _ := mockCmd.Execute("terraform", "show", "-json")
					mockCmdOutput = output
				} else {
					mockCmdOutput = mockCmd.MockOutput
				}
				t.Logf("モックコマンド出力: %s", mockCmdOutput)
				debugJson(t, mockCmdOutput)
			}

			// 以降は元のテストコード
			result, err := GetTerraformResources(tc.env, tc.mockFS, tc.mockCmd)

			// テスト結果の詳細をログに出力
			if result != nil {
				t.Logf("結果.VPC: %d", result.VPC)
				t.Logf("結果.RDS: %d", result.RDS)
				t.Logf("結果.ECSCluster: %d", result.ECSCluster)
				t.Logf("結果.Services: %+v", result.Services)
			}
			// 期待するエラー状態の検証
			if tc.shouldSucceed {
				if err != nil {
					t.Errorf("エラーは期待されていませんでした。得られたエラー: %v", err)
				}

				// 結果の検証
				if result.VPC != tc.expected.VPC {
					t.Errorf("VPC: 期待値 %d, 実際の値 %d", tc.expected.VPC, result.VPC)
				}
				if result.RDS != tc.expected.RDS {
					t.Errorf("RDS: 期待値 %d, 実際の値 %d", tc.expected.RDS, result.RDS)
				}
				if result.ECSCluster != tc.expected.ECSCluster {
					t.Errorf("ECSCluster: 期待値 %d, 実際の値 %d", tc.expected.ECSCluster, result.ECSCluster)
				}

				// サービスの検証
				for svcType, expected := range tc.expected.Services {
					actual, ok := result.Services[svcType]
					if !ok {
						t.Errorf("サービス %s が結果に含まれていません", svcType)
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
				} else if tc.expectedError != "" && !strings.Contains(err.Error(), tc.expectedError) {
					t.Errorf("期待されたエラーメッセージ '%s' を含むエラーですが、得られたエラーは '%v'", tc.expectedError, err)
				}
			}
		})
	}
}

// TestExtractResources はextractResources関数のテストです
func TestExtractResources(t *testing.T) {
	testCases := []struct {
		name               string
		inputJSON          string
		expectedVPC        int
		expectedRDS        int
		expectedECSCluster int
		expectedServices   map[string]models.ServiceResources
	}{
		{
			name: "正常系: 基本的なリソース構造",
			inputJSON: `{
				"values": {
					"root_module": {
						"resources": [],
						"child_modules": [
							{
								"address": "module.networking",
								"resources": [
									{
										"address": "module.networking.aws_vpc.main",
										"type": "aws_vpc",
										"values": {
											"id": "vpc-12345"
										}
									}
								]
							},
							{
								"address": "module.database",
								"resources": [
									{
										"address": "module.database.aws_db_instance.main",
										"type": "aws_db_instance",
										"values": {
											"id": "db-12345"
										}
									}
								]
							}
						]
					}
				}
			}`,
			expectedVPC:        1,
			expectedRDS:        1,
			expectedECSCluster: 0,
			expectedServices:   map[string]models.ServiceResources{},
		},
		{
			name: "正常系: サービスリソースを含む",
			inputJSON: `{
				"values": {
					"root_module": {
						"resources": [],
						"child_modules": [
							{
								"address": "module.service_api",
								"resources": [
									{
										"address": "module.service_api.aws_ecs_service.main",
										"type": "aws_ecs_service",
										"values": {
											"id": "ecs-svc-api-12345"
										}
									}
								]
							},
							{
								"address": "module.loadbalancer_api",
								"resources": [
									{
										"address": "module.loadbalancer_api.aws_lb.main",
										"type": "aws_lb",
										"values": {
											"id": "lb-api-12345"
										}
									}
								]
							}
						]
					}
				}
			}`,
			expectedVPC:        0,
			expectedRDS:        0,
			expectedECSCluster: 0,
			expectedServices: map[string]models.ServiceResources{
				"api": {
					ECSService:  1,
					ALB:         1,
					TargetGroup: 0,
				},
			},
		},
		{
			name: "正常系: 複数サービスタイプ",
			inputJSON: `{
				"values": {
					"root_module": {
						"resources": [],
						"child_modules": [
							{
								"address": "module.service_api",
								"resources": [
									{
										"address": "module.service_api.aws_ecs_service.main",
										"type": "aws_ecs_service",
										"values": {
											"id": "ecs-svc-api-12345"
										}
									}
								]
							},
							{
								"address": "module.service_graphql",
								"resources": [
									{
										"address": "module.service_graphql.aws_ecs_service.main",
										"type": "aws_ecs_service",
										"values": {
											"id": "ecs-svc-graphql-12345"
										}
									}
								]
							}
						]
					}
				}
			}`,
			expectedVPC:        0,
			expectedRDS:        0,
			expectedECSCluster: 0,
			expectedServices: map[string]models.ServiceResources{
				"api": {
					ECSService:  1,
					ALB:         0,
					TargetGroup: 0,
				},
				"graphql": {
					ECSService:  1,
					ALB:         0,
					TargetGroup: 0,
				},
			},
		},
		{
			name: "正常系: リソースIDの重複",
			inputJSON: `{
				"values": {
					"root_module": {
						"resources": [
							{
								"address": "aws_vpc.main",
								"type": "aws_vpc",
								"values": {
									"id": "vpc-12345"
								}
							}
						],
						"child_modules": [
							{
								"address": "module.networking",
								"resources": [
									{
										"address": "module.networking.aws_vpc.main",
										"type": "aws_vpc",
										"values": {
											"id": "vpc-12345"
										}
									}
								]
							}
						]
					}
				}
			}`,
			expectedVPC:        1, // 同じIDなので1つにカウント
			expectedRDS:        0,
			expectedECSCluster: 0,
			expectedServices:   map[string]models.ServiceResources{},
		},
		{
			name: "正常系: 空のモジュール構造",
			inputJSON: `{
				"values": {
					"root_module": {
						"resources": [],
						"child_modules": []
					}
				}
			}`,
			expectedVPC:        0,
			expectedRDS:        0,
			expectedECSCluster: 0,
			expectedServices:   map[string]models.ServiceResources{},
		},
		{
			name: "異常系: values構造がない",
			inputJSON: `{
				"format_version": "1.0",
				"terraform_version": "1.5.7"
			}`,
			expectedVPC:        0,
			expectedRDS:        0,
			expectedECSCluster: 0,
			expectedServices:   map[string]models.ServiceResources{},
		},
		{
			name: "異常系: root_module構造がない",
			inputJSON: `{
				"values": {
					"terraform_version": "1.5.7"
				}
			}`,
			expectedVPC:        0,
			expectedRDS:        0,
			expectedECSCluster: 0,
			expectedServices:   map[string]models.ServiceResources{},
		},
		// 問題のテストケースを修正
		{
			name: "異常系: リソースに必要なキーがない",
			inputJSON: `{
                "values": {
                    "root_module": {
                        "resources": [
                            {
                                "address": "aws_vpc.main"
                            }
                        ],
                        "child_modules": []
                    }
                }
            }`,
			expectedVPC:        0,
			expectedRDS:        0,
			expectedECSCluster: 0,
			expectedServices:   map[string]models.ServiceResources{},
		},

		// 深いネスト構造のテストケースを追加
		{
			name: "正常系: 深いネスト構造のモジュール",
			inputJSON: `{
                "values": {
                    "root_module": {
                        "resources": [],
                        "child_modules": [
                            {
                                "address": "module.networking",
                                "resources": [],
                                "child_modules": [
                                    {
                                        "address": "module.networking.module.vpc",
                                        "resources": [
                                            {
                                                "address": "module.networking.module.vpc.aws_vpc.main",
                                                "type": "aws_vpc",
                                                "values": {
                                                    "id": "vpc-12345"
                                                }
                                            }
                                        ]
                                    }
                                ]
                            }
                        ]
                    }
                }
            }`,
			expectedVPC:        1,
			expectedRDS:        0,
			expectedECSCluster: 0,
			expectedServices:   map[string]models.ServiceResources{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// JSONデータをパース
			var tfState map[string]interface{}
			err := json.Unmarshal([]byte(tc.inputJSON), &tfState)
			if err != nil {
				t.Fatalf("JSONパースエラー: %v", err)
			}

			// テスト用のリソースオブジェクトとIDマップを作成
			resources := &models.Resources{
				Services: make(map[string]models.ServiceResources),
			}
			resourceIDs := make(map[string]map[string]bool)

			// デバッグ用: 入力データの構造をログ出力
			t.Logf("テストケース: %s", tc.name)
			if values, ok := tfState["values"].(map[string]interface{}); ok {
				t.Logf("values構造あり: %T", values)
				if rootModule, ok := values["root_module"].(map[string]interface{}); ok {
					t.Logf("root_module構造あり: %T", rootModule)

					if resources, ok := rootModule["resources"].([]interface{}); ok {
						t.Logf("リソース数: %d", len(resources))
						for i, res := range resources {
							resMap, ok := res.(map[string]interface{})
							if !ok {
								t.Logf("リソース[%d]がマップでない: %T", i, res)
								continue
							}
							t.Logf("リソース[%d]キー: %v", i, getKeys(resMap))
						}
					}

					if childModules, ok := rootModule["child_modules"].([]interface{}); ok {
						t.Logf("子モジュール数: %d", len(childModules))
					}
				}
			}

			// extractResources関数を呼び出し
			extractResources(tfState, resourceIDs, resources)

			// 結果を検証
			t.Logf("結果 - VPC: %d, RDS: %d, ECSCluster: %d, Services: %v",
				resources.VPC, resources.RDS, resources.ECSCluster, resources.Services)

			if resources.VPC != tc.expectedVPC {
				t.Errorf("VPC: 期待値 %d, 実際の値 %d", tc.expectedVPC, resources.VPC)
			}
			if resources.RDS != tc.expectedRDS {
				t.Errorf("RDS: 期待値 %d, 実際の値 %d", tc.expectedRDS, resources.RDS)
			}
			if resources.ECSCluster != tc.expectedECSCluster {
				t.Errorf("ECSCluster: 期待値 %d, 実際の値 %d", tc.expectedECSCluster, resources.ECSCluster)
			}

			// サービスの検証
			if len(resources.Services) != len(tc.expectedServices) {
				t.Errorf("サービス数: 期待値 %d, 実際の値 %d", len(tc.expectedServices), len(resources.Services))
			}

			for svcType, expected := range tc.expectedServices {
				actual, ok := resources.Services[svcType]
				if !ok {
					t.Errorf("サービス %s が結果に含まれていません", svcType)
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

			// 期待していないサービスがないか確認
			for svcType := range resources.Services {
				if _, ok := tc.expectedServices[svcType]; !ok {
					t.Errorf("期待していないサービス %s が結果に含まれています", svcType)
				}
			}
		})
	}
}

// TestProcessModule はprocessModule関数のテストです
func TestProcessModule(t *testing.T) {
	testCases := []struct {
		name               string
		inputModule        map[string]interface{}
		path               string
		expectedVPC        int
		expectedRDS        int
		expectedECSCluster int
		expectedServices   map[string]models.ServiceResources
	}{
		{
			name: "単一レベルのモジュール",
			inputModule: map[string]interface{}{
				"resources": []interface{}{
					map[string]interface{}{
						"address": "module.networking.aws_vpc.main",
						"type":    "aws_vpc",
						"values":  map[string]interface{}{"id": "vpc-12345"},
					},
				},
				"child_modules": []interface{}{},
			},
			path:               "module.networking",
			expectedVPC:        1,
			expectedRDS:        0,
			expectedECSCluster: 0,
			expectedServices:   map[string]models.ServiceResources{},
		},
		{
			name: "複数のリソースタイプを含むモジュール",
			inputModule: map[string]interface{}{
				"resources": []interface{}{
					map[string]interface{}{
						"address": "module.networking.aws_vpc.main",
						"type":    "aws_vpc",
						"values":  map[string]interface{}{"id": "vpc-12345"},
					},
					map[string]interface{}{
						"address": "module.database.aws_db_instance.main",
						"type":    "aws_db_instance",
						"values":  map[string]interface{}{"id": "db-12345"},
					},
				},
				"child_modules": []interface{}{},
			},
			path:               "module.mixed",
			expectedVPC:        1,
			expectedRDS:        1,
			expectedECSCluster: 0,
			expectedServices:   map[string]models.ServiceResources{},
		},
		{
			name: "サービスリソースを含むモジュール",
			inputModule: map[string]interface{}{
				"resources": []interface{}{
					map[string]interface{}{
						"address": "module.service_api.aws_ecs_service.main",
						"type":    "aws_ecs_service",
						"values":  map[string]interface{}{"id": "ecs-svc-api-12345"},
					},
					map[string]interface{}{
						"address": "module.loadbalancer_api.aws_lb.main",
						"type":    "aws_lb",
						"values":  map[string]interface{}{"id": "lb-api-12345"},
					},
					map[string]interface{}{
						"address": "module.target_group_api.aws_lb_target_group.main",
						"type":    "aws_lb_target_group",
						"values":  map[string]interface{}{"id": "tg-api-12345"},
					},
				},
				"child_modules": []interface{}{},
			},
			path:               "module.api",
			expectedVPC:        0,
			expectedRDS:        0,
			expectedECSCluster: 0,
			expectedServices: map[string]models.ServiceResources{
				"api": {
					ECSService:  1,
					ALB:         1,
					TargetGroup: 1,
				},
			},
		},
		{
			name: "子モジュールを含むモジュール",
			inputModule: map[string]interface{}{
				"resources": []interface{}{},
				"child_modules": []interface{}{
					map[string]interface{}{
						"address": "module.parent.module.child",
						"resources": []interface{}{
							map[string]interface{}{
								"address": "module.parent.module.child.aws_vpc.main",
								"type":    "aws_vpc",
								"values":  map[string]interface{}{"id": "vpc-nested"},
							},
						},
					},
				},
			},
			path:               "module.parent",
			expectedVPC:        1,
			expectedRDS:        0,
			expectedECSCluster: 0,
			expectedServices:   map[string]models.ServiceResources{},
		},
		{
			name: "複数階層のネストモジュール",
			inputModule: map[string]interface{}{
				"resources": []interface{}{},
				"child_modules": []interface{}{
					map[string]interface{}{
						"address": "module.level1",
						"resources": []interface{}{
							map[string]interface{}{
								"address": "module.level1.aws_ecs_cluster.main",
								"type":    "aws_ecs_cluster",
								"values":  map[string]interface{}{"id": "cluster-123"},
							},
						},
						"child_modules": []interface{}{
							map[string]interface{}{
								"address": "module.level1.module.level2",
								"resources": []interface{}{
									map[string]interface{}{
										"address": "module.level1.module.level2.aws_db_instance.main",
										"type":    "aws_db_instance",
										"values":  map[string]interface{}{"id": "db-nested"},
									},
								},
								"child_modules": []interface{}{
									map[string]interface{}{
										"address": "module.level1.module.level2.module.level3",
										"resources": []interface{}{
											map[string]interface{}{
												"address": "module.level1.module.level2.module.level3.aws_vpc.main",
												"type":    "aws_vpc",
												"values":  map[string]interface{}{"id": "vpc-deep-nested"},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			path:               "module.root",
			expectedVPC:        1,
			expectedRDS:        1,
			expectedECSCluster: 1,
			expectedServices:   map[string]models.ServiceResources{},
		},
		{
			name: "リソースなし、子モジュールなし",
			inputModule: map[string]interface{}{
				"resources":     []interface{}{},
				"child_modules": []interface{}{},
			},
			path:               "module.empty",
			expectedVPC:        0,
			expectedRDS:        0,
			expectedECSCluster: 0,
			expectedServices:   map[string]models.ServiceResources{},
		},
		{
			name: "リソースタイプがないリソース",
			inputModule: map[string]interface{}{
				"resources": []interface{}{
					map[string]interface{}{
						"address": "module.invalid.missing_type",
						"values":  map[string]interface{}{"id": "invalid-123"},
					},
				},
				"child_modules": []interface{}{},
			},
			path:               "module.invalid",
			expectedVPC:        0,
			expectedRDS:        0,
			expectedECSCluster: 0,
			expectedServices:   map[string]models.ServiceResources{},
		},
		{
			name: "IDがないリソース値",
			inputModule: map[string]interface{}{
				"resources": []interface{}{
					map[string]interface{}{
						"address": "module.noid.aws_vpc.main",
						"type":    "aws_vpc",
						"values":  map[string]interface{}{},
					},
				},
				"child_modules": []interface{}{},
			},
			path:               "module.noid",
			expectedVPC:        1, // IDがなくてもカウントされるはず（アドレスをIDとして使用）
			expectedRDS:        0,
			expectedECSCluster: 0,
			expectedServices:   map[string]models.ServiceResources{},
		},
		{
			name: "valuesがないリソース",
			inputModule: map[string]interface{}{
				"resources": []interface{}{
					map[string]interface{}{
						"address": "module.novalues.aws_vpc.main",
						"type":    "aws_vpc",
					},
				},
				"child_modules": []interface{}{},
			},
			path:               "module.novalues",
			expectedVPC:        1, // valuesがなくてもカウントされるはず（アドレスをIDとして使用）
			expectedRDS:        0,
			expectedECSCluster: 0,
			expectedServices:   map[string]models.ServiceResources{},
		},
		{
			name: "リソース重複（同一モジュール内）",
			inputModule: map[string]interface{}{
				"resources": []interface{}{
					map[string]interface{}{
						"address": "module.dup.aws_vpc.main1",
						"type":    "aws_vpc",
						"values":  map[string]interface{}{"id": "vpc-dup"},
					},
					map[string]interface{}{
						"address": "module.dup.aws_vpc.main2",
						"type":    "aws_vpc",
						"values":  map[string]interface{}{"id": "vpc-dup"},
					},
				},
				"child_modules": []interface{}{},
			},
			path:               "module.dup",
			expectedVPC:        1, // 同じIDなので1つとしてカウント
			expectedRDS:        0,
			expectedECSCluster: 0,
			expectedServices:   map[string]models.ServiceResources{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// テスト用のリソースオブジェクトとIDマップを作成
			resources := &models.Resources{
				Services: make(map[string]models.ServiceResources),
			}
			resourceIDs := make(map[string]map[string]bool)

			// processModule関数を呼び出し
			processModule(tc.inputModule, tc.path, resourceIDs, resources)

			// 結果を検証
			t.Logf("結果 - VPC: %d, RDS: %d, ECSCluster: %d, Services: %v",
				resources.VPC, resources.RDS, resources.ECSCluster, resources.Services)

			if resources.VPC != tc.expectedVPC {
				t.Errorf("VPC: 期待値 %d, 実際の値 %d", tc.expectedVPC, resources.VPC)
			}
			if resources.RDS != tc.expectedRDS {
				t.Errorf("RDS: 期待値 %d, 実際の値 %d", tc.expectedRDS, resources.RDS)
			}
			if resources.ECSCluster != tc.expectedECSCluster {
				t.Errorf("ECSCluster: 期待値 %d, 実際の値 %d", tc.expectedECSCluster, resources.ECSCluster)
			}

			// サービスの検証
			if len(resources.Services) != len(tc.expectedServices) {
				t.Errorf("サービス数: 期待値 %d, 実際の値 %d", len(tc.expectedServices), len(resources.Services))
			}

			for svcType, expected := range tc.expectedServices {
				actual, ok := resources.Services[svcType]
				if !ok {
					t.Errorf("サービス %s が結果に含まれていません", svcType)
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

			// 期待していないサービスがないか確認
			for svcType := range resources.Services {
				if _, ok := tc.expectedServices[svcType]; !ok {
					t.Errorf("期待していないサービス %s が結果に含まれています", svcType)
				}
			}
		})
	}
}

// TestProcessResourceByType はprocessResourceByType関数のテストです
func TestProcessResourceByType(t *testing.T) {
	testCases := []struct {
		name          string
		resType       string
		resourceID    string
		address       string
		initialState  *models.Resources
		initialIDs    map[string]map[string]bool
		expectedState *models.Resources
	}{
		{
			name:       "VPCリソースのカウント",
			resType:    "aws_vpc",
			resourceID: "vpc-12345",
			address:    "module.networking.aws_vpc.main",
			initialState: &models.Resources{
				VPC:        0,
				RDS:        0,
				ECSCluster: 0,
				Services:   map[string]models.ServiceResources{},
			},
			initialIDs: map[string]map[string]bool{},
			expectedState: &models.Resources{
				VPC:        1,
				RDS:        0,
				ECSCluster: 0,
				Services:   map[string]models.ServiceResources{},
			},
		},
		{
			name:       "RDSリソースのカウント",
			resType:    "aws_db_instance",
			resourceID: "db-12345",
			address:    "module.database.aws_db_instance.main",
			initialState: &models.Resources{
				VPC:        0,
				RDS:        0,
				ECSCluster: 0,
				Services:   map[string]models.ServiceResources{},
			},
			initialIDs: map[string]map[string]bool{},
			expectedState: &models.Resources{
				VPC:        0,
				RDS:        1,
				ECSCluster: 0,
				Services:   map[string]models.ServiceResources{},
			},
		},
		{
			name:       "ECSクラスターリソースのカウント",
			resType:    "aws_ecs_cluster",
			resourceID: "cluster-12345",
			address:    "module.shared_ecs_cluster.aws_ecs_cluster.main",
			initialState: &models.Resources{
				VPC:        0,
				RDS:        0,
				ECSCluster: 0,
				Services:   map[string]models.ServiceResources{},
			},
			initialIDs: map[string]map[string]bool{},
			expectedState: &models.Resources{
				VPC:        0,
				RDS:        0,
				ECSCluster: 1,
				Services:   map[string]models.ServiceResources{},
			},
		},
		{
			name:       "APIサービスのECSサービスカウント",
			resType:    "aws_ecs_service",
			resourceID: "ecs-svc-api-12345",
			address:    "module.service_api.aws_ecs_service.main",
			initialState: &models.Resources{
				VPC:        0,
				RDS:        0,
				ECSCluster: 0,
				Services:   map[string]models.ServiceResources{},
			},
			initialIDs: map[string]map[string]bool{},
			expectedState: &models.Resources{
				VPC:        0,
				RDS:        0,
				ECSCluster: 0,
				Services: map[string]models.ServiceResources{
					"api": {
						ECSService:  1,
						ALB:         0,
						TargetGroup: 0,
					},
				},
			},
		},
		{
			name:       "GraphQLサービスのECSサービスカウント",
			resType:    "aws_ecs_service",
			resourceID: "ecs-svc-graphql-12345",
			address:    "module.service_graphql.aws_ecs_service.main",
			initialState: &models.Resources{
				VPC:        0,
				RDS:        0,
				ECSCluster: 0,
				Services:   map[string]models.ServiceResources{},
			},
			initialIDs: map[string]map[string]bool{},
			expectedState: &models.Resources{
				VPC:        0,
				RDS:        0,
				ECSCluster: 0,
				Services: map[string]models.ServiceResources{
					"graphql": {
						ECSService:  1,
						ALB:         0,
						TargetGroup: 0,
					},
				},
			},
		},
		{
			name:       "gRPCサービスのECSサービスカウント",
			resType:    "aws_ecs_service",
			resourceID: "ecs-svc-grpc-12345",
			address:    "module.service_grpc.aws_ecs_service.main",
			initialState: &models.Resources{
				VPC:        0,
				RDS:        0,
				ECSCluster: 0,
				Services:   map[string]models.ServiceResources{},
			},
			initialIDs: map[string]map[string]bool{},
			expectedState: &models.Resources{
				VPC:        0,
				RDS:        0,
				ECSCluster: 0,
				Services: map[string]models.ServiceResources{
					"grpc": {
						ECSService:  1,
						ALB:         0,
						TargetGroup: 0,
					},
				},
			},
		},
		{
			name:       "APIサービスのALBカウント",
			resType:    "aws_lb",
			resourceID: "lb-api-12345",
			address:    "module.loadbalancer_api.aws_lb.main",
			initialState: &models.Resources{
				VPC:        0,
				RDS:        0,
				ECSCluster: 0,
				Services:   map[string]models.ServiceResources{},
			},
			initialIDs: map[string]map[string]bool{},
			expectedState: &models.Resources{
				VPC:        0,
				RDS:        0,
				ECSCluster: 0,
				Services: map[string]models.ServiceResources{
					"api": {
						ECSService:  0,
						ALB:         1,
						TargetGroup: 0,
					},
				},
			},
		},
		{
			name:       "GraphQLサービスのALBカウント",
			resType:    "aws_lb",
			resourceID: "lb-graphql-12345",
			address:    "module.loadbalancer_graphql.aws_lb.main",
			initialState: &models.Resources{
				VPC:        0,
				RDS:        0,
				ECSCluster: 0,
				Services:   map[string]models.ServiceResources{},
			},
			initialIDs: map[string]map[string]bool{},
			expectedState: &models.Resources{
				VPC:        0,
				RDS:        0,
				ECSCluster: 0,
				Services: map[string]models.ServiceResources{
					"graphql": {
						ECSService:  0,
						ALB:         1,
						TargetGroup: 0,
					},
				},
			},
		},
		{
			name:       "gRPCサービスのALBカウント",
			resType:    "aws_lb",
			resourceID: "lb-grpc-12345",
			address:    "module.loadbalancer_grpc.aws_lb.main",
			initialState: &models.Resources{
				VPC:        0,
				RDS:        0,
				ECSCluster: 0,
				Services:   map[string]models.ServiceResources{},
			},
			initialIDs: map[string]map[string]bool{},
			expectedState: &models.Resources{
				VPC:        0,
				RDS:        0,
				ECSCluster: 0,
				Services: map[string]models.ServiceResources{
					"grpc": {
						ECSService:  0,
						ALB:         1,
						TargetGroup: 0,
					},
				},
			},
		},
		{
			name:       "APIサービスのターゲットグループカウント",
			resType:    "aws_lb_target_group",
			resourceID: "tg-api-12345",
			address:    "module.target_group_api.aws_lb_target_group.main",
			initialState: &models.Resources{
				VPC:        0,
				RDS:        0,
				ECSCluster: 0,
				Services:   map[string]models.ServiceResources{},
			},
			initialIDs: map[string]map[string]bool{},
			expectedState: &models.Resources{
				VPC:        0,
				RDS:        0,
				ECSCluster: 0,
				Services: map[string]models.ServiceResources{
					"api": {
						ECSService:  0,
						ALB:         0,
						TargetGroup: 1,
					},
				},
			},
		},
		{
			name:       "GraphQLサービスのターゲットグループカウント",
			resType:    "aws_lb_target_group",
			resourceID: "tg-graphql-12345",
			address:    "module.target_group_graphql.aws_lb_target_group.main",
			initialState: &models.Resources{
				VPC:        0,
				RDS:        0,
				ECSCluster: 0,
				Services:   map[string]models.ServiceResources{},
			},
			initialIDs: map[string]map[string]bool{},
			expectedState: &models.Resources{
				VPC:        0,
				RDS:        0,
				ECSCluster: 0,
				Services: map[string]models.ServiceResources{
					"graphql": {
						ECSService:  0,
						ALB:         0,
						TargetGroup: 1,
					},
				},
			},
		},
		{
			name:       "gRPCサービスのターゲットグループカウント",
			resType:    "aws_lb_target_group",
			resourceID: "tg-grpc-12345",
			address:    "module.target_group_grpc.aws_lb_target_group.main",
			initialState: &models.Resources{
				VPC:        0,
				RDS:        0,
				ECSCluster: 0,
				Services:   map[string]models.ServiceResources{},
			},
			initialIDs: map[string]map[string]bool{},
			expectedState: &models.Resources{
				VPC:        0,
				RDS:        0,
				ECSCluster: 0,
				Services: map[string]models.ServiceResources{
					"grpc": {
						ECSService:  0,
						ALB:         0,
						TargetGroup: 1,
					},
				},
			},
		},
		{
			name:       "APIサービスの既存サービスに追加",
			resType:    "aws_ecs_service",
			resourceID: "ecs-svc-api-12345",
			address:    "module.service_api.aws_ecs_service.main",
			initialState: &models.Resources{
				VPC:        0,
				RDS:        0,
				ECSCluster: 0,
				Services: map[string]models.ServiceResources{
					"api": {
						ECSService:  0,
						ALB:         1,
						TargetGroup: 1,
					},
				},
			},
			initialIDs: map[string]map[string]bool{},
			expectedState: &models.Resources{
				VPC:        0,
				RDS:        0,
				ECSCluster: 0,
				Services: map[string]models.ServiceResources{
					"api": {
						ECSService:  1,
						ALB:         1,
						TargetGroup: 1,
					},
				},
			},
		},
		{
			name:       "無効なリソースタイプ",
			resType:    "aws_invalid_type",
			resourceID: "invalid-12345",
			address:    "module.invalid.aws_invalid_type.main",
			initialState: &models.Resources{
				VPC:        0,
				RDS:        0,
				ECSCluster: 0,
				Services:   map[string]models.ServiceResources{},
			},
			initialIDs: map[string]map[string]bool{},
			expectedState: &models.Resources{
				VPC:        0,
				RDS:        0,
				ECSCluster: 0,
				Services:   map[string]models.ServiceResources{},
			},
		},
		{
			name:       "重複ID（同一リソース）",
			resType:    "aws_vpc",
			resourceID: "vpc-12345",
			address:    "module.networking.aws_vpc.main",
			initialState: &models.Resources{
				VPC:        1,
				RDS:        0,
				ECSCluster: 0,
				Services:   map[string]models.ServiceResources{},
			},
			initialIDs: map[string]map[string]bool{
				"aws_vpc": {"vpc-12345": true},
			},
			expectedState: &models.Resources{
				VPC:        1, // 重複IDなので増加しない
				RDS:        0,
				ECSCluster: 0,
				Services:   map[string]models.ServiceResources{},
			},
		},
		{
			name:       "不明なサービスタイプのECSサービス",
			resType:    "aws_ecs_service",
			resourceID: "ecs-svc-unknown-12345",
			address:    "module.service_unknown.aws_ecs_service.main",
			initialState: &models.Resources{
				VPC:        0,
				RDS:        0,
				ECSCluster: 0,
				Services:   map[string]models.ServiceResources{},
			},
			initialIDs: map[string]map[string]bool{},
			expectedState: &models.Resources{
				VPC:        0,
				RDS:        0,
				ECSCluster: 0,
				Services:   map[string]models.ServiceResources{},
			},
		},
		{
			name:       "不明なサービスタイプのALB",
			resType:    "aws_lb",
			resourceID: "lb-unknown-12345",
			address:    "module.loadbalancer_unknown.aws_lb.main",
			initialState: &models.Resources{
				VPC:        0,
				RDS:        0,
				ECSCluster: 0,
				Services:   map[string]models.ServiceResources{},
			},
			initialIDs: map[string]map[string]bool{},
			expectedState: &models.Resources{
				VPC:        0,
				RDS:        0,
				ECSCluster: 0,
				Services:   map[string]models.ServiceResources{},
			},
		},
		{
			name:       "不明なサービスタイプのターゲットグループ",
			resType:    "aws_lb_target_group",
			resourceID: "tg-unknown-12345",
			address:    "module.target_group_unknown.aws_lb_target_group.main",
			initialState: &models.Resources{
				VPC:        0,
				RDS:        0,
				ECSCluster: 0,
				Services:   map[string]models.ServiceResources{},
			},
			initialIDs: map[string]map[string]bool{},
			expectedState: &models.Resources{
				VPC:        0,
				RDS:        0,
				ECSCluster: 0,
				Services:   map[string]models.ServiceResources{},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// テスト前の状態をコピー
			resources := &models.Resources{
				VPC:        tc.initialState.VPC,
				RDS:        tc.initialState.RDS,
				ECSCluster: tc.initialState.ECSCluster,
				Services:   make(map[string]models.ServiceResources),
			}
			for svcType, svc := range tc.initialState.Services {
				resources.Services[svcType] = svc
			}

			// リソースIDマップをコピー
			resourceIDs := make(map[string]map[string]bool)
			for resType, ids := range tc.initialIDs {
				resourceIDs[resType] = make(map[string]bool)
				for id, val := range ids {
					resourceIDs[resType][id] = val
				}
			}

			// processResourceByType関数を呼び出し
			processResourceByType(tc.resType, tc.resourceID, tc.address, resourceIDs, resources)

			// 結果を検証
			t.Logf("結果 - VPC: %d, RDS: %d, ECSCluster: %d, Services: %v",
				resources.VPC, resources.RDS, resources.ECSCluster, resources.Services)

			if resources.VPC != tc.expectedState.VPC {
				t.Errorf("VPC: 期待値 %d, 実際の値 %d", tc.expectedState.VPC, resources.VPC)
			}
			if resources.RDS != tc.expectedState.RDS {
				t.Errorf("RDS: 期待値 %d, 実際の値 %d", tc.expectedState.RDS, resources.RDS)
			}
			if resources.ECSCluster != tc.expectedState.ECSCluster {
				t.Errorf("ECSCluster: 期待値 %d, 実際の値 %d", tc.expectedState.ECSCluster, resources.ECSCluster)
			}

			// サービスの検証
			if len(resources.Services) != len(tc.expectedState.Services) {
				t.Errorf("サービス数: 期待値 %d, 実際の値 %d", len(tc.expectedState.Services), len(resources.Services))
			}

			for svcType, expected := range tc.expectedState.Services {
				actual, ok := resources.Services[svcType]
				if !ok {
					t.Errorf("サービス %s が結果に含まれていません", svcType)
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

			// 期待していないサービスがないか確認
			for svcType := range resources.Services {
				if _, ok := tc.expectedState.Services[svcType]; !ok {
					t.Errorf("期待していないサービス %s が結果に含まれています", svcType)
				}
			}

			// リソースIDマップの検証（処理済みとしてマークされているか）
			if len(resourceIDs[tc.resType]) == 0 && tc.resType != "aws_invalid_type" {
				t.Errorf("リソースタイプ %s のIDマップが作成されていません", tc.resType)
			}
			if _, ok := resourceIDs[tc.resType]; ok {
				if !resourceIDs[tc.resType][tc.resourceID] && !tc.initialIDs[tc.resType][tc.resourceID] {
					t.Errorf("リソースID %s が処理済みとしてマークされていません", tc.resourceID)
				}
			}
		})
	}
}

// TestCompareResources はCompareResources関数のテストです
func TestCompareResources(t *testing.T) {
	testCases := []struct {
		name          string
		awsResources  *models.Resources
		tfResources   *models.Resources
		expectedCount int // 一致するリソースの数
		allMatch      bool
		shouldSucceed bool
	}{
		{
			name: "正常系: 全リソースが一致",
			awsResources: &models.Resources{
				VPC:        1,
				RDS:        1,
				ECSCluster: 1,
				Services: map[string]models.ServiceResources{
					"api": {
						ECSService:  1,
						ALB:         1,
						TargetGroup: 1,
					},
				},
			},
			tfResources: &models.Resources{
				VPC:        1,
				RDS:        1,
				ECSCluster: 1,
				Services: map[string]models.ServiceResources{
					"api": {
						ECSService:  1,
						ALB:         1,
						TargetGroup: 1,
					},
				},
			},
			expectedCount: 12, // 常に12個の結果が返される
			allMatch:      true,
			shouldSucceed: true,
		},
		{
			name: "正常系: 一部リソースが不一致",
			awsResources: &models.Resources{
				VPC:        1,
				RDS:        0, // 不一致
				ECSCluster: 1,
				Services: map[string]models.ServiceResources{
					"api": {
						ECSService:  1,
						ALB:         1,
						TargetGroup: 0, // 不一致
					},
				},
			},
			tfResources: &models.Resources{
				VPC:        1,
				RDS:        1, // 不一致
				ECSCluster: 1,
				Services: map[string]models.ServiceResources{
					"api": {
						ECSService:  1,
						ALB:         1,
						TargetGroup: 1, // 不一致
					},
				},
			},
			expectedCount: 12, // 常に12個の結果が返される
			allMatch:      false,
			shouldSucceed: true,
		},
		{
			name: "正常系: AWSにないサービスがTerraformにある",
			awsResources: &models.Resources{
				VPC:        1,
				RDS:        1,
				ECSCluster: 1,
				Services: map[string]models.ServiceResources{
					"api": {
						ECSService:  1,
						ALB:         1,
						TargetGroup: 1,
					},
					// graphqlが存在しない
				},
			},
			tfResources: &models.Resources{
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
				},
			},
			expectedCount: 12, // 常に12個の結果が返される
			allMatch:      false,
			shouldSucceed: true,
		},
		{
			name: "正常系: 空のリソース",
			awsResources: &models.Resources{
				VPC:        0,
				RDS:        0,
				ECSCluster: 0,
				Services:   map[string]models.ServiceResources{},
			},
			tfResources: &models.Resources{
				VPC:        0,
				RDS:        0,
				ECSCluster: 0,
				Services:   map[string]models.ServiceResources{},
			},
			expectedCount: 12, // 常に12個の結果が返される
			allMatch:      true,
			shouldSucceed: true,
		},
		{
			name: "エッジケース: nilサービスマップ",
			awsResources: &models.Resources{
				VPC:        1,
				RDS:        1,
				ECSCluster: 1,
				Services:   nil,
			},
			tfResources: &models.Resources{
				VPC:        1,
				RDS:        1,
				ECSCluster: 1,
				Services:   map[string]models.ServiceResources{},
			},
			expectedCount: 12, // 常に12個の結果が返される
			allMatch:      true,
			shouldSucceed: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			results := CompareResources(tc.awsResources, tc.tfResources)

			// 結果の長さを検証
			if len(results) != tc.expectedCount {
				t.Errorf("結果の長さ: 期待値 %d, 実際の値 %d", tc.expectedCount, len(results))
			}

			// 全ての結果が一致するかを検証
			allMatch := true
			matchCount := 0
			for _, result := range results {
				if result.IsMatch {
					matchCount++
				} else {
					allMatch = false
				}
			}

			if allMatch != tc.allMatch {
				t.Errorf("全ての結果が一致するか: 期待値 %v, 実際の値 %v", tc.allMatch, allMatch)
			}
		})
	}
}

// TestRunTerraformPlan はRunTerraformPlan関数のテストです
// testExitError は*exec.ExitErrorをモックするためのカスタム型
// verify_context_test.goでも使用
type testExitError struct {
	exitCode int
}

func (e *testExitError) Error() string {
	return fmt.Sprintf("exit status %d", e.exitCode)
}

func (e *testExitError) ExitCode() int {
	return e.exitCode
}

// TestRunTerraformPlan はRunTerraformPlan関数のテストです
func TestRunTerraformPlan(t *testing.T) {
	testCases := []struct {
		name           string
		env            string
		mockFS         *MockFileSystem
		mockCmd        *MockCommandExecutor
		expectedCode   int
		expectedOutput string
		expectedError  string
		shouldSucceed  bool
	}{
		{
			name: "正常系: 変更なし",
			env:  "development",
			mockFS: NewMockFileSystem().
				WithGetwdFunc(func() (string, error) {
					return "/original/path", nil
				}).
				WithChdirFunc(func(dir string) error {
					return nil
				}),
			mockCmd: NewMockCommandExecutor().
				WithExecuteFunc(func(command string, args ...string) (string, error) {
					if command == "terraform" && args[0] == "plan" {
						return "No changes. Infrastructure is up-to-date.", nil
					}
					return "", errors.New("unexpected command")
				}),
			expectedCode:   0,
			expectedOutput: "No changes. Infrastructure is up-to-date.",
			expectedError:  "",
			shouldSucceed:  true,
		},
		{
			name: "正常系: 変更あり",
			env:  "development",
			mockFS: NewMockFileSystem().
				WithGetwdFunc(func() (string, error) {
					return "/original/path", nil
				}).
				WithChdirFunc(func(dir string) error {
					return nil
				}),
			mockCmd: NewMockCommandExecutor().
				WithExecuteFunc(func(command string, args ...string) (string, error) {
					if command == "terraform" && args[0] == "plan" {
						// ここで実行するコマンドに対して "Plan: 1 to add" を返し、
						// エラーとして普通のエラーを返す（ExitError型ではない）
						// verify.goの実装では、ExitError以外のエラーの場合は
						// 終了コード1とエラーが返されるため、これに合わせる
						return "Plan: 1 to add, 0 to change, 0 to destroy.", nil
					}
					return "", errors.New("unexpected command")
				}),
			expectedCode:   0, // エラーがない場合は0
			expectedOutput: "Plan: 1 to add, 0 to change, 0 to destroy.",
			expectedError:  "",
			shouldSucceed:  true,
		},
		{
			name: "異常系: 環境ディレクトリが見つからない",
			env:  "invalid",
			mockFS: NewMockFileSystem().
				WithGetwdFunc(func() (string, error) {
					return "/original/path", nil
				}).
				WithChdirFunc(func(dir string) error {
					return errors.New("no such directory")
				}),
			mockCmd:        NewMockCommandExecutor(),
			expectedCode:   1,
			expectedOutput: "",
			expectedError:  "環境ディレクトリが見つかりません",
			shouldSucceed:  false,
		},
		{
			name: "異常系: カレントディレクトリ取得エラー",
			env:  "development",
			mockFS: NewMockFileSystem().
				WithGetwdFunc(func() (string, error) {
					return "", errors.New("getwd error")
				}),
			mockCmd:        NewMockCommandExecutor(),
			expectedCode:   1,
			expectedOutput: "",
			expectedError:  "getwd error",
			shouldSucceed:  false,
		},
		{
			name: "異常系: terraformコマンドエラー（非ExitError）",
			env:  "development",
			mockFS: NewMockFileSystem().
				WithGetwdFunc(func() (string, error) {
					return "/original/path", nil
				}).
				WithChdirFunc(func(dir string) error {
					return nil
				}),
			mockCmd: NewMockCommandExecutor().
				WithExecuteFunc(func(command string, args ...string) (string, error) {
					if command == "terraform" && args[0] == "plan" {
						return "Error initializing Terraform", errors.New("terraform init error")
					}
					return "", errors.New("unexpected command")
				}),
			expectedCode:   1,
			expectedOutput: "Error initializing Terraform",
			expectedError:  "terraform init error",
			shouldSucceed:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// デバッグログを追加
			t.Logf("テストケース: %s", tc.name)

			exitCode, output, err := RunTerraformPlan(tc.env, tc.mockFS, tc.mockCmd)

			// デバッグログ
			t.Logf("結果: exitCode=%d, output=%s, err=%v", exitCode, output, err)

			// 終了コードの検証
			if exitCode != tc.expectedCode {
				t.Errorf("終了コード: 期待値 %d, 実際の値 %d", tc.expectedCode, exitCode)
			}

			// 出力の検証
			if !strings.Contains(output, tc.expectedOutput) {
				t.Errorf("出力: 期待値 '%s' を含む, 実際の値 '%s'", tc.expectedOutput, output)
			}

			// エラーの検証
			if tc.shouldSucceed {
				if err != nil {
					t.Errorf("エラーは期待されていませんでした。得られたエラー: %v", err)
				}
			} else {
				if err == nil {
					t.Error("エラーが期待されましたが、エラーはありませんでした")
				} else if tc.expectedError != "" && !strings.Contains(err.Error(), tc.expectedError) {
					t.Errorf("期待されたエラーメッセージ '%s' を含むエラーですが、得られたエラーは '%v'", tc.expectedError, err)
				}
			}
		})
	}
}

// TestVerifyState はVerifyState関数のテストです
func TestVerifyState(t *testing.T) {
	testCases := []struct {
		name          string
		opts          models.VerifyOptions
		mockAwsRunner *MockAWSRunner
		mockFS        *MockFileSystem
		mockCmd       *MockCommandExecutor
		expectedCode  int
		expectedError string
		shouldMatch   bool
		shouldSucceed bool
	}{
		{
			name: "正常系: 全リソースが一致、terraformプランなし",
			opts: models.VerifyOptions{
				Environment:       "development",
				Debug:             false,
				SkipTerraformPlan: true,
			},
			mockAwsRunner: NewMockAWSRunner().
				WithCommandFunc(func(args ...string) (string, error) {
					// すべての引数を結合してパターンマッチに使用
					cmdLine := strings.Join(args, " ")

					// グラフィカルサービスに関するもの
					if strings.Contains(cmdLine, "graphql") || strings.Contains(cmdLine, "grpc") {
						return "0", nil
					}

					// api サービスとコアリソース
					return "1", nil
				}),
			mockFS: NewMockFileSystem().
				WithGetwdFunc(func() (string, error) {
					return "/original/path", nil
				}).
				WithChdirFunc(func(dir string) error {
					return nil
				}),
			mockCmd: NewMockCommandExecutor().
				WithExecuteFunc(func(command string, args ...string) (string, error) {
					if command == "terraform" && args[0] == "state" && args[1] == "list" {
						// api関連のリソースのみを返す
						return `module.networking.aws_vpc.main
module.database.aws_db_instance.main
module.shared_ecs_cluster.aws_ecs_cluster.main
module.service_api.aws_ecs_service.main
module.loadbalancer_api.aws_lb.main
module.target_group_api.aws_lb_target_group.main`, nil
					}
					return "", errors.New("unexpected command")
				}),
			expectedCode:  0,
			expectedError: "",
			shouldMatch:   true,
			shouldSucceed: true,
		},
		{
			name: "正常系: 全リソースが一致、terraformプランあり（変更なし）",
			opts: models.VerifyOptions{
				Environment:       "development",
				Debug:             false,
				SkipTerraformPlan: false,
			},
			mockAwsRunner: NewMockAWSRunner().
				WithCommandFunc(func(args ...string) (string, error) {
					// すべての引数を結合してパターンマッチに使用
					cmdLine := strings.Join(args, " ")

					// グラフィカルサービスに関するもの
					if strings.Contains(cmdLine, "graphql") || strings.Contains(cmdLine, "grpc") {
						return "0", nil
					}

					// api サービスとコアリソース
					return "1", nil
				}),
			mockFS: NewMockFileSystem().
				WithGetwdFunc(func() (string, error) {
					return "/original/path", nil
				}).
				WithChdirFunc(func(dir string) error {
					return nil
				}),
			mockCmd: NewMockCommandExecutor().
				WithExecuteFunc(func(command string, args ...string) (string, error) {
					if command == "terraform" && args[0] == "state" && args[1] == "list" {
						return `module.networking.aws_vpc.main
module.database.aws_db_instance.main
module.shared_ecs_cluster.aws_ecs_cluster.main
module.service_api.aws_ecs_service.main
module.loadbalancer_api.aws_lb.main
module.target_group_api.aws_lb_target_group.main`, nil
					} else if command == "terraform" && args[0] == "plan" {
						return "No changes. Infrastructure is up-to-date.", nil
					}
					return "", errors.New("unexpected command")
				}),
			expectedCode:  0,
			expectedError: "",
			shouldMatch:   true,
			shouldSucceed: true,
		},
		{
			name: "異常系: AWSリソース取得エラー",
			opts: models.VerifyOptions{
				Environment: "development",
			},
			mockAwsRunner: NewMockAWSRunner().
				WithCommandFunc(func(args ...string) (string, error) {
					return "", errors.New("AWS CLI error")
				}),
			mockFS:        NewMockFileSystem(),
			mockCmd:       NewMockCommandExecutor(),
			expectedCode:  1,
			expectedError: "AWS情報取得エラー",
			shouldMatch:   false,
			shouldSucceed: false,
		},
		{
			name: "異常系: Terraform状態取得エラー",
			opts: models.VerifyOptions{
				Environment: "development",
			},
			mockAwsRunner: NewMockAWSRunner().
				WithCommandFunc(func(args ...string) (string, error) {
					return "1", nil
				}),
			mockFS: NewMockFileSystem().
				WithGetwdFunc(func() (string, error) {
					return "/original/path", nil
				}).
				WithChdirFunc(func(dir string) error {
					return errors.New("no such directory")
				}),
			mockCmd:       NewMockCommandExecutor(),
			expectedCode:  1,
			expectedError: "Terraform情報取得エラー",
			shouldMatch:   false,
			shouldSucceed: false,
		},
		{
			name: "正常系: リソース不一致",
			opts: models.VerifyOptions{
				Environment: "development",
			},
			mockAwsRunner: NewMockAWSRunner().
				WithCommandFunc(func(args ...string) (string, error) {
					return "1", nil
				}),
			mockFS: NewMockFileSystem().
				WithGetwdFunc(func() (string, error) {
					return "/original/path", nil
				}).
				WithChdirFunc(func(dir string) error {
					return nil
				}),
			mockCmd: NewMockCommandExecutor().
				WithExecuteFunc(func(command string, args ...string) (string, error) {
					if command == "terraform" && args[0] == "state" && args[1] == "list" {
						// 一部のリソースが存在しない
						return `module.networking.aws_vpc.main`, nil
					}
					return "", errors.New("unexpected command")
				}),
			expectedCode:  2, // 不一致の場合は2を返す
			expectedError: "",
			shouldMatch:   false,
			shouldSucceed: true,
		},
		{
			name: "正常系: 空の環境（リソースなし）",
			opts: models.VerifyOptions{
				Environment: "development",
			},
			mockAwsRunner: NewMockAWSRunner().
				WithCommandFunc(func(args ...string) (string, error) {
					return "0", nil // リソースが存在しない
				}),
			mockFS: NewMockFileSystem().
				WithGetwdFunc(func() (string, error) {
					return "/original/path", nil
				}).
				WithChdirFunc(func(dir string) error {
					return nil
				}),
			mockCmd: NewMockCommandExecutor().
				WithExecuteFunc(func(command string, args ...string) (string, error) {
					if command == "terraform" && args[0] == "state" && args[1] == "list" {
						return "", nil // 状態ファイルも空
					}
					return "", errors.New("unexpected command")
				}),
			expectedCode:  0, // 一致する場合は0
			expectedError: "",
			shouldMatch:   true,
			shouldSucceed: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("テストケース: %s", tc.name)

			// デバッグ: mockAwsRunnerの動作確認
			if tc.name == "正常系: 全リソースが一致、terraformプランなし" {
				// graphqlリソース用のモック動作を確認
				output, err := tc.mockAwsRunner.RunCommand("elbv2", "describe-target-groups", "--query", "length(TargetGroups[?contains(TargetGroupName, 'graphql')])")
				t.Logf("AWS Runner Debug - graphql TargetGroup: output=%s, err=%v", output, err)

				// apiリソース用のモック動作を確認
				output, err = tc.mockAwsRunner.RunCommand("elbv2", "describe-target-groups", "--query", "length(TargetGroups[?contains(TargetGroupName, 'api')])")
				t.Logf("AWS Runner Debug - api TargetGroup: output=%s, err=%v", output, err)
			}

			exitCode, results, err := VerifyState(tc.opts, tc.mockAwsRunner, tc.mockFS, tc.mockCmd)

			// 結果を詳細にログ出力
			t.Logf("結果: exitCode=%d, err=%v", exitCode, err)
			if results != nil {
				for _, r := range results {
					t.Logf("リソース結果: %s, AWS=%d, TF=%d, Match=%v",
						r.ResourceName, r.AWSCount, r.TerraformCount, r.IsMatch)
				}
			}

			// 終了コードの検証
			if exitCode != tc.expectedCode {
				t.Errorf("終了コード: 期待値 %d, 実際の値 %d", tc.expectedCode, exitCode)
			}

			// エラーの検証
			if tc.shouldSucceed {
				if err != nil {
					t.Errorf("エラーは期待されていませんでした。得られたエラー: %v", err)
				}

				// 成功する場合は結果も検証
				if results == nil {
					t.Error("結果がnilです")
				} else {
					// 全て一致するかどうかの検証
					allMatch := true
					for _, result := range results {
						if !result.IsMatch {
							t.Logf("不一致のリソース: %s (AWS=%d, TF=%d)",
								result.ResourceName, result.AWSCount, result.TerraformCount)
							allMatch = false
						}
					}

					if allMatch != tc.shouldMatch {
						t.Errorf("全てのリソースが一致するか: 期待値 %v, 実際の値 %v", tc.shouldMatch, allMatch)
					}
				}
			} else {
				if err == nil {
					t.Error("エラーが期待されましたが、エラーはありませんでした")
				} else if tc.expectedError != "" && !strings.Contains(err.Error(), tc.expectedError) {
					t.Errorf("期待されたエラーメッセージ '%s' を含むエラーですが、得られたエラーは '%v'", tc.expectedError, err)
				}
			}
		})
	}
}
