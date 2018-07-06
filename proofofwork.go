package main

import (
	"math/big"
	"bytes"
	"fmt"
	"math"
	"crypto/sha256"
)

var (
	maxNonce = math.MaxInt64
)

//设置挖矿（区块产生）的难度：
const targetBits = 24

//proof of work
type ProofOfWork struct {
	block 	*Block
	target 	*big.Int
}

//新建PoW
func NewProofOfWork(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256 - targetBits))

	pow := &ProofOfWork{b, target}
	return pow
}

//准备需要计算hash值的数据
// nonce变量在这里就是Hashcash算法中描述的counter（计数器），它是一条加密后的数据。
func (pow *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.block.PrevBlockHash,
			pow.block.Data,
			IntToHex(pow.block.Timestamp),
			IntToHex(int64(targetBits)),
			IntToHex(int64(nonce)),
		},
		[]byte{},
	)
	return data
}


//PoW算法的核心部分
func (pow *ProofOfWork) Run() (int, []byte) {
	var hashInt big.Int	//整型数代表
	var hash [32]byte
	nonce := 0			//counter

	fmt.Printf("Mining the block containing \"%s\"\n", pow.block.Data)
	for nonce < maxNonce {
		data := pow.prepareData(nonce)	//1.准备数据
		hash = sha256.Sum256(data)		//2.用SHA-256算法计算该数据的哈希值
		fmt.Printf("\r%x", hash)
		hashInt.SetBytes(hash[:])		//3.将哈希值转换成Big整型数据

		if hashInt.Cmp(pow.target) == -1 {	//4.将转换后的哈希值与target进行比较

			break
		} else {
			nonce++
		}
	}
	fmt.Print("\n\n")

	return nonce, hash[:]
}

// 实现工作证明的验证
func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int

	data := pow.prepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	isValid := hashInt.Cmp(pow.target) == -1

	return isValid
}