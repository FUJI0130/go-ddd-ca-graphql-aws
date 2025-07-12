#!/bin/bash

# 接続パラメータ
HOST="localhost"
PORT="5433"
USER="test_user"
PASSWORD="test_pass"
DB="test_db"

# 最大試行回数
MAX_ATTEMPTS=30
ATTEMPT=0

echo "PostgreSQLサーバーの準備ができるまで待機中..."

while [ $ATTEMPT -lt $MAX_ATTEMPTS ]; do
    ATTEMPT=$((ATTEMPT+1))
    echo "接続試行 $ATTEMPT/$MAX_ATTEMPTS..."
    
    # PostgreSQLに接続を試みる
    PGPASSWORD=$PASSWORD psql -h $HOST -p $PORT -U $USER -d $DB -c "SELECT 1" >/dev/null 2>&1
    
    if [ $? -eq 0 ]; then
        echo "PostgreSQLサーバーに接続できました！"
        exit 0
    fi
    
    echo "PostgreSQLサーバーはまだ準備ができていません。2秒待機します..."
    sleep 2
done

echo "PostgreSQLサーバーへの接続試行が最大回数に達しました。"
exit 1