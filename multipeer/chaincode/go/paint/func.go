package main

import (
	"encoding/json"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer"
)

func QueryAllWork(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	resultIterator, err := stub.GetStateByRange("", "")
	if err != nil {
		return shim.Error("QueryAllWork Failed.")
	}

	defer resultIterator.Close()
	works := make([]Work, 0, 30)
	for resultIterator.HasNext() {
		queryResponse, err := resultIterator.Next()
		if err != nil {
			return shim.Error("QueryAllWork Get BimFile By Iterator Failed.")
		}
		if queryResponse == nil {
			continue
		}
		value := queryResponse.Value
		if value == nil {
			continue
		}

		var work Work
		err = json.Unmarshal(value, &work)
		if err != nil {
			return shim.Error("QueryAllWork Unmarshal is Wrong")
		}
		if work.Hash1 == "" || work.Hash3 == "" || work.Signature == "" {
			continue //跳过查询出的work数据
		}
		works = append(works, work)
	}
	bytes, _ := json.Marshal(works)
	return shim.Success(bytes) //返回的是查询的bytes 在sdk中unmarshal即可得到对应切片数据
}

func PutWork(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 3 {
		return shim.Error("PutWork invalid parameters,it must be 3.")
	}

	work := Work{
		Hash1:     args[0],
		Hash3:     args[1],
		Signature: args[2],
	}

	var workBytes []byte
	var err error
	workBytes, err = json.Marshal(work)
	if err != nil {
		return shim.Error("PutWork json marshal error")
	}

	err = stub.PutState("Hash1:"+work.Hash1, workBytes)
	if err != nil {
		return shim.Error("PutWork put to chain failed")
	}

	tx_id := stub.GetTxID()
	return shim.Success([]byte(tx_id))
}

func QueryWork(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 1 {
		return shim.Error("QueryWork: invalid parameters, it must be 1")
	}
	Bytes, err := stub.GetState("Hash1:" + args[0])
	if err != nil {
		return shim.Error("QueryWork: get Work Failed")
	}
	return shim.Success(Bytes) // 直接返回查询的Bytes sdk进行转化为结构体
}

func QueryAllOperation(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	resultIterator, err := stub.GetStateByRange("", "")
	if err != nil {
		return shim.Error("QueryAllOperation Failed.")
	}

	defer resultIterator.Close()
	operations := make([]Operation, 0, 30)
	for resultIterator.HasNext() {
		queryResponse, err := resultIterator.Next()
		if err != nil {
			return shim.Error("QueryAllOperation Get BimFile By Iterator Failed.")
		}
		if queryResponse == nil {
			continue
		}
		value := queryResponse.Value
		if value == nil {
			continue
		}

		var operation Operation
		err = json.Unmarshal(value, &operation)
		if err != nil {
			return shim.Error("QueryAllOperation Unmarshal is Wrong")
		}
		if operation.OperationID == "" || operation.UserId == "" || operation.UserName == "" ||
			operation.InstitutionName == "" || operation.ImageId == "" || operation.ImageName == "" ||
			operation.OperationType == "" || operation.OperationContent == "" || operation.OperationTime == "" {
			continue //跳过查询出的work数据
		}
		operations = append(operations, operation)
	}
	bytes, _ := json.Marshal(operations)
	return shim.Success(bytes) //返回的是查询的bytes 在sdk中unmarshal即可得到对应切片数据
}

func PutOperation(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 9 {
		return shim.Error(" PutOperation: invalid parameters, it must be 9")
	}
	if args[0] == "" || args[1] == "" || args[2] == "" || args[3] == "" ||
		args[4] == "" || args[5] == "" || args[6] == "" || args[7] == "" || args[8] == "" {
		return shim.Error("PutOperation: input args have empty value")
	}
	operation := Operation{
		OperationID:      args[0],
		UserId:           args[1],
		UserName:         args[2],
		InstitutionName:  args[3],
		ImageId:          args[4],
		ImageName:        args[5],
		OperationType:    args[6],
		OperationContent: args[7],
		OperationTime:    args[8],
	}

	Bytes, err := json.Marshal(operation)
	if err != nil {
		return shim.Error("PutOperation: Marshal Operation Failed")
	}
	err = stub.PutState("Operation:"+operation.OperationID, Bytes)
	if err != nil {
		return shim.Error("PutOperation: Put Operation Failed")
	}
	tx_id := stub.GetTxID()
	return shim.Success([]byte(tx_id))
}

func QueryOperation(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 1 {
		return shim.Error("GetOperation: invalid parameters, it must be 1:OperationId")
	}
	Bytes, err := stub.GetState("Operation:" + args[0])
	if err != nil {
		return shim.Error("GetOperation: get Operation Failed")
	}
	return shim.Success(Bytes) // 直接返回查询的Bytes sdk进行转化为结构体
}
