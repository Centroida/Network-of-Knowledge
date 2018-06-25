package main

import (
	"encoding/binary"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
	"encoding/json"
)

type Balance struct {
	User  string `json:"user"`
	Value uint64 `json:"value"`
}


const BalanceIndex = "user~balance"

func (t *TokenChaincode) setBalance(stub shim.ChaincodeStubInterface, user string, balance uint64) error {
	key, _ := stub.CreateCompositeKey(BalanceIndex, []string{user})
	data := make([]byte, 8)
	binary.LittleEndian.PutUint64(data, balance)
	return stub.PutState(key, data)
}

func (t *TokenChaincode) getBalance(stub shim.ChaincodeStubInterface, user string) (uint64, error) {
	key, _ := stub.CreateCompositeKey(BalanceIndex, []string{user})
	data, err := stub.GetState(key)
	if err != nil {
		return 0, err
	}

	if data == nil {
		return 0, nil
	}

	return binary.LittleEndian.Uint64(data), nil
}

func (t *TokenChaincode) balanceAsJSON(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 1 {
		return shim.Error("Only 1 argument expected.")
	}

	balance := Balance{}
	err := json.Unmarshal([]byte(args[0]), &balance)
	if err != nil {
		return shim.Error("Error unmarshalling balance query input: " + args[0])
	}

	value, err := t.getBalance(stub, balance.User)
	if err != nil {
		return shim.Error("Error getting balance.")
	}

	balance.Value = value
	balanceJSON, err := json.Marshal(balance)
	if err != nil {
		return shim.Error("Error marshalling balance to json.")
	}

	return shim.Success(balanceJSON)
}
