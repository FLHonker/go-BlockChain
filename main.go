package main

import (
	"fmt"
)

//test
func main() {
	bc := NewBlockChain()
	bc.AddBlock("Frank")
	bc.AddBlock("Ghost")

	for _, block := range bc.blocks {
		fmt.Printf("Prev.hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Println()
	}
}