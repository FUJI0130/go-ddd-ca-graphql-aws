#!/bin/bash
# 1. NAT Gatewayの確認と削除（コスト発生の主な原因）
NAT_GATEWAYS=$(aws ec2 describe-nat-gateways --filter "Name=vpc-id,Values=vpc-08e6de8485b5f5eed" --query "NatGateways[*].NatGatewayId" --output text)
for NAT in $NAT_GATEWAYS; do
  echo "NAT Gateway $NAT を削除しています..."
  aws ec2 delete-nat-gateway --nat-gateway-id $NAT
done