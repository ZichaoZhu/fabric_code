#!/bin/bash
# filepath: /home/zzc/hyperledger/fabric/scripts/fabric-samples/chaincode/fabcar/go/test_functions.sh

# --- 环境设置 ---
# 这是一个好习惯，确保脚本从 test-network 目录执行 peer 命令
cd ~/hyperledger/fabric/scripts/fabric-samples/test-network

# 设置基础环境变量
export PATH=${PWD}/../bin:$PATH
export FABRIC_CFG_PATH=$PWD/../config/
export CORE_PEER_TLS_ENABLED=true
CHANNEL_NAME="mychannel"
CC_NAME="fabcar" # 你的链码名称

# 定义 Orderer 的地址和 TLS 证书路径
ORDERER_ADDRESS=localhost:7050
ORDERER_TLS_ROOTCERT_FILE=${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem

# --- 辅助函数 ---
# 这个函数用于设置调用命令时所使用的身份（属于哪个组织，是哪个用户）
set_identity() {
  ORG_NAME=$1
  USER_NAME=$2
  
  if [ "$ORG_NAME" == "Org1" ]; then
    export CORE_PEER_LOCALMSPID="Org1MSP"
    export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
    export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org1.example.com/users/${USER_NAME}@org1.example.com/msp
    export CORE_PEER_ADDRESS=localhost:7051
    echo -e "\n==== 身份切换为: ${USER_NAME}@org1.example.com (验证者) ===="
  elif [ "$ORG_NAME" == "Org2" ]; then
    export CORE_PEER_LOCALMSPID="Org2MSP"
    export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt
    export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org2.example.com/users/${USER_NAME}@org2.example.com/msp
    export CORE_PEER_ADDRESS=localhost:9051
    echo -e "\n==== 身份切换为: ${USER_NAME}@org2.example.com (申请人) ===="
  else
    echo "错误: 无效的组织名称 '$ORG_NAME'，只能是 Org1 或 Org2"
    exit 1
  fi
}

# 封装 peer chaincode invoke/query 命令，使其更简洁
invoke_chaincode() {
  local FUNCTION_CALL=$1
  local EXPECTED_RESULT=$2 # "invoke" or "query"

  echo "--> 正在调用: $FUNCTION_CALL"
  
  if [ "$EXPECTED_RESULT" == "invoke" ]; then
    peer chaincode invoke -o $ORDERER_ADDRESS --ordererTLSHostnameOverride orderer.example.com \
    --tls --cafile $ORDERER_TLS_ROOTCERT_FILE -C $CHANNEL_NAME -n $CC_NAME \
    --peerAddresses localhost:7051 --tlsRootCertFiles ${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt \
    --peerAddresses localhost:9051 --tlsRootCertFiles ${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt \
    -c "$FUNCTION_CALL"
  else
    peer chaincode query -C $CHANNEL_NAME -n $CC_NAME -c "$FUNCTION_CALL"
  fi
  sleep 2 # 等待交易上链
}

# --- 测试区域 ---

# === 阶段一: 学生身份申请与审批 ===

echo -e "\n############### 1. 申请人 (Org2) 申请学生身份 ###############"
set_identity Org2 User1
FUNCTION_CALL='{"function":"addStudent","Args":["zju", "cs", "3180100001", "Tom"]}'
invoke_chaincode "$FUNCTION_CALL" "invoke"

echo -e "\n############### 2. 申请人查询自己 (预期失败，因为状态是 Pending) ###############"
set_identity Org2 User1
FUNCTION_CALL='{"function":"queryStudent","Args":["zju", "3180100001"]}'
invoke_chaincode "$FUNCTION_CALL" "query"

echo -e "\n############### 3. 验证者 (Org1) 审批通过学生申请 ###############"
set_identity Org1 Admin
FUNCTION_CALL='{"function":"validateStudent","Args":["zju", "3180100001", "Approved"]}'
invoke_chaincode "$FUNCTION_CALL" "invoke"

echo -e "\n############### 4. 任何人查询该学生 (预期成功) ###############"
set_identity Org2 User1 # 用谁查询都可以
FUNCTION_CALL='{"function":"queryStudent","Args":["zju", "3180100001"]}'
invoke_chaincode "$FUNCTION_CALL" "query"
echo "--> 预期能看到 Tom 的完整信息，并且 status 为 Approved"


# === 阶段二: 成绩申请与审批 ===

echo -e "\n############### 5. 已验证的学生 (Tom) 为自己添加成绩 ###############"
set_identity Org2 User1 # 必须是 Tom 自己 (Org2 User1)
FUNCTION_CALL='{"function":"addGrade","Args":["OS", "C001", "Prof.Lee", "zju", "3180100001", "2025", "95", "1"]}'
invoke_chaincode "$FUNCTION_CALL" "invoke"

echo -e "\n############### 6. 验证者 (Org1) 审批通过该成绩 ###############"
set_identity Org1 Admin
FUNCTION_CALL='{"function":"validateGrade","Args":["zju", "3180100001", "C001", "2025", "1", "Approved"]}'
invoke_chaincode "$FUNCTION_CALL" "invoke"

echo -e "\n############### 7. 任何人查询该成绩 (预期成功) ###############"
set_identity Org1 Admin # 用谁查询都可以
FUNCTION_CALL='{"function":"queryGrade","Args":["zju", "3180100001", "C001", "2025", "1"]}'
invoke_chaincode "$FUNCTION_CALL" "query"
echo "--> 预期能看到分数为 95.0 的成绩信息"


# === 阶段三: 奖项申请与审批 ===

echo -e "\n############### 8. 已验证的学生 (Tom) 为自己添加奖项 ###############"
set_identity Org2 User1 # 必须是 Tom 自己 (Org2 User1)
FUNCTION_CALL='{"function":"addPrice","Args":["zju", "3180100001", "National Scholarship", "PRICE-001", "2025", "National", "MOE"]}'
invoke_chaincode "$FUNCTION_CALL" "invoke"

echo -e "\n############### 9. 验证者 (Org1) 审批通过该奖项 ###############"
set_identity Org1 Admin
FUNCTION_CALL='{"function":"validatePrice","Args":["PRICE-001", "Approved"]}'
invoke_chaincode "$FUNCTION_CALL" "invoke"

echo -e "\n############### 10. 任何人查询该奖项 (预期成功) ###############"
set_identity Org2 User1 # 用谁查询都可以
FUNCTION_CALL='{"function":"queryPrice","Args":["PRICE-001"]}'
invoke_chaincode "$FUNCTION_CALL" "query"
echo "--> 预期能看到 National Scholarship 的奖项信息"


echo -e "\n\n===================== 所有测试执行完毕 ====================="