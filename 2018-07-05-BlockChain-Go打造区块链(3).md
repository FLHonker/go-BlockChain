---
layout:     post
title:      用Go打造区块链（3）—— 数据存储及命令行（CLI）
subtitle:   
date:       2018-07-05
author:     Frank Liu
header-img: img/post-bg-golang1.png
catalog: true
tags:
    - Go
    - BloackChain
---

# 用Go打造区块链（3）——数据存储及命令行（CLI）

这一系列的文章是由[Ivan Kuznetsov][1]所写，第一篇文章的翻译稿由[李笑来][2]在其公众号学习学习再学习首发，本人觉得是一个结合Go语言学习区块链技术的好资料，后面将用自己的语言翻译一遍，从第一篇开始，顺便对Go语言以及区块链有一个初步的认识。

## 1 介绍

到目前为止，我们已经构建了一个带PoW机制的区块链系统，这也使得挖矿变得可能。我们的实现离完整功能的区块链又近了一步，但是依然缺少一些重要的特性。今天开始我们将区块链存储到一个数据库当中，在这之后我们会做一个命令行接口（CLI）来实现与区块链的互动。虽然区块链的核心是分布式数据库，但是目前我们将暂时忽略“分布式”而专注于数据库本身。

本节过后，项目结构为：

![tree][3]

## 2 数据库选择

目前，我们的实现当中并没有数据库；而是在每次运行程序的时候将区块链存储在内存当中，程序结束后所有的数据便消失了。我们无法重复使用一个区块链，也无法与其他人分享。因此后续我们需要将它存储在硬盘当中。

我们需要哪个数据库呢？事实上，任何一个都可以。在[比特币白皮书][4]当中，并没有特别提到需要使用哪一个数据库，这将取决于每个开发者本身。中本聪首发的[比特币内核][5]且也当前比特币的一个参考实现，采用[google/leveldb][6]。这里我们将采用[BoltDB][7]...

## 3 BoltDB

因为：

1. 简单、简洁
2. 用Go实现
3. 不需要运行一个服务
4. 允许用户构建自己需要的数据结构

以下是来自Github上[BoltDB Readme][7]文件的介绍：

> Blot 是一个纯Go的键/值（key/value）存储系统，来自于Howard Chu的LMDB项目。该项目的目的是为那些并不需要像Postgres或者MySQL这样的全功能服务数据库的项目提供一款简单、快速且可靠的数据库。
正因为 Bolt 注定要用在这样一个队功能要求层次较低的项目上，简单是最关键的。API也很小，也仅仅关注于取值和赋值。就这样小而美。
听起来能够完美地满足我们的需求，让我们花几分钟再看一下。

Bolt 是一个key/value数据库，这意味着没有像其它SQL系的关系型数据库管理系统（RDMBS，比如MySQL，PostgreSQL等）的表（table）、行（row）、列（column）等元素。数据以键/值（key/value）的方式进行存储（类似Go语言中的maps数据结构）。key/value对存储在类似于关系型数据库中的表（table）的 buckets 当中，bucket会将相似的数据对进行分组。因此，要想获得value，你需要知道对应的bucket和key。

[BoltDB][7] 的重要特点是无数据类型：key/value是byte类型的数组（array）。而我们要存储Go语言的结构体数据（特别是Block结构体），因此我们需要将结构体序列化（serialize）。实现一个首先将Go语言的结构体转化成byte数组，然后将byte数组恢复成结构体的机制。我们将采用[encoding/gob][8]来完成这一工作，当然，**JSON, XML, Protocol Buffers** 等工具也是可以的，我们之所以采用gob包是因为它足够简单并且是Go语言标准库的一部分。

## 4 数据结构

在实现数据持续存储之前，首先需要确定如何在数据库存储数据。就此，我们将参考[比特币内核][5]的实现方式。

简单地讲，比特币内核用了两个“Buckets”来存储数据：

1. blocks：存储区块链中描述区块的元数据
2. chainstate：存储链的状态，包括所有未成交的交易输出（Outputs）记录和一些元数据

并且，每一个blocks在硬盘中以单独的文件来保存，这样做是为了提高性能：读取一个单独的区块并不需要载入所有（或者部分）区块到内存。不过我们并不这样实现。

在blocks 当中，Key -> Value 数据对如下：

1. 'b' + 32-byte block hash -> block index record
2. 'f' + 4-byte file number -> file information record
3. 'l' -> 4-byte file number: the last block file number used
4. 'R' -> 1-byte boolean: whether we're in the process of reindexing
5. 'F' + 1-byte flag name length + flag name string -> 1 byte boolean: various flags that can be on or off
6. 't' + 32-byte transaction hash -> transaction index record

在chainstate 当中，Key -> Value 数据对如下：

1. 'c' + 32-byte transaction hash -> unspent transaction output record for that transaction
2. 'B' -> 32-byte block hash: the block hash up to which the database represents the unspent transaction outputs

（关于以上数据的详细解释可以访问这个链接[Bitcoin Core 0.11 (ch 2): Data Storage][9]）

因为目前我们还没有交易记录，我们的数据库当中仅有 blocks 这个bucket。并且，正如前面所提到的，我们将全部的数据库存储在一个文件当中，并不将单个区块存储在单个文件当中。因此，我们也没有什么信息是与文件数（file number）有个的，所能够利用的 key -> value 数据对如下：

1. 'b' + 32-byte block-hash -> Block structure (serialized)区块结构体（序列化后的）
2. 'l' -> the hash of the last block in a chain（区块链中最近一个区块的哈希值）

以上是目前开始实现数据存储机制所需要了解的全部内容。

## 5 序列化（Serialization）

接下来，为简化程序结构，错误处理不在冗余，我们在工具包`utils.go`中添加新函数用于错误日志记录：
```go
func logErr(err error)  {
	if err != nil {
		log.Panic(err)
	}
}
```

正如前面所言，BoltDB 数据的值只能是Byte 类型的数组，并且我们想要将Block 结构体的数据存储到数据库当中。我们将采用 encoding/gob 来对结构体进行序列化。

让我们来实现将 Block 序列化的方法（为了简化，将异常处理部分的代码略去）：
```go
func (b *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(b)
    logErr(err)

	return result.Bytes()
}
```
因为这个只是针对 Block 结构体的，所以这里只是一个方法，而不是一个函数（only a method belongs to Block structure particularly, not a general function）。

这一段简单明了：一开始，我们申明了一个bytes.Buffer (bytes 包中用于接收字节读入的缓存）的变量 result，用来存储序列化后的数据；然后初始化了一个gob包的 NewEncoder 然后对区块进行编码（Encode）；最后 变量 result 以 Bytes 的类型返回。

接下来，我们需要一个反序列化函数，在我们将一个Byte 数组传递给它以后返回对应的一个Block 结构体。这不是专属于特定一个结构体或者数据类型的方法，而是一个独立的函数：
```go
func DeserializeBlock(d []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block)

	return &block
}
```
以上是序列化部分的解释和代码。

## 6 数据存储（Persistence）

让我们从`NewBlockchain`函数开始。目前，它创建了一个 Blockchain 结构体的实例（instance）并将创始区块（genesis block）加入链中。以下是我们接下来要做的：

1. 打开一个数据库（DB）文件

2. 检查数据库文件是否已有其他区块链文件存储，是或者否分两种情况不同论述

3. 如果已经有了一个区块链：
    * 创建一个新的 Blockchain 实例
    * 将新建 Blockchain 实例的顶端设置为原数据库当中的区块链最后一个区块的哈希值

4. 假如还有没有区块链：

    * 创建创始区块
    * 将新建的创始区块存储到数据当中
    * 将创始区块的哈希保存为最后区块的哈希
    * 创建一个新的 Blockchain 的实例，然后将其顶端指向创始区块

可以用以下代码来表示：
```go
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
```

让我们一条一条地过这段代码。

```go
db, err := bolt.Open(dbFile, 0600, nil)
```
这是打开BoltDB 数据文件的标准方式。需要注意的是如果文件不存在，它也不会报错。
```go
err = db.Update(func(tx *bolt.Tx) error {
...
})
```
在BoltDB当中，对数据库的操作是通过一个个记录（transactions）来实现的。在此有两种记录模式，只读模式和读写模式。这里我们通过db.Updata(...) 以读写的模式打开，因为我们会将创始区块加入到数据库文件当中。
```go
b := tx.Bucket([]byte(blocksBucket))

if b == nil {
	genesis := NewGenesisBlock()
	b, err := tx.CreateBucket([]byte(blocksBucket))
	err = b.Put(genesis.Hash, genesis.Serialize())
	err = b.Put([]byte("l"), genesis.Hash)
	tip = genesis.Hash
} else {
	tip = b.Get([]byte("l"))
}
```
这段代码是整个函数的核心，这里我们获取了用于存储区块的 bucket：假如它已经存在，我们读取它的 “l”值；如果还不存在，我们创建创始区块，创建 bucket，将创始区块保存到新建的 bucket当红，然后将“l”值更新为区块链当中最近区块的哈希值。

特别注意到创建 Blockchain 的新方式：
```go
bc := Blockchain{tip, db}
```
我们并不将所有的区块存储进去，而仅仅是存储区块链的顶端。同时，也保存一个数据库文件的链接（DB），因为我们希望打开一次以后只要程序在运行就一直保持打开。这样，现在 Blockchain 结构体看起来像这个样子：
```go
type Blockchain struct {
	tip []byte
	db  *bolt.DB
}
```

下一件事情我们需要更新修改的是 AddBlock 方法：往一个区块链当中增加一个区块已经不像是像数组添加一个新的元素那么简单了。从现在开始我将区块保存在数据库文件（DB）当中：
```go
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
```
让我们一段一段过代码：
```go
err := bc.db.View(func(tx *bolt.Tx) error {
	b := tx.Bucket([]byte(blocksBucket))
	lastHash = b.Get([]byte("l"))

	return nil
})
```
这是BoltDB 另外一种只读的记录（操作数据库文件）的模式。这里我们从数据库文件获得了最近区块的哈希然后用它来挖一个新的区块哈希。
```go
newBlock := NewBlock(data, lastHash)
b := tx.Bucket([]byte(blocksBucket))
err := b.Put(newBlock.Hash, newBlock.Serialize())
err = b.Put([]byte("l"), newBlock.Hash)
bc.tip = newBlock.Hash
```
挖到一个新的区块以后，我们将其序列化后的数据保存到数据库当中然后更新其中的“l”值，这时“l”保存新区块的哈希值。

完成了！看起来不是太难吧？


## 7 查看区块链（Inspecting Blockchain）

所有新的区块现在都保存在数据当中，于是我们可以重复打开一个区块链然后往里面添加新的区块。但是，实现了存储以后，我们失去了一个新的特性：因为我们已经不再将区块存储在数组当中，于是便不能将区块链打印出来了。让我们来消除这个缺陷吧。

BoltDB 允许对bucket 中的key 进行遍历，但是keys是以字节顺序进行保存的，而我们想将区块按区块链中的顺序进行打印。并且，我们也不想将所有的区块载入到内存当中（我们的区块链文件会非常大！。。。或者让我假装它会很大），我们将逐个读取它们。为了这个目的，我们需要一个`区块链遍历器（iterator）`：
```go
type BlockchainIterator struct {
	currentHash []byte
	db          *bolt.DB
}
```
让我们需要遍历区块链中的区块的时候就创建一个遍历器，并且BlockchainIterator 结构体将存储当前遍历为止的区块哈希值和数据库文件的连接。因为后面这个因素，一个遍历器逻辑上将依附于一个区块链（它是一个存储数据库连接的区块链实例），因此，它将在一个属于 Blockchain 结构体的方法中创建：
```go
func (bc *Blockchain) Iterator() *BlockchainIterator {
	bci := &BlockchainIterator{bc.tip, bc.db}

	return bci
}
```
注意到遍历器初始化的时候指向了区块链的顶端，这样可以从顶部到底部从最新到最老的顺序获取区块。实际上，**选择顶部是在为一个区块链“投票”。** 一个区块链可以有很多个分支，最长的那个分支被认识是主分支。取得一个顶部以后（可以是区块链当中的任意一个区块）我们可以重构整个区块链然后确定它的长度以及构建它所需要的工作量。这也意味着顶端数据是一个区块链的识别器。

BlockchainIterator 只做一件事情：从区块链当中返回下一个区块：
```go
func (i *BlockchainIterator) Next() *Block {
	var block *Block

	err := i.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get(i.currentHash)
		block = DeserializeBlock(encodedBlock)

		return nil
	})

	i.currentHash = block.PrevBlockHash

	return block
}
```
数据库部分就这样完成了！

## 8 命令行接口（CLI）

到目前为止我们还没有实现提供一个接口能够与程序互动：我们只是简单的执行 NewBlockchain bc.AddBlock在 main 函数当中。是时候来有所提高了，我们要有这些命令：
```
$ go_build_goBlockChain addblock "Pay 0.031337 for a coffee"
$ go_build_goBlockChain printchain
```
所有的命令行相关的操作由CLI 结构体参与执行：
```go
type CLI struct {
	bc *Blockchain
}
```
它的“切入点”在 Run函数：
```go
func (cli *CLI) Run() {
	cli.validateArgs()

	addBlockCmd := flag.NewFlagSet("addblock", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)

	addBlockData := addBlockCmd.String("data", "", "Block data")

	switch os.Args[1] {
	case "addblock":
        err := addBlockCmd.Parse(os.Args[2:])
        logErr(err)
	case "printchain":
        err := printChainCmd.Parse(os.Args[2:])
        logErr(err)
	default:
		cli.printUsage()
		os.Exit(1)
	}

	if addBlockCmd.Parsed() {
		if *addBlockData == "" {
			addBlockCmd.Usage()
			os.Exit(1)
		}
		cli.addBlock(*addBlockData)
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}
}
```
我们使用Go语言的标准库当中的 flag 包来做命令行的词法分析。

addBlockCmd := flag.NewFlagSet("addblock", flag.ExitOnError)
printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
addBlockData := addBlockCmd.String("data", "", "Block data")
首先，我们创建两个子命令，addblock 和 printchain，然后在前面一个命令当中加入 -data 参数。printchain 命令不带任何参数。
```
switch os.Args[1] {
case "addblock":
	err := addBlockCmd.Parse(os.Args[2:])
    logErr(err)
case "printchain":
	err := printChainCmd.Parse(os.Args[2:])
    logErr(err)
default:
	cli.printUsage()
	os.Exit(1)
}
```
接下来，我们检查用户和词法分析提供的 flag 子命令。
```
if addBlockCmd.Parsed() {
	if *addBlockData == "" {
		addBlockCmd.Usage()
		os.Exit(1)
	}
	cli.addBlock(*addBlockData)
}

if printChainCmd.Parsed() {
	cli.printChain()
}
```
然后我再检查哪一个子命令将被分析然后运行相关函数。
```go
func (cli *CLI) addBlock(data string) {
	cli.bc.AddBlock(data)
	fmt.Println("Success!")
}

func (cli *CLI) printChain() {
	bci := cli.bc.Iterator()

	for {
		block := bci.Next()

		fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)
		pow := NewProofOfWork(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}
```
这一段刚之前那一段非常类似。唯一的不同使我们现在用 BlockchainIterator 去遍历区块链中的区块。

并且，也不要忘记相应的修改main 函数：
```
func main() {
	bc := NewBlockchain()
	defer bc.db.Close()

	cli := CLI{bc}
	cli.Run()
}
```
注意无论输入什么样的命令行参数都将会新建 Blockchain 实例。

好了，就这样，让我们来检查一下一切是否与所期望的相同：
```
$ go_build_goBlockChain printchain
No existing blockchain found.Creating a new one...
Mining the block containing "Frank's Genesis Block"
000000c7c586a90f973c7c44332f5b76ee19eb0cb52e5a610a8625983695c6be

Prev.hash: 
Data: Frank's Genesis Block
Hash: 000000c7c586a90f973c7c44332f5b76ee19eb0cb52e5a610a8625983695c6be
PoW: true

$ go_build_goBlockChain addblock -data "Golang"
Mining the block containing "Golang"
00000062acf2a2f20821181beaa6e2659112a12ece704e2ba451543de7fbd115

Success!

$ go_build_goBlockChain printchain             
Prev.hash: 000000c7c586a90f973c7c44332f5b76ee19eb0cb52e5a610a8625983695c6be
Data: Golang
Hash: 00000062acf2a2f20821181beaa6e2659112a12ece704e2ba451543de7fbd115
PoW: true

Prev.hash: 
Data: Frank's Genesis Block
Hash: 000000c7c586a90f973c7c44332f5b76ee19eb0cb52e5a610a8625983695c6be
PoW: true
 
```

## 9 结论（Conclusion）

下次我们将实现地址、钱包以及交易记录（很有可能）。保持好节奏！

## 10 链接（Links）

1. [Full source codes][10]
2. [Bitcoin Core Data Sorage][9]
3. [boltdb][7]
4. [encoding/gob][8]




[1]:https://link.zhihu.com/?target=https%3A//jeiwan.cc/	"Ivan Kuznetsov"

[2]:https://www.zhihu.com/people/xiaolai/activities	"李笑来"

[3]:https://res.cloudinary.com/flhonker/image/upload/v1530941971/githubio/go/goBlockChain/blockchain_part3-tree.png	"dir tree"

[4]:https://bitcoin.org/bitcoin.pdf	"比特币白皮书"

[5]:https://github.com/bitcoin/bitcoin	"比特币内核"

[6]:https://github.com/google/leveldb	"google/leveldb"

[7]:https://github.com/boltdb/bolt	"BoltDB & README"

[8]:https://golang.org/pkg/encoding/gob/	"encoding/gob"

[9]:https://en.bitcoin.it/wiki/Bitcoin_Core_0.11_%28ch_2%29%3A_Data_Storage	   "Bitcoin Core 0.11 (ch 2): Data Storage"

[10]:https://github.com/Jeiwan/blockchain_go/tree/part_3	"Full source codes"

