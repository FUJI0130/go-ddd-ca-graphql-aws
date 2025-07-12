#!/bin/bash
set -e

# タイトルを表示
echo "========================================"
echo "   REST API サービスのテスト実行"
echo "========================================"

# 変数設定
SERVICE_NAME="test-management-api"
CONTAINER_NAME="test-api-container"
DB_CONTAINER_NAME="test_management_db"
PORT=8080
# Docker Composeが作成するネットワーク名を指定（プロジェクト名_ネットワーク名）
NETWORK_NAME="docker_app_network"

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

# すでに起動しているAPIコンテナを停止・削除
echo "既存のAPIコンテナをクリーンアップ中..."
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
docker build -t $SERVICE_NAME --build-arg SERVICE_TYPE=api .

# コンテナの起動
echo "コンテナを起動中: $CONTAINER_NAME on port $PORT..."
if [ "$NETWORK_NAME" == "bridge" ]; then
  # bridgeネットワークを使用する場合、コンテナ間通信のためにDB_HOSTにホストのIPを指定
  HOST_IP=$(hostname -I | awk '{print $1}')
  echo "ホストのIPアドレス '$HOST_IP' を使用してDBに接続します"
  docker run -d --name $CONTAINER_NAME -p $PORT:8080 \
    -e DB_HOST=$HOST_IP \
    -e DB_PORT=5432 \
    -e DB_USER=testuser \
    -e DB_PASSWORD=testpass \
    -e DB_NAME=test_management \
    -e DB_SSL_MODE=disable \
    $SERVICE_NAME
else
  # カスタムネットワークを使用する場合、コンテナ名でDB_HOSTを指定
  docker run -d --name $CONTAINER_NAME -p $PORT:8080 \
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

# ヘルスチェックの実行
echo "ヘルスチェックを実行中..."
if ! curl -s http://localhost:$PORT/health; then
  echo "ヘルスチェックに失敗しました。コンテナログを確認します:"
  docker logs $CONTAINER_NAME
  cleanup "with_db"
  exit 1
fi

# API呼び出しのテスト
echo "APIエンドポイントのテスト実行中..."
if ! curl -s -X GET http://localhost:$PORT/api/v1/test-suites; then
  echo "APIテストに失敗しました。コンテナログを確認します:"
  docker logs $CONTAINER_NAME
  cleanup "with_db"
  exit 1
fi

# テスト成功メッセージ
echo -e "\nテスト成功！"

# コンテナのログを表示
echo -e "\nコンテナのログ:"
docker logs $CONTAINER_NAME

# テスト終了時の選択肢を提供
echo -e "\nテスト完了！以下のオプションを選択してください:"
echo "1) コンテナを実行したままにする"
echo "2) APIコンテナのみ停止して削除する"
echo "3) APIコンテナとデータベースコンテナを停止して削除する"
read -p "選択 (デフォルト: 3): " choice

case $choice in
  1)
    echo "コンテナはバックグラウンドで実行中です。終了するには:"
    echo "  docker stop $CONTAINER_NAME"
    echo "  docker rm $CONTAINER_NAME"
    ;;
  2)
    echo "APIコンテナを停止・削除中..."
    docker stop $CONTAINER_NAME
    docker rm $CONTAINER_NAME
    echo "APIコンテナのクリーンアップ完了"
    ;;
  *)
    echo "APIコンテナとデータベースコンテナを停止・削除中..."
    docker stop $CONTAINER_NAME
    docker rm $CONTAINER_NAME
    make db-down
    echo "すべてのコンテナのクリーンアップ完了"
    ;;
esac

echo "テスト完了"