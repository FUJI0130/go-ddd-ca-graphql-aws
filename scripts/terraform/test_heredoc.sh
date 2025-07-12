#!/bin/bash

# テスト環境変数設定
export DB_USERNAME="admin"
export DB_PASSWORD="SecurePassword123!"
export DB_HOST="test-host"
export DB_NAME="test_db"

echo "=== 環境変数確認 ==="
echo "DB_USERNAME: ${DB_USERNAME}"
echo "DB_PASSWORD: ${DB_PASSWORD}"
echo "DB_HOST: ${DB_HOST}"
echo "DB_NAME: ${DB_NAME}"

echo -e "\n=== Here Document テスト1: クォートなし ==="
RESULT1=$(cat <<EOF
postgresql://\\$DB_USERNAME:\\$DB_PASSWORD@\\$DB_HOST:5432/\\$DB_NAME
EOF
)
echo "結果1: ${RESULT1}"

echo -e "\n=== Here Document テスト2: 'EOF'クォートあり ==="
RESULT2=$(cat <<'EOF'
postgresql://\$DB_USERNAME:\$DB_PASSWORD@\$DB_HOST:5432/\$DB_NAME
EOF
)
echo "結果2: ${RESULT2}"

echo -e "\n=== Here Document テスト3: 変数展開なし ==="
RESULT3=$(cat <<EOF
postgresql://${DB_USERNAME}:${DB_PASSWORD}@${DB_HOST}:5432/${DB_NAME}
EOF
)
echo "結果3: ${RESULT3}"

echo -e "\n=== JSON生成テスト ==="
JSON_TEST=$(cat <<EOF
{
  "command": [
    "sh", "-c",
    "migrate -database 'postgresql://\\$DB_USERNAME:\\$DB_PASSWORD@\\$DB_HOST:5432/\\$DB_NAME' up"
  ]
}
EOF
)
echo "JSON結果:"
echo "${JSON_TEST}"
SCRIPT