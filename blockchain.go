package main

import (
	"github.com/boltdb/bolt"
	"fmt"
)

const dbFile = "blockchain.db"
const blocksBucket = "blocks"

//区块链
type BlockChain struct {
	tip []byte
	db *bolt.DB
}

// 区块链迭代器
type BlockChainIterator struct {
	currentHash	[]byte
	db 			*bolt.DB
}


//向区块链中添加区块
func (bc *BlockChain) AddBlock(data string) {
	var lastHash []byte

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))
		return nil
	})
	logErr(err)

	newBlock := NewBlock(data, lastHash)

	err = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err := b.Put(newBlock.Hash, newBlock.Serialize())
		logErr(err)

		err = b.Put([]byte("l"), newBlock.Hash)
		logErr(err)

		bc.tip = newBlock.Hash

		return nil
	})
}

//创建“创始块”（Genesis Block）
func NewGenesisBlock() *Block {
	return NewBlock("Frank's Genesis Block", []byte{})
}

//用创始块创建一个区块链的函数
func NewBlockChain() *BlockChain {
	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)	//如果文件不存在，它也不会报错

	logErr(err)

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		if b == nil {
			fmt.Println("No existing blockchain found.Creating a new one...")
			genesis := NewGenesisBlock()

			b, err := tx.CreateBucket([]byte(blocksBucket))
			logErr(err)

			err = b.Put(genesis.Hash, genesis.Serialize())
			logErr(err)

			err = b.Put([]byte("l"), genesis.Hash)
			logErr(err)

			tip = genesis.Hash
		} else {
			tip = b.Get([]byte("l"))
		}

		return nil
	})

	logErr(err)

	bc := BlockChain{tip, db}

	return &bc
}

//New Iterator...
func (bc *BlockChain) Iterator() *BlockChainIterator {
	bci := &BlockChainIterator{bc.tip, bc.db}

	return bci
}

// BlockchainIterator is used to iterate over blockchain blocks
func (i *BlockChainIterator) Next() *Block {
	var block *Block

	err := i.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get([]byte(i.currentHash))
		block = DeseralizeBlock(encodedBlock)

		return nil
	})

	logErr(err)

	i.currentHash = block.PrevBlockHash

	return block
}