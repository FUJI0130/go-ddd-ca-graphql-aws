package terraform

import (
	"strings"
	"testing"

	"github.com/FUJI0130/go-ddd-ca/internal/tools/terraform/models"
)

// TestFormatComparisonTable はFormatComparisonTable関数のテスト
// TestFormatComparisonTable改善版
func TestFormatComparisonTable(t *testing.T) {
	testCases := []struct {
		name          string
		results       []models.ComparisonResult
		expected      string
		shouldSucceed bool
	}{
		{
			name: "正常系：複数リソースの比較結果がある",
			results: []models.ComparisonResult{
				{
					ResourceName:   "VPC",
					AWSCount:       1,
					TerraformCount: 1,
					IsMatch:        true,
				},
				{
					ResourceName:   "RDS",
					AWSCount:       1,
					TerraformCount: 0,
					IsMatch:        false,
				},
			},
			expected:      "リソース\t\tAWS\tTerraform\t状態\n--------------------------------\nVPC\t\t1\t1\t一致\nRDS\t\t1\t0\t不一致\n",
			shouldSucceed: true,
		},
		{
			name:          "正常系：空のリスト",
			results:       []models.ComparisonResult{},
			expected:      "リソース\t\tAWS\tTerraform\t状態\n--------------------------------\n",
			shouldSucceed: true,
		},
		{
			name: "正常系：長いリソース名",
			results: []models.ComparisonResult{
				{
					ResourceName:   "非常に長いリソース名AAAAAAAAAAAAAAAAAAAA",
					AWSCount:       1,
					TerraformCount: 1,
					IsMatch:        true,
				},
			},
			expected:      "リソース\t\tAWS\tTerraform\t状態\n--------------------------------\n非常に長いリソース名AAAAAAAAAAAAAAAAAAAA\t\t1\t1\t一致\n",
			shouldSucceed: true,
		},
		{
			name:          "異常系：nilのリソースリスト",
			results:       nil,
			expected:      "リソース\t\tAWS\tTerraform\t状態\n--------------------------------\n",
			shouldSucceed: true, // 現実装ではnilでもエラーにならず空の結果を返すと想定
		},
		{
			name: "異常系：リソース名がnil",
			results: []models.ComparisonResult{
				{
					ResourceName:   "", // 空のリソース名
					AWSCount:       1,
					TerraformCount: 0,
					IsMatch:        false,
				},
			},
			expected:      "リソース\t\tAWS\tTerraform\t状態\n--------------------------------\n\t\t1\t0\t不一致\n",
			shouldSucceed: true,
		},
		{
			name: "異常系：極端な数値",
			results: []models.ComparisonResult{
				{
					ResourceName:   "極端なリソース",
					AWSCount:       9999999,
					TerraformCount: -1,
					IsMatch:        false,
				},
			},
			expected:      "リソース\t\tAWS\tTerraform\t状態\n--------------------------------\n極端なリソース\t\t9999999\t-1\t不一致\n",
			shouldSucceed: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// モックの準備
			mockWriter := NewMockOutputWriter()
			formatter := NewDefaultOutputFormatter(mockWriter)

			// テスト対象関数の実行
			result := formatter.FormatComparisonTable(tc.results)

			// 期待するフォーマット結果の検証
			if result != tc.expected {
				t.Errorf("期待されたフォーマット結果:\n%s\n\n実際の結果:\n%s", tc.expected, result)
			}
		})
	}
}

// TestDisplayResults はDisplayResults関数のテスト
// TestDisplayResults改善版
func TestDisplayResults(t *testing.T) {
	testCases := []struct {
		name           string
		results        []models.ComparisonResult
		env            string
		expectedOutput []string
		shouldSucceed  bool
	}{
		{
			name: "正常系：複数リソースの比較結果がある",
			results: []models.ComparisonResult{
				{
					ResourceName:   "VPC",
					AWSCount:       1,
					TerraformCount: 1,
					IsMatch:        true,
				},
				{
					ResourceName:   "RDS",
					AWSCount:       1,
					TerraformCount: 0,
					IsMatch:        false,
				},
			},
			env: "development",
			expectedOutput: []string{
				"\n■ 検証結果: development\n",
				"--------------------------------\n",
				"リソース\t\tAWS\tTerraform\t状態\n",
				"--------------------------------\n",
				"VPC\t\t1\t1\t一致\n",
				"RDS\t\t1\t0\t不一致\n",
			},
			shouldSucceed: true,
		},
		{
			name:    "正常系：空のリスト",
			results: []models.ComparisonResult{},
			env:     "production",
			expectedOutput: []string{
				"\n■ 検証結果: production\n",
				"--------------------------------\n",
				"リソース\t\tAWS\tTerraform\t状態\n",
				"--------------------------------\n",
			},
			shouldSucceed: true,
		},
		{
			name:    "異常系：nilのリソースリスト",
			results: nil,
			env:     "development",
			expectedOutput: []string{
				"\n■ 検証結果: development\n",
				"--------------------------------\n",
				"リソース\t\tAWS\tTerraform\t状態\n",
				"--------------------------------\n",
			},
			shouldSucceed: true, // nilも空のリストとして処理されると想定
		},
		{
			name: "異常系：空の環境名",
			results: []models.ComparisonResult{
				{
					ResourceName:   "VPC",
					AWSCount:       1,
					TerraformCount: 1,
					IsMatch:        true,
				},
			},
			env: "",
			expectedOutput: []string{
				"\n■ 検証結果: \n",
				"--------------------------------\n",
				"リソース\t\tAWS\tTerraform\t状態\n",
				"--------------------------------\n",
				"VPC\t\t1\t1\t一致\n",
			},
			shouldSucceed: true,
		},
		{
			name: "異常系：Printf呼び出しエラー",
			results: []models.ComparisonResult{
				{
					ResourceName:   "VPC",
					AWSCount:       1,
					TerraformCount: 1,
					IsMatch:        true,
				},
			},
			env:            "development",
			expectedOutput: nil,   // 出力の検証は行わない
			shouldSucceed:  false, // 失敗を期待
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// モックの準備
			var mockWriter *MockOutputWriter

			if tc.name == "異常系：Printf呼び出しエラー" {
				// エラーをシミュレートするモック
				mockWriter = NewMockOutputWriter().
					WithPrintfFunc(func(format string, args ...interface{}) {
						if strings.Contains(format, "検証結果") {
							panic("Printf error simulated")
						}
					})
			} else {
				mockWriter = NewMockOutputWriter()
			}

			formatter := NewDefaultOutputFormatter(mockWriter)

			// 異常系のテストで予期されるパニックを処理
			if !tc.shouldSucceed {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("期待されたパニックが発生しませんでした")
					}
				}()
			}

			// テスト対象関数の実行
			formatter.DisplayResults(tc.results, tc.env)

			// 成功するケースのみ出力を検証
			if tc.shouldSucceed && tc.expectedOutput != nil {
				outputs := mockWriter.GetOutputs()
				for i, expectedLine := range tc.expectedOutput {
					if i >= len(outputs) {
						t.Errorf("期待された出力行数: %d, 実際の出力行数: %d", len(tc.expectedOutput), len(outputs))
						break
					}
					if outputs[i] != expectedLine {
						t.Errorf("期待された出力 [%d]: %q, 実際の出力: %q", i, expectedLine, outputs[i])
					}
				}
			}
		})
	}
}

// TestShowMismatchRemediation はShowMismatchRemediation関数のテスト
// TestShowMismatchRemediation改善版
func TestShowMismatchRemediation(t *testing.T) {
	testCases := []struct {
		name           string
		env            string
		expectedOutput []string
		printfError    bool
		shouldSucceed  bool
	}{
		{
			name: "正常系：developmentに対する修復オプション",
			env:  "development",
			expectedOutput: []string{
				"\n■ 修復オプション:\n",
				"1. terraform importで不足リソースをインポート: make terraform-import TF_ENV=development\n",
				"2. terraform state rmで余分なリソースを削除: terraform state rm <リソースパス>\n",
				"3. タグベース削除を使用: make tag-cleanup TF_ENV=development\n",
			},
			printfError:   false,
			shouldSucceed: true,
		},
		{
			name: "正常系：productionに対する修復オプション",
			env:  "production",
			expectedOutput: []string{
				"\n■ 修復オプション:\n",
				"1. terraform importで不足リソースをインポート: make terraform-import TF_ENV=production\n",
				"2. terraform state rmで余分なリソースを削除: terraform state rm <リソースパス>\n",
				"3. タグベース削除を使用: make tag-cleanup TF_ENV=production\n",
			},
			printfError:   false,
			shouldSucceed: true,
		},
		{
			name: "異常系：空の環境名",
			env:  "",
			expectedOutput: []string{
				"\n■ 修復オプション:\n",
				"1. terraform importで不足リソースをインポート: make terraform-import TF_ENV=\n",
				"2. terraform state rmで余分なリソースを削除: terraform state rm <リソースパス>\n",
				"3. タグベース削除を使用: make tag-cleanup TF_ENV=\n",
			},
			printfError:   false,
			shouldSucceed: true,
		},
		{
			name:           "異常系：Printf呼び出しエラー",
			env:            "development",
			expectedOutput: nil, // 出力の検証は行わない
			printfError:    true,
			shouldSucceed:  false, // 失敗を期待
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// モックの準備
			var mockWriter *MockOutputWriter

			if tc.printfError {
				// エラーをシミュレートするモック
				mockWriter = NewMockOutputWriter().
					WithPrintfFunc(func(format string, args ...interface{}) {
						if strings.Contains(format, "修復オプション") {
							panic("Printf error simulated")
						}
					})
			} else {
				mockWriter = NewMockOutputWriter()
			}

			formatter := NewDefaultOutputFormatter(mockWriter)

			// 異常系のテストで予期されるパニックを処理
			if !tc.shouldSucceed {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("期待されたパニックが発生しませんでした")
					}
				}()
			}

			// テスト対象関数の実行
			formatter.ShowMismatchRemediation(tc.env)

			// 成功するケースのみ出力を検証
			if tc.shouldSucceed && tc.expectedOutput != nil {
				outputs := mockWriter.GetOutputs()
				for i, expectedLine := range tc.expectedOutput {
					if i >= len(outputs) {
						t.Errorf("期待された出力行数: %d, 実際の出力行数: %d", len(tc.expectedOutput), len(outputs))
						break
					}
					if outputs[i] != expectedLine {
						t.Errorf("期待された出力 [%d]: %q, 実際の出力: %q", i, expectedLine, outputs[i])
					}
				}
			}
		})
	}
}

// TestPrintFunctions はPrint系関数のテスト
// TestPrintFunctions改善版
func TestPrintFunctions(t *testing.T) {
	testCases := []struct {
		name           string
		function       string
		message        string
		args           []interface{}
		expectedOutput string
		printfError    bool
		shouldSucceed  bool
	}{
		{
			name:           "正常系：PrintDebugInfo",
			function:       "PrintDebugInfo",
			message:        "デバッグ情報: %s",
			args:           []interface{}{"テスト"},
			expectedOutput: "DEBUG: デバッグ情報: テスト\n",
			printfError:    false,
			shouldSucceed:  true,
		},
		{
			name:           "正常系：PrintError",
			function:       "PrintError",
			message:        "エラーが発生しました: %d",
			args:           []interface{}{404},
			expectedOutput: "✖ エラーが発生しました: 404\n",
			printfError:    false,
			shouldSucceed:  true,
		},
		{
			name:           "正常系：PrintWarning",
			function:       "PrintWarning",
			message:        "警告: %s",
			args:           []interface{}{"リソースが存在しません"},
			expectedOutput: "⚠️ 警告: リソースが存在しません\n",
			printfError:    false,
			shouldSucceed:  true,
		},
		{
			name:           "正常系：PrintSuccess",
			function:       "PrintSuccess",
			message:        "成功: %s",
			args:           []interface{}{"処理が完了しました"},
			expectedOutput: "✅ 成功: 処理が完了しました\n",
			printfError:    false,
			shouldSucceed:  true,
		},
		{
			name:           "正常系：PrintInfo",
			function:       "PrintInfo",
			message:        "情報: %s",
			args:           []interface{}{"システム状態"},
			expectedOutput: "■ 情報: システム状態\n",
			printfError:    false,
			shouldSucceed:  true,
		},
		{
			name:           "異常系：空のメッセージ",
			function:       "PrintInfo",
			message:        "",
			args:           []interface{}{},
			expectedOutput: "■ \n",
			printfError:    false,
			shouldSucceed:  true,
		},
		{
			name:           "異常系：無効なフォーマット",
			function:       "PrintDebugInfo",
			message:        "無効なフォーマット: %d",
			args:           []interface{}{"文字列"},                  // 数値フォーマットに文字列
			expectedOutput: "DEBUG: 無効なフォーマット: %!d(string=文字列)\n", // フォーマットエラーのメッセージ
			printfError:    false,
			shouldSucceed:  true, // fmt.Sprintfはpanicを発生させず、エラーインジケータを含む文字列を返す
		},
		{
			name:           "異常系：Printf呼び出しエラー",
			function:       "PrintError",
			message:        "エラーメッセージ",
			args:           []interface{}{},
			expectedOutput: "",
			printfError:    true,
			shouldSucceed:  false, // 失敗を期待
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// モックの準備
			var mockWriter *MockOutputWriter

			if tc.printfError {
				// エラーをシミュレートするモック
				mockWriter = NewMockOutputWriter().
					WithPrintfFunc(func(format string, args ...interface{}) {
						panic("Printf error simulated")
					})
			} else {
				mockWriter = NewMockOutputWriter()
			}

			formatter := NewDefaultOutputFormatter(mockWriter)

			// 異常系のテストで予期されるパニックを処理
			if !tc.shouldSucceed {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("期待されたパニックが発生しませんでした")
					}
				}()
			}

			// テスト対象関数の実行
			switch tc.function {
			case "PrintDebugInfo":
				formatter.PrintDebugInfo(tc.message, tc.args...)
			case "PrintError":
				formatter.PrintError(tc.message, tc.args...)
			case "PrintWarning":
				formatter.PrintWarning(tc.message, tc.args...)
			case "PrintSuccess":
				formatter.PrintSuccess(tc.message, tc.args...)
			case "PrintInfo":
				formatter.PrintInfo(tc.message, tc.args...)
			}

			// 成功するケースのみ出力を検証
			if tc.shouldSucceed {
				output := mockWriter.GetOutput()
				if output != tc.expectedOutput {
					t.Errorf("期待された出力: %q, 実際の出力: %q", tc.expectedOutput, output)
				}
			}
		})
	}
}

// TestNewDefaultOutputFormatter はNewDefaultOutputFormatterのテスト
// TestNewDefaultOutputFormatter改善版
// TestNewDefaultOutputFormatter修正版
func TestNewDefaultOutputFormatter(t *testing.T) {
	testCases := []struct {
		name           string
		mockWriter     *MockOutputWriter
		testMessage    string
		expectedOutput string
		shouldSucceed  bool
	}{
		{
			name:           "正常系：コンストラクタの動作検証",
			mockWriter:     NewMockOutputWriter(),
			testMessage:    "テストメッセージ",
			expectedOutput: "■ テストメッセージ\n",
			shouldSucceed:  true,
		},
		{
			name:           "異常系：nilのwriter",
			mockWriter:     nil,
			testMessage:    "テストメッセージ",
			expectedOutput: "",
			shouldSucceed:  true, // コンストラクタはnilチェックを行わないため一旦成功
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// コンストラクタ呼び出し
			formatter := NewDefaultOutputFormatter(tc.mockWriter)

			// nilのwriterケースは、メソッド呼び出し時にパニックが発生する可能性がある
			if tc.mockWriter == nil {
				// nilポインタ参照による遅延パニックを検証
				defer func() {
					if r := recover(); r == nil {
						t.Error("nilのwriterでメソッド呼び出し時にパニックが発生するべきですが、発生しませんでした")
					}
				}()

				// メソッド呼び出し（実行時にパニックが発生する可能性）
				formatter.PrintInfo(tc.testMessage)
			} else {
				// 正常系のメソッド呼び出しとアサーション
				formatter.PrintInfo(tc.testMessage)

				actual := tc.mockWriter.GetOutput()
				if actual != tc.expectedOutput {
					t.Errorf("コンストラクタがwriterを正しく設定していません。期待値: %q, 実際: %q", tc.expectedOutput, actual)
				}
			}
		})
	}
}
