# ベースMakefile - 共通変数と基本コマンド
# makefiles/base.mk

# 環境変数のデフォルト値設定
NODE_ENV ?= development
BUILD_MODE ?= development

# カラー設定
BLUE := \033[0;34m
GREEN := \033[0;32m
YELLOW := \033[0;33m
RED := \033[0;31m
NC := \033[0m # No Color

# フロントエンド用ツリー表示（node_modules等を除外）
.PHONY: tree
tree:
	tree -a -I 'node_modules|dist|build|.git|coverage|.next|.nuxt|.output'

# ヘルプタスク
.PHONY: help
help:
	@echo -e "${BLUE}フロントエンド開発 - Makefile使用ガイド${NC}"
	@echo ""
	@echo "=== 開発用コマンド ==="
	@echo "  dev               - 開発サーバーを起動"
	@echo "  dev-host          - ホストネットワークで開発サーバーを起動"
	@echo "  build             - プロダクション用ビルド"
	@echo "  preview           - ビルド結果をプレビュー"
	@echo "  tree              - プロジェクトファイル構造を表示"
	@echo ""
	@echo "=== 依存関係管理 ==="
	@echo "  install           - 依存パッケージをインストール"
	@echo "  install-clean     - node_modulesを削除してクリーンインストール"
	@echo "  update            - パッケージを更新"
	@echo "  audit             - セキュリティ監査"
	@echo ""
	@echo "=== コード品質 ==="
	@echo "  lint              - ESLintでコードチェック"
	@echo "  lint-fix          - ESLintで自動修正"
	@echo "  format            - Prettierでコード整形"
	@echo "  type-check        - TypeScript型チェック"
	@echo ""
	@echo "=== テスト関連 ==="
	@echo "  test              - テストを実行"
	@echo "  test-watch        - テストをウォッチモードで実行"
	@echo "  test-coverage     - カバレッジ付きでテスト実行"
	@echo ""
	@echo "=== GraphQL関連 ==="
	@echo "  generate          - GraphQL Code Generatorを実行"
	@echo "  generate-watch    - GraphQL Code Generatorをウォッチモードで実行"
	@echo ""
	@echo "=== 環境変数 ==="
	@echo "  NODE_ENV          - 環境 (development|production) デフォルト: ${NODE_ENV}"
	@echo "  BUILD_MODE        - ビルドモード (development|production) デフォルト: ${BUILD_MODE}"
	@echo ""