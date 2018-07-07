package main

import "time"

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
	block := &Block{time.Now().Unix(), []byte(data), prevBlockHash, []byte{}, 0}
	block.SetHash()

	return block
}
