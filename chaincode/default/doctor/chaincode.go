package main

import (
	"bytes"
	"encoding/json"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

type Contract struct {
	Token          string `json:"token"`
	PatientAddress string `json:"patientAddress"`
	MedicationSold string `json:"medicationSold"`
}

func (cn *Contract) Init(stub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

func (cn *Contract) Invoke(stub shim.ChaincodeStubInterface) sc.Response {
	fcn, params := stub.GetFunctionAndParameters()

	switch fcn {
	case "send":
		return cn.Send(stub, params)
	case "query":
		return cn.QueryByPatientAddress(stub, params)
	}

	return shim.Success(nil)
}

func (cn *Contract) Send(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 3 {
		return shim.Error("Wrong number of arguments! Expected 3 but got " + strconv.Itoa(len(args)))
	}

	contract := &Contract{
		Token:          args[0],
		PatientAddress: args[1],
		MedicationSold: args[2],
	}

	contractBytes, err := json.Marshal(contract)

	if err != nil {
		return shim.Error("Error serializing data: " + err.Error())
	}

	err = stub.PutState(args[0], contractBytes)

	if err != nil {
		return shim.Error("Error putting contract on state: " + err.Error())
	}

	patientTokenKey, err := stub.CreateCompositeKey("patient~token", []string{contract.PatientAddress, contract.Token})

	if err != nil {
		return shim.Error("Error associating token to patient: " + err.Error())
	}

	err = stub.PutState(patientTokenKey, []byte{0x00})

	if err != nil {
		return shim.Error("Error associating token to patient: " + err.Error())
	}

	return shim.Success(contractBytes)
}

func (cn *Contract) QueryByPatientAddress(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 1 {
		return shim.Error("Wrong number of arguments! Expected 1 but got " + strconv.Itoa(len(args)))
	}

	indexName := "patient~token"

	iterator, err := stub.GetStateByPartialCompositeKey(indexName, []string{args[0]})

	if err != nil {
		return shim.Error("Error getting state: " + err.Error())
	}

	var buffer bytes.Buffer
	buffer.WriteString("[")

	firstToken := true

	for iterator.HasNext() {
		compositeKey, err := iterator.Next()

		if err != nil {
			return shim.Error("Error getting state: " + err.Error())
		}

		_, compositeKeyArgs, err := stub.SplitCompositeKey(compositeKey.Key)

		tokenID := compositeKeyArgs[1]

		contractBytes, err := stub.GetState(tokenID)

		if err != nil {
			return shim.Error("Error getting state: " + err.Error())
		}

		if !firstToken {
			buffer.WriteString(",")
		}

		buffer.WriteString("{\"Token\":")
		buffer.WriteString("\"")
		buffer.WriteString(tokenID)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Contract\":")
		buffer.WriteString(string(contractBytes))
		buffer.WriteString("}")

		firstToken = false
	}
	buffer.WriteString("]")

	return shim.Success(buffer.Bytes())
}
