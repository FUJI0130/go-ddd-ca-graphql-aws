#!/bin/bash
echo "=== AWS環境診断開始 ==="

# 1. Terraform状態確認
cd deployments/terraform/environments/development
echo "--- Terraform State ---"
terraform show -json | jq '.values.outputs' 2>/dev/null || echo "Terraform出力取得失敗"

# 2. 個別出力確認
echo "--- 個別出力値確認 ---"
echo "RDS Host: $(terraform output -raw db_instance_address 2>/dev/null || echo 'EMPTY')"
echo "DB Name: $(terraform output -raw db_name 2>/dev/null || echo 'EMPTY')"
echo "VPC ID: $(terraform output -raw vpc_id 2>/dev/null || echo 'EMPTY')"
echo "GraphQL ALB: $(terraform output -raw graphql_alb_dns_name 2>/dev/null || echo 'EMPTY')"

# 3. Terraformモジュール状態確認
echo "--- モジュール状態 ---"
terraform state list | grep -E "(database|graphql|networking)" | head -10

cd ../../../  # プロジェクトルートに戻る