#!/bin/bash
set -e

# タイトルを表示
echo "========================================"
echo "   GraphQL サービスのテスト実行"
echo "========================================"

# 変数設定
SERVICE_NAME="test-management-graphql"
CONTAINER_NAME="test-graphql-container"
DB_CONTAINER_NAME="test_management_db"
PORT=8080
QUERY_DIR="scripts/docker/graphql-queries"

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

# クエリディレクトリが存在するか確認、なければ作成
mkdir -p $QUERY_DIR

# サンプルクエリファイルが存在しなければ作成
SAMPLE_QUERY_FILE="$QUERY_DIR/list_test_suites.graphql"
if [ ! -f "$SAMPLE_QUERY_FILE" ]; then
  echo 'query ListTestSuites {
  testSuites {
    edges {
      node {
        id
        name
        status
        progress
      }
    }
    totalCount
  }
}' > "$SAMPLE_QUERY_FILE"
  echo "サンプルクエリファイルを作成しました: $SAMPLE_QUERY_FILE"
fi

# すでに起動しているコンテナを停止・削除
echo "既存のGraphQLコンテナをクリーンアップ中..."
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
docker build -t $SERVICE_NAME --build-arg SERVICE_TYPE=graphql .

# コンテナの起動
echo "コンテナを起動中: $CONTAINER_NAME on port $PORT..."
if [ "$NETWORK_NAME" == "bridge" ]; then
  # bridgeネットワークを使用する場合、コンテナ間通信のためにDB_HOSTにホストのIPを指定
  HOST_IP=$(hostname -I | awk '{print $1}')
  echo "ホストのIPアドレス '$HOST_IP' を使用してDBに接続します"
  docker run -d --name $CONTAINER_NAME -p $PORT:$PORT \
    -e DB_HOST=$HOST_IP \
    -e DB_PORT=5432 \
    -e DB_USER=testuser \
    -e DB_PASSWORD=testpass \
    -e DB_NAME=test_management \
    -e DB_SSL_MODE=disable \
    -e SERVICE_TYPE=graphql \
    -e PORT=$PORT \
    $SERVICE_NAME
else
  # カスタムネットワークを使用する場合、コンテナ名でDB_HOSTを指定
  docker run -d --name $CONTAINER_NAME -p $PORT:$PORT \
    --network $NETWORK_NAME \
    -e DB_HOST=$DB_CONTAINER_NAME \
    -e DB_PORT=5432 \
    -e DB_USER=testuser \
    -e DB_PASSWORD=testpass \
    -e DB_NAME=test_management \
    -e DB_SSL_MODE=disable \
    -e SERVICE_TYPE=graphql \
    -e PORT=$PORT \
    $SERVICE_NAME
fi

# コンテナの起動を待機
echo "サービスの起動を待機中..."
sleep 5

# GraphQL Playgroundへのアクセス確認
echo "GraphQLサービスのテスト実行中..."
echo "GraphQL Playground: http://localhost:$PORT/"

# GraphQLエンドポイントのヘルスチェック
echo "GraphQLエンドポイントのチェック中..."
RESPONSE=$(curl -s http://localhost:$PORT/)
ENDPOINT_STATUS=$?

if [ $ENDPOINT_STATUS -eq 0 ] && [ -n "$RESPONSE" ]; then
  echo "✅ GraphQLエンドポイントは正常です"
  # レスポンスに特定のキーワードが含まれているか確認（GraphQL Playgroundの特徴）
  if echo "$RESPONSE" | grep -q "GraphQL" || echo "$RESPONSE" | grep -q "playground"; then
    echo "✅ GraphQL Playgroundが応答しています"
  fi
else
  echo "❌ GraphQLエンドポイントの確認に失敗しました (ステータス: $ENDPOINT_STATUS)"
  echo "コンテナログを確認します:"
  docker logs $CONTAINER_NAME
  cleanup "with_db"
  exit 1
fi

# GraphQLクエリの実行
echo -e "\n----- GraphQLクエリテスト -----"

# クエリファイルがあれば実行
if ls $QUERY_DIR/*.graphql 1> /dev/null 2>&1; then
  for query_file in $QUERY_DIR/*.graphql; do
    query_name=$(basename "$query_file" .graphql)
    echo -e "\n▶ クエリの実行: $query_name"
    echo "📄 クエリファイル: $query_file"
    
    query_content=$(cat "$query_file")
    echo -e "📝 クエリ内容:\n$query_content"
    
    # jqコマンドの有無を確認し、適切なJSONフォーマットを使用
    if command -v jq &> /dev/null; then
      # jqを使用してJSONをフォーマット
      query_json="{\"query\": $(echo "$query_content" | jq -Rs .)}"
    else
      # jqがない場合は単純なエスケープを使用
      query_json="{\"query\": \"$(echo "$query_content" | sed 's/"/\\"/g' | sed ':a;N;$!ba;s/\n/\\n/g')\"}"
    fi
    
    echo -e "📊 実行結果:"
    echo "送信するJSONクエリ: $query_json"
    
    # クエリの実行
    response=$(curl -s -X POST \
      -H "Content-Type: application/json" \
      -d "$query_json" \
      http://localhost:$PORT/query)
    
    # jqが利用可能ならJSONをフォーマットして表示
    if command -v jq &> /dev/null; then
      echo "$response" | jq '.' || echo "$response"
    else
      echo "$response"
    fi
    
    # エラーチェック
    if echo "$response" | grep -q "errors"; then
      echo "❌ クエリ実行中にエラーが発生しました"
      echo "コンテナログを確認します:"
      docker logs $CONTAINER_NAME
      cleanup "with_db"
      exit 1
    else
      echo "✅ クエリは正常に実行されました"
    fi
  done
else
  echo "テスト用のGraphQLクエリファイルが見つかりません。"
  echo "以下にクエリファイルを作成してください: $QUERY_DIR/*.graphql"
fi

# テスト成功メッセージ
echo -e "\nテスト成功！"

# コンテナのログを表示
echo -e "\nコンテナのログ:"
docker logs $CONTAINER_NAME

# テスト終了時の選択肢を提供
echo -e "\nテスト完了！以下のオプションを選択してください:"
echo "1) コンテナを実行したままにする"
echo "2) GraphQLコンテナのみ停止して削除する"
echo "3) GraphQLコンテナとデータベースコンテナを停止して削除する"
echo "4) クエリを手動で実行する (GraphQL Playground)"
echo "5) 新しいクエリをテストに追加する"
read -p "選択 (デフォルト: 3): " choice

case $choice in
  1)
    echo "コンテナはバックグラウンドで実行中です。終了するには:"
    echo "  docker stop $CONTAINER_NAME"
    echo "  docker rm $CONTAINER_NAME"
    echo "GraphQL Playgroundにアクセスするには: http://localhost:$PORT/"
    ;;
  2)
    echo "GraphQLコンテナを停止・削除中..."
    docker stop $CONTAINER_NAME
    docker rm $CONTAINER_NAME
    echo "GraphQLコンテナのクリーンアップ完了"
    ;;
  4)
    echo "GraphQL Playgroundを開きます: http://localhost:$PORT/"
    # プラットフォームによってブラウザオープンコマンドを選択
    if [[ "$OSTYPE" == "darwin"* ]]; then
      open "http://localhost:$PORT/"
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
      xdg-open "http://localhost:$PORT/" &> /dev/null || echo "ブラウザを手動で開いてください"
    else
      echo "ブラウザを手動で開き、以下にアクセスしてください: http://localhost:$PORT/"
    fi
    ;;
  5)
    read -p "新しいクエリの名前 (例: create_test_suite): " query_name
    echo "新しいクエリファイルを作成します: $QUERY_DIR/${query_name}.graphql"
    echo "クエリを入力してください (入力終了後にCtrl+Dを押してください):"
    cat > "$QUERY_DIR/${query_name}.graphql"
    echo "クエリが保存されました: $QUERY_DIR/${query_name}.graphql"
    ;;
  *)
    echo "GraphQLコンテナとデータベースコンテナを停止・削除中..."
    docker stop $CONTAINER_NAME
    docker rm $CONTAINER_NAME
    make db-down
    echo "すべてのコンテナのクリーンアップ完了"
    ;;
esac

echo "テスト完了"