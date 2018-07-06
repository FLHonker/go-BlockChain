package main

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
