package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hyperledger/fabric-chaincode-go/shim"
)

func WriteLedger(obj interface{}, stub shim.ChaincodeStubInterface, objectType string, keys []string) error {
	var key string
	if val, err := stub.CreateCompositeKey(objectType, keys); err != nil {
		return errors.New(fmt.Sprintf("%s-create key error-%s", objectType, err))
	} else {
		key = val
	}
	bytes, err := json.Marshal(obj)
	if err != nil {
		return errors.New(fmt.Sprintf("%s-sequential json value error-%s", objectType, err))
	}
	if err := stub.PutState(key, bytes); err != nil {
		return errors.New(fmt.Sprintf("%s-write fabric ledger error: %s", objectType, err))
	}
	return nil
}
