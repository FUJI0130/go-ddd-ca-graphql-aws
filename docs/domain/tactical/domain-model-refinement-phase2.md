# テストケース管理システム ドメインモデル詳細化記録

## 格納場所
`docs/domain/tactical/domain-model-refinement-phase2.md`

## 1. 検討の流れ

### フェーズ1：状態管理の詳細化
1. テストケースの状態遷移フローを定義
   - 作成 → テスト → 修正 → レビュー待ち → レビュー中 → 完了
   - 各状態での進捗率を設定（0%, 25%, 50%, 75%, 100%）
2. 再テスト時の扱いを決定
   - 完了から再テストになった場合は50%に戻す
   - 遅延表示：「再テスト中（○日遅延）」形式

### フェーズ2：権限と工数管理
1. 権限管理の方針を決定
   - テスト中は実行者のみ編集可能
   - レビュー中はレビュワーがコメント・軽微な修正が可能
2. 工数管理の仕組みを設計
   - 日単位での記録
   - 追加工数は都度更新
   - コメントは任意（スイート単位で必須/任意を切り替え可能）

### フェーズ3：グループ管理と進捗計算
1. グループ化の方針を決定
   - 単一階層のシンプルな構造
   - グループ間の移動を可能に
2. 進捗計算ロジックを確立
   - テストケース：状態による進捗率 × 重要度の重み
   - グループ：所属テストケースの単純平均
   - 表示：進捗率と完了件数/全件数を併記

## 2. 主要コンポーネントの設計

### TestSuite（テストスイート）
```
- ID
- Name
- Description
- Status
- RequireEffortComment: bool
- CalculateOverallProgress()
- ReorderGroups()
```

### TestGroup（テストグループ）
```
- ID
- Name
- Description
- DisplayOrder: int
- Status
- CalculateProgress()
- GetProgressSummary()
- UpdateDisplayOrder()
```

### TestCase（テストケース）
```
- ID
- Title
- Description
- GroupID
- Status
- Priority
- PlannedEffort: float
- ActualEffort: float
- IsDelayed: bool
- DelayDays: int
- CalculateProgress()
- MoveToGroup()
```

### 進捗率の定義
```
作成: 0%
テスト/修正: 25%
レビュー待ち: 50%
レビュー中: 75%
完了: 100%
再テスト: 50%
```

### Priority（重要度）
```
Critical: 4.0
High: 3.0
Medium: 2.0
Low: 1.0
```

## 3. 実装上の重要ポイント

### 進捗計算
1. テストケースレベル
   - 状態による基本進捗率
   - 重要度による重み付け

2. グループレベル
   - 所属テストケースの単純平均
   - 完了件数/全件数の表示

### 表示順序管理
1. DisplayOrderによる単純な番号管理
2. 順序変更時の自動調整機能
3. 重複チェック

### 工数管理
1. 日単位での記録
2. 追加工数の都度更新
3. コメントの必須/任意をスイート単位で設定

## 4. 次のステップ候補
1. ガントチャートとの連携部分の詳細設計
2. 権限管理の具体的な実装方式の検討
3. レポート機能の詳細設計

## 5. 未決定事項
1. アーカイブデータの保持期間
2. 詳細な権限設定の範囲
3. レポート出力の形式

## 6. 関連ドキュメント
- コンテキストマップ: `docs/domain/strategic/context-map.md`
- 初期ドメインモデル: `docs/domain/tactical/domain-model.md`
- プロジェクト概要: `docs/_meta/progress-summary.md`
