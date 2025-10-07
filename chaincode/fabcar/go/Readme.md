## 2025.10.07
### 代码更新
在这一次的更新中，我切换了一下思路：
1. 核心设计思路
 * 状态 (Status)：为 Student, Grade, Price 结构体统一增加 Status 字段，可能的值为 Pending (待审核), Approved (已批准), Rejected (已拒绝)。
2. 角色定义:
 * all (任何人)：不需要特定角色。任何连接到网络的身份都可以是 all。
 * student (已验证学生)：我们不在证书里设置 student 角色。而是在链码中动态判断：如果一个调用者，其身份ID（Owner ID）对应的学生记录状态为 Approved，那么我们就认为他具有 student 身份。
 * validator (验证者)：我们指定一个组织的管理员为验证者。在 test-network 中，我们可以指定 Org1MSP 的管理员为唯一的验证者。这是最安全、最符合 Fabric 设计的“内置”方式。
3. 函数职责:
 * addStudent: 任何人可调用，创建状态为 Pending 的学生记录。
 * addGrade/addPrice: 只有身份为 student 的用户可调用，创建状态为 Pending 的成绩/奖项记录。
 * validate...: 新增一系列 validate 函数，只有 validator 可调用，用于将记录状态从 Pending 修改为 Approved 或 Rejected。
 * query...: 任何人可调用，但只返回状态为 Approved 的记录。

### 测试运行
同样的，我更新了测试程序test_function.sh，我自己的测试结果如下：
```bash
zzc@zzc-virtual-machine:~/hyperledger/fabric/scripts/fabric-samples/chaincode/fabcar/go$ ./test_function.sh

############### 1. 申请人 (Org2) 申请学生身份 ###############

==== 身份切换为: User1@org2.example.com (申请人) ====
--> 正在调用: {"function":"addStudent","Args":["zju", "cs", "3180100001", "Tom"]}
2025-10-07 10:55:20.530 CST 0001 INFO [chaincodeCmd] chaincodeInvokeOrQuery -> Chaincode invoke successful. result: status:200 

############### 2. 申请人查询自己 (预期失败，因为状态是 Pending) ###############

==== 身份切换为: User1@org2.example.com (申请人) ====
--> 正在调用: {"function":"queryStudent","Args":["zju", "3180100001"]}
Error: endorsement failure during query. response: status:500 message:"\350\257\245\345\255\246\347\224\237\344\277\241\346\201\257\345\260\232\346\234\252\351\200\232\350\277\207\351\252\214\350\257\201\346\210\226\345\267\262\350\242\253\346\213\222\347\273\235" 

############### 3. 验证者 (Org1) 审批通过学生申请 ###############

==== 身份切换为: Admin@org1.example.com (验证者) ====
--> 正在调用: {"function":"validateStudent","Args":["zju", "3180100001", "Approved"]}
2025-10-07 10:55:24.745 CST 0001 INFO [chaincodeCmd] chaincodeInvokeOrQuery -> Chaincode invoke successful. result: status:200 

############### 4. 任何人查询该学生 (预期成功) ###############

==== 身份切换为: User1@org2.example.com (申请人) ====
--> 正在调用: {"function":"queryStudent","Args":["zju", "3180100001"]}
{"school":"zju","major":"cs","id":3180100001,"name":"Tom","owner":"eDUwOTo6Q049VXNlcjFAb3JnMi5leGFtcGxlLmNvbSxPVT1jbGllbnQsTD1TYW4gRnJhbmNpc2NvLFNUPUNhbGlmb3JuaWEsQz1VUzo6Q049Y2Eub3JnMi5leGFtcGxlLmNvbSxPPW9yZzIuZXhhbXBsZS5jb20sTD1TYW4gRnJhbmNpc2NvLFNUPUNhbGlmb3JuaWEsQz1VUw==","status":"Approved"}
--> 预期能看到 Tom 的完整信息，并且 status 为 Approved

############### 5. 已验证的学生 (Tom) 为自己添加成绩 ###############

==== 身份切换为: User1@org2.example.com (申请人) ====
--> 正在调用: {"function":"addGrade","Args":["OS", "C001", "Prof.Lee", "zju", "3180100001", "2025", "95", "1"]}
2025-10-07 10:55:29.090 CST 0001 INFO [chaincodeCmd] chaincodeInvokeOrQuery -> Chaincode invoke successful. result: status:200 

############### 6. 验证者 (Org1) 审批通过该成绩 ###############

==== 身份切换为: Admin@org1.example.com (验证者) ====
--> 正在调用: {"function":"validateGrade","Args":["zju", "3180100001", "C001", "2025", "1", "Approved"]}
2025-10-07 10:55:31.339 CST 0001 INFO [chaincodeCmd] chaincodeInvokeOrQuery -> Chaincode invoke successful. result: status:200 

############### 7. 任何人查询该成绩 (预期成功) ###############

==== 身份切换为: Admin@org1.example.com (验证者) ====
--> 正在调用: {"function":"queryGrade","Args":["zju", "3180100001", "C001", "2025", "1"]}
{"course":"OS","courseId":"C001","teacher":"Prof.Lee","school":"zju","studentId":3180100001,"year":2025,"semester":1,"score":95,"owner":"eDUwOTo6Q049VXNlcjFAb3JnMi5leGFtcGxlLmNvbSxPVT1jbGllbnQsTD1TYW4gRnJhbmNpc2NvLFNUPUNhbGlmb3JuaWEsQz1VUzo6Q049Y2Eub3JnMi5leGFtcGxlLmNvbSxPPW9yZzIuZXhhbXBsZS5jb20sTD1TYW4gRnJhbmNpc2NvLFNUPUNhbGlmb3JuaWEsQz1VUw==","status":"Approved"}
--> 预期能看到分数为 95.0 的成绩信息

############### 8. 已验证的学生 (Tom) 为自己添加奖项 ###############

==== 身份切换为: User1@org2.example.com (申请人) ====
--> 正在调用: {"function":"addPrice","Args":["zju", "3180100001", "National Scholarship", "PRICE-001", "2025", "National", "MOE"]}
2025-10-07 10:55:35.604 CST 0001 INFO [chaincodeCmd] chaincodeInvokeOrQuery -> Chaincode invoke successful. result: status:200 

############### 9. 验证者 (Org1) 审批通过该奖项 ###############

==== 身份切换为: Admin@org1.example.com (验证者) ====
--> 正在调用: {"function":"validatePrice","Args":["PRICE-001", "Approved"]}
2025-10-07 10:55:37.687 CST 0001 INFO [chaincodeCmd] chaincodeInvokeOrQuery -> Chaincode invoke successful. result: status:200 

############### 10. 任何人查询该奖项 (预期成功) ###############

==== 身份切换为: User1@org2.example.com (申请人) ====
--> 正在调用: {"function":"queryPrice","Args":["PRICE-001"]}
{"name":"National Scholarship","id":"PRICE-001","year":2025,"level":"National","institution":"MOE","owner":"eDUwOTo6Q049VXNlcjFAb3JnMi5leGFtcGxlLmNvbSxPVT1jbGllbnQsTD1TYW4gRnJhbmNpc2NvLFNUPUNhbGlmb3JuaWEsQz1VUzo6Q049Y2Eub3JnMi5leGFtcGxlLmNvbSxPPW9yZzIuZXhhbXBsZS5jb20sTD1TYW4gRnJhbmNpc2NvLFNUPUNhbGlmb3JuaWEsQz1VUw==","status":"Approved"}
--> 预期能看到 National Scholarship 的奖项信息


===================== 所有测试执行完毕 =====================
```
可以看到测试还是挺成功的。

如果是第一次运行的话，需要给脚本赋予权限，也就是使用`chmod`命令。

使用方法就是先使用deploy_gradechain.sh部署链码，然后再使用test_function.sh测试。