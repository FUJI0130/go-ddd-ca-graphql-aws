# VSCode Remote Development with Go - Setup and Best Practices

## 目次
- [概要](#概要)
- [アーキテクチャ解説](#アーキテクチャ解説)
- [セットアップ手順](#セットアップ手順)
- [必要な拡張機能](#必要な拡張機能)
- [トラブルシューティング](#トラブルシューティング)
- [ベストプラクティス](#ベストプラクティス)

## 概要

VSCodeを使用してリモートマシン（例：Raspberry Pi）上でのGo開発を行う場合の設定と注意点をまとめています。
特にSSH経由での開発における重要なポイントと、開発環境の正しいセットアップ方法を解説します。

## アーキテクチャ解説

### クライアント・サーバーモデル

VSCode Remote SSHは以下の2つの主要コンポーネントで構成されています：

1. **クライアント側（ローカルPC）**
   - VSCodeのUIインターフェース
   - ファイルエクスプローラー
   - エディタ表示
   - キーボード入力の処理

2. **サーバー側（リモートマシン）**
   - VSCode Server
   - 言語サービス（Go言語の場合）
   - 拡張機能
   - 実行環境

### 処理の流れ

1. エディタでのコード編集 → クライアント側で処理
2. コードの解析・補完 → サーバー側で処理
3. テストの実行 → サーバー側で処理
4. デバッグ実行 → サーバー側で処理

## セットアップ手順

### 1. リモートマシン（サーバー側）の準備

```bash
# Goのインストール
sudo apt-get update
sudo apt-get install golang-go

# GOPATHの設定
echo 'export GOPATH=$HOME/go' >> ~/.bashrc
echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.bashrc
source ~/.bashrc

# 開発ツールのインストール
go install golang.org/x/tools/gopls@latest
go install golang.org/x/tools/cmd/godoc@latest
go install github.com/go-delve/delve/cmd/dlv@latest
```

### 2. VSCodeの設定

1. ローカルPCにVSCodeをインストール
2. 以下の拡張機能をインストール：
   - Remote - SSH
   - Go

### 3. リモート接続の設定

1. VSCodeで`F1`キーを押してコマンドパレットを開く
2. `Remote-SSH: Connect to Host...`を選択
3. SSHの接続情報を入力

## 必要な拡張機能

### リモートマシン側（重要）
- Go
- Go Test Explorer
- Go Outline

これらの拡張機能がリモートマシン側にインストールされていないと、以下の機能が制限されます：
- テストの実行
- デバッグ
- コード補完
- シンボル参照

### ローカルマシン側
- Remote - SSH（必須）
- Remote - SSH: Editing Configuration Files（推奨）

## トラブルシューティング

### よくある問題と解決方法

1. **テストが実行できない**
   - リモート側のGo拡張機能の確認
   - `go test`コマンドの動作確認
   - GOPATHの設定確認

2. **デバッグができない**
   - delveのインストール確認
   - ポート開放確認
   - launch.jsonの設定確認

3. **コード補完が効かない**
   - goplsの動作確認
   - GOPATHにプロジェクトが含まれているか確認

## ベストプラクティス

### プロジェクト構成

```
project/
├── .vscode/
│   ├── launch.json
│   └── settings.json
├── cmd/
├── pkg/
└── internal/
```

### settings.jsonの推奨設定

```json
{
  "go.useLanguageServer": true,
  "go.testOnSave": true,
  "go.coverOnSave": true,
  "go.buildOnSave": "workspace"
}
```

### セキュリティ考慮事項

1. SSH鍵の適切な管理
2. リモートマシンのファイアウォール設定
3. 適切なパーミッション設定

### パフォーマンス最適化

1. **ネットワーク遅延の最小化**
   - 安定したネットワーク接続の確保
   - 必要なファイルのみの同期

2. **リソース使用の最適化**
   - 不要な拡張機能の無効化
   - ワークスペースのインデックス範囲の適切な設定

## 補足情報

### なぜリモート側に拡張機能が必要か

1. **実行環境の一貫性**
   - すべてのツールチェーンがコードと同じ環境に存在
   - 環境依存の問題を防止

2. **パフォーマンスの最適化**
   - ファイルアクセスの局所性
   - ネットワーク遅延の最小化

3. **デバッグの正確性**
   - 実際の実行環境でのデバッグ
   - 環境の差異による問題の防止

### リモート開発のメリット

1. 一貫した開発環境の維持
2. リソースの効率的な使用
3. チーム間での環境の統一
4. 本番環境に近い開発環境の実現

## 参考リンク

- [VSCode Remote Development](https://code.visualstudio.com/docs/remote/remote-overview)
- [Go Tools](https://golang.org/doc/editors.html)
- [Delve Debugger](https://github.com/go-delve/delve)
