package main

import (
	"fmt"
	"strconv"
)

//test
func main() {
	bc := NewBlockChain()

	bc.AddBlock("Frank")
	bc.AddBlo
	ck("Ghost")

	for _, block := range bc.blocks {
		fmt.Printf("Prev.hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)

		pow := NewProofOfWork(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()
	}
}