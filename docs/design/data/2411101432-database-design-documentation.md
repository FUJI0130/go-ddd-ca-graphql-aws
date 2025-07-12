# データベース設計完了図 (2024-11-10)

## 1. ER図

```mermaid
erDiagram
    test_suites ||--o{ test_groups : "contains"
    test_groups ||--o{ test_cases : "contains"
    test_cases ||--o{ effort_records : "tracks"
    test_cases ||--o{ status_history : "logs"

    test_suites {
        varchar(50) id PK
        varchar(100) name
        text description
        date estimated_start_date
        date estimated_end_date
        boolean require_effort_comment
        timestamp created_at
        timestamp updated_at
        suite_status_enum status
    }

    test_groups {
        varchar(50) id PK
        varchar(50) suite_id FK
        varchar(100) name
        text description
        integer display_order
        timestamp created_at
        timestamp updated_at
        suite_status_enum status
    }

    test_cases {
        varchar(50) id PK
        varchar(50) group_id FK
        varchar(200) title
        text description
        test_status_enum status
        priority_enum priority
        float planned_effort
        float actual_effort
        boolean is_delayed
        integer delay_days
        varchar(100) current_editor
        boolean is_locked
        timestamp created_at
        timestamp updated_at
    }

    effort_records {
        integer id PK
        varchar(50) test_case_id FK
        date record_date
        float effort_amount
        boolean is_additional
        text comment
        varchar(100) recorded_by
        timestamp created_at
    }

    status_history {
        integer id PK
        varchar(50) test_case_id FK
        test_status_enum old_status
        test_status_enum new_status
        timestamp changed_at
        varchar(100) changed_by
        text reason
    }
```

## 2. ENUM定義

```mermaid
classDiagram
    class priority_enum {
        <<enumeration>>
        Critical
        High
        Medium
        Low
    }

    class suite_status_enum {
        <<enumeration>>
        準備中
        実行中
        完了
        中断
    }

    class test_status_enum {
        <<enumeration>>
        作成
        テスト
        修正
        レビュー待ち
        レビュー中
        完了
        再テスト
    }
```

## 3. インデックス定義

```mermaid
%%{init: {'theme': 'dark'}}%%
classDiagram
    class PrimaryKeys {
        effort_records_pkey (id)
        status_history_pkey (id)
        test_cases_pkey (id)
        test_groups_pkey (id)
        test_suites_pkey (id)
    }
    
    class SecondaryIndexes {
        idx_effort_records_date (record_date)
        idx_test_cases_priority (priority)
        idx_test_cases_status (status)
        idx_test_groups_order (display_order, suite_id)
    }
```

## 4. 制約一覧

### 4.1 CHECK制約
- effort_records: `check_positive_effort (effort_amount > 0)`

### 4.2 トリガー
1. update_test_cases_updated_at
   - テーブル: test_cases
   - イベント: UPDATE
   - タイミング: BEFORE

2. update_test_groups_updated_at
   - テーブル: test_groups
   - イベント: UPDATE
   - タイミング: BEFORE

3. update_test_suites_updated_at
   - テーブル: test_suites
   - イベント: UPDATE
   - タイミング: BEFORE

### 4.3 外部キー制約
1. test_groups.suite_id → test_suites.id
2. test_cases.group_id → test_groups.id
3. effort_records.test_case_id → test_cases.id
4. status_history.test_case_id → test_cases.id

## 5. シーケンス
1. effort_records_id_seq
   - 開始値: 1
   - 最小値: 1
   - 最大値: 2147483647
   - 増分: 1

2. status_history_id_seq
   - 開始値: 1
   - 最小値: 1
   - 最大値: 2147483647
   - 増分: 1

## 6. 残作業項目

### 6.1 確認が必要な項目
- [ ] パーティショニング戦略の必要性検討
- [ ] バックアップ・リストア手順の確認
- [ ] パフォーマンスチューニングの必要性評価
- [ ] 監視・メンテナンス計画の策定
- [ ] データアーカイブ戦略の決定

### 6.2 推奨アクション
1. 短期的なアクション
   - [ ] 各テーブルのサンプルデータ作成
   - [ ] 主要クエリのパフォーマンステスト
   - [ ] バックアップ・リストア手順の文書化

2. 中長期的な検討項目
   - [ ] データアーカイブ戦略の策定
   - [ ] パーティショニング戦略の検討
   - [ ] 監視・アラート体制の構築
