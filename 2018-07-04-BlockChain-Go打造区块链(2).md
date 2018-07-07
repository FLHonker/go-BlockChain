---
layout:     post
title:      用Go打造区块链（2）—— 工作证明机制（PoW）
subtitle:   Proof of Work
date:       2018-07-04
author:     Frank Liu
header-img: img/post-bg-golang1.png
catalog: true
tags:
    - Go
    - BloackChain
---

# 用Go打造区块链（2）——工作证明机制（PoW）

这一系列的文章是由[Ivan Kuznetsov][1]所写，第一篇文章的翻译稿由[李笑来][2]在其公众号学习学习再学习首发，本人觉得是一个结合Go语言学习区块链技术的好资料，后面将用自己的语言翻译一遍，从第一篇开始，顺便对Go语言以及区块链有一个初步的认识。

## 1 介绍

在[前面一篇文章][3]当中我们构建了一个简单的但却是区块链数据库的核心的数据结构。同时，我们也实现了向该数据库当中添加链式关系（chain-like relation）区块的方法：每一个区块都链接到它的前一个区块。令人遗憾的是，我们所实现的区块链有一个致命缺陷：添加新的区块太容易，成本也太低。区块和比特币的重要基本特征之一就是添加新的区块是一项非常难的工作。今天我们将处理完善这个缺陷。

本项目part_2的目录结构为：

![tree][4]

## 2 工作证明（Proof-of-Work）

区块链的重要设想就是如果你要往里面添加新的区块就要完成一些艰难的工作。而正是这种机制确保了区块链的安全和数据的一致性。同时，给这些艰难的工作适当的奖励（这也是人们挖矿获取比特币的机制）。

这种机制与现实非常类似：一个人必须通过努力工作获得回报以维持生计。在区块链当中，网络上的参与者（矿工）的工作维持网络的正常运行，向区块链中加入新的区块，并因为他们的努力工作而获得回报。他们的工作结果是将一个个区块以安全的方式连成一个完整的区块链，这也维护了整个区块链数据库的稳定性。更有价值的是，谁完成了工作必须进行自我证明。

这一整个的“努力工作并证明”的机制被称为“工作证明”（PoW）。它难在需要大量的计算资源：即便是高性能的计算机，也无法快速完成工作。甚至，为了保证新的区块增加速度维持在6个每小时，这个计算工作会越来越繁重。在比特币当中，计算工作的目的是为了给区块找一个匹配的并满足一些特定要求的哈希值。同时这个哈希值也为工作服务。因此，实际的工作就是寻找证明。

最后一点需要注意的，PoW算法必须满足一项要求：虽然计算很困难，但是工作证明的验证要非常容易。因为证明通常会传给网络上的其他参与者进行，不应该消耗他们的太多时间了验证这个证明。

## 3 哈希计算（hashing）

在这一段当中，我们将讨论下哈希值及其计算。如果你对这一概念已经熟悉，可以跳过这部分内容。

哈希计算是取得特定数据对应的哈希值的过程。对于计算出来的哈希值可以作为相应数据的特定代表。哈希函数是针对任意大小的数据产生固定长度的哈希值。哈希的主要特征如下：

1. 元数据无法通过哈希恢复。这样，哈希本身并不是加密的过程。
2. 特定数据只能有唯一的哈希值，哈希值是独一无二的。
3. 即便只是修改输入数据的一个字节，也会导致完全不同的哈希值。

![hash原理][5]

哈希函数被广泛应用于检验数据的一致性。一些软件提供商除了软件包以外会额外发布软件包对应的哈希检验值。在你下载了软件包以后可以将其代入一个哈希函数看生成的哈希值与软件商所提供的是否一致。

在区块链当中，哈希过程被用于确保一个区块的一致性。哈希算法的输入数据包含前一个区块的哈希值，使得修改区块链当中的区块变得不太可能（至少，非常困难），因为修改便意味着你必须计算该区块以及其之后所有区块的哈希值，而这个计算工作量是非常之大的。

## 4 Hashcash算法

关于哈希值的计算，比特币采用[Hashcash算法][6]，一种最早用于垃圾邮件过滤的带PoW机制的算法。它可以分解为以下的几个步骤：

采用一些公开数据（在邮件过滤当中，比如接收者的邮箱地址；在比特币当中，区块头部数据）
给它加一个计数器（counter）。计数器从0开始计数
将公开数据和计数器组合在一起（Data + counter），并获取组合数据的哈希值
检验获得的哈希值是否满足特定的要求
如果满足，则计算结束
如果不满足，计数器加1并重复步骤3、4
显然，这是一个暴力求解算法：改变计数器计算新的哈希值，检验，递增计数器，再计算一个哈希值，周而复始。这是其计算开销高昂的原因所在。

现在我们从头细看一个哈希值需要满足的具体要求。在Hashcash算法的原始实现当中，对哈希值的要求是“前20位必须为0”。在比特币当中，要求随时间变化有所调整，因为，根据设计，必须每10分钟产生一个区块，不管算力如何增加并有越来越多的矿工加入网络当中。

为了证明这个算法，我以前面例子当中的数据（“I like donuts”）为例并产生前面三个字节为0开头的哈希值：

![hashcah][7]

ca07ca是计数器的十六进制数，对应的十进制数为13240266。

## 5 实现（Implementation）

好了，理论部门已经清晰明了，让我们开始写代码吧。首先，让我们来设置下挖矿（区块产生）的难度：
```go
const targetBits = 24
```
在比特币当中，“目标位数”（targetBits）是存储在区块头部数据用以指示挖矿难度的指标。目前，我们并不打算实现难度可调的算法。因此，我们可以将难度系数定义为一个全局常量。

24是一个任意的数字，我们的目的是有一个数在内存中所占的存储空间在256位以下。并且这个差异值能够让我们明显感受到挖矿的难度，但不必太大，因为差异值设置的越大，将越难去找一个合适的哈希值。

```go
// proofofwork.go

const targetBits = 24   //设置挖矿（区块产生）的难度：

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
```
上述代码创建了 ProofOfWork 结构体来保存一个区块的指针和一个目标（target）的指针。这里的“target”与我们之前讨论的对哈希值的要求等同。我们之所以用Big整型来定义“target”在于我们将哈希值与目标比较的方式。我们将一个哈希值转换为一个Big整型然后检验其是否小于目标值。

然后在NewProofOfWork函数当中，我们初始化了一个Big整型数据为1并将其赋值给target随后将其左移（二进制的位操作）（256-targetBits）位。256是SHA-256哈希值的总体位数，在本文当中targetBits是24，因此这里总共左移了232位。后面我们也将采用SHA-256算法来产生哈希值。计算后的target的十六进制表示如下：
```
0x10000000000000000000000000000000000000000000000000000000000
```
在内存当中占据29个字节空间，下面是将其与我们之前例子中产生的哈希值的直观比较：
```
0fac49161af82ed938add1d8725835cc123a1a87b1b196488360e58d4bfb51e3
0000010000000000000000000000000000000000000000000000000000000000
0000008b0f41ec78bab747864db66bcb9fb89920ee75f43fdaaeb5544f7f76ca
```
第一个哈希（基于“I like donuts”计算）比target大，不是一个有效的工作证明。第二个哈希（基于“I like donutsca07ca”）比target小，是一个有效的证明。

你可以将target理解成一个范围的上边界：假如一个数（一个哈希值）比这个边界小，有效，反之，则无效。减小边界的数值，会导致更少的有效数字的个数，这样得到一个有效哈希值的难度将加大。

现在，我们来准备需要计算哈希值的数据。
```go
// proofofwork.go

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
```
对于工具函数IntToHex()，是将64bit的int整数转化为16进制的字符串格式我们在utils.go中定义：
```go
func IntToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}
```
上述的 prepareData 方法（因为在prepareData之前有ProofOfWork的结构体声明，这样的一种特殊的函数在Go语言当中叫做method）比较简单明了：我们通过调用bytes包的Join函数将区块信息与targetBits（在比特币当中，难度系数也是属于区块头部数据，这里也把它当做公开数据的一部分）以及nonce（临时值）相合并。nonce变量在这里就是前面Hashcash算法中描述的counter（计数器），它是一条加密后的数据。

OK，所有的准备已经就绪，来让我们实现PoW算法的核心部分吧：
```go
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
```
首先，我们初始化了几个变量：**hashInt** 是 hash 的整数型代表；**nonce** 是计数器。接下来我们开始一个循环：循环次数由 **maxNonce** 来控制，maxNonce等同于`math.MaxIn64`；这样是为了避免 nonce 的溢出。虽然我们所设置的难度想溢出还是比较困难的，以防万一，最好还是这样设置一下。

在这个循环体内，我们主要做了以下几件事情：

1. 准备数据
2. 用SHA-256算法计算该数据的哈希值
3. 将哈希值转换成Big整型数据
4. 将转换后的哈希值与target进行比较

像之前解释的一样简单。现在我们可以删除 Block 结构体的 SetHash 方法然后修改NewBlock 函数：
```go
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
```
在此你可以看到 nonce 存储为 Block 的一部分。这是很有必要的，因为 nonce 要用于工作证明的验证。同时 Block 结构体也按照以下代码进行修改：
```go
type Block struct {
	Timestamp     int64
	Data          []byte
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
}
```

好了，让我运行下程序看看是不是一切工作正常。

```
Mining the block containing "Frank's Genesis Block"
0000009dddcc2d4f9056ce01359f01360d26564702b4a5f2c1ecf963a9a0b3e2

Mining the block containing "Frank"
000000a8d357c56b788d7f357679b7e973e039c8e297c0b182d56bcd53a7a627

Mining the block containing "Ghost"
000000b42a12a08c40baea1e122c45e18b7763e0f3885fe31f5e0884d5dcb791

Prev.hash: 
Data: Frank's Genesis Block
Hash: 0000009dddcc2d4f9056ce01359f01360d26564702b4a5f2c1ecf963a9a0b3e2

Prev.hash: 0000009dddcc2d4f9056ce01359f01360d26564702b4a5f2c1ecf963a9a0b3e2
Data: Frank
Hash: 000000a8d357c56b788d7f357679b7e973e039c8e297c0b182d56bcd53a7a627

Prev.hash: 000000a8d357c56b788d7f357679b7e973e039c8e297c0b182d56bcd53a7a627
Data: Ghost
Hash: 000000b42a12a08c40baea1e122c45e18b7763e0f3885fe31f5e0884d5dcb791
```

(＾－＾)V 现在可以可以看到每一个哈希值以三个字节的0值开头，而且需要花一点时间才能获得这些哈希值。

还剩下一件事情需要完成：让我们来实现工作证明的验证。
```go
func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int

	data := pow.prepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	isValid := hashInt.Cmp(pow.target) == -1

	return isValid
}
```
这里也是我们需要将 nonce 保存下来的原因。

让我们再来测试一下一切是否正常：
```go
func main() {
    ...

    for _, block := range bc.blocks {
        ...
        pow := NewProofOfWork(block)
        fmt.Printf("PoW: %s\n", strconv.FormatBoo   (pow.Validate()))
        fmt.Println()
    }
}
```

输出结果：
```
... ...

Prev.hash: 
Data: Frank's Genesis Block
Hash: 0000003f7eed4aad038f74016c8d25b8bc193ab1e9062d7a61f55ff6c8a5144a
PoW: true

Prev.hash: 0000003f7eed4aad038f74016c8d25b8bc193ab1e9062d7a61f55ff6c8a5144a
Data: Frank
Hash: 000000f94ffbbb9b649bf2df6de57255e93db6f4ef0c1677a91ca6d378645b7b
PoW: true

Prev.hash: 000000f94ffbbb9b649bf2df6de57255e93db6f4ef0c1677a91ca6d378645b7b
Data: Ghost
Hash: 000000ea43c9de25c0e098171559bbc2dd3046da25e518cd53dd82a238427310
PoW: true

```

## 6 结论

Our blockchain is a step closer to its actual architecture: adding blocks now requires hard work, thus mining is possible. But it still lacks some crucial features: the blockchain database is not persistent, there are no wallets, addresses, transactions, and there’s no consensus mechanism. All these things we’ll implement in future articles, and for now, happy mining!

我们的区块链离它的实际架构又近了一步：往区块链里面增加区块需要复杂的计算工作，这样挖矿变得可能。然后它依然缺少一些关键的特性：区块链数据库无法持续存在，程序完成后数据就丢失了，也没有**钱包、地址、交易记录、共识机制（consensus mechanism）**。所有这些特性将会在后面的文章中实现，然后现在，先让我们快乐的挖矿吧！


## 7 链接

1. [Full source codes](https://github.com/Jeiwan/blockchain_go/tree/part_2)
2. [Blockchain hashing algorithm](https://en.bitcoin.it/wiki/Block_hashing_algorithm)
3. [Proof of work](https://en.bitcoin.it/wiki/Proof_of_work)
4. [Hashcash][6]

[1]:https://link.zhihu.com/?target=https%3A//jeiwan.cc/

[2]:https://www.zhihu.com/people/xiaolai/activities

[3]:http://frankliu624.me/2018/07/03/BlockChain-Go%E6%89%93%E9%80%A0%E5%8C%BA%E5%9D%97%E9%93%BE(1)/		"last post"

[4]:https://res.cloudinary.com/flhonker/image/upload/v1530879206/githubio/go/goBlockChain/blockchain_part2-tree.png	  "tree img"

[5]:https://res.cloudinary.com/flhonker/image/upload/v1530880143/githubio/go/goBlockChain/blockchain_part2-hash.jpg	  "hash img"

[6]:https://en.wikipedia.org/wiki/Hashcash	"hashcash"

[7]:https://res.cloudinary.com/flhonker/image/upload/v1530880307/githubio/go/goBlockChain/blockchain_part2-hashcash.jpg	"hashcash img"
