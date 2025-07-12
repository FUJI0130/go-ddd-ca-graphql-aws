package models

import (
	"testing"
)

func TestResourcesInitialization(t *testing.T) {
	// 基本的な初期化
	resources := &Resources{
		VPC:        1,
		RDS:        2,
		ECSCluster: 3,
		Services:   map[string]ServiceResources{},
	}

	// フィールド値の検証
	if resources.VPC != 1 {
		t.Errorf("Expected VPC count to be 1, got %d", resources.VPC)
	}

	if resources.RDS != 2 {
		t.Errorf("Expected RDS count to be 2, got %d", resources.RDS)
	}

	if resources.ECSCluster != 3 {
		t.Errorf("Expected ECSCluster count to be 3, got %d", resources.ECSCluster)
	}

	if len(resources.Services) != 0 {
		t.Errorf("Expected Services map to be empty, got %d items", len(resources.Services))
	}
}

func TestResources(t *testing.T) {
	tests := []struct {
		name            string
		vpc             int
		rds             int
		ecsCluster      int
		services        map[string]ServiceResources
		expectedVPC     int
		expectedRDS     int
		expectedCluster int
		expectedSvcLen  int
		shouldMatch     bool // 期待値と実際の値が一致すべきか
	}{
		{
			name:            "正常系: 基本初期化",
			vpc:             1,
			rds:             2,
			ecsCluster:      3,
			services:        map[string]ServiceResources{},
			expectedVPC:     1,
			expectedRDS:     2,
			expectedCluster: 3,
			expectedSvcLen:  0,
			shouldMatch:     true,
		},
		{
			name:            "正常系: サービスマップあり",
			vpc:             1,
			rds:             1,
			ecsCluster:      1,
			services:        map[string]ServiceResources{"api": {ECSService: 1, ALB: 1, TargetGroup: 1}},
			expectedVPC:     1,
			expectedRDS:     1,
			expectedCluster: 1,
			expectedSvcLen:  1,
			shouldMatch:     true,
		},
		{
			name:            "異常系: 期待値と不一致",
			vpc:             2,
			rds:             0,
			ecsCluster:      1,
			services:        map[string]ServiceResources{},
			expectedVPC:     1, // 実際の値と異なる期待値を設定
			expectedRDS:     0,
			expectedCluster: 1,
			expectedSvcLen:  0,
			shouldMatch:     false, // 不一致を期待
		},
		{
			name:            "エッジケース: nilサービスマップ",
			vpc:             1,
			rds:             1,
			ecsCluster:      1,
			services:        nil,
			expectedVPC:     1,
			expectedRDS:     1,
			expectedCluster: 1,
			expectedSvcLen:  0,
			shouldMatch:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resources := &Resources{
				VPC:        tt.vpc,
				RDS:        tt.rds,
				ECSCluster: tt.ecsCluster,
				Services:   tt.services,
			}

			vpcMatch := resources.VPC == tt.expectedVPC
			rdsMatch := resources.RDS == tt.expectedRDS
			clusterMatch := resources.ECSCluster == tt.expectedCluster

			var servicesLenMatch bool
			if tt.services == nil {
				servicesLenMatch = len(resources.Services) == 0
			} else {
				servicesLenMatch = len(resources.Services) == tt.expectedSvcLen
			}

			// すべての条件がマッチするかどうか
			allMatch := vpcMatch && rdsMatch && clusterMatch && servicesLenMatch

			// shouldMatchフラグに基づいて検証
			if tt.shouldMatch && !allMatch {
				// 一致すべきなのに一致しない場合、エラーの詳細を出力
				if !vpcMatch {
					t.Errorf("VPC: expected %d, got %d", tt.expectedVPC, resources.VPC)
				}
				if !rdsMatch {
					t.Errorf("RDS: expected %d, got %d", tt.expectedRDS, resources.RDS)
				}
				if !clusterMatch {
					t.Errorf("ECSCluster: expected %d, got %d", tt.expectedCluster, resources.ECSCluster)
				}
				if !servicesLenMatch {
					var servicesLen int
					if resources.Services == nil {
						servicesLen = 0
					} else {
						servicesLen = len(resources.Services)
					}
					t.Errorf("Services length: expected %d, got %d", tt.expectedSvcLen, servicesLen)
				}
			} else if !tt.shouldMatch && allMatch {
				// 一致すべきでないのに一致する場合
				t.Errorf("Expected values to NOT match actual values, but they all matched")
			}
		})
	}
}

func TestServiceResources(t *testing.T) {
	tests := []struct {
		name               string
		ecsService         int
		alb                int
		targetGroup        int
		expectedECSService int
		expectedALB        int
		expectedTG         int
		shouldMatch        bool // 期待値と実際の値が一致すべきか
	}{
		{
			name:               "正常系: 基本初期化",
			ecsService:         1,
			alb:                1,
			targetGroup:        1,
			expectedECSService: 1,
			expectedALB:        1,
			expectedTG:         1,
			shouldMatch:        true,
		},
		{
			name:               "正常系: ゼロ値",
			ecsService:         0,
			alb:                0,
			targetGroup:        0,
			expectedECSService: 0,
			expectedALB:        0,
			expectedTG:         0,
			shouldMatch:        true,
		},
		{
			name:               "正常系: 複数リソース",
			ecsService:         2,
			alb:                3,
			targetGroup:        4,
			expectedECSService: 2,
			expectedALB:        3,
			expectedTG:         4,
			shouldMatch:        true,
		},
		{
			name:               "異常系: 期待値と不一致",
			ecsService:         1,
			alb:                2,
			targetGroup:        3,
			expectedECSService: 2, // 実際の値と異なる期待値
			expectedALB:        2,
			expectedTG:         3,
			shouldMatch:        false, // 不一致を期待
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svcResources := ServiceResources{
				ECSService:  tt.ecsService,
				ALB:         tt.alb,
				TargetGroup: tt.targetGroup,
			}

			ecsMatch := svcResources.ECSService == tt.expectedECSService
			albMatch := svcResources.ALB == tt.expectedALB
			tgMatch := svcResources.TargetGroup == tt.expectedTG

			// すべての条件がマッチするかどうか
			allMatch := ecsMatch && albMatch && tgMatch

			// shouldMatchフラグに基づいて検証
			if tt.shouldMatch && !allMatch {
				// 一致すべきなのに一致しない場合、エラーの詳細を出力
				if !ecsMatch {
					t.Errorf("ECSService: expected %d, got %d", tt.expectedECSService, svcResources.ECSService)
				}
				if !albMatch {
					t.Errorf("ALB: expected %d, got %d", tt.expectedALB, svcResources.ALB)
				}
				if !tgMatch {
					t.Errorf("TargetGroup: expected %d, got %d", tt.expectedTG, svcResources.TargetGroup)
				}
			} else if !tt.shouldMatch && allMatch {
				// 一致すべきでないのに一致する場合
				t.Errorf("Expected values to NOT match actual values, but they all matched")
			}
		})
	}
}

func TestComparisonResult(t *testing.T) {
	tests := []struct {
		name             string
		resourceName     string
		awsCount         int
		terraformCount   int
		isMatch          bool
		expectedName     string
		expectedAWS      int
		expectedTF       int
		expectedIsMatch  bool
		shouldMatchField bool   // 特定のフィールドが期待値と一致すべきか
		fieldToTest      string // テストする特定のフィールド (全部の場合は "all")
	}{
		{
			name:             "正常系: リソース一致",
			resourceName:     "VPC",
			awsCount:         1,
			terraformCount:   1,
			isMatch:          true,
			expectedName:     "VPC",
			expectedAWS:      1,
			expectedTF:       1,
			expectedIsMatch:  true,
			shouldMatchField: true,
			fieldToTest:      "all",
		},
		{
			name:             "正常系: リソース不一致",
			resourceName:     "RDS",
			awsCount:         0,
			terraformCount:   1,
			isMatch:          false,
			expectedName:     "RDS",
			expectedAWS:      0,
			expectedTF:       1,
			expectedIsMatch:  false,
			shouldMatchField: true,
			fieldToTest:      "all",
		},
		{
			name:             "異常系: IsMatchフィールドが不適切",
			resourceName:     "ECSクラスター",
			awsCount:         1,
			terraformCount:   2,
			isMatch:          true, // AWSとTerraformの値が異なるのにtrueは不適切
			expectedName:     "ECSクラスター",
			expectedAWS:      1,
			expectedTF:       2,
			expectedIsMatch:  false, // 本来はfalseであるべき
			shouldMatchField: false, // 一致しないことを期待
			fieldToTest:      "isMatch",
		},
		{
			name:             "異常系: ResourceName不一致",
			resourceName:     "ALB",
			awsCount:         1,
			terraformCount:   1,
			isMatch:          true,
			expectedName:     "LoadBalancer", // 実際と異なる値
			expectedAWS:      1,
			expectedTF:       1,
			expectedIsMatch:  true,
			shouldMatchField: false,
			fieldToTest:      "resourceName",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ComparisonResult{
				ResourceName:   tt.resourceName,
				AWSCount:       tt.awsCount,
				TerraformCount: tt.terraformCount,
				IsMatch:        tt.isMatch,
			}

			nameMatch := result.ResourceName == tt.expectedName
			awsMatch := result.AWSCount == tt.expectedAWS
			tfMatch := result.TerraformCount == tt.expectedTF
			isMatchMatch := result.IsMatch == tt.expectedIsMatch

			var fieldMatch bool
			switch tt.fieldToTest {
			case "resourceName":
				fieldMatch = nameMatch
			case "awsCount":
				fieldMatch = awsMatch
			case "terraformCount":
				fieldMatch = tfMatch
			case "isMatch":
				fieldMatch = isMatchMatch
			case "all":
				fieldMatch = nameMatch && awsMatch && tfMatch && isMatchMatch
			default:
				t.Fatalf("Invalid fieldToTest: %s", tt.fieldToTest)
			}

			if tt.shouldMatchField && !fieldMatch {
				// 一致すべきなのに一致しない場合
				switch tt.fieldToTest {
				case "resourceName":
					t.Errorf("ResourceName: expected %s, got %s", tt.expectedName, result.ResourceName)
				case "awsCount":
					t.Errorf("AWSCount: expected %d, got %d", tt.expectedAWS, result.AWSCount)
				case "terraformCount":
					t.Errorf("TerraformCount: expected %d, got %d", tt.expectedTF, result.TerraformCount)
				case "isMatch":
					t.Errorf("IsMatch: expected %v, got %v", tt.expectedIsMatch, result.IsMatch)
				case "all":
					if !nameMatch {
						t.Errorf("ResourceName: expected %s, got %s", tt.expectedName, result.ResourceName)
					}
					if !awsMatch {
						t.Errorf("AWSCount: expected %d, got %d", tt.expectedAWS, result.AWSCount)
					}
					if !tfMatch {
						t.Errorf("TerraformCount: expected %d, got %d", tt.expectedTF, result.TerraformCount)
					}
					if !isMatchMatch {
						t.Errorf("IsMatch: expected %v, got %v", tt.expectedIsMatch, result.IsMatch)
					}
				}
			} else if !tt.shouldMatchField && fieldMatch {
				// 一致すべきでないのに一致する場合
				t.Errorf("Expected field %s to NOT match, but it did match", tt.fieldToTest)
			}
		})
	}
}

func TestVerifyOptions(t *testing.T) {
	tests := []struct {
		name                 string
		environment          string
		debug                bool
		skipTerraformPlan    bool
		forceCleanup         bool
		expectedEnvironment  string
		expectedDebug        bool
		expectedSkipPlan     bool
		expectedForceCleanup bool
		shouldMatch          bool
	}{
		{
			name:                 "正常系: 基本オプション",
			environment:          "development",
			debug:                true,
			skipTerraformPlan:    false,
			forceCleanup:         false,
			expectedEnvironment:  "development",
			expectedDebug:        true,
			expectedSkipPlan:     false,
			expectedForceCleanup: false,
			shouldMatch:          true,
		},
		{
			name:                 "正常系: 本番環境オプション",
			environment:          "production",
			debug:                false,
			skipTerraformPlan:    true,
			forceCleanup:         true,
			expectedEnvironment:  "production",
			expectedDebug:        false,
			expectedSkipPlan:     true,
			expectedForceCleanup: true,
			shouldMatch:          true,
		},
		{
			name:                 "異常系: 環境名不一致",
			environment:          "staging",
			debug:                false,
			skipTerraformPlan:    false,
			forceCleanup:         false,
			expectedEnvironment:  "test", // 実際と異なる値
			expectedDebug:        false,
			expectedSkipPlan:     false,
			expectedForceCleanup: false,
			shouldMatch:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := VerifyOptions{
				Environment:       tt.environment,
				Debug:             tt.debug,
				SkipTerraformPlan: tt.skipTerraformPlan,
				ForceCleanup:      tt.forceCleanup,
			}

			envMatch := opts.Environment == tt.expectedEnvironment
			debugMatch := opts.Debug == tt.expectedDebug
			skipPlanMatch := opts.SkipTerraformPlan == tt.expectedSkipPlan
			forceCleanupMatch := opts.ForceCleanup == tt.expectedForceCleanup

			allMatch := envMatch && debugMatch && skipPlanMatch && forceCleanupMatch

			if tt.shouldMatch && !allMatch {
				// 一致すべきなのに一致しない場合
				if !envMatch {
					t.Errorf("Environment: expected %s, got %s", tt.expectedEnvironment, opts.Environment)
				}
				if !debugMatch {
					t.Errorf("Debug: expected %v, got %v", tt.expectedDebug, opts.Debug)
				}
				if !skipPlanMatch {
					t.Errorf("SkipTerraformPlan: expected %v, got %v", tt.expectedSkipPlan, opts.SkipTerraformPlan)
				}
				if !forceCleanupMatch {
					t.Errorf("ForceCleanup: expected %v, got %v", tt.expectedForceCleanup, opts.ForceCleanup)
				}
			} else if !tt.shouldMatch && allMatch {
				// 一致すべきでないのに一致する場合
				t.Errorf("Expected values to NOT match actual values, but they all matched")
			}
		})
	}
}
