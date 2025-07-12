# データベース設計検証の経緯 (2024-11-10)

## 1. セットアップフェーズ (11/09)

### 1.1 環境構築
1. PostgreSQLのインストールと初期設定
   - PostgreSQL 14.13を選択
   - 基本設定の調整（listen_addresses, pg_hba.conf）

2. DBeaverのセットアップ
   - バージョン24.2.4を導入
   - 接続テスト完了

### 1.2 初期スキーマ作成
1. データベース作成
   ```sql
   CREATE DATABASE test_management;
   ```

2. 基本型の定義
   ```sql
   CREATE TYPE test_status_enum AS ENUM (...);
   CREATE TYPE priority_enum AS ENUM (...);
   ```

## 2. テーブル作成フェーズ (11/09-11/10)

### 2.1 共通機能の実装
1. 更新日時管理の関数作成
   - 初回作成時に問題発生
   - スキーマ指定で解決

### 2.2 テーブル作成
1. 基本テーブルの作成
   - test_suites
   - test_groups
   - test_cases

2. 関連テーブルの作成
   - effort_records
   - status_history

## 3. 機能検証フェーズ (11/10)

### 3.1 基本機能の検証
1. ステータス変更テスト
   ```sql
   -- テストケース作成
   INSERT INTO test_cases ...
   
   -- ステータス変更
   UPDATE test_cases SET status = 'テスト' ...
   ```

2. 工数記録テスト
   ```sql
   INSERT INTO effort_records ...
   ```

### 3.2 進捗計算の検証
1. テストデータの準備
   - 複数のテストケース作成
   - 異なる状態と優先度の組み合わせ

2. 集計機能の確認
   - グループレベルの集計
   - スイート全体の進捗確認

## 4. 課題対応の経緯

### 4.1 工数の負値問題
1. 問題検出
   ```sql
   INSERT INTO effort_records (effort_amount) VALUES (-1.0);
   -- 成功してしまう
   ```

2. 対応実施
   ```sql
   ALTER TABLE effort_records 
   ADD CONSTRAINT positive_effort CHECK (effort_amount > 0);
   ```

### 4.2 その他の課題
1. Critical優先度の進捗未定義
   - 仕様の明確化が必要と判断
   - 次フェーズでの対応を決定

2. グループ移動の履歴
   - 現状の仕組みでは不十分と判断
   - 拡張方法を検討中

## 5. 最終確認内容 (11/10)
- テーブル構造の確認
- 制約の動作確認
- トリガーの動作確認
- ENUMの制約確認

## 6. 更新履歴
- 2024-11-09: 環境構築と基本設計開始
- 2024-11-10: 
  - AM: テーブル作成と基本機能の検証
  - PM: 進捗計算機能の検証と課題対応
