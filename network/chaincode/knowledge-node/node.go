package main

import (
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type SimpleChaincode struct {
}

type NodeID string
type UserID string
type W float64

type Owner struct {
	ID     UserID `json:"id"`
	Weight W      `json:"weight"`
}

type Reference struct {
	ID     NodeID `json:"id"`
	Weight W      `json:"weight"`
}

type KnowledgeNode struct {
	ID         NodeID      `json:"id"`
	Ownership  []Owner     `json:"ownership"`
	References []Reference `json:"references"`
	Tokens     float64     `json:"tokens"`
        FileID     string      `json:"fileId"`
}

// Main
func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode - %s", err)
	}
}

// Init - initialize the chaincode
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

// Invoke - Our entry point for Invocations
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println(" ")
	fmt.Println("starting invoke, for - " + function)

	// Handle different functions
	if function == "init" {
		return t.Init(stub)
	} else if function == "query" {
		return t.query(stub, args)
	} else if function == "addNode" {
		return addNode(stub, args)
	} else if function == "distribute" {
		return distribute(stub, args)
	}
	// error out
	fmt.Println("Received unknown invoke function name - " + function)
	return shim.Error("Received unknown invoke function name - '" + function + "'")
}

func (t *SimpleChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("Query Start...")
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1.")
	}

	nodeID := NodeID(args[0])

	node, err := stub.GetState(string(nodeID))
	if err != nil {
		return shim.Error("Failed to get node")
	}
	fmt.Println("Query End...")

	return shim.Success(node)
}
