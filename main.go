package main

import (
	"strconv"
	"bytes"
	"crypto/sha256"
	"time"
	"fmt"
)

//区块
type Block struct {
	Timestamp 		int64	//当前时间戳
	Data 			[]byte	//数据
	PrevBlockHash	[]byte	//前一个区块的hash值
	Hash			[]byte	//当前区块的hash值
}

func (b *Block) SetHash() {
	timestamp := []byte(strconv.FormatInt(b.Timestamp, 10))
	headers := bytes.Join([][]byte{b.PrevBlockHash, b.Data, timestamp}, []byte{})
	hash := sha256.Sum256(headers)

	b.Hash = hash[:]	//det hash
}

//创建区块
func NewBlock(data string, prevBlockHash []byte) *Block {
	block := &Block{time.Now().Unix(), []byte(data), prevBlockHash, []byte{}}
	block.SetHash()
	return block
}


//区块链
type BlockChain struct {
	blocks []*Block
}

//向区块链中添加区块
func (bc *BlockChain) AddBlock(data string) {
	prevBlock := bc.blocks[len(bc.blocks) - 1]
	newBlock := NewBlock(data, prevBlock.Hash)
	bc.blocks = append(bc.blocks, newBlock)
}

//创建“创始块”（Genesis Block）
func NewGenesisBlock() *Block {
	return NewBlock("Frank's Genesis Block", []byte{})
}

//用创始块创建一个区块链的函数
func NewBlockChain() *BlockChain {
	return &BlockChain{[]*Block{NewGenesisBlock()}}
}

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