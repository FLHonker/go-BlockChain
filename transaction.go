package main

import (
	"bytes"
	"encoding/gob"
	"log"
	"crypto/sha256"
	"fmt"
	"encoding/hex"
)

const subsidy = 10

// Transaction represents a Bitcoin transaction
type Transaction struct {
	ID 		[]byte
	Vin 	[]TXInput
	Vout 	[]TXOutput
}

// TXInput represents a transaction input
type TXInput struct {
	Txid		[]byte	//交易记录的ID
	Vout		int		//交易记录当中输出的索引
	ScriptSig 	string	//用户定义的钱包地址,向输出的ScriptPubKey字段中提供数据的脚本
}

// TXOutput represents a transaction output
type TXOutput struct {
	Value 			int		//保存satoshis（聪）的数量,而不是比特币（BTC1）的数量
	ScriptPubKey 	string	//口令,保存便意味着用一段口令将它锁定
}

// IsCoinbase checks whether the transaction is coinbase
func (tx Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}

// SetID sets ID of a transaction
func (tx *Transaction) SetID() {
	var encoded bytes.Buffer
	var hash 	[32]byte

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}
	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]
}

// CanUnlockOutputWith checks whether the address initiated the transaction
func (in *TXInput) CanUnlockOutputWith(unlockingData string) bool {
	return in.ScriptSig == unlockingData
}

// CanBeUnlockedWith checks if the output can be unlocked with the provided data
func (out *TXOutput) CanBeUnlockedWith(unlockingData string) bool {
	return out.ScriptPubKey == unlockingData
}

// NewCoinbaseTX creates a new coinbase transaction
func NewCoinbaseTX(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}

	txin := TXInput{[]byte{}, -1, data}
	txout := TXOutput{subsidy, to}
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{txout}}
	tx.SetID()

	return &tx
}

// NewUTXOTransaction creates a new transaction
func NewUTXOTransaction(from, to string, amount int, bc *BlockChain) *Transaction {
	var inputs	[]TXInput
	var outputs	[]TXOutput

	acc, validOutputs := bc.FindSpendableOutputs(from, amout)
	if acc < amount {
		log.Panic("ERROR: Not enough funds.")
	}

	//Build a list of inputs
	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		if err != nil {
			log.Panic(err)
		}

		for _, out := range outs {
			input := TXInput{txID, out, from}
			inputs = append(inputs, input)
		}
	}

	//Build a list of outputs
	outputs := append(outputs, TXOutput{amount, to})
	if acc > amount {
		outputs = append(outputs, TXOutput{acc - amount, from}) //a change
	}

	tx := Transaction{nil, inputs, outputs}
	tx.SetID()

	return &tx
}
