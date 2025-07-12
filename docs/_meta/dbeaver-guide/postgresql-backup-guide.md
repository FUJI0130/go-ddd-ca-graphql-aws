# PostgreSQLデータベースのバックアップ手順書

## 1. 環境準備

### 必要なパッケージの確認
```bash
# PostgreSQLクライアントツールのインストール状態確認
sudo apt install postgresql-client

# pg_dumpの場所確認
which pg_dump
# 通常は /usr/bin/pg_dump にインストールされている
```

## 2. DBeaverでのバックアップ手順

1. データベースを右クリック
2. ツール → バックアップ を選択
3. バックアップ設定画面で以下を設定：

### 基本設定
- Format: Custom
- Compression: デフォルト
- Encoding: デフォルト

### 出力設定
- Output folder: バックアップファイルの保存先を指定
- File name pattern: `dump-${database}-${timestamp}.sql`
  
### オプション設定（チェックボックスの説明）
- [  ] Use SQL INSERT instead of COPY for rows
  - チェックを入れると、COPYの代わりにINSERT文を使用
  - データ量が多い場合はチェックを外した方が効率的
- [  ] Do not backup privileges (GRANT/REVOKE)
  - 権限情報のバックアップを除外する場合にチェック
- [  ] Discard objects owner
  - オブジェクトの所有者情報を除外する場合にチェック
- [  ] Add drop database statement
  - リストア時に既存のデータベースを削除するステートメントを含める
- [  ] Add create database statement
  - データベース作成のステートメントを含める

※基本的にはチェックボックスは全て未選択で問題ありません

## 3. 実行とログの確認

正常なバックアップ実行時のログ例：
```
/usr/lib/postgresql/14/bin/pg_dump --verbose --host=localhost --port=5432 --username=postgres --format=c --file /home/fuji0130/dump-test_management-202411101327.sql -n public test_management
Task 'PostgreSQL dump' started at Sun Nov 10 13:27:30 JST 2024
...
pg_dump: dumping contents of table "public.effort_records"
pg_dump: dumping contents of table "public.status_history"
pg_dump: dumping contents of table "public.test_cases"
pg_dump: dumping contents of table "public.test_groups"
pg_dump: dumping contents of table "public.test_suites"
Task 'PostgreSQL dump' finished at Sun Nov 10 13:27:30 JST 2024
```

## 4. トラブルシューティング

### エラー1: pg_dump not found
```
IO error: Utility 'pg_dump' not found in client home '/home/fuji0130/workspace/postgreSQL'
```

**解決策**:
1. PostgreSQLクライアントツールが正しくインストールされているか確認
2. pg_dumpの場所を確認（`which pg_dump`）
3. 必要に応じてパスを環境変数に追加

### エラー2: 権限の問題
```
Unable to examine folder /etc/alternatives while looking for a client home
java.nio.file.AccessDeniedException: /etc/alternatives
```

**解決策**:
```bash
sudo chmod +r /etc/alternatives
```

## 5. 代替手順（コマンドラインでのバックアップ）

DBeaverでの実行に問題がある場合は、直接コマンドラインでバックアップを取ることも可能：

```bash
pg_dump -U postgres -F c -f backup_$(date +%Y%m%d).sql database_name
```

オプションの説明：
- `-U`: ユーザー名
- `-F c`: カスタムフォーマット
- `-f`: 出力ファイル名
- 最後の引数はデータベース名

## 6. 注意事項

- バックアップ前にデータベースの接続状態を確認
- 十分なディスク容量があることを確認
- 定期的なバックアップスケジュールの設定を推奨
- バックアップファイルの保管場所の管理と定期的な確認
