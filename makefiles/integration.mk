# インストラクション管理システム統合Makefile
# makefiles/integration.mk

#----------------------------------------
# インストラクション管理システム統合
#----------------------------------------
.PHONY: instructions-aws instructions-backend instructions-frontend instructions-problem
.PHONY: instructions-aws-problem instructions-backend-problem instructions-frontend-problem
.PHONY: instructions-custom instructions-all instructions-clean

# 基本インストラクション生成コマンド
instructions-aws:
	$(MAKE) -f makefiles/instructions.mk aws

instructions-backend:
	$(MAKE) -f makefiles/instructions.mk backend

instructions-frontend:
	$(MAKE) -f makefiles/instructions.mk frontend

instructions-problem:
	$(MAKE) -f makefiles/instructions.mk problem

# 複合インストラクション生成コマンド
instructions-aws-problem:
	$(MAKE) -f makefiles/instructions.mk aws_problem

instructions-backend-problem:
	$(MAKE) -f makefiles/instructions.mk backend_problem

instructions-frontend-problem:
	$(MAKE) -f makefiles/instructions.mk frontend_problem

# カスタムインストラクション生成
instructions-custom:
	@if [ -z "$(MODULES)" ]; then \
		echo -e "${RED}エラー: MODULES環境変数を指定してください${NC}"; \
		echo "例: MODULES=\"aws backend\" make instructions-custom"; \
		exit 1; \
	fi
	$(MAKE) -f makefiles/instructions.mk custom MODULES="$(MODULES)"

# すべてのインストラクション生成
instructions-all:
	$(MAKE) -f makefiles/instructions.mk all

# インストラクションのクリーンアップ
instructions-clean:
	$(MAKE) -f makefiles/instructions.mk clean