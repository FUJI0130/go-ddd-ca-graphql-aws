#!/bin/bash
# ===================================================================
# ファイル名: aws-iam-cleaner.sh
# 説明: AWSのIAMリソースをクリーンアップするスクリプト（デバッグ強化版）
#
# 用途:
#  - IAMポリシーのデタッチと削除
#  - IAMロールから特定のポリシーをデタッチ
#  - IAMリソースの依存関係を考慮した削除
#  - 環境タグに基づくリソースのフィルタリング
#  - Terraform destroyで完全に削除されないIAMリソースの手動クリーンアップ
#
# デバッグ機能:
#  - 詳細なログ出力
#  - コマンド実行結果の表示
#  - エラー発生時の継続処理
#  - 環境変数によるデバッグモード制御
#
# 使用方法:
#  ./aws-iam-cleaner.sh <環境名> [auto] [debug]
#  例: ./aws-iam-cleaner.sh development
#      ./aws-iam-cleaner.sh production auto
#      ./aws-iam-cleaner.sh development debug
# ===================================================================

# デバッグモードの設定（set -x で実行時にコマンドを表示）
if [[ "$3" == "debug" || "$2" == "debug" || "$IAM_CLEANER_DEBUG" == "true" ]]; then
  DEBUG_MODE=true
  set -x
else
  DEBUG_MODE=false
fi

# エラー発生時にも継続する（set -e を使用しない）
# set -e

# 色の設定
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 引数の解析
ENV=${1:-development}
# 2番目のパラメータがautoかdebugか判断
if [ "$2" == "auto" ]; then
  AUTO_MODE="auto"
elif [ "$2" == "debug" ]; then
  DEBUG_MODE=true
  AUTO_MODE=""
else
  AUTO_MODE=${2:-""}
fi
ENV_PREFIX=${ENV}  # プレフィックスの設定

# デバッグ関数
debug_log() {
  if [ "$DEBUG_MODE" = true ]; then
    echo -e "${BLUE}[DEBUG] $1${NC}"
  fi
}

# コマンド実行結果をデバッグ表示する関数
debug_cmd() {
  if [ "$DEBUG_MODE" = true ]; then
    echo -e "${BLUE}[DEBUG CMD] $1${NC}"
    eval "$1" 2>&1 | while IFS= read -r line; do echo -e "${BLUE}[DEBUG OUT] $line${NC}"; done
    return ${PIPESTATUS[0]}
  else
    eval "$1"
    return $?
  fi
}

# 自動モードでない場合は確認を求める
if [ "$AUTO_MODE" != "auto" ]; then
  echo -e "${RED}警告: このスクリプトはAWS IAMリソースを削除します。この操作は元に戻せません！${NC}"
  read -p "続行しますか？(y/n) " CONTINUE

  if [ "$CONTINUE" != "y" ]; then
    echo "中止します"
    exit 1
  fi
fi

echo -e "${BLUE}IAMリソースのクリーンアップを開始します（環境: $ENV）...${NC}"
debug_log "環境: $ENV, 自動モード: $AUTO_MODE, デバッグモード: $DEBUG_MODE"

# 環境情報を表示
debug_log "AWS CLI バージョン:"
if [ "$DEBUG_MODE" = true ]; then
  aws --version
fi

# IAMポリシーをロールからデタッチして削除する関数
cleanup_iam_policy() {
  local policy_name=$1
  
  echo -e "${BLUE}IAMポリシー '${policy_name}' を検索しています...${NC}"
  debug_log "ポリシー検索コマンド: aws iam list-policies --query \"Policies[?PolicyName=='${policy_name}'].Arn\" --output text"
  
  # ポリシーARNの取得
  POLICY_ARN=$(aws iam list-policies --query "Policies[?PolicyName=='${policy_name}'].Arn" --output text)
  debug_log "取得したポリシーARN: ${POLICY_ARN}"
  
  if [ -z "$POLICY_ARN" ] || [ "$POLICY_ARN" == "None" ]; then
    echo -e "${GREEN}IAMポリシー '${policy_name}' は見つかりませんでした。${NC}"
    return 0
  fi
  
  echo -e "${YELLOW}IAMポリシー '${policy_name}' (${POLICY_ARN}) が見つかりました。${NC}"
  
  # ポリシーを使用しているエンティティを確認
  echo -e "${BLUE}ポリシーを使用しているエンティティを確認しています...${NC}"
  debug_log "エンティティ確認コマンド: aws iam list-entities-for-policy --policy-arn ${POLICY_ARN}"
  
  # エンティティ情報を取得して表示
  ENTITIES_OUTPUT=$(aws iam list-entities-for-policy --policy-arn ${POLICY_ARN} 2>&1)
  ENTITIES_EXIT_CODE=$?
  
  debug_log "エンティティ取得結果 (終了コード: ${ENTITIES_EXIT_CODE}):"
  debug_log "$ENTITIES_OUTPUT"
  
  # エラーチェック
  if [ $ENTITIES_EXIT_CODE -ne 0 ]; then
    echo -e "${RED}エラー: ポリシーエンティティの取得に失敗しました${NC}"
    echo -e "${RED}エラー詳細: $ENTITIES_OUTPUT${NC}"
    return 1
  fi
  
  # JQを使用せずにAWS CLIの出力から直接情報を抽出
  echo -e "${BLUE}エンティティ情報を解析しています...${NC}"
  
  # ロールからポリシーをデタッチ
  # AWS CLIのテキスト出力からロール名を抽出
  debug_log "ロール情報を抽出します"
  aws iam list-entities-for-policy --policy-arn ${POLICY_ARN} --query "PolicyRoles[].RoleName" --output text > /tmp/roles_list.txt
  
  # 一時ファイルからロール名を読み込み
  if [ -s /tmp/roles_list.txt ]; then
    echo -e "${YELLOW}ポリシーをロールからデタッチします...${NC}"
    while read -r ROLE_NAME; do
      debug_log "ロール名: $ROLE_NAME"
      if [ ! -z "$ROLE_NAME" ]; then
        echo -e "${YELLOW} - ロール '${ROLE_NAME}' からポリシーをデタッチします...${NC}"
        debug_log "デタッチコマンド: aws iam detach-role-policy --role-name ${ROLE_NAME} --policy-arn ${POLICY_ARN}"
        
        DETACH_OUTPUT=$(aws iam detach-role-policy --role-name ${ROLE_NAME} --policy-arn ${POLICY_ARN} 2>&1)
        DETACH_EXIT_CODE=$?
        
        debug_log "デタッチ結果 (終了コード: ${DETACH_EXIT_CODE}):"
        debug_log "$DETACH_OUTPUT"
        
        if [ $DETACH_EXIT_CODE -ne 0 ]; then
          echo -e "${RED}エラー: ロール '${ROLE_NAME}' からのポリシーデタッチに失敗しました${NC}"
          echo -e "${RED}エラー詳細: $DETACH_OUTPUT${NC}"
          # 継続処理のためreturnしない
        else
          echo -e "${GREEN} - ロール '${ROLE_NAME}' からポリシーを正常にデタッチしました${NC}"
        fi
      fi
    done < /tmp/roles_list.txt
  else
    echo -e "${GREEN}ポリシーはどのロールにもアタッチされていません${NC}"
  fi
  
  # ユーザーからポリシーをデタッチ
  debug_log "ユーザー情報を抽出します"
  aws iam list-entities-for-policy --policy-arn ${POLICY_ARN} --query "PolicyUsers[].UserName" --output text > /tmp/users_list.txt
  
  if [ -s /tmp/users_list.txt ]; then
    echo -e "${YELLOW}ポリシーをユーザーからデタッチします...${NC}"
    while read -r USER_NAME; do
      debug_log "ユーザー名: $USER_NAME"
      if [ ! -z "$USER_NAME" ]; then
        echo -e "${YELLOW} - ユーザー '${USER_NAME}' からポリシーをデタッチします...${NC}"
        debug_log "デタッチコマンド: aws iam detach-user-policy --user-name ${USER_NAME} --policy-arn ${POLICY_ARN}"
        
        DETACH_OUTPUT=$(aws iam detach-user-policy --user-name ${USER_NAME} --policy-arn ${POLICY_ARN} 2>&1)
        DETACH_EXIT_CODE=$?
        
        debug_log "デタッチ結果 (終了コード: ${DETACH_EXIT_CODE}):"
        debug_log "$DETACH_OUTPUT"
        
        if [ $DETACH_EXIT_CODE -ne 0 ]; then
          echo -e "${RED}エラー: ユーザー '${USER_NAME}' からのポリシーデタッチに失敗しました${NC}"
          echo -e "${RED}エラー詳細: $DETACH_OUTPUT${NC}"
          # 継続処理のためreturnしない
        else
          echo -e "${GREEN} - ユーザー '${USER_NAME}' からポリシーを正常にデタッチしました${NC}"
        fi
      fi
    done < /tmp/users_list.txt
  else
    echo -e "${GREEN}ポリシーはどのユーザーにもアタッチされていません${NC}"
  fi
  
  # グループからポリシーをデタッチ
  debug_log "グループ情報を抽出します"
  aws iam list-entities-for-policy --policy-arn ${POLICY_ARN} --query "PolicyGroups[].GroupName" --output text > /tmp/groups_list.txt
  
  if [ -s /tmp/groups_list.txt ]; then
    echo -e "${YELLOW}ポリシーをグループからデタッチします...${NC}"
    while read -r GROUP_NAME; do
      debug_log "グループ名: $GROUP_NAME"
      if [ ! -z "$GROUP_NAME" ]; then
        echo -e "${YELLOW} - グループ '${GROUP_NAME}' からポリシーをデタッチします...${NC}"
        debug_log "デタッチコマンド: aws iam detach-group-policy --group-name ${GROUP_NAME} --policy-arn ${POLICY_ARN}"
        
        DETACH_OUTPUT=$(aws iam detach-group-policy --group-name ${GROUP_NAME} --policy-arn ${POLICY_ARN} 2>&1)
        DETACH_EXIT_CODE=$?
        
        debug_log "デタッチ結果 (終了コード: ${DETACH_EXIT_CODE}):"
        debug_log "$DETACH_OUTPUT"
        
        if [ $DETACH_EXIT_CODE -ne 0 ]; then
          echo -e "${RED}エラー: グループ '${GROUP_NAME}' からのポリシーデタッチに失敗しました${NC}"
          echo -e "${RED}エラー詳細: $DETACH_OUTPUT${NC}"
          # 継続処理のためreturnしない
        else
          echo -e "${GREEN} - グループ '${GROUP_NAME}' からポリシーを正常にデタッチしました${NC}"
        fi
      fi
    done < /tmp/groups_list.txt
  else
    echo -e "${GREEN}ポリシーはどのグループにもアタッチされていません${NC}"
  fi
  
  # 一時ファイルのクリーンアップ
  rm -f /tmp/roles_list.txt /tmp/users_list.txt /tmp/groups_list.txt
  
  # ポリシーの削除
  echo -e "${YELLOW}IAMポリシー '${policy_name}' を削除します...${NC}"
  debug_log "ポリシー削除コマンド: aws iam delete-policy --policy-arn ${POLICY_ARN}"
  
  DELETE_OUTPUT=$(aws iam delete-policy --policy-arn ${POLICY_ARN} 2>&1)
  DELETE_EXIT_CODE=$?
  
  debug_log "削除結果 (終了コード: ${DELETE_EXIT_CODE}):"
  debug_log "$DELETE_OUTPUT"
  
  if [ $DELETE_EXIT_CODE -ne 0 ]; then
    echo -e "${RED}エラー: ポリシー '${policy_name}' の削除に失敗しました${NC}"
    echo -e "${RED}エラー詳細: $DELETE_OUTPUT${NC}"
    return 1
  else
    echo -e "${GREEN}IAMポリシー '${policy_name}' のクリーンアップが完了しました。${NC}"
    return 0
  fi
}

# 特定の環境のIAMロールを削除する関数
cleanup_iam_role() {
  local role_name=$1
  
  echo -e "${BLUE}IAMロール '${role_name}' を検索しています...${NC}"
  debug_log "ロール検索コマンド: aws iam get-role --role-name ${role_name}"
  
  # ロールが存在するか確認
  ROLE_OUTPUT=$(aws iam get-role --role-name ${role_name} 2>&1)
  ROLE_EXIT_CODE=$?
  
  debug_log "ロール検索結果 (終了コード: ${ROLE_EXIT_CODE}):"
  debug_log "$ROLE_OUTPUT"
  
  if [ $ROLE_EXIT_CODE -ne 0 ]; then
    echo -e "${GREEN}IAMロール '${role_name}' は見つかりませんでした。${NC}"
    return 0
  fi
  
  echo -e "${YELLOW}IAMロール '${role_name}' が見つかりました。${NC}"
  
  # アタッチされているポリシーを取得してデタッチ
  echo -e "${BLUE}アタッチされているポリシーを確認しています...${NC}"
  debug_log "アタッチポリシー確認コマンド: aws iam list-attached-role-policies --role-name ${role_name}"
  
  aws iam list-attached-role-policies --role-name ${role_name} --query "AttachedPolicies[].PolicyArn" --output text > /tmp/policies_list.txt
  
  if [ -s /tmp/policies_list.txt ]; then
    echo -e "${YELLOW}ポリシーをデタッチします...${NC}"
    while read -r POLICY_ARN; do
      debug_log "ポリシーARN: $POLICY_ARN"
      if [ ! -z "$POLICY_ARN" ]; then
        echo -e "${YELLOW} - ポリシー '${POLICY_ARN}' をデタッチします...${NC}"
        debug_log "デタッチコマンド: aws iam detach-role-policy --role-name ${role_name} --policy-arn ${POLICY_ARN}"
        
        DETACH_OUTPUT=$(aws iam detach-role-policy --role-name ${role_name} --policy-arn ${POLICY_ARN} 2>&1)
        DETACH_EXIT_CODE=$?
        
        debug_log "デタッチ結果 (終了コード: ${DETACH_EXIT_CODE}):"
        debug_log "$DETACH_OUTPUT"
        
        if [ $DETACH_EXIT_CODE -ne 0 ]; then
          echo -e "${RED}エラー: ポリシー '${POLICY_ARN}' のデタッチに失敗しました${NC}"
          echo -e "${RED}エラー詳細: $DETACH_OUTPUT${NC}"
        else
          echo -e "${GREEN} - ポリシー '${POLICY_ARN}' を正常にデタッチしました${NC}"
        fi
      fi
    done < /tmp/policies_list.txt
  else
    echo -e "${GREEN}アタッチされているポリシーはありません${NC}"
  fi
  
  # インラインポリシーを取得して削除
  echo -e "${BLUE}インラインポリシーを確認しています...${NC}"
  debug_log "インラインポリシー確認コマンド: aws iam list-role-policies --role-name ${role_name}"
  
  aws iam list-role-policies --role-name ${role_name} --query "PolicyNames" --output text > /tmp/inline_policies.txt
  
  if [ -s /tmp/inline_policies.txt ] && [ "$(cat /tmp/inline_policies.txt)" != "None" ]; then
    echo -e "${YELLOW}インラインポリシーを削除します...${NC}"
    while read -r POLICY_NAME; do
      debug_log "インラインポリシー名: $POLICY_NAME"
      if [ ! -z "$POLICY_NAME" ] && [ "$POLICY_NAME" != "None" ]; then
        echo -e "${YELLOW} - インラインポリシー '${POLICY_NAME}' を削除します...${NC}"
        debug_log "削除コマンド: aws iam delete-role-policy --role-name ${role_name} --policy-name ${POLICY_NAME}"
        
        DELETE_OUTPUT=$(aws iam delete-role-policy --role-name ${role_name} --policy-name ${POLICY_NAME} 2>&1)
        DELETE_EXIT_CODE=$?
        
        debug_log "削除結果 (終了コード: ${DELETE_EXIT_CODE}):"
        debug_log "$DELETE_OUTPUT"
        
        if [ $DELETE_EXIT_CODE -ne 0 ]; then
          echo -e "${RED}エラー: インラインポリシー '${POLICY_NAME}' の削除に失敗しました${NC}"
          echo -e "${RED}エラー詳細: $DELETE_OUTPUT${NC}"
        else
          echo -e "${GREEN} - インラインポリシー '${POLICY_NAME}' を正常に削除しました${NC}"
        fi
      fi
    done < /tmp/inline_policies.txt
  else
    echo -e "${GREEN}インラインポリシーはありません${NC}"
  fi
  
  # インスタンスプロファイルからロールを削除
  echo -e "${BLUE}インスタンスプロファイルを確認しています...${NC}"
  debug_log "インスタンスプロファイル確認コマンド: aws iam list-instance-profiles-for-role --role-name ${role_name}"
  
  PROFILES_OUTPUT=$(aws iam list-instance-profiles-for-role --role-name ${role_name} 2>&1)
  PROFILES_EXIT_CODE=$?
  
  debug_log "プロファイル確認結果 (終了コード: ${PROFILES_EXIT_CODE}):"
  debug_log "$PROFILES_OUTPUT"
  
  if [ $PROFILES_EXIT_CODE -eq 0 ]; then
    aws iam list-instance-profiles-for-role --role-name ${role_name} --query "InstanceProfiles[].InstanceProfileName" --output text > /tmp/profiles_list.txt
    
    if [ -s /tmp/profiles_list.txt ] && [ "$(cat /tmp/profiles_list.txt)" != "None" ]; then
      echo -e "${YELLOW}インスタンスプロファイルからロールを削除します...${NC}"
      while read -r PROFILE_NAME; do
        debug_log "プロファイル名: $PROFILE_NAME"
        if [ ! -z "$PROFILE_NAME" ] && [ "$PROFILE_NAME" != "None" ]; then
          echo -e "${YELLOW} - プロファイル '${PROFILE_NAME}' からロールを削除します...${NC}"
          debug_log "削除コマンド: aws iam remove-role-from-instance-profile --instance-profile-name ${PROFILE_NAME} --role-name ${role_name}"
          
          REMOVE_OUTPUT=$(aws iam remove-role-from-instance-profile --instance-profile-name ${PROFILE_NAME} --role-name ${role_name} 2>&1)
          REMOVE_EXIT_CODE=$?
          
          debug_log "削除結果 (終了コード: ${REMOVE_EXIT_CODE}):"
          debug_log "$REMOVE_OUTPUT"
          
          if [ $REMOVE_EXIT_CODE -ne 0 ]; then
            echo -e "${RED}エラー: プロファイル '${PROFILE_NAME}' からのロール削除に失敗しました${NC}"
            echo -e "${RED}エラー詳細: $REMOVE_OUTPUT${NC}"
          else
            echo -e "${GREEN} - プロファイル '${PROFILE_NAME}' からロールを正常に削除しました${NC}"
          fi
        fi
      done < /tmp/profiles_list.txt
    else
      echo -e "${GREEN}インスタンスプロファイルはありません${NC}"
    fi
  else
    echo -e "${YELLOW}インスタンスプロファイル情報の取得に失敗しました。続行します...${NC}"
  fi
  
  # 一時ファイルのクリーンアップ
  rm -f /tmp/policies_list.txt /tmp/inline_policies.txt /tmp/profiles_list.txt
  
  # ロールの削除
  echo -e "${YELLOW}IAMロール '${role_name}' を削除します...${NC}"
  debug_log "ロール削除コマンド: aws iam delete-role --role-name ${role_name}"
  
  DELETE_OUTPUT=$(aws iam delete-role --role-name ${role_name} 2>&1)
  DELETE_EXIT_CODE=$?
  
  debug_log "削除結果 (終了コード: ${DELETE_EXIT_CODE}):"
  debug_log "$DELETE_OUTPUT"
  
  if [ $DELETE_EXIT_CODE -ne 0 ]; then
    echo -e "${RED}エラー: ロール '${role_name}' の削除に失敗しました${NC}"
    echo -e "${RED}エラー詳細: $DELETE_OUTPUT${NC}"
    return 1
  else
    echo -e "${GREEN}IAMロール '${role_name}' のクリーンアップが完了しました。${NC}"
    return 0
  fi
}

# 特定の環境用のIAMポリシーのクリーンアップ
echo -e "\n${BLUE}環境 '${ENV}' 用のIAMポリシーをクリーンアップしています...${NC}"

# SSMパラメータアクセスポリシーのクリーンアップ
cleanup_iam_policy "${ENV}-ssm-parameter-access"

# 他のポリシーのクリーンアップ（必要に応じて追加）
# cleanup_iam_policy "${ENV}-additional-policy"

# 特定の環境用のIAMロールのクリーンアップ（必要な場合）
# echo -e "\n${BLUE}環境 '${ENV}' 用のIAMロールをクリーンアップしています...${NC}"
# cleanup_iam_role "${ENV}-specific-role"

# クリーンアップ後の確認
echo -e "\n${BLUE}IAMリソースのクリーンアップ後の状態を確認しています...${NC}"

# SSMパラメータアクセスポリシーの確認
debug_log "ポリシー確認コマンド: aws iam list-policies --query \"Policies[?PolicyName=='${ENV}-ssm-parameter-access'].Arn\" --output text"
SSM_POLICY_EXISTS=$(aws iam list-policies --query "Policies[?PolicyName=='${ENV}-ssm-parameter-access'].Arn" --output text)
debug_log "確認結果: '$SSM_POLICY_EXISTS'"

if [ -z "$SSM_POLICY_EXISTS" ] || [ "$SSM_POLICY_EXISTS" == "None" ]; then
  echo -e "${GREEN}SSMパラメータアクセスポリシーは正常に削除されました。${NC}"
else
  echo -e "${RED}警告: SSMパラメータアクセスポリシーが残っています: ${SSM_POLICY_EXISTS}${NC}"
  echo -e "${YELLOW}手動で削除する必要があるかもしれません。${NC}"
  
  # 手動削除のコマンド例を表示
  echo -e "${YELLOW}手動削除コマンド例:${NC}"
  echo -e "${YELLOW}1. ポリシーアタッチメント確認:${NC}"
  echo -e "  aws iam list-entities-for-policy --policy-arn ${SSM_POLICY_EXISTS}"
  echo -e "${YELLOW}2. ロールからデタッチ（ROLENAMEを置き換え）:${NC}"
  echo -e "  aws iam detach-role-policy --role-name ROLENAME --policy-arn ${SSM_POLICY_EXISTS}"
  echo -e "${YELLOW}3. ポリシー削除:${NC}"
  echo -e "  aws iam delete-policy --policy-arn ${SSM_POLICY_EXISTS}"
fi

# 残りのIAMリソースの確認（環境タグ付きのもの）
echo -e "\n${BLUE}環境タグ '${ENV}' が付いた残りのIAMリソースを確認しています...${NC}"
debug_log "タグリソース確認コマンド: aws resourcegroupstaggingapi get-resources --resource-type-filters \"iam:*\" --tag-filters Key=Environment,Values=${ENV}"

IAM_RESOURCES_OUTPUT=$(aws resourcegroupstaggingapi get-resources --resource-type-filters "iam:*" --tag-filters Key=Environment,Values=${ENV} 2>&1)
IAM_RESOURCES_EXIT_CODE=$?

debug_log "確認結果 (終了コード: ${IAM_RESOURCES_EXIT_CODE}):"
debug_log "$IAM_RESOURCES_OUTPUT"

if [ $IAM_RESOURCES_EXIT_CODE -eq 0 ]; then
  IAM_RESOURCES=$(aws resourcegroupstaggingapi get-resources --resource-type-filters "iam:*" --tag-filters Key=Environment,Values=${ENV} --query "ResourceTagMappingList[].ResourceARN" --output text)
  
  if [ -z "$IAM_RESOURCES" ] || [ "$IAM_RESOURCES" == "None" ]; then
    echo -e "${GREEN}環境タグ '${ENV}' が付いたIAMリソースは見つかりませんでした。${NC}"
  else
    echo -e "${YELLOW}環境タグ '${ENV}' が付いた残りのIAMリソース:${NC}"
    for RESOURCE in $IAM_RESOURCES; do
      echo -e "${YELLOW} - ${RESOURCE}${NC}"
    done
    echo -e "${YELLOW}これらのリソースは手動で確認し、必要に応じてクリーンアップしてください。${NC}"
  fi
else
  echo -e "${YELLOW}タグ付きリソースの確認に失敗しました。${NC}"
  echo -e "${YELLOW}エラー詳細: $IAM_RESOURCES_OUTPUT${NC}"
fi

echo -e "\n${GREEN}IAMリソースのクリーンアップが完了しました。${NC}"
echo -e "${YELLOW}注意: Terraformの状態ファイルは更新されていません。必要に応じて terraform state rm コマンドを実行してください。${NC}"

debug_log "スクリプト実行完了"