# ビルドステージ
FROM golang:1.22-alpine AS builder
ENV GOTOOLCHAIN=auto

# 必要なツールのインストール
RUN apk add --no-cache git

# 作業ディレクトリの設定
WORKDIR /app

# 依存関係のダウンロード
COPY go.mod go.sum ./
RUN go mod download

# ソースコードのコピー
COPY . .

# ビルド変数
ARG SERVICE_TYPE=api
ARG VERSION=dev

# 選択されたサービスタイプに基づいて適切なバイナリをビルド
# ARM64アーキテクチャを明示的に指定
RUN case $SERVICE_TYPE in \
    api) \
      echo "Building REST API server for ARM64" && \
      CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o /app/server -ldflags="-X main.version=$VERSION" ./cmd/api/main.go \
      ;; \
    graphql) \
      echo "Building GraphQL server for ARM64" && \
      CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o /app/server -ldflags="-X main.version=$VERSION" ./cmd/graphql/main.go \
      ;; \
    grpc) \
      echo "Building gRPC server for ARM64" && \
      CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o /app/server -ldflags="-X main.version=$VERSION" ./cmd/grpc/main.go \
      ;; \
    *) \
      echo "Unknown service type: $SERVICE_TYPE, defaulting to REST API for ARM64" && \
      CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o /app/server -ldflags="-X main.version=$VERSION" ./cmd/api/main.go \
      ;; \
    esac

# 実行ステージ
FROM --platform=linux/arm64 alpine:3.17

# タイムゾーン設定
RUN apk add --no-cache tzdata
ENV TZ=Asia/Tokyo

# 必要なパッケージのインストール
RUN apk add --no-cache ca-certificates

# バイナリコピー
COPY --from=builder /app/server /app/server

# 設定ファイルのコピー - 12-Factor原則に従い、環境変数を優先するため含めない
# ローカル開発では configs/development.yml を使用し、クラウド環境では環境変数のみを使用
# COPY --from=builder /app/configs /app/configs

# 設定ファイルが含まれていないことを明示的に保証（通常は不要だが意図を明確化）
# RUN rm -rf /app/configs
# マイグレーションスクリプトのコピー（オプション）
COPY --from=builder /app/scripts/migrations /app/scripts/migrations

# ワーキングディレクトリ設定
WORKDIR /app

# デフォルト環境変数の設定
ENV APP_ENVIRONMENT=production \
    SERVICE_TYPE=api \
    PORT=8080 \
    GRPC_PORT=50051

# ポート公開
EXPOSE 8080 50051

# ヘルスチェック設定 - サービスタイプに応じて適切なエンドポイントを使用
# このヘルスチェックはコンテナ内部から実行されるため、localhostの使用は適切
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD case $SERVICE_TYPE in \
      api) \
        wget -qO- http://localhost:$PORT/health || exit 1 \
        ;; \
      graphql) \
        wget -qO- http://localhost:$PORT/health || exit 1 \
        ;; \
      grpc) \
        # gRPCのヘルスチェックオプション
        # オプション1: ポートチェック（基本的な接続確認のみ）
        nc -z localhost $GRPC_PORT || exit 1 \
        # オプション2: 専用のHTTPヘルスチェックが実装されている場合
        # wget -qO- http://localhost:8080/health || exit 1 \
        ;; \
      *) \
        wget -qO- http://localhost:$PORT/health || exit 1 \
        ;; \
      esac

# 起動コマンド
CMD ["./server"]