package main

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

func sanitizeArguments(strs []string) error {
	for i, val := range strs {
		if len(val) <= 0 {
			return errors.New("Argument " + strconv.Itoa(i) + " must be a non-empty string")
		}
		if len(val) > 32 {
			return errors.New("Argument " + strconv.Itoa(i) + " must be <= 32 characters")
		}
	}
	return nil
}

func readNode(stub shim.ChaincodeStubInterface, nodeID NodeID) (*KnowledgeNode, error) {
	nodeAsBytes, err := stub.GetState(string(nodeID))
	if err != nil {
		return nil, errors.New("Error")
	}

	var node KnowledgeNode
	json.Unmarshal(nodeAsBytes, &node)
	return &node, nil
}

func saveNode(stub shim.ChaincodeStubInterface, node KnowledgeNode) error {
	jsonAsBytes, _ := json.Marshal(node)
	err := stub.PutState(string(node.ID), jsonAsBytes)
	if err != nil {
		return errors.New("Error")
	}
	return nil
}
