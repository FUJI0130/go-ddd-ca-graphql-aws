# インストラクション生成用Makefile
# makefiles/instructions.mk

# 変数定義
INST_DIR = ./docs/claude/instructions
BASE = $(INST_DIR)/base/core.md
MOD_DIR = $(INST_DIR)/modules
OUT_DIR = $(INST_DIR)/combined

# ディレクトリ作成
$(OUT_DIR):
	mkdir -p $(OUT_DIR)

# 基本インストラクション
base: $(BASE) $(OUT_DIR)
	cp $(BASE) $(OUT_DIR)/base_instruction.md
	@echo "基本インストラクションを生成しました: $(OUT_DIR)/base_instruction.md"

# AWS作業用
aws: $(BASE) $(MOD_DIR)/aws.md $(OUT_DIR)
	@echo "" > $(OUT_DIR)/aws_instruction.md
	cat $(BASE) >> $(OUT_DIR)/aws_instruction.md
	@echo "" >> $(OUT_DIR)/aws_instruction.md
	@echo "---" >> $(OUT_DIR)/aws_instruction.md
	@echo "" >> $(OUT_DIR)/aws_instruction.md
	cat $(MOD_DIR)/aws.md >> $(OUT_DIR)/aws_instruction.md
	@echo "AWS作業用インストラクションを生成しました: $(OUT_DIR)/aws_instruction.md"

# バックエンド開発用
backend: $(BASE) $(MOD_DIR)/backend.md $(OUT_DIR)
	@echo "" > $(OUT_DIR)/backend_instruction.md
	cat $(BASE) >> $(OUT_DIR)/backend_instruction.md
	@echo "" >> $(OUT_DIR)/backend_instruction.md
	@echo "---" >> $(OUT_DIR)/backend_instruction.md
	@echo "" >> $(OUT_DIR)/backend_instruction.md
	cat $(MOD_DIR)/backend.md >> $(OUT_DIR)/backend_instruction.md
	@echo "バックエンド開発用インストラクションを生成しました: $(OUT_DIR)/backend_instruction.md"

# フロントエンド開発用
frontend: $(BASE) $(MOD_DIR)/frontend-md.md $(OUT_DIR)
	@echo "" > $(OUT_DIR)/frontend_instruction.md
	cat $(BASE) >> $(OUT_DIR)/frontend_instruction.md
	@echo "" >> $(OUT_DIR)/frontend_instruction.md
	@echo "---" >> $(OUT_DIR)/frontend_instruction.md
	@echo "" >> $(OUT_DIR)/frontend_instruction.md
	cat $(MOD_DIR)/frontend-md.md >> $(OUT_DIR)/frontend_instruction.md
	@echo "フロントエンド開発用インストラクションを生成しました: $(OUT_DIR)/frontend_instruction.md"

# 問題解決用
problem: $(BASE) $(MOD_DIR)/problem-solving-md.md $(OUT_DIR)
	@echo "" > $(OUT_DIR)/problem_solving_instruction.md
	cat $(BASE) >> $(OUT_DIR)/problem_solving_instruction.md
	@echo "" >> $(OUT_DIR)/problem_solving_instruction.md
	@echo "---" >> $(OUT_DIR)/problem_solving_instruction.md
	@echo "" >> $(OUT_DIR)/problem_solving_instruction.md
	cat $(MOD_DIR)/problem-solving-md.md >> $(OUT_DIR)/problem_solving_instruction.md
	@echo "問題解決用インストラクションを生成しました: $(OUT_DIR)/problem_solving_instruction.md"

# AWS開発問題解決向け（複合）
aws_problem: $(BASE) $(MOD_DIR)/aws.md $(MOD_DIR)/problem-solving-md.md $(OUT_DIR)
	@echo "" > $(OUT_DIR)/aws_problem_instruction.md
	cat $(BASE) >> $(OUT_DIR)/aws_problem_instruction.md
	@echo "" >> $(OUT_DIR)/aws_problem_instruction.md
	@echo "---" >> $(OUT_DIR)/aws_problem_instruction.md
	@echo "" >> $(OUT_DIR)/aws_problem_instruction.md
	cat $(MOD_DIR)/aws.md >> $(OUT_DIR)/aws_problem_instruction.md
	@echo "" >> $(OUT_DIR)/aws_problem_instruction.md
	@echo "---" >> $(OUT_DIR)/aws_problem_instruction.md
	@echo "" >> $(OUT_DIR)/aws_problem_instruction.md
	cat $(MOD_DIR)/problem-solving-md.md >> $(OUT_DIR)/aws_problem_instruction.md
	@echo "AWS開発問題解決用インストラクションを生成しました: $(OUT_DIR)/aws_problem_instruction.md"

# バックエンド開発問題解決向け（複合）
backend_problem: $(BASE) $(MOD_DIR)/backend.md $(MOD_DIR)/problem-solving-md.md $(OUT_DIR)
	@echo "" > $(OUT_DIR)/backend_problem_instruction.md
	cat $(BASE) >> $(OUT_DIR)/backend_problem_instruction.md
	@echo "" >> $(OUT_DIR)/backend_problem_instruction.md
	@echo "---" >> $(OUT_DIR)/backend_problem_instruction.md
	@echo "" >> $(OUT_DIR)/backend_problem_instruction.md
	cat $(MOD_DIR)/backend.md >> $(OUT_DIR)/backend_problem_instruction.md
	@echo "" >> $(OUT_DIR)/backend_problem_instruction.md
	@echo "---" >> $(OUT_DIR)/backend_problem_instruction.md
	@echo "" >> $(OUT_DIR)/backend_problem_instruction.md
	cat $(MOD_DIR)/problem-solving-md.md >> $(OUT_DIR)/backend_problem_instruction.md
	@echo "バックエンド開発問題解決用インストラクションを生成しました: $(OUT_DIR)/backend_problem_instruction.md"

# フロントエンド開発問題解決向け（複合）
frontend_problem: $(BASE) $(MOD_DIR)/frontend-md.md $(MOD_DIR)/problem-solving-md.md $(OUT_DIR)
	@echo "" > $(OUT_DIR)/frontend_problem_instruction.md
	cat $(BASE) >> $(OUT_DIR)/frontend_problem_instruction.md
	@echo "" >> $(OUT_DIR)/frontend_problem_instruction.md
	@echo "---" >> $(OUT_DIR)/frontend_problem_instruction.md
	@echo "" >> $(OUT_DIR)/frontend_problem_instruction.md
	cat $(MOD_DIR)/frontend-md.md >> $(OUT_DIR)/frontend_problem_instruction.md
	@echo "" >> $(OUT_DIR)/frontend_problem_instruction.md
	@echo "---" >> $(OUT_DIR)/frontend_problem_instruction.md
	@echo "" >> $(OUT_DIR)/frontend_problem_instruction.md
	cat $(MOD_DIR)/problem-solving-md.md >> $(OUT_DIR)/frontend_problem_instruction.md
	@echo "フロントエンド開発問題解決用インストラクションを生成しました: $(OUT_DIR)/frontend_problem_instruction.md"

# カスタム組み合わせ (例: make custom MODULES="aws backend")
custom: $(BASE) $(OUT_DIR)
	@echo "" > $(OUT_DIR)/custom_instruction.md
	cat $(BASE) >> $(OUT_DIR)/custom_instruction.md
	@for module in $(MODULES); do \
		echo "" >> $(OUT_DIR)/custom_instruction.md; \
		echo "---" >> $(OUT_DIR)/custom_instruction.md; \
		echo "" >> $(OUT_DIR)/custom_instruction.md; \
		if [ -f "$(MOD_DIR)/$$module.md" ]; then \
			cat $(MOD_DIR)/$$module.md >> $(OUT_DIR)/custom_instruction.md; \
		elif [ -f "$(MOD_DIR)/$$module-md.md" ]; then \
			cat $(MOD_DIR)/$$module-md.md >> $(OUT_DIR)/custom_instruction.md; \
		else \
			echo "モジュール $$module が見つかりません"; \
			exit 1; \
		fi; \
	done
	@echo "カスタムインストラクションを生成しました: $(OUT_DIR)/custom_instruction.md"

# すべての事前定義モジュールを生成
all: aws backend frontend problem aws_problem backend_problem frontend_problem

# クリーンアップ
clean:
	rm -rf $(OUT_DIR)/*

.PHONY: base aws backend frontend problem aws_problem backend_problem frontend_problem custom all clean