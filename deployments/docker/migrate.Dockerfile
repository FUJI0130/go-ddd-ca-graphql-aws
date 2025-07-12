# ===================================================================
# ファイル名: migrate.Dockerfile
# 配置場所: docker/migrate.Dockerfile
# 説明: データベースマイグレーション専用Dockerコンテナ（x86_64対応版）
# 
# 修正内容:
#  - GOOS, GOARCH環境変数の明示的設定
#  - x86_64バイナリビルドの強制
#  - アーキテクチャ検証の追加
# ===================================================================

# =====================================
# Stage 1: migrate CLIツールのビルド
# =====================================
FROM --platform=linux/amd64 golang:1.23-alpine AS builder

# 必要なパッケージのインストール
RUN apk --no-cache add git file

# x86_64バイナリビルドの明示的設定
ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0

# golang-migrate/migrate CLIツールのインストール
# PostgreSQL対応版を明示的にインストール
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# ビルドされたバイナリのアーキテクチャ確認
RUN echo "=== ビルドバイナリ検証 ===" && \
    file /go/bin/migrate && \
    /go/bin/migrate --version && \
    echo "Platform: $(uname -m)" && \
    echo "Go version: $(go version)"

# =====================================
# Stage 2: 実行用軽量イメージの構築
# =====================================
FROM --platform=linux/amd64 alpine:3.18

# パッケージ情報の更新と必要なパッケージのインストール
RUN apk --no-cache update && \
    apk --no-cache add \
    ca-certificates \
    postgresql-client \
    tzdata \
    file && \
    rm -rf /var/cache/apk/*

# タイムゾーンの設定（日本時間）
ENV TZ=Asia/Tokyo

# migrate CLIバイナリをコピー
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate

# migrate CLIに実行権限を付与
RUN chmod +x /usr/local/bin/migrate

# コピーしたバイナリのアーキテクチャ最終確認
RUN echo "=== 最終バイナリ検証 ===" && \
    file /usr/local/bin/migrate && \
    /usr/local/bin/migrate --version && \
    echo "Container platform: $(uname -m)"

# マイグレーションファイルをコンテナ内にコピー
COPY scripts/migrations /migrations

# マイグレーションファイルの権限設定
RUN chmod -R 644 /migrations

# マイグレーションファイルの構造確認用スクript
RUN echo '#!/bin/sh' > /usr/local/bin/list-migrations && \
    echo 'echo "=== マイグレーションファイル一覧 ==="' >> /usr/local/bin/list-migrations && \
    echo 'ls -la /migrations' >> /usr/local/bin/list-migrations && \
    echo 'echo "=== ファイル内容確認 ==="' >> /usr/local/bin/list-migrations && \
    echo 'for file in /migrations/*.up.sql; do' >> /usr/local/bin/list-migrations && \
    echo '  echo "--- $(basename $file) ---"' >> /usr/local/bin/list-migrations && \
    echo '  head -5 "$file"' >> /usr/local/bin/list-migrations && \
    echo '  echo ""' >> /usr/local/bin/list-migrations && \
    echo 'done' >> /usr/local/bin/list-migrations && \
    chmod +x /usr/local/bin/list-migrations

# デバッグ用スクリプト（接続テスト）
RUN echo '#!/bin/sh' > /usr/local/bin/test-db-connection && \
    echo 'echo "=== データベース接続テスト ==="' >> /usr/local/bin/test-db-connection && \
    echo 'if [ -z "$1" ]; then' >> /usr/local/bin/test-db-connection && \
    echo '  echo "使用方法: test-db-connection <database-url>"' >> /usr/local/bin/test-db-connection && \
    echo '  exit 1' >> /usr/local/bin/test-db-connection && \
    echo 'fi' >> /usr/local/bin/test-db-connection && \
    echo 'psql "$1" -c "SELECT version();"' >> /usr/local/bin/test-db-connection && \
    chmod +x /usr/local/bin/test-db-connection

# ヘルスチェック用スクリプト
RUN echo '#!/bin/sh' > /usr/local/bin/healthcheck && \
    echo 'migrate --version > /dev/null 2>&1' >> /usr/local/bin/healthcheck && \
    chmod +x /usr/local/bin/healthcheck

# 作業ディレクトリの設定
WORKDIR /migrations

# migrate CLI用の環境変数設定
ENV MIGRATION_PATH=/migrations

# migrate CLIの設定確認用情報出力
RUN echo "=== マイグレーションコンテナ情報 ===" && \
    echo "migrate version: $(migrate --version)" && \
    echo "PostgreSQL client version: $(psql --version)" && \
    echo "Available migrations:" && \
    ls -la /migrations && \
    echo "Container timezone: $(date)" && \
    echo "Binary architecture: $(file /usr/local/bin/migrate)" && \
    echo "======================================"

# デフォルトのエントリーポイント
ENTRYPOINT ["migrate"]

# デフォルトコマンド（ヘルプ表示）
CMD ["--help"]

# メタデータラベル
LABEL maintainer="test-management-system"
LABEL description="Database migration container for PostgreSQL (x86_64)"
LABEL version="1.1"
LABEL architecture="x86_64"

# ヘルスチェック設定
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD healthcheck