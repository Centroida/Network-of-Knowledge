package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

func addNode(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1.")
	}

	var node KnowledgeNode
	json.Unmarshal([]byte(args[0]), &node)
	fmt.Println("Adding node:", node.ID)

	// Checks

	//

	jsonAsBytes, _ := json.Marshal(node)
	err = stub.PutState(string(node.ID), jsonAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("Done...")
	return shim.Success(nil)
}
