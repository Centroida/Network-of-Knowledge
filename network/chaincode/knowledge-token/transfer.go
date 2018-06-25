package main

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
	"encoding/json"
)

type Transfer struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Value uint64 `json:"value"`
}

func (t *TokenChaincode) transfer(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 1 {
		return shim.Error("Only 1 argument expected.")
	}

	transfer := Transfer{}
	err := json.Unmarshal([]byte(args[0]), &transfer)
	if err != nil {
		return shim.Error("Error unmarshalling tranfer data: " + args[0])
	}

	from, err := CallerCN(stub)
	if err != nil {
		return shim.Error("Error getting caller identity.")
	}

	transfer.From = from

	if transfer.From == transfer.To {
		return shim.Success(nil)
	}

	fromBalance, err := t.getBalance(stub, transfer.From)
	if err != nil {
		return shim.Error("Error getting transaction sender balance.")
	}
	toBalance, err := t.getBalance(stub, transfer.To)
	if err != nil {
		return shim.Error("Error getting transaction receiver balance.")
	}

	if fromBalance < transfer.Value {
		return shim.Error("Insufficient tokens in sender balance.")
	}

	if toBalance + transfer.Value < toBalance {
		return shim.Error("Balance overflow.")
	}

	err = t.setBalance(stub, transfer.From, fromBalance - transfer.Value)
	if err != nil {
		return shim.Error("Error changing sender balance.")
	}
	err = t.setBalance(stub, transfer.To, toBalance + transfer.Value)
	if err != nil {
		return shim.Error("Error changing receiver balance.")
	}

	transferData, _ := json.Marshal(transfer)
	return shim.Success(transferData)
}
