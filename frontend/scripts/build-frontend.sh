#!/bin/bash
# frontend/scripts/build-frontend.sh
# フロントエンドビルドスクリプト（環境変数動的生成対応）

set -e

# 色の設定
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

ENVIRONMENT=${1:-development}

echo -e "${BLUE}フロントエンドをビルドしています (環境: ${ENVIRONMENT})...${NC}"

# 環境変数ファイルの読み込み
if [ -f ~/.env.terraform ]; then
    source ~/.env.terraform
    echo -e "${GREEN}✓ ~/.env.terraform を読み込みました${NC}"
else
    echo -e "${YELLOW}⚠ ~/.env.terraform が見つかりません${NC}"
fi

# GraphQL ALB DNS名の確認
if [ -z "${GRAPHQL_ALB_DNS_NAME}" ]; then
    echo -e "${YELLOW}⚠ GRAPHQL_ALB_DNS_NAME が設定されていません${NC}"
    echo -e "${YELLOW}  バックエンドがデプロイされているか確認してください${NC}"
    GRAPHQL_ALB_DNS_NAME="localhost:8080"
    echo -e "${YELLOW}  フォールバック値を使用します: ${GRAPHQL_ALB_DNS_NAME}${NC}"
fi

# アプリケーションバージョンの取得
APP_VERSION=$(cat package.json | grep '"version"' | sed 's/.*"version": "\(.*\)".*/\1/' 2>/dev/null || echo "0.1.0")

# 環境変数ファイル生成（.env.production を動的生成）
echo -e "${BLUE}環境変数ファイルを生成しています...${NC}"

cat > .env.production << EOF
# 動的生成された本番環境変数ファイル
# 生成日時: $(date)
# 環境: ${ENVIRONMENT}

# GraphQL API設定
# VITE_GRAPHQL_API_URL=http://${GRAPHQL_ALB_DNS_NAME}/query
# VITE_GRAPHQL_API_URL=https://graphql.grpc-dev-fuji0130.com/query
VITE_GRAPHQL_API_URL=https://example-graphql-api.com/query

# アプリケーション設定
VITE_APP_NAME=テスト管理システム
VITE_APP_VERSION=${APP_VERSION}
VITE_APP_ENV=${ENVIRONMENT}

# 認証設定（Cookie認証のためトークンキーは参考値）
VITE_JWT_LOCAL_STORAGE_KEY=auth_token
VITE_REFRESH_TOKEN_LOCAL_STORAGE_KEY=refresh_token

# デプロイ情報
VITE_BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
VITE_DEPLOYMENT_ENV=${ENVIRONMENT}
EOF

echo -e "${GREEN}✓ .env.production を生成しました${NC}"
echo -e "${BLUE}  GraphQL API URL: https://${GRAPHQL_ALB_DNS_NAME}/query${NC}"

# package.json の存在確認
if [ ! -f "package.json" ]; then
    echo -e "${RED}エラー: package.json が見つかりません${NC}"
    echo -e "${RED}フロントエンドディレクトリで実行してください${NC}"
    exit 1
fi

# node_modules の確認とインストール
if [ ! -d "node_modules" ]; then
    echo -e "${BLUE}依存関係をインストールしています...${NC}"
    npm ci
    echo -e "${GREEN}✓ 依存関係のインストールが完了しました${NC}"
else
    echo -e "${GREEN}✓ node_modules が存在します${NC}"
fi

# TypeScript型チェック
echo -e "${BLUE}TypeScript型チェックを実行しています...${NC}"
npm run type-check || {
    echo -e "${YELLOW}⚠ TypeScript型エラーがありますが、ビルドを続行します${NC}"
}

# GraphQL Code Generatorの実行
if [ -f "codegen.yml" ]; then
    echo -e "${BLUE}GraphQL型定義を生成しています...${NC}"
    npm run codegen || {
        echo -e "${YELLOW}⚠ GraphQL Code Generator でエラーが発生しましたが、続行します${NC}"
    }
else
    echo -e "${YELLOW}⚠ codegen.yml が見つかりません。GraphQL型生成をスキップします${NC}"
fi

# ビルド実行
echo -e "${BLUE}Reactアプリケーションをビルドしています...${NC}"
npm run build

# ビルド結果の確認
if [ -d "dist" ]; then
    DIST_SIZE=$(du -sh dist | cut -f1)
    FILE_COUNT=$(find dist -type f | wc -l)
    echo -e "${GREEN}✓ ビルドが完了しました${NC}"
    echo -e "${GREEN}  ビルドサイズ: ${DIST_SIZE}${NC}"
    echo -e "${GREEN}  ファイル数: ${FILE_COUNT}${NC}"
    
    # 主要ファイルの確認
    echo -e "${BLUE}ビルド成果物:${NC}"
    if [ -f "dist/index.html" ]; then
        echo -e "  ✓ index.html"
    else
        echo -e "  ✗ index.html が見つかりません"
    fi
    
    if [ -d "dist/assets" ]; then
        JS_FILES=$(find dist/assets -name "*.js" | wc -l)
        CSS_FILES=$(find dist/assets -name "*.css" | wc -l)
        echo -e "  ✓ assets/ (JS: ${JS_FILES}, CSS: ${CSS_FILES})"
    else
        echo -e "  ✗ assets/ ディレクトリが見つかりません"
    fi
else
    echo -e "${RED}エラー: ビルドに失敗しました。dist/ ディレクトリが作成されていません${NC}"
    exit 1
fi

# ビルド設定情報の出力
echo -e "${BLUE}=== ビルド設定情報 ===${NC}"
echo -e "環境: ${ENVIRONMENT}"
echo -e "GraphQL API: https://${GRAPHQL_ALB_DNS_NAME}/query"
echo -e "アプリバージョン: ${APP_VERSION}"
echo -e "ビルド時刻: $(date)"

echo -e "${GREEN}フロントエンドビルドが正常に完了しました！${NC}"