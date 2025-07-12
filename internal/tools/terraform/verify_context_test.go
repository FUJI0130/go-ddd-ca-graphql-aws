package terraform

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/FUJI0130/go-ddd-ca/internal/tools/terraform/models"
)

// TestRunTerraformPlanWithContext はRunTerraformPlanWithContext関数のテストです
func TestRunTerraformPlanWithContext(t *testing.T) {
	testCases := []struct {
		name          string
		env           string
		mockFS        *MockFileSystem
		mockCmd       *MockContextCommandExecutor
		timeout       time.Duration
		delay         time.Duration
		expectedCode  int
		expectTimeout bool
		success       bool
	}{
		{
			name:    "正常系: 変更なし",
			env:     "development",
			timeout: 5 * time.Second,
			delay:   0,
			mockFS: NewMockFileSystem().
				WithGetwdFunc(func() (string, error) {
					return "/original/path", nil
				}).
				WithChdirFunc(func(dir string) error {
					return nil
				}),
			mockCmd: NewMockContextCommandExecutor().
				WithExecuteWithContextFunc(func(ctx context.Context, command string, args ...string) (string, error) {
					if command == "terraform" && args[0] == "plan" {
						return "No changes. Infrastructure is up-to-date.", nil
					}
					return "", errors.New("unexpected command")
				}),
			expectedCode:  0,
			expectTimeout: false,
			success:       true,
		},
		{
			name:    "タイムアウト発生",
			env:     "development",
			timeout: 1 * time.Second,
			delay:   2 * time.Second,
			mockFS: NewMockFileSystem().
				WithGetwdFunc(func() (string, error) {
					return "/original/path", nil
				}).
				WithChdirFunc(func(dir string) error {
					return nil
				}),
			mockCmd: NewMockContextCommandExecutor().
				WithExecuteWithContextFunc(func(ctx context.Context, command string, args ...string) (string, error) {
					if command == "terraform" && args[0] == "plan" {
						// 遅延を入れて意図的にタイムアウトさせる
						time.Sleep(2 * time.Second)
						// ここには到達しないはず
						return "", errors.New("should not reach here")
					}
					return "", errors.New("unexpected command")
				}),
			expectedCode:  1,
			expectTimeout: true,
			success:       true,
		},
		{
			name:    "正常系: 変更あり",
			env:     "development",
			timeout: 5 * time.Second,
			delay:   0,
			mockFS: NewMockFileSystem().
				WithGetwdFunc(func() (string, error) {
					return "/original/path", nil
				}).
				WithChdirFunc(func(dir string) error {
					return nil
				}),
			mockCmd: NewMockContextCommandExecutor().
				WithExecuteWithContextFunc(func(ctx context.Context, command string, args ...string) (string, error) {
					if command == "terraform" && args[0] == "plan" {
						// 終了コード2を返すためのエラー
						return "Plan: 1 to add, 0 to change, 0 to destroy.", &testExitError{exitCode: 2}
					}
					return "", errors.New("unexpected command")
				}),
			expectedCode:  2,
			expectTimeout: false,
			success:       true,
		},
		{
			name:    "コンテキストキャンセル発生",
			env:     "development",
			timeout: 5 * time.Second, // タイムアウトは長めに設定
			delay:   0,
			mockFS: NewMockFileSystem().
				WithGetwdFunc(func() (string, error) {
					return "/original/path", nil
				}).
				WithChdirFunc(func(dir string) error {
					return nil
				}),
			mockCmd: NewMockContextCommandExecutor().
				WithExecuteWithContextFunc(func(ctx context.Context, command string, args ...string) (string, error) {
					if command == "terraform" && args[0] == "plan" {
						// コンテキストキャンセルを確認
						ctx, cancel := context.WithCancel(ctx)
						cancel() // 明示的にキャンセル
						return "", ctx.Err()
					}
					return "", errors.New("unexpected command")
				}),
			expectedCode:  1,
			expectTimeout: false, // タイムアウトではなくキャンセル
			success:       true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// タイムアウト付きコンテキスト作成
			ctx, cancel := context.WithTimeout(context.Background(), tc.timeout)
			defer cancel()

			exitCode, output, err := RunTerraformPlanWithContext(ctx, tc.env, tc.mockFS, tc.mockCmd)

			// タイムアウト確認
			isTimeout := err != nil && ctx.Err() == context.DeadlineExceeded

			if isTimeout != tc.expectTimeout {
				t.Errorf("タイムアウト状態: 期待=%v, 実際=%v", tc.expectTimeout, isTimeout)
			}

			// 終了コードの検証
			if !isTimeout && exitCode != tc.expectedCode {
				t.Errorf("終了コード: 期待値 %d, 実際の値 %d", tc.expectedCode, exitCode)
			}

			// 成功/失敗の検証
			testSuccess := isTimeout == tc.expectTimeout
			if !isTimeout {
				testSuccess = testSuccess && exitCode == tc.expectedCode
			}

			if testSuccess != tc.success {
				t.Errorf("テスト成功判定: 期待=%v, 実際=%v", tc.success, testSuccess)
				if err != nil {
					t.Logf("エラー: %v", err)
				}
				t.Logf("出力: %s", output)
			}
		})
	}
}

// TestVerifyStateWithContext はVerifyStateWithContext関数のテストです
// TestVerifyStateWithContext はVerifyStateWithContext関数のテストです
func TestVerifyStateWithContext(t *testing.T) {
	testCases := []struct {
		name          string
		opts          models.VerifyOptions
		mockAWSRunner *MockAWSRunner
		mockFS        *MockFileSystem
		mockCmd       *MockContextCommandExecutor
		timeout       time.Duration
		expectedCode  int
		expectTimeout bool
		success       bool
	}{
		{
			name: "正常系: リソース一致かつスキップplan",
			opts: models.VerifyOptions{
				Environment:       "development",
				SkipTerraformPlan: true,
			},
			timeout: 5 * time.Second,
			mockAWSRunner: NewMockAWSRunner().
				WithCommandFunc(func(args ...string) (string, error) {
					// AWSリソース数を調整
					if strings.Contains(args[0], "ec2") && strings.Contains(args[0], "vpcs") {
						return "0", nil // VPC
					} else if strings.Contains(args[0], "rds") {
						return "0", nil // RDS
					} else if strings.Contains(args[0], "ecs") && strings.Contains(args[0], "clusters") {
						return "0", nil // ECSクラスター
					} else if strings.Contains(args[0], "ecs") && strings.Contains(args[0], "services") {
						return "0", nil // ECSサービス
					} else if strings.Contains(args[0], "elbv2") && strings.Contains(args[0], "load-balancers") {
						return "0", nil // ALB
					} else if strings.Contains(args[0], "elbv2") && strings.Contains(args[0], "target-groups") {
						return "0", nil // ターゲットグループ
					}
					return "0", nil
				}),
			mockFS: NewMockFileSystem().
				WithGetwdFunc(func() (string, error) {
					return "/original/path", nil
				}).
				WithChdirFunc(func(dir string) error {
					return nil
				}),
			mockCmd: NewMockContextCommandExecutor().
				WithExecuteFunc(func(command string, args ...string) (string, error) {
					if command == "terraform" && args[0] == "show" && args[1] == "-json" {
						// 空のTerraform状態（すべて0）
						return `{
                            "values": {
                                "root_module": {
                                    "resources": []
                                }
                            }
                        }`, nil
					}
					return "", fmt.Errorf("unexpected command in Execute: %s %v", command, args)
				}).
				WithExecuteWithContextFunc(func(ctx context.Context, command string, args ...string) (string, error) {
					if command == "terraform" && args[0] == "show" && args[1] == "-json" {
						// 空のTerraform状態（すべて0）
						return `{
                            "values": {
                                "root_module": {
                                    "resources": []
                                }
                            }
                        }`, nil
					}
					return "", fmt.Errorf("unexpected command in ExecuteWithContext: %s %v", command, args)
				}),
			expectedCode:  0, // リソースがすべて0で一致する場合は0
			expectTimeout: false,
			success:       true,
		},
		{
			name: "正常系: リソース一致かつplan実行",
			opts: models.VerifyOptions{
				Environment:       "development",
				SkipTerraformPlan: false,
			},
			timeout: 5 * time.Second,
			mockAWSRunner: NewMockAWSRunner().
				WithCommandFunc(func(args ...string) (string, error) {
					// AWSリソース数を調整
					if strings.Contains(args[0], "ec2") && strings.Contains(args[0], "vpcs") {
						return "0", nil // VPC
					} else if strings.Contains(args[0], "rds") {
						return "0", nil // RDS
					} else if strings.Contains(args[0], "ecs") && strings.Contains(args[0], "clusters") {
						return "0", nil // ECSクラスター
					} else if strings.Contains(args[0], "ecs") && strings.Contains(args[0], "services") {
						return "0", nil // ECSサービス
					} else if strings.Contains(args[0], "elbv2") && strings.Contains(args[0], "load-balancers") {
						return "0", nil // ALB
					} else if strings.Contains(args[0], "elbv2") && strings.Contains(args[0], "target-groups") {
						return "0", nil // ターゲットグループ
					}
					return "0", nil
				}),
			mockFS: NewMockFileSystem().
				WithGetwdFunc(func() (string, error) {
					return "/original/path", nil
				}).
				WithChdirFunc(func(dir string) error {
					return nil
				}),
			mockCmd: NewMockContextCommandExecutor().
				WithExecuteFunc(func(command string, args ...string) (string, error) {
					if command == "terraform" && args[0] == "show" && args[1] == "-json" {
						// 空のTerraform状態（すべて0）
						return `{
                            "values": {
                                "root_module": {
                                    "resources": []
                                }
                            }
                        }`, nil
					}
					return "", fmt.Errorf("unexpected command in Execute: %s %v", command, args)
				}).
				WithExecuteWithContextFunc(func(ctx context.Context, command string, args ...string) (string, error) {
					if command == "terraform" && args[0] == "show" && args[1] == "-json" {
						// 空のTerraform状態（すべて0）
						return `{
                            "values": {
                                "root_module": {
                                    "resources": []
                                }
                            }
                        }`, nil
					} else if command == "terraform" && args[0] == "plan" {
						// 空の環境でplan実行（変更なし）
						return "No changes. Infrastructure is up-to-date.", nil
					}
					return "", fmt.Errorf("unexpected command in ExecuteWithContext: %s %v", command, args)
				}),
			expectedCode:  0, // 一致＋planで変更なしの場合は0
			expectTimeout: false,
			success:       true,
		},
		{
			name: "タイムアウト発生",
			opts: models.VerifyOptions{
				Environment:       "development",
				SkipTerraformPlan: false,
			},
			timeout: 200 * time.Millisecond,
			mockAWSRunner: NewMockAWSRunner().
				WithCommandFunc(func(args ...string) (string, error) {
					// デバッグ出力
					t.Logf("[DEBUG-AWS] AWS CLI args: %v", args)

					// VPC数を正確に設定
					if strings.Contains(strings.Join(args, " "), "ec2 describe-vpcs") {
						t.Logf("[DEBUG-AWS] Returning VPC count: 1")
						return "1", nil // VPC=1
					} else if strings.Contains(strings.Join(args, " "), "rds describe-db-instances") {
						// RDS数を設定
						t.Logf("[DEBUG-AWS] Returning RDS count: 1")
						return "1", nil // RDS=1
					} else if strings.Contains(strings.Join(args, " "), "ecs list-clusters") {
						// ECSクラスター数を設定
						t.Logf("[DEBUG-AWS] Returning ECS cluster count: 1")
						return "1", nil // ECSクラスター=1
					} else {
						// その他のリソースは0を返す
						t.Logf("[DEBUG-AWS] Returning default count: 0 for: %v", args)
						return "0", nil
					}
				}),
			mockFS: NewMockFileSystem().
				WithGetwdFunc(func() (string, error) {
					t.Logf("[DEBUG-FS] GetWD called")
					return "/original/path", nil
				}).
				WithChdirFunc(func(dir string) error {
					t.Logf("[DEBUG-FS] Chdir called with dir: %s", dir)
					return nil
				}),
			mockCmd: NewMockContextCommandExecutor().
				WithExecuteFunc(func(command string, args ...string) (string, error) {
					t.Logf("[DEBUG-CMD] Execute called: %s %v", command, args)

					// terraform show -json の呼び出しを適切に処理
					if command == "terraform" && args[0] == "show" && args[1] == "-json" {
						// 適切なJSON形式を返す
						return `{
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
                                            "address": "module.shared_ecs_cluster.aws_ecs_cluster.main",
                                            "values": { "id": "cluster-12345" }
                                        }
                                    ]
                                }
                            }
                        }`, nil
					}

					return "", fmt.Errorf("unexpected command in Execute: %s %v", command, args)
				}).
				WithExecuteWithContextFunc(func(ctx context.Context, command string, args ...string) (string, error) {
					t.Logf("[DEBUG-CMD-CTX] ExecuteWithContext called: %s %v", command, args)

					// terraform show -json の呼び出しも適切に処理
					if command == "terraform" && args[0] == "show" && args[1] == "-json" {
						// 適切なJSON形式を返す
						return `{
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
                                            "address": "module.shared_ecs_cluster.aws_ecs_cluster.main",
                                            "values": { "id": "cluster-12345" }
                                        }
                                    ]
                                }
                            }
                        }`, nil
					} else if command == "terraform" && args[0] == "plan" {
						t.Logf("[DEBUG-PLAN] Terraform plan called, sleeping for 300ms (> timeout)")
						// タイムアウト発生をシミュレート (300ms > 200ms)
						time.Sleep(300 * time.Millisecond)

						// コンテキストエラーを明示的に確認
						if ctx.Err() != nil {
							t.Logf("[DEBUG-PLAN] Context error detected: %v", ctx.Err())
							return "", ctx.Err()
						}

						t.Logf("[DEBUG-PLAN] WARNING: Plan completed without timeout")
						return "No changes. Infrastructure is up-to-date.", nil
					}

					return "", fmt.Errorf("unexpected command in ExecuteWithContext: %s %v", command, args)
				}),
			expectedCode:  1, // タイムアウトエラーの期待コード
			expectTimeout: true,
			success:       true,
		},
		{
			name: "リソース不一致の検出",
			opts: models.VerifyOptions{
				Environment:       "development",
				SkipTerraformPlan: true,
			},
			timeout: 5 * time.Second,
			mockAWSRunner: NewMockAWSRunner().
				WithCommandFunc(func(args ...string) (string, error) {
					// AWS側のリソースカウントをすべて2に設定
					return "2", nil
				}),
			mockFS: NewMockFileSystem().
				WithGetwdFunc(func() (string, error) {
					return "/original/path", nil
				}).
				WithChdirFunc(func(dir string) error {
					return nil
				}),
			mockCmd: NewMockContextCommandExecutor().
				WithExecuteFunc(func(command string, args ...string) (string, error) {
					if command == "terraform" && args[0] == "show" && args[1] == "-json" {
						// 空のTerraform状態（すべて0）
						return `{
                            "values": {
                                "root_module": {
                                    "resources": []
                                }
                            }
                        }`, nil
					}
					return "", fmt.Errorf("unexpected command in Execute: %s %v", command, args)
				}).
				WithExecuteWithContextFunc(func(ctx context.Context, command string, args ...string) (string, error) {
					if command == "terraform" && args[0] == "show" && args[1] == "-json" {
						// 空のTerraform状態（すべて0）
						return `{
                            "values": {
                                "root_module": {
                                    "resources": []
                                }
                            }
                        }`, nil
					}
					return "", fmt.Errorf("unexpected command in ExecuteWithContext: %s %v", command, args)
				}),
			expectedCode:  2, // 不一致の場合は2を返す
			expectTimeout: false,
			success:       true,
		},
		{
			name: "AWS情報取得エラー",
			opts: models.VerifyOptions{
				Environment:       "development",
				SkipTerraformPlan: true,
			},
			timeout: 5 * time.Second,
			mockAWSRunner: NewMockAWSRunner().
				WithCommandFunc(func(args ...string) (string, error) {
					return "", fmt.Errorf("AWS CLI エラー: 認証情報の取得に失敗しました")
				}),
			mockFS: NewMockFileSystem().
				WithGetwdFunc(func() (string, error) {
					return "/original/path", nil
				}),
			mockCmd: NewMockContextCommandExecutor().
				WithExecuteFunc(func(command string, args ...string) (string, error) {
					if command == "terraform" && args[0] == "show" && args[1] == "-json" {
						return `{
                            "values": {
                                "root_module": {
                                    "resources": []
                                }
                            }
                        }`, nil
					}
					return "", fmt.Errorf("unexpected command in Execute: %s %v", command, args)
				}).
				WithExecuteWithContextFunc(func(ctx context.Context, command string, args ...string) (string, error) {
					if command == "terraform" && args[0] == "show" && args[1] == "-json" {
						return `{
                            "values": {
                                "root_module": {
                                    "resources": []
                                }
                            }
                        }`, nil
					}
					return "", fmt.Errorf("unexpected command in ExecuteWithContext: %s %v", command, args)
				}),
			expectedCode:  1, // エラーの期待コード
			expectTimeout: false,
			success:       true,
		},
		{
			name: "Terraform情報取得エラー",
			opts: models.VerifyOptions{
				Environment:       "development",
				SkipTerraformPlan: true,
			},
			timeout: 5 * time.Second,
			mockAWSRunner: NewMockAWSRunner().
				WithCommandFunc(func(args ...string) (string, error) {
					return "1", nil // 正常なAWS応答
				}),
			mockFS: NewMockFileSystem().
				WithGetwdFunc(func() (string, error) {
					return "/original/path", nil
				}).
				WithChdirFunc(func(dir string) error {
					return fmt.Errorf("環境ディレクトリが見つかりません: %s", dir)
				}),
			mockCmd: NewMockContextCommandExecutor().
				WithExecuteFunc(func(command string, args ...string) (string, error) {
					return "", fmt.Errorf("unexpected command in Execute: %s %v", command, args)
				}).
				WithExecuteWithContextFunc(func(ctx context.Context, command string, args ...string) (string, error) {
					return "", fmt.Errorf("unexpected command in ExecuteWithContext: %s %v", command, args)
				}),
			expectedCode:  1, // エラーの期待コード
			expectTimeout: false,
			success:       true,
		},
		{
			name: "コンテキストキャンセル",
			opts: models.VerifyOptions{
				Environment:       "development",
				SkipTerraformPlan: false,
			},
			timeout: 5 * time.Second,
			mockAWSRunner: NewMockAWSRunner().
				WithCommandFunc(func(args ...string) (string, error) {
					// AWS側のリソースカウントをすべて1に設定
					if strings.Contains(strings.Join(args, " "), "vpc") {
						return "1", nil // VPC=1
					} else if strings.Contains(strings.Join(args, " "), "db-instance") {
						return "1", nil // RDS=1
					} else if strings.Contains(strings.Join(args, " "), "cluster") {
						return "1", nil // ECSクラスター=1
					}
					// その他のサービスも1を返す
					return "1", nil
				}),
			mockFS: NewMockFileSystem().
				WithGetwdFunc(func() (string, error) {
					return "/original/path", nil
				}).
				WithChdirFunc(func(dir string) error {
					return nil
				}),
			mockCmd: NewMockContextCommandExecutor().
				WithExecuteFunc(func(command string, args ...string) (string, error) {
					if command == "terraform" && args[0] == "show" && args[1] == "-json" {
						// Terraform側も1のリソースを返す
						return `{
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
                                            "address": "module.shared_ecs_cluster.aws_ecs_cluster.main",
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
                                        },
                                        {
                                            "type": "aws_ecs_service",
                                            "address": "module.service_grpc.aws_ecs_service.main",
                                            "values": { "id": "service-grpc-12345" }
                                        },
                                        {
                                            "type": "aws_lb",
                                            "address": "module.loadbalancer_grpc.aws_lb.main",
                                            "values": { "id": "alb-grpc-12345" }
                                        },
                                        {
                                            "type": "aws_lb_target_group",
                                            "address": "module.target_group_grpc.aws_lb_target_group.main",
                                            "values": { "id": "tg-grpc-12345" }
                                        }
                                    ]
                                }
                            }
                        }`, nil
					}
					return "", fmt.Errorf("unexpected command in Execute: %s %v", command, args)
				}).
				WithExecuteWithContextFunc(func(ctx context.Context, command string, args ...string) (string, error) {
					if command == "terraform" && args[0] == "show" && args[1] == "-json" {
						// 同じJSONをExecuteWithContextでも返す
						return `{
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
                                            "address": "module.shared_ecs_cluster.aws_ecs_cluster.main",
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
                                        },
                                        {
                                            "type": "aws_ecs_service",
                                            "address": "module.service_grpc.aws_ecs_service.main",
                                            "values": { "id": "service-grpc-12345" }
                                        },
                                        {
                                            "type": "aws_lb",
                                            "address": "module.loadbalancer_grpc.aws_lb.main",
                                            "values": { "id": "alb-grpc-12345" }
                                        },
                                        {
                                            "type": "aws_lb_target_group",
                                            "address": "module.target_group_grpc.aws_lb_target_group.main",
                                            "values": { "id": "tg-grpc-12345" }
                                        }
                                    ]
                                }
                            }
                        }`, nil
					} else if command == "terraform" && args[0] == "plan" {
						// terraform planが呼ばれたら明示的にキャンセルエラーを返す
						return "", context.Canceled
					}
					return "", fmt.Errorf("unexpected command in ExecuteWithContext: %s %v", command, args)
				}),
			expectedCode:  1, // キャンセルの場合は1
			expectTimeout: false,
			success:       true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// タイムアウト付きコンテキスト作成
			ctx, cancel := context.WithTimeout(context.Background(), tc.timeout)
			defer cancel()

			exitCode, results, err := VerifyStateWithContext(ctx, tc.opts, tc.mockAWSRunner, tc.mockFS, tc.mockCmd)

			// タイムアウト確認
			isTimeout := err != nil && ctx.Err() == context.DeadlineExceeded

			if isTimeout != tc.expectTimeout {
				t.Errorf("タイムアウト状態: 期待=%v, 実際=%v", tc.expectTimeout, isTimeout)
			}

			// 終了コードの検証
			if !isTimeout && exitCode != tc.expectedCode {
				t.Errorf("終了コード: 期待値 %d, 実際の値 %d", tc.expectedCode, exitCode)
			}

			// 成功/失敗の検証
			testSuccess := isTimeout == tc.expectTimeout
			if !isTimeout {
				testSuccess = testSuccess && exitCode == tc.expectedCode
			}

			if testSuccess != tc.success {
				t.Errorf("テスト成功判定: 期待=%v, 実際=%v", tc.success, testSuccess)
				if err != nil {
					t.Logf("エラー: %v", err)
				}
				if results != nil {
					for _, r := range results {
						t.Logf("リソース結果: %s, AWS=%d, TF=%d, Match=%v",
							r.ResourceName, r.AWSCount, r.TerraformCount, r.IsMatch)
					}
				}
			}
		})
	}
}

// TestVerifyStateWithContextTimeout 関数の修正
func TestVerifyStateWithContextTimeout(t *testing.T) {
	// 非常に短いタイムアウト設定
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// AWS Runnerの初期化
	mockAwsRunner := NewMockAWSRunner().
		WithCommandFunc(func(args ...string) (string, error) {
			// デバッグログを追加
			t.Logf("[DEBUG-AWS] AWS CLI args: %v", args)

			// VPC数を設定
			if strings.Contains(strings.Join(args, " "), "ec2 describe-vpcs") {
				t.Logf("[DEBUG-AWS] Returning VPC count: 1")
				return "1", nil // VPC=1
			} else if strings.Contains(strings.Join(args, " "), "rds describe-db-instances") {
				// RDS数を設定
				t.Logf("[DEBUG-AWS] Returning RDS count: 1")
				return "1", nil // RDS=1
			} else if strings.Contains(strings.Join(args, " "), "ecs list-clusters") {
				// ECSクラスター数を設定
				t.Logf("[DEBUG-AWS] Returning ECS cluster count: 1")
				return "1", nil // ECSクラスター=1
			} else {
				// その他のリソースは0を返す
				t.Logf("[DEBUG-AWS] Returning default count: 0 for: %v", args)
				return "0", nil
			}
		})

	// モックファイルシステム
	mockFS := NewMockFileSystem().
		WithGetwdFunc(func() (string, error) {
			t.Logf("[DEBUG-FS] GetWD called")
			return "/test/path", nil
		}).
		WithChdirFunc(func(dir string) error {
			t.Logf("[DEBUG-FS] Chdir called: %s", dir)
			return nil
		})

	// 実際のタイムアウトを発生させるモック
	mockCmd := NewMockContextCommandExecutor().
		WithExecuteFunc(func(command string, args ...string) (string, error) {
			t.Logf("[DEBUG-CMD] Execute called: %s %v", command, args)

			// terraform show -json の呼び出しを適切に処理
			if command == "terraform" && args[0] == "show" && args[1] == "-json" {
				// 正しいJSON形式を返す
				return `{
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
                                    "address": "module.shared_ecs_cluster.aws_ecs_cluster.main",
                                    "values": { "id": "cluster-12345" }
                                }
                            ]
                        }
                    }
                }`, nil
			}

			return "", fmt.Errorf("unexpected command in Execute: %s %v", command, args)
		}).
		WithExecuteWithContextFunc(func(ctx context.Context, command string, args ...string) (string, error) {
			t.Logf("[DEBUG-CMD-CTX] ExecuteWithContext called: %s %v", command, args)

			// terraform show -json の呼び出しも適切に処理
			if command == "terraform" && args[0] == "show" && args[1] == "-json" {
				return `{
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
                                    "address": "module.shared_ecs_cluster.aws_ecs_cluster.main",
                                    "values": { "id": "cluster-12345" }
                                }
                            ]
                        }
                    }
                }`, nil
			} else if command == "terraform" && args[0] == "plan" {
				t.Logf("[DEBUG-PLAN] Terraform plan called, sleeping for 100ms (> timeout)")
				// タイムアウト発生をシミュレート (100ms > 50ms)
				time.Sleep(100 * time.Millisecond)

				// コンテキストエラーを明示的に確認
				if ctx.Err() != nil {
					t.Logf("[DEBUG-PLAN] Context error detected: %v", ctx.Err())
					return "", ctx.Err()
				}

				return "No changes. Infrastructure is up-to-date.", nil
			}

			return "", fmt.Errorf("unexpected command in ExecuteWithContext: %s %v", command, args)
		})

	// オプション設定
	opts := models.VerifyOptions{
		Environment:       "development",
		SkipTerraformPlan: false,
	}

	// テスト実行
	exitCode, results, err := VerifyStateWithContext(ctx, opts, mockAwsRunner, mockFS, mockCmd)

	// デバッグログ
	t.Logf("Test result: exitCode=%d, err=%v", exitCode, err)
	if results != nil {
		t.Logf("Results: %+v", results)
	}

	// 検証
	if err == nil || !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Expected timeout error, got: %v", err)
	}

	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got: %d", exitCode)
	}
}

// canceledContext はキャンセル済みのコンテキストを返す
func canceledContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

type testLogger struct {
	t *testing.T
}

func (l *testLogger) Debug(format string, args ...interface{}) {
	l.t.Logf("[DEBUG] "+format, args...)
}

func (l *testLogger) Info(format string, args ...interface{}) {
	l.t.Logf("[INFO] "+format, args...)
}

func (l *testLogger) Warn(format string, args ...interface{}) {
	l.t.Logf("[WARN] "+format, args...)
}

func (l *testLogger) Error(format string, args ...interface{}) {
	l.t.Logf("[ERROR] "+format, args...)
}

// テストケースの構造を変更し、専用のタイムアウトテストを追加
func TestRunTerraformPlanWithContextTimeout(t *testing.T) {
	// 非常に短いタイムアウトを設定（実際のタイムアウトを発生させる）
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// モックファイルシステム
	mockFS := NewMockFileSystem().
		WithGetwdFunc(func() (string, error) {
			t.Logf("GetWD called")
			return "/test/path", nil
		}).
		WithChdirFunc(func(dir string) error {
			t.Logf("Chdir called: %s", dir)
			return nil
		})

	// 実際のタイムアウトを発生させるモック
	mockCmd := NewMockContextCommandExecutor().
		WithExecuteWithContextFunc(func(ctx context.Context, command string, args ...string) (string, error) {
			t.Logf("ExecuteWithContext called for command: %s %v", command, args)

			// タイムアウト発生をシミュレート
			if command == "terraform" && args[0] == "plan" {
				// 十分に長く待機してタイムアウトを発生させる
				time.Sleep(300 * time.Millisecond)

				// 実際のコンテキストエラーを確認
				if ctx.Err() == context.DeadlineExceeded {
					t.Logf("Context deadline exceeded detected")
					return "", ctx.Err()
				}

				t.Logf("Unexpected: Command completed without timeout")
				return "No timeout occurred", nil
			}

			return "", fmt.Errorf("unexpected command: %s", command)
		})

	// テスト実行
	exitCode, output, err := RunTerraformPlanWithContext(ctx, "development", mockFS, mockCmd)

	// コンテキストの状態確認
	t.Logf("After execution: ctx.Err()=%v, err=%v, exitCode=%d, output=%s", ctx.Err(), err, exitCode, output)

	// 期待結果の確認
	if err == nil || ctx.Err() != context.DeadlineExceeded {
		t.Errorf("expected timeout error, got: err=%v, ctx.Err()=%v", err, ctx.Err())
	}

	if exitCode != 1 {
		t.Errorf("expected exit code 1, got: %d", exitCode)
	}
}

// モックの動作を確認するための単体テスト
func TestMockExecutorTimeoutBehavior(t *testing.T) {
	// 非常に短いタイムアウト
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// ExecuteWithContextの実装をテスト
	mockCmd := NewMockContextCommandExecutor().
		WithExecuteWithContextFunc(func(ctx context.Context, command string, args ...string) (string, error) {
			// タイムアウトを発生させる遅延
			time.Sleep(100 * time.Millisecond)

			// コンテキストの状態確認
			if ctx.Err() != nil {
				t.Logf("Context error detected: %v", ctx.Err())
				return "", ctx.Err()
			}

			return "completed", nil
		})

	// 実行
	output, err := mockCmd.ExecuteWithContext(ctx, "test", "args")

	// 検証
	t.Logf("Result: output=%s, err=%v, ctx.Err()=%v", output, err, ctx.Err())

	if ctx.Err() != context.DeadlineExceeded {
		t.Errorf("Expected context.DeadlineExceeded, got: %v", ctx.Err())
	}

	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

// TestContextCancellation はコンテキストキャンセルのテストです
// TestContextCancellation はコンテキストキャンセルのテストです
func TestContextCancellation(t *testing.T) {
	// キャンセル可能なコンテキストを作成
	ctx, cancel := context.WithCancel(context.Background())

	// AWS Runnerの初期化
	mockAwsRunner := NewMockAWSRunner().
		WithCommandFunc(func(args ...string) (string, error) {
			// 一律1を返すのではなく、パターンマッチングを行う
			if strings.Contains(strings.Join(args, " "), "describe-vpcs") {
				return "1", nil // VPC=1
			} else if strings.Contains(strings.Join(args, " "), "describe-db-instances") {
				return "1", nil // RDS=1
			} else if strings.Contains(strings.Join(args, " "), "list-clusters") {
				return "1", nil // ECSクラスター=1
			}
			return "0", nil // その他=0
		})

	// モックファイルシステム
	mockFS := NewMockFileSystem().
		WithGetwdFunc(func() (string, error) {
			return "/test/path", nil
		}).
		WithChdirFunc(func(dir string) error {
			return nil
		})

	// コンテキストキャンセルを検出するモック
	mockCmd := NewMockContextCommandExecutor().
		WithExecuteFunc(func(command string, args ...string) (string, error) {
			if command == "terraform" && args[0] == "show" && args[1] == "-json" {
				// 適切なJSON形式を返す
				return `{
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
                                    "address": "module.shared_ecs_cluster.aws_ecs_cluster.main",
                                    "values": { "id": "cluster-12345" }
                                }
                            ]
                        }
                    }
                }`, nil
			}
			return "", errors.New("unexpected command")
		}).
		WithExecuteWithContextFunc(func(ctx context.Context, command string, args ...string) (string, error) {
			if command == "terraform" && args[0] == "show" && args[1] == "-json" {
				// 同じJSON形式をExecuteWithContextでも返す
				return `{
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
                                    "address": "module.shared_ecs_cluster.aws_ecs_cluster.main",
                                    "values": { "id": "cluster-12345" }
                                }
                            ]
                        }
                    }
                }`, nil
			} else if command == "terraform" && args[0] == "plan" {
				// terraform planが呼ばれたら明示的にキャンセル
				cancel()
				// キャンセルがすぐに検出されるよう少し待機
				time.Sleep(10 * time.Millisecond)

				// キャンセルされたことを確認
				if ctx.Err() == context.Canceled {
					t.Logf("Context canceled detected in mock")
					return "", ctx.Err()
				}

				return "plan executed", nil
			}
			return "", errors.New("unexpected command in ExecuteWithContext")
		})

	// オプション設定
	opts := models.VerifyOptions{
		Environment:       "development",
		SkipTerraformPlan: false,
	}

	// テスト実行
	exitCode, _, err := VerifyStateWithContext(ctx, opts, mockAwsRunner, mockFS, mockCmd)

	// 検証
	if err == nil {
		t.Errorf("Expected canceled error, got nil")
	} else if !errors.Is(err, context.Canceled) {
		t.Errorf("Expected error to be context.Canceled, got: %v", err)
	}

	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got: %d", exitCode)
	}
}
func TestVerifyStateWithContextWithIgnoreOption(t *testing.T) {
	testCases := []struct {
		name               string
		opts               models.VerifyOptions
		mockAWSError       error
		mockTerraformError error
		expectedExitCode   int
		expectError        bool
	}{
		{
			name: "リソースエラー（無視オプションあり）",
			opts: models.VerifyOptions{
				Environment:          "development",
				IgnoreResourceErrors: true,
			},
			mockAWSError:       errors.New("ClusterNotFoundException"),
			mockTerraformError: nil,
			expectedExitCode:   0, // 無視するので成功
			expectError:        false,
		},
		{
			name: "リソースエラー（無視オプションなし）",
			opts: models.VerifyOptions{
				Environment:          "development",
				IgnoreResourceErrors: false,
			},
			mockAWSError:       errors.New("ClusterNotFoundException"),
			mockTerraformError: nil,
			expectedExitCode:   1, // エラーコード
			expectError:        true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// モックAWSRunner
			mockAWSRunner := NewMockAWSRunner().
				WithCommandFunc(func(args ...string) (string, error) {
					return "", tc.mockAWSError
				})

			// モックファイルシステム
			mockFS := NewMockFileSystem().
				WithGetwdFunc(func() (string, error) {
					return "/test/path", nil
				}).
				WithChdirFunc(func(dir string) error {
					return tc.mockTerraformError
				})

			// モックコマンド実行
			mockCmd := NewMockContextCommandExecutor()

			// テスト実行
			exitCode, _, err := VerifyStateWithContext(context.Background(), tc.opts, mockAWSRunner, mockFS, mockCmd)

			// 結果検証
			if tc.expectError && err == nil {
				t.Errorf("エラーが期待されていましたが、エラーはありませんでした")
			} else if !tc.expectError && err != nil {
				t.Errorf("エラーは期待されていませんでした。得られたエラー: %v", err)
			}

			if exitCode != tc.expectedExitCode {
				t.Errorf("終了コード: 期待値 %d, 実際の値 %d", tc.expectedExitCode, exitCode)
			}
		})
	}
}
