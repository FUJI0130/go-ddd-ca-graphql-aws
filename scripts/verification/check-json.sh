# JSON生成テスト
echo "=== JSON生成テスト ==="

# 変数設定（実際の値で）
ENVIRONMENT="development"
ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
IMAGE_URI="${ACCOUNT_ID}.dkr.ecr.ap-northeast-1.amazonaws.com/development-test-management-migration:latest"
DB_HOST=""  # 現在空の状態
DB_NAME=""  # 現在空の状態
DB_USERNAME="admin"

echo "変数確認:"
echo "  ACCOUNT_ID: $ACCOUNT_ID"
echo "  IMAGE_URI: $IMAGE_URI"
echo "  DB_HOST: '$DB_HOST'"
echo "  DB_NAME: '$DB_NAME'"

# JSON生成テスト（簡略版）
TASK_DEFINITION_TEST=$(cat <<EOF
{
  "family": "development-migration",
  "cpu": "256",
  "memory": "512",
  "containerDefinitions": [
    {
      "name": "migration-container",
      "image": "${IMAGE_URI}",
      "command": [
        "migrate",
        "-path", "/migrations", 
        "-database", "postgresql://${DB_USERNAME}:\${DB_PASSWORD}@${DB_HOST}:5432/${DB_NAME}?sslmode=require",
        "up"
      ]
    }
  ]
}
EOF
)

echo "生成されたJSON（最初の500文字）:"
echo "$TASK_DEFINITION_TEST" | head -c 500
echo ""
echo "JSON妥当性チェック:"
echo "$TASK_DEFINITION_TEST" | jq . > /dev/null 2>&1 && echo "✓ JSON有効" || echo "✗ JSON無効"