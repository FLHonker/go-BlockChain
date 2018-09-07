---
title: "用Go打造区块链（4）—— 交易记录（一）"
date: 2018-07-06T22:25:04+08:00
bigimg: [{src: "https://res.cloudinary.com/flhonker/image/upload/flhonker-hugo/share_img/post-bg-yourname.jpg", desc: "WHUT|Wuhan|Jul 6,2018"}]
notoc: true
draft: false
categories: "BlockChain"
tags: ["go","blockchain"]
---

这一系列的文章是由[Ivan Kuznetsov][1]所写，第一篇文章的翻译稿由李笑来在其公众号学习学习再学习首发，本人觉得是一个结合Go语言学习区块链技术的好资料，后面将用自己的语言翻译一遍，从第一篇开始，顺便对Go语言以及区块链有一个初步的认识。

## 1 介绍（Introduction）

交易记录是比特币的核心，用区块链的目的是想以一种安全和可靠的方式来存储交易记录，使得无人可以在它们被创建以后再修改它们。今天我们将开始实现交易记录。不过这是一个非常大的话题，我将把这部分内容分成两部分：在这部分当中，我们会实现交易记录的通用机制，在第二部分会实现具体细节。

然后，因为代码改动非常大，对所有代码进行描述变得不太有意义。要查阅所有的代码变得可以[点击这里][2]。

## 2 这里没有调羹（There is no spoon）

假如你曾经开发过web应用，为了实现支付你应该会在数据库中创建这样的一些表：账户（accounts）和交易记录（transactions）。一个账户会存储每个用户的信息，包括他们的个人信息和资产负债表，而一个交易记录会存储货币如何从一个账户转到另外一个账户。在比特币当中，支付是采用完全不同的方式实现的。主要特点如下：

    * 没有账户
    * 没有资产负债表
    * 没有地址
    * 没有货币
    * 没有支付方和接收方

因为区块链是一个公共、开放的数据库，我们不打算存储钱包拥有者的敏感信息。币也不集中在某一个账户。交易记录也把钱从一个账户转移到另一个账户。也没有账户资产负债情况的信息和特性。仅仅只有交易记录（transactions）。然后在交易记录当中究竟有什么呢？

## 3 比特币交易记录（Bitcoin Transaction）

一条交易记录是所有输入（inputs）和输出（outputs）的合并：
```go
type Transaction struct {
	ID   []byte
	Vin  []TXInput
	Vout []TXOutput
}
```
一条新的交易记录的所有输入与之前交易记录的所有输出相对应（有一个例外，一会儿我们会讨论）。输出是实际存储比特币的地方。下面的示意图展示了交易记录之间的内部关系：

![transaction][3]

注意以下几点：

1. 有一些输出并不与输入相对应；
2. 在一个交易记录当中，输入可以与来自不同交易记录的输出相对应；
3. 一个输入只能对于一个输出。

在这边文章当中，我们将会用到这样的名词：“金钱（money）”，“货币（coins）”，“消费（spend）”，“发送（send）”，“账户（account）”等。但是在比特币当中没有这样对于的概念。交易记录通过一段脚本将价值锁定，并且只能由锁定它们的人来解锁。



## 4 交易记录输出（Transaction Outputs）

让我们从输出开始：
```go
type TXOutput struct {
	Value 			int		//保存satoshis（聪）的数量,而不是比特币（BTC1）的数量
	ScriptPubKey 	string	//口令,保存便意味着用一段口令将它锁定
}
```
事实上，这里的输出存储着“币值（coins）”（可以参阅上面的 Value 变量）。保存便意味着用一段口令将它锁定， 而这段口令保存在 ScriptPubKey中。在比特币内部使用一个叫做Script的脚本语言，用来定义输出的锁定和解锁逻辑。这个脚本语言非常接近底层（刻意如此设计，是为了避免可能的滥用和攻击），在此我们不对它进行详细讨论。你可以[点击这里][4]查看详细的解释。

> 在比特币中，value 字段用来保存 satoshis（聪）的数量，而不是比特币（BTC1）的数量。一比特币等于1000万个聪，聪目前是比特币系统中最小的货币单位（就像美分在美元系统中的地位一样）。

由于我们没有实现任何地址相关的内容，目前我们就避免讨论。ScriptPubKey 字段会存储一段随机字符串（用户自定义的钱包地址）

顺便提一句，因为有这样的脚本语言的存在意味着比特币也能够用作智能合约平台。

输出的一个重要特性是`不可分割性`，意味着你能够提及其值的一部分。当一个输出在一个新的交易记录当中被提到，它会花掉它的全部。当它的价值比所需要的大，会产生找零然后返回给发出的人。这与现实生活中的场景类似，当你为某价值1元的东西支付5元会得到4元的找零。

## 5 交易记录输入（Transaction Inputs）

然后接下来是输入（inputs）
```go
type TXInput struct {
	Txid		[]byte	//交易记录的ID
	Vout		int		//交易记录当中输出的索引
	ScriptSig 	string	//用户定义的钱包地址,向输出的ScriptPubKey字段中提供数据的脚本
}
```
之前已经提到，一个输入对应前一个输出：Txid 字段存储这样一条交易记录的ID，Vout 字段存储交易记录当中输出的索引。ScriptSig 字段是一段向输出的 ScriptPubKey 字段中提供数据的脚本。如果数据正确的话，输出将会被解锁，然后其含的价值（value）可以用来产生新的输出；如果不正确的话，输出将无法被输入引用，无法建立连接。这个机制是为了保证用户不能花属于别人的比特币。

再一次的，因为我们没有实现地址，在我们的实现当中 ScriptSig 字段只是一段随机字符串保存用户定义的钱包地址。我们将在下一篇文章当中实现公钥（public keys）和签名（signatures）检查。

让我们先总结一下：输出是“币”存储的地方。每一个输出会带一段解锁的脚本（字符串），决定了解锁输出的逻辑。每一个新的交易记录只是有一个输入和一个输出。一个输入对应一个从前一个交易记录（transaction）来的输出并提供用于解锁（unlocking）输出的数据以便使用输出中的币值（value）来创建新的输出（outputs）。

但是哪个先出现呢：输入还是输出？

## 6 “先有鸡还是先有蛋”（The egg）

在比特币当中，蛋是在鸡之前出现的。“输入对应输出”的逻辑与“鸡或者蛋”的场景类似：输入产生输出，输出让输入变得可能。在比特币当中，输出在输入之前出现。

当一个矿工开始挖坑（mining a block），将在区块当中添加`币基交易记录（coinbase transaction）`。**币基交易记录是一种特殊的交易记录，并不需要已经存在的输出就能够产生，它可以从无直接创造出输出，相当于不需要鸡的蛋。这是矿工在挖新区块时获得的奖励。**

正如你所知，在区块链当中有一个创世区块。正是这个创世区块在区块链中产生最新的输出。并不需要已有的输出，因为也没有已存在的交易记录和这样的输出可供使用。
```go
func NewCoinbaseTX(to, data string)  *Transaction {
    if data == "" {
        data = fmt.Sprintf("Reward to '%s'", to)
    }

    txin := TXInput{[]byte{}, -1, data}
    txout := TXOutput{subsidy, to}
    tx := Transaction{nil, []TXInput{txin}, []TXOutput{txout}}
    tx.SetID()

    return &tx
}
```
一个币基交易记录只有一个输入。在我们的实现当中它的Txid 字段是空的，Vout 字段为-1。而且，币基交易记录的ScriptSig 字段并不保存脚本。相应的，存储随机数据。

> 在比特币当中，最早的币基交易记录包含以下信息：“The Times 03/Jan/2009 Chancellor on brink of second bailout for banks”。你可以[点击这里][5]自我查阅。

subsidy 字段是奖励的数量。在比特币当中，这个数据不在任何地方保存，只是根据总的区块数进行计算：区块数除以210000。挖到创世区块产生50个比特币，然后每 210000 个区块奖励减半。在我们的实现当中，我们将奖励设定为常数（只是目前是这个样子的）。

## 7 在区块链中存储交易记录（Storing Transactions in Blockchain）

从此以后，每一个区块必须存储至少一条交易记录，并且没有交易记录的挖坑也变得不再可能。这意味着我们将从 Block 结构体中移除Data 字段，添加transactions 代替：
```go
//区块
type Block struct {
	Timestamp 		int64	//当前时间戳
	Transactions	[]*Transaction	//交易记录
	PrevBlockHash	[]byte	//前一个区块的hash值
	Hash			[]byte	//当前区块的hash值
	Nonce         	int		//nonce计数器存储为block的一部分，要用于手工的验证
}
```
NewBlock 和 NewGenesisBlock 也必须相应地进行修改：
```go
// block.go

// 创建区块
func NewBlock(transactions []*Transaction, prevBlockHash []byte) *Block {
	block := &Block{time.Now().Unix(), transactions, prevBlockHash, []byte{}, 0}
	// Pow
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

// NewGenesisBlock creates and returns genesis Block
func NewGenesisBlock(coinbase *Transaction) *Block {
	return NewBlock([]*Transaction{coinbase}, []byte{})
}
```
下一步需要改变的是修改区块链的创建方式：
```go
// blockchain.go

func CreateBlockchain(address string) *BlockChain {
	if dbExists() {
		fmt.Println("Blockchain already exists.")
		os.Exit(1)
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	logErr(err)

	err = db.Update(func(tx *bolt.Tx) error {
		cbtx := NewCoinbaseTX(address, genesisCoinbaseData)
		genesis := NewGenesisBlock(cbtx)

		b, err := tx.CreateBucket([]byte(blocksBucket))
		logErr(err)

		err = b.Put(genesis.Hash, genesis.Serialize())
		logErr(err)

		err = b.Put([]byte("l"), genesis.Hash)
		logErr(err)
		tip = genesis.Hash

		return nil
	})

	logErr(err)
	bc := BlockChain{tip, db}

	return &bc
}
```
现在，这个函数将接收一个会收到挖到创世区块奖励的地址。

## 8 工作证明（Proof-of-Work）

工作证明算法（PoW）必须考虑区块链中存储的交易记录，为了确保保存交易记录的区块链的一致性和可靠性。因此，现在我们必须修改 ProofOfWork.prepareData 方法：
```go
// proofofwork.go

func (pow *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.block.PrevBlockHash,
			pow.block.HashTransactions(), // This line was changed
			IntToHex(pow.block.Timestamp),
			IntToHex(int64(targetBits)),
			IntToHex(int64(nonce)),
		},
		[]byte{},
	)

	return data
}
```
我们现在用pow.block.HashTransactions() 代替pow.block.Data：
```go
func (b *Block) HashTransactions() []byte {
	var txHashes [][]byte
	var txHash [32]byte

	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.ID)
	}
	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))

	return txHash[:]
}
```

再一次，我们用哈希来作为一种提供数据唯一标示的工具。我们要一个区块中的所有交易记录都让一个唯一的哈希值来标示。为了达到这个目的，我们获取了每一条交易记录的哈希，然后将它们串联起来，最后获得串联后数据的哈希值。

> 比特币用一个更加复杂的技术：它采用[Merkle tree][[6]来组织一个区块中的所有交易记录，然后用树的根哈希值来确保PoW系统的运行。这种方法能够让我们快速的检验一个区块是否包含确定的交易记录，只要有根哈希值而不需要下载所有的交易记录。

让我们检验一下到目前为止一切照常进行：
```Bash
./goBlockChain createblockchain -address Ivan
00000093450837f8b52b78c25f8163bb6137caf43ff4d9a01d1b731fa8ddcc8a

Done!
```
非常好！我们获得了第一个挖矿奖励。但是我们如何检查余额呢？

## 9 未花费交易记录输出（Unspent Transaction Outputs）

我们需要找出所有的未花费交易记录输出（UTXO）。未花费相当于这些输出并有与任何输入有关联。在上面的示意图中，以下为未花费交易记录：

    1. tx0, output 1;
    2. tx1, output 0;
    3. tx3, output 0;
    4. tx4, output 0.

当然，当我们检查余额时，我们并不需要全部，值需要那些我们有私钥可以解锁的部分。（目前我们还没有实现私钥，将用用户定义的地址来代替）。首先，让我们在输入和输出上定义锁定-解锁方法：
```go
// transaction.go

func (in *TXInput) CanUnlockOutputWith(unlockingData string) bool {
	return in.ScriptSig == unlockingData
}

func (out *TXOutput) CanBeUnlockedWith(unlockingData string) bool {
	return out.ScriptPubKey == unlockingData
}
```

这里，我们只是将unlockingData 与script 字段进行了比较。这部分代码会在后续的文章中所有提升，在我们实现基于私钥的地址以后。

下一步-寻找含有未花费输出的交易记录-实现起来非常的难：
```go
// FindUnspentTransactions 返回一组包含未花费的产出的交易
func (bc *BlockChain) FindUnspentTransactions(address string) []Transaction {
	var unspentTXs []Transaction
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()
	for {
		block := bci.Next()
		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)
		Outputs:
			for outIdx, out := range tx.Vout {
				// Was the output spent?
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}

				if out.CanBeUnlockedWith(address) {
					unspentTXs = append(unspentTXs, *tx)
				}
			}

			if tx.IsCoinbase() == false {
				for _, in := range tx.Vin {
					if in.CanUnlockOutputWith(address) {
						inTxID := hex.EncodeToString(in.Txid)
						spentTXOs[inTxID] = append(spentTXOs[inTxID])
					}
				}
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return unspentTXs
}
```

由于交易记录存在在区块当中，我们必须检查区块链中的每一个区块。我们从输出开始：
```go
if out.CanBeUnlockedWith(address) {
	unspentTXs = append(unspentTXs, tx)
}
```
当一个输出由我们用来选择未花费交易记录的地址上的锁，那么这个输出就是我们想要的。

但是在我们取得它之前我们需要确认它是否已经与一个输入相关联：
```go
if spentTXOs[txID] != nil {
    for _, spentOut := range spentTXOs[txID] {
        if spentOut == outIdx {
            continue Outputs
        }
    }
}
```
我们将忽略哪些已经与输入相关的的输出（它们的价值已经转移到其它的输出，所以不能再统计它们）。检查输出以后，我们手机所有那些可以结果被提供的地址锁上的输出的输入（这对币基交易记录不适用，因为它们不解锁任何输出）：
```go
if tx.IsCoinbase() == false {
    for _, in := range tx.Vin {
        if in.CanUnlockOutputWith(address) {
            inTxID := hex.EncodeToString(in.Txid)
            spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
        }
    }
}
```
下面的函数返回一系列包含未花费输出的交易记录。为了计算余额，我们需要额外的一个以交易记录为参数并只返回输出的函数：
```go
func (bc *Blockchain) FindUTXO(address string) []TXOutput {
    var UTXOs []TXOutput
    unspentTransactions := bc.FindUnspentTransactions(address)

    for _, tx := range unspentTransactions {
        for _, out := range tx.Vout {
            if out.CanBeUnlockedWith(address {
                UTXOs = append(UTXOs, out)
            }
        }
    }

    return UTXOs
}
```

就这样，现在我可以实现 getbalance 命令：
```go
func (cli *CLI) getBalance(address string) {
    bc := NewBlockchain(address)
    defer bc.db.Close()

    balance := 0
    UTXOs := bc.FindUTXO(address)

    for _, out := range UTXOs {
        balance += out.Value
    }

    fmt.Printf("Balance of '%s': %d\n", address, balance)
}
```

账户余额就是账户地址所锁定的所有交易记录输出价值的总和。

让我们检查在挖了创世区块的矿以后的余额：
```Bash
$ ./goBlockChain getbalance -address Ivan
Balance of 'Ivan': 10
```
这是我们的第一份钱！

## 10 发送币（Sending Coins）

现在，我们想要给其他人发一些币过去。为此，我们需要创建一个新的交易记录，把它放到区块当中，然后挖坑。到目前为止，我们只实现了币基交易记录（特殊的一种交易记录），现在我们需要一个普通的交易记录：
```go
// transaction.go

func NewUTXOTransaction(from, to string, amount int, bc *BlockChain) *Transaction {
	var inputs	[]TXInput
	var outputs	[]TXOutput

	acc, validOutputs := bc.FindSpendableOutputs(from, amount)
	if acc < amount {
		log.Panic("ERROR: Not enough funds.")
	}

	//Build a list of inputs
	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		logErr(err)

		for _, out := range outs {
			input := TXInput{txID, out, from}
			inputs = append(inputs, input)
		}
	}

	//Build a list of outputs
	outputs = append(outputs, TXOutput{amount, to})
	if acc > amount {
		outputs = append(outputs, TXOutput{acc - amount, from}) //a change
	}

	tx := Transaction{nil, inputs, outputs}
	tx.SetID()

	return &tx
}
```

在创建新的输出之前，我们首先找出所有的未消费输出并且确保它们存有足够的币值。这是FindSpendableOutputs 方法的功能。在这之后，每一个找到的输出创建一个与之对应的输入。然后，我们创建两个输出：

1. 一个用接受者的地址进行锁定。这是实际需要转移到其它地址的币值。
2. 一个用发送者的地址进行锁定。这是找零。当且仅当剩余未花费输出持有的总币值比新的交易记录所要求的多。记住：输出是不可分割的。

FindSpendableOutput 方法是基于我们早先定义的 FindUnspentTransactions 方法：
```go
// blockchain.go

func (bc *BlockChain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	unspentTXs := bc.FindUnspentTransactions(address)
	accumulated := 0

Work:
	for _, tx := range unspentTXs {
		txID := hex.EncodeToString(tx.ID)

		for outIdx, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) && accumulated < amount {
				accumulated += out.Value
				unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)

				if accumulated >= amount {
					break Work
				}
			}
		}
	}

	return accumulated, unspentOutputs
}
```

这个方法遍历所有的未花费交易记录并累计它们的币值。当累计的币值大于或者等于我们所有转移的量时，它停止工作然后返回累计币值（accumulated value）以及按交易记录ID进行分组的输出索引。我们不打算取比我们打算要花费的多。

现在我们可以修改 Blockchain.MineBlock 方法：
```go
//blockchain.go

// MineBlock 使用已有的交易挖掘新的区块
func (bc *BlockChain) MineBlock(transactions []*Transaction) {
	var lastHash []byte

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))
		return nil
	})
	logErr(err)

	newBlock := NewBlock(transactions, lastHash)

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
```

最终，让我们实现 send 命令：
```go
func (cli *CLI) send(from, to string, amount int) {
    bc := NewBlockchain(from)
    defer bc.db.Close()

    tx := NewUTXOTransaction(from, to, amount, bc)
    bc.MineBlock([]*Transaction{tx})
    fmt.Println("Success!")
}
```
发送币意味着创建一个交易记录然后以对一个区块以挖矿的形式将它加入到区块链当中。但是比特币并不立即执行，正如我们所实现的。反而将所有新交易记录放到内存池，当一个矿工准备去挖矿时，它将从内存池中拿走所有的交易记录并创建一个候选区块。当且仅当包含它们的区块被挖出被加入到区块链时所有的交易记录才会被确认。

让我们确认下发送币的功能是否工作正常：
```Bash
$ ./goBlockChain send -from Ivan -to Pedro -amount 6
00000001b56d60f86f72ab2a59fadb197d767b97d4873732be505e0a65cc1e37

Success!

$ ./goBlockChain getbalance -address Ivan
Balance of 'Ivan': 4

$ ./goBlockChain getbalance -address Pedro
Balance of 'Pedro': 6
```

非常好！现在，让我们创建更多的交易记录然后确保从不同的输出发出币也通用工作正常：
```Bash
$ ./goBlockChain send -from Pedro -to Helen -amount 2
00000099938725eb2c7730844b3cd40209d46bce2c2af9d87c2b7611fe9d5bdf

Success!

$ ./goBlockChain send -from Ivan -to Helen -amount 2
000000a2edf94334b1d94f98d22d7e4c973261660397dc7340464f7959a7a9aa

Success!
```

现在 Helen的币被锁定在两个输出当中：一个来自Pedro，另外一个来之Ivan。让我们将它们转移到其它人：
```Bash
$ ./goBlockChain send -from Helen -to Rachel -amount 3
000000c58136cffa669e767b8f881d16e2ede3974d71df43058baaf8c069f1a0

Success!

$ ./goBlockChain getbalance -address Ivan
Balance of 'Ivan': 2

$ ./goBlockChain getbalance -address Pedro
Balance of 'Pedro': 4

$ ./goBlockChain getbalance -address Helen
Balance of 'Helen': 1

$ ./goBlockChain getbalance -address Rachel
Balance of 'Rachel': 3
```

看起来非常棒！让我们测试一个失败案例：
```Bash
$ ./goBlockChain send -from Pedro -to Ivan -amount 5
panic: ERROR: Not enough funds

$ ./goBlockChain getbalance -address Pedro
Balance of 'Pedro': 4

$ ./goBlockChain getbalance -address Ivan
Balance of 'Ivan': 2
```

## 11 结论（Conclusion）

哎呀！这并不容易，但现在我们还是有了交易记录！虽然，一些类比特币的加密币的特性还暂时缺失：

1. 地址。我们并没有实际、私有的地址
2. 奖励。挖矿时绝对无利可图的
3. UTXO集合。获取余额需要扫描整个区块链，当有很多区块时，这会非常耗时。并且，验证后面的交易记录也非常耗时。UTXO集试图解决这些问题并让通过交易记录的运营更加快。
4. 内存池。这是交易记录在被推入区块之前存储交易记录的地方。在我们当前的实现当中，一个区块仅包含一个交易记录，然则这非常的低效。

## 12 链接

1. [Full source codes][2]
2. [Transaction][7]
3. [Merkle tree][6]
4. [Coinbase][8]



[1]:https://link.zhihu.com/?target=https%3A//jeiwan.cc/
[2]:https://github.com/FLHonker/go-BlockChain/tree/part_4
[3]:https://res.cloudinary.com/flhonker/image/upload/githubio/go/goBlockChain/blockchain_part4-transaction.jpg
[4]:https://link.zhihu.com/?target=https%3A//en.bitcoin.it/wiki/Script
[5]:https://link.zhihu.com/?target=https%3A//blockchain.info/tx/4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b%3Fshow_adv%3Dtrue
[6]:https://link.zhihu.com/?target=https%3A//en.wikipedia.org/wiki/Merkle_tree
[7]:https://link.zhihu.com/?target=https%3A//en.bitcoin.it/wiki/Transaction
[8]:https://link.zhihu.com/?target=https%3A//en.bitcoin.it/wiki/Coinbase