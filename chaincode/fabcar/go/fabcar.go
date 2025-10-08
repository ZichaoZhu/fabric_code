// 好的，我完全理解了你的新想法。这是一个非常典型的“申请-审批”工作流，非常适合用区块链来实现。将用户角色分为 all（任何人）、student（已验证学生）和 validator（验证者），并引入 status 属性，是实现这个逻辑的关键。

// 为了代码的清晰和逻辑的严谨，我将为你重写这个链码。新的代码会更清晰地体现你设计的角色和流程。

// 核心设计思路
// 状态 (Status)：为 Student, Grade, Price 结构体统一增加 Status 字段，可能的值为 Pending (待审核), Approved (已批准), Rejected (已拒绝)。
// 角色定义:
// all (任何人)：不需要特定角色。任何连接到网络的身份都可以是 all。
// student (已验证学生)：我们不在证书里设置 student 角色。而是在链码中动态判断：如果一个调用者，其身份ID（Owner ID）对应的学生记录状态为 Approved，那么我们就认为他具有 student 身份。
// validator (验证者)：我们指定一个组织的管理员为验证者。在 test-network 中，我们可以指定 Org1MSP 的管理员为唯一的验证者。这是最安全、最符合 Fabric 设计的“内置”方式。
// 函数职责:
// addStudent: 任何人可调用，创建状态为 Pending 的学生记录。
// addGrade/addPrice: 只有身份为 student 的用户可调用，创建状态为 Pending 的成绩/奖项记录。
// validate...: 新增一系列 validate 函数，只有 validator 可调用，用于将记录状态从 Pending 修改为 Approved 或 Rejected。
// query...: 任何人可调用，但只返回状态为 Approved 的记录。



package main

import (
    "encoding/json"
    "fmt"
    "strconv"

    "github.com/hyperledger/fabric-chaincode-go/pkg/cid"
    "github.com/hyperledger/fabric-chaincode-go/shim"
    sc "github.com/hyperledger/fabric-protos-go/peer"
)

// --- 常量定义 ---
const (
    StatusPending  = "Pending"
    StatusApproved = "Approved"
    StatusRejected = "Rejected"

    ValidatorMSP = "Org1MSP" // 指定 Org1MSP 为验证者组织的 MSP ID
)

type SmartContract struct{}

// --- 数据结构定义 ---
type Student struct {
    School string `json:"school"`
    Major  string `json:"major"`
    Id     int    `json:"id"`
    Name   string `json:"name"`
    Owner  string `json:"owner"` // 创建者的唯一ID
    Status string `json:"status"`// 状态: Pending, Approved, Rejected
}

type Grade struct {
    Course_name string  `json:"course"`
    Course_id   string  `json:"courseId"`
    Teacher     string  `json:"teacher"`
    School      string  `json:"school"`
    Student_id  int     `json:"studentId"`
    Year        int     `json:"year"`
    Semester    int     `json:"semester"`
    Score       float64 `json:"score"`
    Owner       string  `json:"owner"`
    Status      string  `json:"status"`
}

type Price struct {
    Name        string `json:"name"`
    Id          string `json:"id"`
    Year        int    `json:"year"`
    Level       string `json:"level"`
    Institution string `json:"institution"`
    Owner       string `json:"owner"`
    Status      string `json:"status"`
}

// --- 辅助函数 ---

// requireValidator 检查调用者是否是指定的验证者组织
func requireValidator(stub shim.ChaincodeStubInterface) error {
    mspID, err := cid.GetMSPID(stub)
    if err != nil {
        return fmt.Errorf("获取 MSP ID 失败: %v", err)
    }
    if mspID != ValidatorMSP {
        return fmt.Errorf("权限拒绝: 只有 %s 的成员才能执行此操作", ValidatorMSP)
    }
    return nil
}

// requireStudent 检查调用者是否是一个已被批准的学生
func requireStudent(stub shim.ChaincodeStubInterface, school string, studentId string) error {
    callerID, err := cid.GetID(stub)
    if err != nil {
        return fmt.Errorf("获取调用者ID失败: %v", err)
    }

    // 检查该调用者对应的学生记录是否存在且已被批准
    studentKey := school + studentId
    studentAsBytes, err := stub.GetState(studentKey)
    if err != nil || studentAsBytes == nil {
        return fmt.Errorf("权限拒绝: 找不到对应的学生记录")
    }

    var student Student
    json.Unmarshal(studentAsBytes, &student)

    if student.Owner != callerID {
        return fmt.Errorf("权限拒绝: 你只能为自己添加信息")
    }
    if student.Status != StatusApproved {
        return fmt.Errorf("权限拒绝: 你的学生身份尚未被验证通过")
    }

    return nil
}

func getCallerID(stub shim.ChaincodeStubInterface) (string, error) {
    return cid.GetID(stub)
}

func atoi(str string) int { i, _ := strconv.Atoi(str); return i }
func atof(str string) float64 { f, _ := strconv.ParseFloat(str, 64); return f }

// --- 链码生命周期函数 ---

func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
    return shim.Success(nil)
}

func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {
    function, args := APIstub.GetFunctionAndParameters()
    switch function {
    // "all" 角色调用的函数
    case "addStudent":
        return s.addStudent(APIstub, args)
    // "student" 角色调用的函数
    case "addGrade":
        return s.addGrade(APIstub, args)
    case "addPrice":
        return s.addPrice(APIstub, args)
    // "validator" 角色调用的函数
    case "validateStudent":
        return s.validateStudent(APIstub, args)
    case "validateGrade":
        return s.validateGrade(APIstub, args)
    case "validatePrice":
        return s.validatePrice(APIstub, args)
    // 公共查询函数
    case "queryStudent":
        return s.queryStudent(APIstub, args)
    case "queryGrade":
        return s.queryGrade(APIstub, args)
    case "queryPrice":
        return s.queryPrice(APIstub, args)
    default:
        return shim.Error("无效的链码函数名")
    }
}

// --- 业务逻辑函数 ---

// addStudent 任何人都可以调用，申请创建一个学生身份
// 参数顺序：school, major, id, name
func (s *SmartContract) addStudent(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
    if len(args) != 4 { return shim.Error("参数数量错误，需要4个 (school, major, id, name)") }

    callerID, err := getCallerID(APIstub)
    if err != nil { return shim.Error(err.Error()) }

    student := Student{
        School: args[0], Major: args[1], Id: atoi(args[2]), Name: args[3],
        Owner:  callerID,
        Status: StatusPending, // 初始状态为待审核
    }

    studentAsBytes, _ := json.Marshal(student)
    key := args[0] + args[2] // school + id as key
    if err := APIstub.PutState(key, studentAsBytes); err != nil {
        return shim.Error(fmt.Sprintf("保存学生申请失败: %s", key))
    }
    return shim.Success(nil)
}

// addGrade 只有被批准的学生才能为自己添加成绩
// 参数顺序：course_name, course_id, teacher, school, studentId, year, score, semester
func (s *SmartContract) addGrade(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
    if len(args) != 8 { return shim.Error("参数数量错误，需要8个") }
    
    // 权限检查：必须是已验证的学生，且只能为自己操作
    school, studentId := args[3], args[4]
    if err := requireStudent(APIstub, school, studentId); err != nil {
        return shim.Error(err.Error())
    }

    callerID, _ := getCallerID(APIstub) // 在 requireStudent 中已检查过错误
    grade := Grade{
        Course_name: args[0], Course_id: args[1], Teacher: args[2],
        School: school, Student_id: atoi(studentId), Year: atoi(args[5]),
        Score: atof(args[6]), Semester: atoi(args[7]),
        Owner:  callerID,
        Status: StatusPending,
    }

    gradeAsBytes, _ := json.Marshal(grade)
    key := school + studentId + args[1] + args[5] + args[7] // school+studentid+courseid+year+semester
    if err := APIstub.PutState(key, gradeAsBytes); err != nil {
        return shim.Error(fmt.Sprintf("保存成绩申请失败: %s", key))
    }
    return shim.Success(nil)
}

// addPrice 只有被批准的学生才能为自己添加奖项记录
// 参数顺序：school, studentId, prizeName, prizeId, year, level, institution
func (s *SmartContract) addPrice(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
    // 1. 检查参数数量，现在需要7个参数
    if len(args) != 7 {
        return shim.Error("参数数量错误，需要7个 (school, studentId, prizeName, prizeId, year, level, institution)")
    }

    // 2. 提取参数，使其更具可读性
    school := args[0]
    studentId := args[1]
    prizeName := args[2]
    prizeId := args[3]
    year := args[4]
    level := args[5]
    institution := args[6]

    // 3. 权限检查：使用传入的 school 和 studentId 验证调用者是否为合法的、已批准的学生
    if err := requireStudent(APIstub, school, studentId); err != nil {
        return shim.Error(err.Error())
    }

    // 获取调用者ID，用于设置 Owner 字段
    callerID, _ := getCallerID(APIstub) // 在 requireStudent 中已检查过错误，这里可以忽略

    // 4. 创建 Price 对象
    price := Price{
        Name:        prizeName,
        Id:          prizeId,
        Year:        atoi(year),
        Level:       level,
        Institution: institution,
        Owner:       callerID,      // 记录数据所有者
        Status:      StatusPending, // 初始状态为待审核
    }

    priceAsBytes, _ := json.Marshal(price)
    
    // 5. 使用唯一的奖项ID作为键（Key）存储到账本
    key := prizeId
    if err := APIstub.PutState(key, priceAsBytes); err != nil {
        return shim.Error(fmt.Sprintf("保存奖项申请失败: %s", key))
    }
    
    return shim.Success(nil)
}

// validateStudent 验证者调用，审批学生身份申请
// 参数顺序：school, studentId, newStatus
func (s *SmartContract) validateStudent(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
    if err := requireValidator(APIstub); err != nil { return shim.Error(err.Error()) }
    if len(args) != 3 { return shim.Error("参数数量错误，需要3个 (school, studentId, newStatus)") }

    newStatus := args[2]
    if newStatus != StatusApproved && newStatus != StatusRejected {
        return shim.Error("无效的状态，只能是 'Approved' 或 'Rejected'")
    }

    key := args[0] + args[1]
    studentAsBytes, err := APIstub.GetState(key)
    if err != nil || studentAsBytes == nil { return shim.Error("找不到待审批的学生记录") }

    var student Student
    json.Unmarshal(studentAsBytes, &student)
    student.Status = newStatus // 更新状态

    studentAsBytes, _ = json.Marshal(student)
    if err := APIstub.PutState(key, studentAsBytes); err != nil {
        return shim.Error("更新学生状态失败")
    }
    return shim.Success(nil)
}

// validateGrade 验证者调用，审批成绩
// 参数顺序：school, studentId, courseId, year, semester, newStatus
func (s *SmartContract) validateGrade(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
    if err := requireValidator(APIstub); err != nil { return shim.Error(err.Error()) }
    if len(args) != 6 { return shim.Error("参数数量错误，需要6个 (school, studentId, courseId, year, semester, newStatus)") }
    
    newStatus := args[5]
    if newStatus != StatusApproved && newStatus != StatusRejected {
        return shim.Error("无效的状态，只能是 'Approved' 或 'Rejected'")
    }

    key := args[0] + args[1] + args[2] + args[3] + args[4]
    gradeAsBytes, err := APIstub.GetState(key)
    if err != nil || gradeAsBytes == nil { return shim.Error("找不到待审批的成绩记录") }

    var grade Grade
    json.Unmarshal(gradeAsBytes, &grade)
    grade.Status = newStatus

    gradeAsBytes, _ = json.Marshal(grade)
    if err := APIstub.PutState(key, gradeAsBytes); err != nil {
        return shim.Error("更新成绩状态失败")
    }
    return shim.Success(nil)
}

// validatePrice 验证者调用，审批奖项
// 参数顺序：priceId, newStatus
func (s *SmartContract) validatePrice(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
    if err := requireValidator(APIstub); err != nil { return shim.Error(err.Error()) }
    if len(args) != 2 { return shim.Error("参数数量错误，需要2个 (priceId, newStatus)") }

    newStatus := args[1]
    if newStatus != StatusApproved && newStatus != StatusRejected {
        return shim.Error("无效的状态，只能是 'Approved' 或 'Rejected'")
    }

    key := args[0]
    priceAsBytes, err := APIstub.GetState(key)
    if err != nil || priceAsBytes == nil { return shim.Error("找不到待审批的奖项记录") }

    var price Price
    json.Unmarshal(priceAsBytes, &price)
    price.Status = newStatus

    priceAsBytes, _ = json.Marshal(price)
    if err := APIstub.PutState(key, priceAsBytes); err != nil {
        return shim.Error("更新奖项状态失败")
    }
    return shim.Success(nil)
}

// queryStudent 任何人可调用，但只返回已批准的学生信息
// 参数顺序：school, studentId
func (s *SmartContract) queryStudent(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
    if len(args) != 2 { return shim.Error("参数数量错误，需要2个 (school, studentId)") }

    key := args[0] + args[1]
    studentAsBytes, err := APIstub.GetState(key)
    if err != nil || studentAsBytes == nil { return shim.Error("找不到学生信息") }

    var student Student
    json.Unmarshal(studentAsBytes, &student)

    // 关键：只返回已批准的记录
    if student.Status != StatusApproved {
        return shim.Error("该学生信息尚未通过验证或已被拒绝")
    }

    return shim.Success(studentAsBytes)
}

// queryGrade 逻辑同 queryStudent
// 参数顺序：school, studentId, courseId, year, semester
func (s *SmartContract) queryGrade(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
    if len(args) != 5 { return shim.Error("参数数量错误，需要5个") }

    key := args[0] + args[1] + args[2] + args[3] + args[4]
    gradeAsBytes, err := APIstub.GetState(key)
    if err != nil || gradeAsBytes == nil { return shim.Error("找不到成绩信息") }

    var grade Grade
    json.Unmarshal(gradeAsBytes, &grade)

    if grade.Status != StatusApproved {
        return shim.Error("该成绩信息尚未通过验证或已被拒绝")
    }

    return shim.Success(gradeAsBytes)
}

// queryPrice 逻辑同 queryStudent
// 参数顺序：priceId
func (s *SmartContract) queryPrice(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
    if len(args) != 1 { return shim.Error("参数数量错误，需要1个 (priceId)") }

    key := args[0]
    priceAsBytes, err := APIstub.GetState(key)
    if err != nil || priceAsBytes == nil { return shim.Error("找不到奖项信息") }

    var price Price
    json.Unmarshal(priceAsBytes, &price)

    if price.Status != StatusApproved {
        return shim.Error("该奖项信息尚未通过验证或已被拒绝")
    }

    return shim.Success(priceAsBytes)
}

// --- main 函数 ---
func main() {
    if err := shim.Start(new(SmartContract)); err != nil {
        fmt.Printf("创建新的智能合约失败: %s", err)
    }
}