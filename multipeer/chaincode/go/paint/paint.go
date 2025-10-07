package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer"
)

type SmartContract struct {
	//contractapi.Contract
}

type Operation struct {
	OperationID      string `json:"operationid"`      //操作ID
	UserId           string `json:"userid"`           //用户ID
	UserName         string `json:"username"`         //用户名
	InstitutionName  string `json:"institutionname"`  //机构名
	ImageId          string `json:"imageid"`          //图片ID
	ImageName        string `json:"imagename"`        //图片名
	OperationType    string `json:"operationtype"`    //操作类型 增删改查
	OperationContent string `json:"operationcontent"` //操作内容
	OperationTime    string `json:"operationtime"`    //操作的时间
}

type Work struct {
	//WordkId   string `json:"workid"`
	Hash1     string `json:"hash1"`     //第一个hash
	Hash3     string `json:"hash3"`     //第二个hash
	Signature string `json:"signature"` //签名
}

func main() {
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Println("chaincode start failed.", err)
	}
}

func (SmartContract) Init(stub shim.ChaincodeStubInterface) peer.Response {
	var works = []Work{
		{
			//WordkId:   "0",
			Hash1:     "test",
			Hash3:     "test",
			Signature: "test",
		},
	}

	for _, work := range works {
		workAsBytes, err := json.Marshal(work)
		if err != nil {
			return shim.Error("Failed to marshal work")
		}
		err = stub.PutState("Hash1:"+work.Hash1, workAsBytes) // 使用hash1 作为 key 进行存储数据
		if err != nil {
			return shim.Error("Failed to put work to chain")
		}
	}

	var operations = []Operation{
		{
			OperationID:      "test",
			UserId:           "test",
			UserName:         "test",
			InstitutionName:  "test",
			ImageId:          "test",
			ImageName:        "test",
			OperationType:    "test",
			OperationContent: "test",
			OperationTime:    "test",
		},
	}

	for _, operation := range operations {
		operBytes, err := json.Marshal(operation)
		if err != nil {
			return shim.Error("Failed to marshal operation")
		}
		err = stub.PutState("Operation:"+operation.OperationID, operBytes)
		if err != nil {
			return shim.Error("Failed to put operation to chain")
		}
	}
	return shim.Success([]byte("Init success"))
}

func (SmartContract) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	fun, args := stub.GetFunctionAndParameters()
	switch fun {
	case "queryAllWork":
		return QueryAllWork(stub, args)
	case "putWork":
		return PutWork(stub, args)
	case "queryWork":
		return QueryWork(stub, args)
	case "queryAllOperation":
		return QueryAllOperation(stub, args)
	case "putOperation":
		return PutOperation(stub, args)
	case "queryOperation":
		return QueryOperation(stub, args)
	default:
		return shim.Error("invalid function name")
	}
}

