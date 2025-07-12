#!/bin/bash
set -e

# タイトルを表示
echo "========================================"
echo "   gRPC サービスのテスト実行"
echo "========================================"

# 変数設定
SERVICE_NAME="test-management-grpc"
CONTAINER_NAME="test-grpc-container"
DB_CONTAINER_NAME="test_management_db"
PORT=50051

# クリーンアップ関数
cleanup() {
  echo "クリーンアップを実行します..."
  docker stop $CONTAINER_NAME 2>/dev/null || true
  docker rm $CONTAINER_NAME 2>/dev/null || true
  
  # 選択に応じてDBコンテナを停止
  if [ "$1" == "with_db" ]; then
    echo "データベースコンテナを停止します..."
    make db-down
  fi
}

# エラー時にもクリーンアップを実行
trap 'cleanup' ERR

# すでに起動しているコンテナを停止・削除
echo "既存のgRPCコンテナをクリーンアップ中..."
docker stop $CONTAINER_NAME 2>/dev/null || true
docker rm $CONTAINER_NAME 2>/dev/null || true

# データベース起動（すでに実行中の場合は何もしない）
echo "データベースコンテナを確認/起動中..."
if ! docker ps | grep -q $DB_CONTAINER_NAME; then
  echo "データベースコンテナを起動します..."
  make db-up
  # DBの初期化を待つ
  echo "データベースの初期化を待機中..."
  sleep 10
else
  echo "データベースコンテナはすでに実行中です"
fi

# 現在のネットワーク名を確認
echo "利用可能なDocker Network一覧:"
docker network ls

# ネットワーク名を決定（Docker Composeのネットワーク名を確認）
COMPOSE_NETWORK=$(docker network ls | grep app_network | awk '{print $2}')
if [ -n "$COMPOSE_NETWORK" ]; then
  NETWORK_NAME=$COMPOSE_NETWORK
  echo "Docker Composeのネットワーク '$NETWORK_NAME' を使用します"
else
  echo "Docker Composeのネットワークが見つかりません。デフォルトのbridgeネットワークを使用します"
  NETWORK_NAME="bridge"
fi

# イメージのビルド
echo "イメージをビルド中: $SERVICE_NAME..."
docker build -t $SERVICE_NAME --build-arg SERVICE_TYPE=grpc .

# コンテナの起動
echo "コンテナを起動中: $CONTAINER_NAME on port $PORT..."
if [ "$NETWORK_NAME" == "bridge" ]; then
  # bridgeネットワークを使用する場合、コンテナ間通信のためにDB_HOSTにホストのIPを指定
  HOST_IP=$(hostname -I | awk '{print $1}')
  echo "ホストのIPアドレス '$HOST_IP' を使用してDBに接続します"
  docker run -d --name $CONTAINER_NAME -p $PORT:50051 \
    -e DB_HOST=$HOST_IP \
    -e DB_PORT=5432 \
    -e DB_USER=testuser \
    -e DB_PASSWORD=testpass \
    -e DB_NAME=test_management \
    -e DB_SSL_MODE=disable \
    $SERVICE_NAME
else
  # カスタムネットワークを使用する場合、コンテナ名でDB_HOSTを指定
  docker run -d --name $CONTAINER_NAME -p $PORT:50051 \
    --network $NETWORK_NAME \
    -e DB_HOST=$DB_CONTAINER_NAME \
    -e DB_PORT=5432 \
    -e DB_USER=testuser \
    -e DB_PASSWORD=testpass \
    -e DB_NAME=test_management \
    -e DB_SSL_MODE=disable \
    $SERVICE_NAME
fi

# コンテナの起動を待機
echo "サービスの起動を待機中..."
sleep 5

# gRPCの接続テスト (grpcurlがある場合)
GRPC_TEST_SUCCESS=false
GRPC_TEST_PARTIAL=false

if command -v grpcurl &> /dev/null; then
    echo "gRPCサービスのテスト実行中 (grpcurl使用)..."
    echo "利用可能なサービス一覧:"
    if grpcurl -plaintext localhost:$PORT list; then
        echo "✅ gRPCサービスは正常に動作しています"
        GRPC_TEST_SUCCESS=true
        
        # オプション: TestSuiteサービスの詳細を表示
        echo -e "\nTestSuiteサービスの詳細:"
        grpcurl -plaintext localhost:$PORT list testsuite.v1.TestSuiteService || echo "サービス詳細の取得に失敗しました"
    else
        echo "❌ gRPCサービス一覧の取得に失敗しました"
        echo "コンテナログを確認します:"
        docker logs $CONTAINER_NAME
    fi
else
    echo "⚠️ grpcurlがインストールされていないため、部分的な検証のみ実施します"
    if docker ps | grep -q $CONTAINER_NAME; then
        echo "✅ gRPCコンテナは起動していますが、エンドポイントの機能テストは実施されていません"
        GRPC_TEST_PARTIAL=true
        GRPC_TEST_SUCCESS=true
    else
        echo "❌ gRPCコンテナの起動に失敗しました"
    fi
fi

# コンテナのログを表示
echo -e "\nコンテナのログ:"
docker logs $CONTAINER_NAME

# テスト失敗の場合は終了
if [ "$GRPC_TEST_SUCCESS" != "true" ]; then
    echo "❌ gRPCサービスのテストに失敗しました"
    cleanup "with_db"
    exit 1
fi

# テスト結果表示
if [ "$GRPC_TEST_PARTIAL" == "true" ]; then
    echo -e "\n⚠️ 部分的なテスト検証のみ成功しました。完全なテストにはgrpcurlのインストールが必要です。"
else
    echo -e "\nテスト成功！"
fi

# テスト終了時の選択肢を提供
echo -e "\nテスト完了！以下のオプションを選択してください:"
echo "1) コンテナを実行したままにする"
echo "2) gRPCコンテナのみ停止して削除する"
echo "3) gRPCコンテナとデータベースコンテナを停止して削除する"
read -p "選択 (デフォルト: 3): " choice

case $choice in
  1)
    echo "コンテナはバックグラウンドで実行中です。終了するには:"
    echo "  docker stop $CONTAINER_NAME"
    echo "  docker rm $CONTAINER_NAME"
    echo "gRPCサーバーアドレス: localhost:$PORT"
    ;;
  2)
    echo "gRPCコンテナを停止・削除中..."
    docker stop $CONTAINER_NAME
    docker rm $CONTAINER_NAME
    echo "gRPCコンテナのクリーンアップ完了"
    ;;
  *)
    echo "gRPCコンテナとデータベースコンテナを停止・削除中..."
    docker stop $CONTAINER_NAME
    docker rm $CONTAINER_NAME
    make db-down
    echo "すべてのコンテナのクリーンアップ完了"
    ;;
esac

echo "テスト完了"