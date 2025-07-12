# AWS手動クリーンアップ手順と参考コマンド

このドキュメントは、自動クリーンアップスクリプトが機能しない場合に備えた、AWS環境の手動クリーンアップ手順をまとめたものです。前のスレッドで実際に実行したコマンドをベースにしています。

## 1. リソース削除の基本原則と順序

AWS環境のリソースは依存関係があるため、正しい順序で削除する必要があります：

1. ECSサービスとタスク
2. ロードバランサーとターゲットグループ
3. RDSインスタンスとセキュリティグループ
4. ルートテーブル関連付け
5. ルートテーブル
6. Internet Gateway（デタッチしてから削除）
7. サブネット
8. VPC
9. NAT Gateway
10. Elastic IP

## 2. 手動クリーンアップの具体的な手順

### 2.1 ECSサービスの削除

```bash
# クラスターの確認
aws ecs list-clusters

# サービスの一覧表示
aws ecs list-services --cluster development-shared-cluster

# サービスの更新（タスク数を0に設定）
aws ecs update-service --cluster development-shared-cluster --service development-api --desired-count 0

# サービスの削除
aws ecs delete-service --cluster development-shared-cluster --service development-api --force
```

### 2.2 ロードバランサーとターゲットグループの削除

```bash
# ロードバランサーの確認
aws elbv2 describe-load-balancers --query "LoadBalancers[?contains(LoadBalancerName,'development')].{Name:LoadBalancerName,ARN:LoadBalancerArn}"

# リスナーの確認
aws elbv2 describe-listeners --load-balancer-arn <LOAD_BALANCER_ARN> --query "Listeners[].{ARN:ListenerArn,Port:Port,Protocol:Protocol}"

# リスナーの削除（リスナーごとに実行）
aws elbv2 delete-listener --listener-arn <LISTENER_ARN>

# ロードバランサーの削除
aws elbv2 delete-load-balancer --load-balancer-arn <LOAD_BALANCER_ARN>

# ターゲットグループの確認
aws elbv2 describe-target-groups --query "TargetGroups[?contains(TargetGroupName,'development')].{Name:TargetGroupName,ARN:TargetGroupArn}"

# ターゲットグループの削除
aws elbv2 delete-target-group --target-group-arn <TARGET_GROUP_ARN>
```

### 2.3 RDSインスタンスの削除

```bash
# RDSインスタンスの確認
aws rds describe-db-instances --query "DBInstances[?DBInstanceIdentifier=='development-postgres'].{ID:DBInstanceIdentifier,Status:DBInstanceStatus}"

# RDSインスタンスの削除（最終スナップショットなし）
aws rds delete-db-instance --db-instance-identifier development-postgres --skip-final-snapshot

# 削除が完了するまで待機
aws rds wait db-instance-deleted --db-instance-identifier development-postgres

# RDS関連のセキュリティグループの確認
aws ec2 describe-security-groups --filters "Name=group-name,Values=*development*rds*,*development*postgres*" --query "SecurityGroups[].{ID:GroupId,Name:GroupName}"

# セキュリティグループの削除
aws ec2 delete-security-group --group-id <SECURITY_GROUP_ID>
```

### 2.4 VPCとネットワークリソースの削除

以下は前のスレッドで実際に実行したVPC削除のコマンド履歴です。2つのVPCを削除する例です：

#### VPC 1 (vpc-06456a8edcefd3f4e)の削除

```bash
# ルートテーブル関連付けの確認
aws ec2 describe-route-tables --filters "Name=vpc-id,Values=vpc-06456a8edcefd3f4e" --query "RouteTables[].Associations[?SubnetId!=null].{ID:RouteTableAssociationId,SubnetId:SubnetId}"

# ルートテーブル関連付けの解除
aws ec2 disassociate-route-table --association-id rtbassoc-0eba1589758ffee0e
aws ec2 disassociate-route-table --association-id rtbassoc-02b19a75886c7c854
aws ec2 disassociate-route-table --association-id rtbassoc-0a7c9c118c41481a2
aws ec2 disassociate-route-table --association-id rtbassoc-0bbcd318b15400c4f

# ルートテーブルの削除
aws ec2 delete-route-table --route-table-id rtb-0a8f1d832745236a7
aws ec2 delete-route-table --route-table-id rtb-03c699b88957c861a

# Internet Gatewayのデタッチと削除
aws ec2 detach-internet-gateway --internet-gateway-id igw-0617ca95de5f511ba --vpc-id vpc-06456a8edcefd3f4e
aws ec2 delete-internet-gateway --internet-gateway-id igw-0617ca95de5f511ba

# サブネットの削除
aws ec2 delete-subnet --subnet-id subnet-01e6e4168c26c5bf8
aws ec2 delete-subnet --subnet-id subnet-0287e7b1ef05d5d47
aws ec2 delete-subnet --subnet-id subnet-0b62202efb3d3e053
aws ec2 delete-subnet --subnet-id subnet-014d3a2381e669beb

# VPCの削除
aws ec2 delete-vpc --vpc-id vpc-06456a8edcefd3f4e

# 削除確認
aws ec2 describe-vpcs --vpc-ids vpc-06456a8edcefd3f4e || echo "VPC successfully deleted"
```

#### VPC 2 (vpc-0404db4d1543ae2f7)の削除

```bash
# ルートテーブル関連付けの確認
aws ec2 describe-route-tables --filters "Name=vpc-id,Values=vpc-0404db4d1543ae2f7" --query "RouteTables[].Associations[?SubnetId!=null].{ID:RouteTableAssociationId,SubnetId:SubnetId}"

# ルートテーブル関連付けの解除
aws ec2 disassociate-route-table --association-id rtbassoc-00d5d8e38ad568e2c
aws ec2 disassociate-route-table --association-id rtbassoc-0afce4be25cc5c915
aws ec2 disassociate-route-table --association-id rtbassoc-02b2c3d79ff46370b
aws ec2 disassociate-route-table --association-id rtbassoc-021d14172769d3647

# ルートテーブルの削除
aws ec2 delete-route-table --route-table-id rtb-026a6d086e1d3ccc1
aws ec2 delete-route-table --route-table-id rtb-0e3b254c250450a08

# Internet Gatewayのデタッチと削除
aws ec2 detach-internet-gateway --internet-gateway-id igw-0eed145a807804c9c --vpc-id vpc-0404db4d1543ae2f7
aws ec2 delete-internet-gateway --internet-gateway-id igw-0eed145a807804c9c

# サブネットの削除
aws ec2 delete-subnet --subnet-id subnet-0ebcab731371ab6a5
aws ec2 delete-subnet --subnet-id subnet-0d87ec2a7a555cd80
aws ec2 delete-subnet --subnet-id subnet-0ed6be24a04714532
aws ec2 delete-subnet --subnet-id subnet-0822cc71b3ee9120d

# VPCの削除
aws ec2 delete-vpc --vpc-id vpc-0404db4d1543ae2f7

# 削除確認
aws ec2 describe-vpcs --vpc-ids vpc-0404db4d1543ae2f7 || echo "VPC successfully deleted"
```

### 2.5 残りのVPCの確認

```bash
# カスタムVPCの確認
aws ec2 describe-vpcs --query "Vpcs[?IsDefault==\`false\`].{ID:VpcId,CIDR:CidrBlock,Name:Tags[?Key=='Name'].Value|[0],Environment:Tags[?Key=='Environment'].Value|[0]}"

# Internet Gatewayの確認
aws ec2 describe-internet-gateways --query "InternetGateways[].{ID:InternetGatewayId,Attachments:Attachments}"
```

## 3. リソースID取得のためのクエリコマンド

手動削除の際に必要なリソースIDを取得するためのコマンド一覧です：

### 3.1 ECSリソース

```bash
# クラスター一覧
aws ecs list-clusters

# 特定のクラスターのサービス一覧
aws ecs list-services --cluster <CLUSTER_NAME>

# 特定のクラスターのタスク一覧
aws ecs list-tasks --cluster <CLUSTER_NAME>
```

### 3.2 ロードバランサーとターゲットグループ

```bash
# ロードバランサー一覧
aws elbv2 describe-load-balancers --query "LoadBalancers[?contains(LoadBalancerName,'development')].{Name:LoadBalancerName,ARN:LoadBalancerArn}"

# ターゲットグループ一覧
aws elbv2 describe-target-groups --query "TargetGroups[?contains(TargetGroupName,'development')].{Name:TargetGroupName,ARN:TargetGroupArn}"
```

### 3.3 RDSリソース

```bash
# RDSインスタンス一覧
aws rds describe-db-instances --query "DBInstances[].{ID:DBInstanceIdentifier,Status:DBInstanceStatus}"
```

### 3.4 VPCとネットワークリソース

```bash
# VPC一覧
aws ec2 describe-vpcs --query "Vpcs[?IsDefault==\`false\`].{ID:VpcId,CIDR:CidrBlock,Name:Tags[?Key=='Name'].Value|[0]}"

# 特定VPCのサブネット
aws ec2 describe-subnets --filters "Name=vpc-id,Values=<VPC_ID>" --query "Subnets[].{ID:SubnetId,CIDR:CidrBlock,Name:Tags[?Key=='Name'].Value|[0]}"

# 特定VPCのルートテーブル
aws ec2 describe-route-tables --filters "Name=vpc-id,Values=<VPC_ID>" --query "RouteTables[].{ID:RouteTableId,Main:Associations[0].Main}"

# 特定VPCのルートテーブル関連付け
aws ec2 describe-route-tables --filters "Name=vpc-id,Values=<VPC_ID>" --query "RouteTables[].Associations[?SubnetId!=null].{ID:RouteTableAssociationId,SubnetId:SubnetId}"

# Internet Gateway
aws ec2 describe-internet-gateways --query "InternetGateways[].{ID:InternetGatewayId,VPC:Attachments[0].VpcId}"

# NAT Gateway
aws ec2 describe-nat-gateways --filter "Name=state,Values=available" --query "NatGateways[].{ID:NatGatewayId,SubnetId:SubnetId}"

# Elastic IP
aws ec2 describe-addresses --query "Addresses[].{ID:AllocationId,IP:PublicIp}"
```

### 3.5 セキュリティグループ

```bash
# 特定のパターンに一致するセキュリティグループ
aws ec2 describe-security-groups --filters "Name=group-name,Values=*development*" --query "SecurityGroups[].{ID:GroupId,Name:GroupName,VPC:VpcId}"
```

## 4. エラー対処法

### 4.1 依存リソースの問題

VPCやサブネットなどを削除しようとして「依存リソースがある」エラーが出た場合：

```bash
# VPCの依存リソースの確認
aws ec2 describe-network-interfaces --filters "Name=vpc-id,Values=<VPC_ID>" --query "NetworkInterfaces[].{ID:NetworkInterfaceId,Description:Description}"

# サブネットの依存リソースの確認
aws ec2 describe-network-interfaces --filters "Name=subnet-id,Values=<SUBNET_ID>" --query "NetworkInterfaces[].{ID:NetworkInterfaceId,Description:Description}"
```

### 4.2 ロードバランサー削除エラー

ロードバランサー削除時に「Listener exists」エラーが出た場合：

```bash
# リスナーの確認
aws elbv2 describe-listeners --load-balancer-arn <LOAD_BALANCER_ARN> --query "Listeners[].ListenerArn"

# すべてのリスナーを削除
for arn in $(aws elbv2 describe-listeners --load-balancer-arn <LOAD_BALANCER_ARN> --query "Listeners[].ListenerArn" --output text); do
  aws elbv2 delete-listener --listener-arn $arn
done
```

### 4.3 セキュリティグループ削除エラー

セキュリティグループ削除時に依存関係エラーが出た場合：

```bash
# セキュリティグループの依存関係確認
aws ec2 describe-network-interfaces --filters "Name=group-id,Values=<SECURITY_GROUP_ID>" --query "NetworkInterfaces[].{ID:NetworkInterfaceId,Description:Description}"
```

## 5. 便利なシェル関数

繰り返し使用するコマンドを関数化すると作業が効率的になります：

```bash
# VPCとその関連リソースを一括削除する関数
delete_vpc() {
  VPC_ID=$1
  echo "Deleting VPC $VPC_ID and all its resources..."
  
  # ルートテーブル関連付けの解除
  for assoc in $(aws ec2 describe-route-tables --filters "Name=vpc-id,Values=$VPC_ID" --query "RouteTables[].Associations[?SubnetId!=null].RouteTableAssociationId" --output text); do
    echo "Disassociating route table association $assoc"
    aws ec2 disassociate-route-table --association-id $assoc
  done
  
  # ルートテーブルの削除（メインではないもの）
  for rt in $(aws ec2 describe-route-tables --filters "Name=vpc-id,Values=$VPC_ID" --query "RouteTables[?Associations[0].Main!=\`true\`].RouteTableId" --output text); do
    echo "Deleting route table $rt"
    aws ec2 delete-route-table --route-table-id $rt
  done
  
  # Internet Gatewayのデタッチと削除
  for igw in $(aws ec2 describe-internet-gateways --filters "Name=attachment.vpc-id,Values=$VPC_ID" --query "InternetGateways[].InternetGatewayId" --output text); do
    echo "Detaching and deleting internet gateway $igw"
    aws ec2 detach-internet-gateway --internet-gateway-id $igw --vpc-id $VPC_ID
    aws ec2 delete-internet-gateway --internet-gateway-id $igw
  done
  
  # サブネットの削除
  for subnet in $(aws ec2 describe-subnets --filters "Name=vpc-id,Values=$VPC_ID" --query "Subnets[].SubnetId" --output text); do
    echo "Deleting subnet $subnet"
    aws ec2 delete-subnet --subnet-id $subnet
  done
  
  # セキュリティグループの削除（デフォルト以外）
  for sg in $(aws ec2 describe-security-groups --filters "Name=vpc-id,Values=$VPC_ID" --query "SecurityGroups[?GroupName!=\`default\`].GroupId" --output text); do
    echo "Deleting security group $sg"
    aws ec2 delete-security-group --group-id $sg
  done
  
  # VPCの削除
  echo "Deleting VPC $VPC_ID"
  aws ec2 delete-vpc --vpc-id $VPC_ID
  
  echo "VPC deletion completed. Verifying..."
  aws ec2 describe-vpcs --vpc-ids $VPC_ID 2>/dev/null || echo "VPC $VPC_ID successfully deleted"
}
```

## 6. まとめと教訓

### 6.1 手動削除の主な手順

1. ECSサービスのタスク数を0にしてから削除
2. ロードバランサーのリスナーを先に削除してからロードバランサー本体を削除
3. ターゲットグループの削除
4. RDSインスタンスの削除（waitコマンドで完了を待機）
5. VPC関連リソースの正しい順序での削除：
   - ルートテーブル関連付け解除
   - ルートテーブル削除
   - Internet Gatewayデタッチと削除
   - サブネット削除
   - セキュリティグループ削除
   - VPC削除

### 6.2 教訓

1. AWSリソース間の依存関係を理解することが重要
2. 正しい削除順序を守らないとエラーが発生する
3. 削除前にリソースの状態や依存関係を確認する習慣をつける
4. エラーメッセージを注意深く読み、依存リソースを特定する
5. 大規模な削除作業は自動化スクリプトを検討する

### 6.3 自動クリーンアップとの関係

このドキュメントは、`scripts/terraform/aws-cleanup.sh`で実装された自動クリーンアップシステムの基盤となる手順をまとめたものです。自動クリーンアップスクリプトが何らかの理由で機能しない場合は、このドキュメントの手順に従って手動でリソースを削除することができます。

また、新しいタイプのリソースやエラーケースに対応するために自動クリーンアップスクリプトを拡張する際の参考資料としても活用できます。