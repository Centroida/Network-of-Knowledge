// TOKEN
package main

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
	"fmt"
)

type TokenChaincode struct {
}

const TokenSupply uint64 = 1000

func (t *TokenChaincode) Init(stub shim.ChaincodeStubInterface) peer.Response {
	fn, args := stub.GetFunctionAndParameters()

	if fn != "init" {
		return shim.Error("Function 'init' expected.")
	}

	if len(args) != 0 {
		return shim.Error("No arguments expected.")
	}

	caller, err := CallerCN(stub)
	if err != nil {
		return shim.Error("Couldn't get caller identity.")
	}

	err = t.setBalance(stub, caller, TokenSupply)
	if err != nil {
		return shim.Error("Couldn't set user balance.")
	}

	return shim.Success([]byte("Caller: " + caller))
}

func (t *TokenChaincode) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	fn, args := stub.GetFunctionAndParameters()

	switch fn {
	case "balance":
		return t.balanceAsJSON(stub, args)
	case "transfer":
		return t.transfer(stub, args)
	}

	return shim.Error("Unsupported function: " + fn)
}

func main() {
	err := shim.Start(&TokenChaincode{})
	if err != nil {
		fmt.Errorf("Error starting chaincode: %s", err)
	}
}
