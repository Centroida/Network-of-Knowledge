package main

import (
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

func distribute(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var nodeID NodeID
	var tokens float64
	var err error

	fmt.Println("Validate input...")
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments.")
	}

	nodeID = NodeID(args[0])

	tokens, err = strconv.ParseFloat(args[1], 64)
	if err != nil {
		return shim.Error("Failed to convert to float")
	}

	node, _ := readNode(stub, nodeID)
	fmt.Println("Done...")
	fmt.Println("Start value distribution...")
	distributeTokens(stub, *node, tokens)
	fmt.Println("Done.")

	return shim.Success(nil)
}

func distributeTokens(stub shim.ChaincodeStubInterface, node KnowledgeNode, tokens float64) {
	queue := make([]KnowledgeNode, 0)
	distTo := make(map[NodeID]int)
	tokensMap := make(map[NodeID]float64)

	tokensMap[node.ID] = tokens
	distTo[node.ID] = 0
	queue = append(queue, node)
	bfsStep(stub, tokensMap, queue, distTo)
}

func bfsStep(stub shim.ChaincodeStubInterface, tokens map[NodeID]float64, queue []KnowledgeNode, distTo map[NodeID]int) {

	if len(queue) == 0 {
		return
	}

	currentNode := queue[0]
	queue = queue[1:]

	if stopCondition(tokens[currentNode.ID], distTo[currentNode.ID]) {
		printTransfer(currentNode.ID, tokens[currentNode.ID])
		currentNode.Tokens += tokens[currentNode.ID]
		saveNode(stub, currentNode)
		return
	}

	distrNodes := make([]NodeID, 0)
	distrWeights := make([]float64, 0)
	for _, ref := range currentNode.References {
		refNode, _ := readNode(stub, ref.ID)
		if _, ok := distTo[refNode.ID]; !ok {
			distrNodes = append(distrNodes, refNode.ID)
			distrWeights = append(distrWeights, float64(ref.Weight))
			queue = append(queue, *refNode)
			distTo[refNode.ID] = distTo[currentNode.ID] + 1
		}
	}

	if len(distrNodes) == 0 {
		printTransfer(currentNode.ID, tokens[currentNode.ID])
		currentNode.Tokens += tokens[currentNode.ID]
		saveNode(stub, currentNode)
		bfsStep(stub, tokens, queue, distTo)
		return
	}

	saveTokens, distrTokensValue := splitTokens(tokens[currentNode.ID], 0.7)

	printTransfer(currentNode.ID, saveTokens)
	currentNode.Tokens += saveTokens
	saveNode(stub, currentNode)

	tokensDistr := computeTokensDistr(distrWeights, distrTokensValue)

	for i, w := range tokensDistr {
		tokens[distrNodes[i]] = w
	}
	bfsStep(stub, tokens, queue, distTo)
}

func stopCondition(tokens float64, depth int) bool {
	return false
}

func splitTokens(tokens float64, weight float64) (float64, float64) {
	a := tokens * weight
	b := tokens - a
	return a, b
}

func computeTokensDistr(weights []float64, tokens float64) []float64 {
	// Compute the sum of the weights
	sum := 0.0
	for _, w := range weights {
		sum += w
	}

	// Normalize the weights
	for i, w := range weights {
		weights[i] = w / sum
	}

	// Compute the distribution.
	tokensDistr := make([]float64, len(weights))
	for i := range weights {
		tokensDistr[i] = float64(tokens) * weights[i]
	}

	return tokensDistr
}

func printTransfer(id NodeID, tokens float64) {
	fmt.Printf("%s <--- %f\n", id, tokens)
}
