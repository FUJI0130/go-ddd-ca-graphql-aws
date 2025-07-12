# プラットフォーム指定（x86_64アーキテクチャ）
FROM --platform=linux/amd64 postgres:14-alpine

# テストデータファイルをコピー
COPY scripts/testdata/aws-test-users.sql /sql/

# 実行用エントリーポイント  
ENTRYPOINT ["psql"]