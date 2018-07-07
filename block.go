package main

import (
	"time"
	"bytes"
	"encoding/gob"
)

//区块
type Block struct {
	Timestamp 		int64	//当前时间戳
	Data 			[]byte	//数据
	PrevBlockHash	[]byte	//前一个区块的hash值
	Hash			[]byte	//当前区块的hash值
	Nonce         	int		//nonce计数器存储为block的一部分，要用于手工的验证
}

/* 从part_2开始废弃使用，已在NewBlock()中重新设置 */
//func (b *Block) SetHash() {
//	timestamp := []byte(strconv.FormatInt(b.Timestamp, 10))
//	headers := bytes.Join([][]byte{b.PrevBlockHash, b.Data, timestamp}, []byte{})
//	hash := sha256.Sum256(headers)
//
//	b.Hash = hash[:]	//det hash
//}

// 创建区块
func NewBlock(data string, prevBlockHash []byte) *Block {
	block := &Block{time.Now().Unix(), []byte(data), prevBlockHash, []byte{}, 0}
	//block.SetHash()	//废弃,由下面的代替
	// Pow
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

// 序列化
func (b *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(b)
	logErr(err)

	return result.Bytes()
}

// 反序列化
func DeseralizeBlock(d []byte) *Block {
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block)
	logErr(err)

	return &block
}