package main

import (
	"bytes"
	"encoding/gob"
	"log"
	"time"
)

//type block struct {
//	Timestamp int64
//	Data []byte // payload like transaction details
//	PrevBlockHash []byte
//	Hash []byte // current block hash
//	Nonce int
//	// all members are capitalized first
//}

// struct method
//func (b *block) SetHash(){
//	timestamp := []byte(strconv.FormatInt(b.Timestamp, 10))
//	header := bytes.Join([][]byte{b.PrevBlockHash, b.Data, timestamp}, []byte{})
//	currentHash := sha256.Sum256(header) // hash(prevHash, data, time)
//	b.Hash = currentHash[:] // Hash is slice but currentHash is array
//}

/*
right now, each block contains at least one piece of transaction
*/
type block struct {
	Timestamp int64
	//Data []byte // payload like transaction details
	Transactions  []*Transaction
	PrevBlockHash []byte
	Hash          []byte // current block hash
	Nonce         int
	Height        int // current blockchain length
	// all members are capitalized first
}

// convert a block to a byte array
func (b *block) Serialize() []byte {
	var res bytes.Buffer
	encoder := gob.NewEncoder(&res) // address
	err := encoder.Encode(b)        // here b is the source
	if err != nil {
		log.Panic(err)
	}
	return res.Bytes()
}

// convert byte array to a block
func Deserialize(buffer []byte) *block {
	var b block
	decoder := gob.NewDecoder(bytes.NewReader(buffer)) // convert []byte to io.Reader
	err := decoder.Decode(&b)                          // here b is the dest
	if err != nil {
		log.Panic(err)
	}
	return &b
}

// initialize a new block
//func NewBlock(data string, prevBlockHash []byte) *block {
//	currentTime := time.Now().Unix()
//	newBlock := &block{currentTime, []byte(data), prevBlockHash, []byte{}, 0}
//	// a pointer; but not a pointer seems ok
//
//	// remove the old SetHash
//	pow := NewProofOfWork(newBlock)
//	nonce, hash := pow.Run()
//	newBlock.Hash = hash[:] // set the small hash as the block hash
//	newBlock.Nonce = nonce
//	return newBlock
//}

// the transaction version
func NewBlock(transactions []*Transaction, prevBlockHash []byte, height int) *block {
	currentTime := time.Now().Unix()
	newBlock := &block{currentTime, transactions, prevBlockHash, []byte{}, 0, height}
	// a pointer; but not a pointer seems ok

	// remove the old SetHash
	pow := NewProofOfWork(newBlock) // remember that newBlock will have to mine first (proof of work)
	nonce, hash := pow.Run()
	newBlock.Hash = hash[:] // set the small hash as the block hash
	newBlock.Nonce = nonce
	return newBlock
}

//func NewGenesisBlock() *block {
//	return NewBlock("Genesis Block", []byte{})
//}

/*
the transaction version
*/
func NewGenesisBlock(coinbase *Transaction) *block {
	return NewBlock([]*Transaction{coinbase}, []byte{}, 0) //  there is no blockchain yet
}

func (b *block) HashTransactions() []byte {
	var transaction [][]byte
	for _, tx := range b.Transactions {
		transaction = append(transaction, tx.Serialize())
	}
	mTree := NewMerkelTree(transaction)
	return mTree.Root.Data
}
