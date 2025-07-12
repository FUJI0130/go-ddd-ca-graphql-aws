# テストケース管理システム - メインMakefile
# 各機能別Makefileをインクルード

# 基本変数とコマンド
include makefiles/base.mk

# テスト関連
include makefiles/test.mk

# データベース関連
include makefiles/db.mk

# Docker関連
include makefiles/docker.mk

# AWS/Terraform関連
include makefiles/aws.mk

# Terraform関連
include makefiles/terraform.mk

# 検証関連
include makefiles/verification.mk

# インストラクション管理システム統合
include makefiles/instructions.mk

# 統合テスト
include makefiles/integration.mk

include makefiles/update-image-only.mk