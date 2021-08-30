package main

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
	"os"
)

const dbFile = "blockchain_%s.db"
const blocksBucket = "blocks"
const genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"

//type BlockChain struct {
//	blocks []*block // array is a member
//} // array formed by blocks

type BlockChain struct {
	tip []byte   // only stored the last block hash
	db  *bolt.DB // along with its specific db
}

//func (chain *BlockChain) AddBlock(data string) {
//	prevBlock := chain.blocks[len(chain.blocks) - 1] // blocks is a member of blockchain
//	newBlock := NewBlock(data, prevBlock.Hash)
//	chain.blocks = append(chain.blocks, newBlock) // append a new block
//}

//func (chain *BlockChain) AddBlock(data string) {
//	var prevHash []byte
//	err := chain.db.View(func(tx *bolt.Tx) error { // only View not edit
//		bucket := tx.Bucket([]byte(blocksBucket))
//		if bucket == nil {
//			log.Panic("Bucket is Null")
//		}
//		prevHash = bucket.Get([]byte("l")) // get the prev block (aka. last block) hash
//		return nil
//	})
//	block := NewBlock(data, prevHash)
//	err = chain.db.Update(func(tx *bolt.Tx) error {
//		bucket := tx.Bucket([]byte(blocksBucket))
//		e := bucket.Put(block.Hash, block.Serialize()) // store the new block : hash to block
//		if e != nil {
//			log.Panic(e)
//		}
//		e = bucket.Put([]byte("l"), block.Hash) // update the l entry to the new block hash
//		if e != nil {
//			log.Panic(e)
//		}
//		chain.tip = block.Hash // dont forget update the chain tip
//		return nil
//	})
//	if err != nil{
//		log.Panic(err)
//	}
//}

func dbExists(thisdbFile string) bool {
	if _, err := os.Stat(thisdbFile); os.IsNotExist(err) {
		return false
	}
	return true
}

// transaction version of AddBlock, but the two are essentially the same
func (chain *BlockChain) MineBlock(transactions []*Transaction) *block {
	var prevHash []byte
	var prevHeight int

	for _, tx := range transactions {
		if chain.VerifyTransaction(tx) != true {
			log.Panic("ERROR: Invalid transaction")
		}
	}

	err := chain.db.View(func(tx *bolt.Tx) error { // only View not edit
		bucket := tx.Bucket([]byte(blocksBucket))
		if bucket == nil {
			log.Panic("Bucket is Null")
		}
		prevHash = bucket.Get([]byte("l")) // get the prev block (aka. last block) hash
		prevBlockData := bucket.Get(prevHash)
		prevBlock := *Deserialize(prevBlockData)
		prevHeight = prevBlock.Height //  get the current height
		return nil
	})
	newBlock := NewBlock(transactions, prevHash, prevHeight+1) // the new block extends the chain
	err = chain.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucket))
		e := bucket.Put(newBlock.Hash, newBlock.Serialize()) // store the new block : hash to block
		if e != nil {
			log.Panic(e)
		}
		e = bucket.Put([]byte("l"), newBlock.Hash) // update the l entry to the new block hash
		if e != nil {
			log.Panic(e)
		}
		chain.tip = newBlock.Hash // dont forget update the chain tip, that is add a block to the whole chain
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return newBlock
}

func (bc *BlockChain) AddBlock(block *block) {
	err := bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		blockInDb := b.Get(block.Hash)

		if blockInDb != nil {
			return nil
		}

		blockData := block.Serialize()
		err := b.Put(block.Hash, blockData)
		if err != nil {
			log.Panic(err)
		}

		lastHash := b.Get([]byte("l"))
		lastBlockData := b.Get(lastHash)
		lastBlock := Deserialize(lastBlockData)

		if block.Height > lastBlock.Height {
			err = b.Put([]byte("l"), block.Hash)
			if err != nil {
				log.Panic(err)
			}
			bc.tip = block.Hash
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

//func NewBlockChain() *BlockChain {
//	var tip []byte
//	db, err := bolt.Open(dbFile, 0600, nil)
//	if err != nil{
//		log.Panic(err)
//	}
//	// dp Update is a transaction involving reading and updating
//	err = db.Update(func(tx *bolt.Tx) error {
//		b := tx.Bucket([]byte(blocksBucket)) // get bucket by string id
//		if b == nil {
//			fmt.Println("No blockchain existed, creating one...")
//			genesis := NewGenesisBlock()
//			// create a new genesis block and add it to the database
//			b, err := tx.CreateBucket([]byte(blocksBucket)) // create "blocks"bucket
//			if err != nil {
//				log.Panic(err)
//			}
//			err = b.Put(genesis.Hash, genesis.Serialize())
//			if err != nil {
//				log.Panic(err)
//			}
//			err = b.Put([]byte{'l'}, genesis.Hash)
//			if err != nil {
//				log.Panic(err)
//			}
//			// two keys : hash to block(serialized), tip
//			tip = genesis.Hash
//		}else{
//			// block chain existed
//			tip = b.Get([]byte{'l'}) // the key 'l' is the last block hash
//		}
//		return nil
//	})
//	if err != nil {
//		log.Panic(err)
//	}
//
//	bc := BlockChain{tip, db}
//	return &bc // initialize a new block
//}

// create a blockchain database and genesis block
func CreateBlockChain(address string, nodeId string) *BlockChain {
	thisdbFile := fmt.Sprintf(dbFile, nodeId)
	if dbExists(thisdbFile) {
		fmt.Println("Blockchain already exists.")
		os.Exit(1)
	}
	var tip []byte
	db, err := bolt.Open(thisdbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}
	fmt.Println("CreateBlockChain1")
	err = db.Update(func(tx *bolt.Tx) error {
		coinbasetx := NewCoinbaseTX(address, genesisCoinbaseData)
		genesis := NewGenesisBlock(coinbasetx)
		bucket, err := tx.CreateBucket([]byte(blocksBucket)) // there is none yet, create one
		if err != nil {
			log.Panic(err)
		}
		err = bucket.Put(genesis.Hash, genesis.Serialize())
		if err != nil {
			log.Panic(err)
		}
		err = bucket.Put([]byte{'l'}, genesis.Hash)
		if err != nil {
			log.Panic(err)
		}
		tip = genesis.Hash
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	fmt.Println("CreateBlockChain2")
	bc := &BlockChain{tip, db}
	return bc
}

// the "create" function now is handled by CreateBlockChain, so this function only reserves the "get" function
// that is, use the created genesis block to initialize a new block chain
// 与其说是new，不如说是get，这个函数本质上就是从现有db的最后hash开始，新建一个blockchain对象，然后开始后续操作，新的概念被弱化了
func NewBlockChain(nodeId string) *BlockChain {
	thisdbFile := fmt.Sprintf(dbFile, nodeId)
	if dbExists(thisdbFile) == false {
		fmt.Println("No existing blockchain found. Create one first.")
		os.Exit(1)
	}

	var tip []byte
	db, err := bolt.Open(thisdbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}
	// dp Update is a transaction involving reading and updating
	err = db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucket))
		tip = bucket.Get([]byte("l"))
		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	bc := BlockChain{tip, db}
	return &bc // initialize a new block
}

// find those transactions containing unspent output (other outputs in this tx may be spent)

/*
func (bc *BlockChain) FindUnspentTransactions(pubKeyHash []byte) []Transaction{
	bci := bc.Iterator()
	spentTX := make(map[string][]int) // current tx id -> out index ; record output location
	var unspentTX []Transaction
	for {
		block := bci.Next()
		for _, tx := range block.Transactions {
			currId := hex.EncodeToString(tx.ID)
			// if there is an outIdx not in the spentTx, then this whole TX itself will be recorded
			Outputs:
			for outIdx, out := range tx.VOut {
				if spentTX[currId] != nil { // referenced by some tx
					for _, refid := range spentTX[currId] {
						if refid == outIdx {
							continue Outputs// indeed referenced or appeared
						}
					}
					// check if all outid are included and appeared
				}
				// if not, that means we found a left one, which is an output not being referenced
				if out.isLockedWithKey(pubKeyHash) {
					unspentTX = append(unspentTX, *tx) // record it
				}
				// 明明这里只是out没有被引用，但是会把整个TX都加进去，是不是有点不太公平？或者说两个的筛选粒度不一样
				// 而且如果一个tx有多个output没被引用，那是否会被append多次
			}

			if tx.isCoinbaseTX() == false {
				for _, vin := range tx.VIn {
					if vin.UseKey(pubKeyHash) {
						prevTxId := hex.EncodeToString(vin.TXid)
						spentTX[prevTxId] = append(spentTX[prevTxId], vin.Vout) // inner index
					}
				}
			}
		}
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	// this function should be called further to specify which output are exactly unspent
	//fmt.Println(len(unspentTX))
	return unspentTX
}
*/
//almost identical to Blockchain.FindUnspentTransactions, but now it returns a map of TransactionID → TransactionOutputs pairs
func (bc *BlockChain) FindUTXO() map[string]TXOutputs {
	UTXO := make(map[string]TXOutputs)

	bci := bc.Iterator()
	spentTX := make(map[string][]int) // current tx id -> out index ; record output location

	for {
		block := bci.Next()
		for _, tx := range block.Transactions {
			currId := hex.EncodeToString(tx.ID)
			// if there is an outIdx not in the spentTx, then this whole TX itself will be recorded
		Outputs:
			for outIdx, out := range tx.VOut {
				if spentTX[currId] != nil { // referenced by some tx
					for _, refid := range spentTX[currId] {
						if refid == outIdx {
							continue Outputs // indeed referenced or appeared
						}
					}

					// check if all outid are included and appeared
				}
				outs := UTXO[currId]
				outs.Outputs = append(outs.Outputs, out)
				UTXO[currId] = outs
				// if not, that means we found a left one, which is an output not being referenced

				// 明明这里只是out没有被引用，但是会把整个TX都加进去，是不是有点不太公平？或者说两个的筛选粒度不一样
				// 而且如果一个tx有多个output没被引用，那是否会被append多次
			}

			if tx.isCoinbaseTX() == false {
				for _, vin := range tx.VIn {
					prevTxId := hex.EncodeToString(vin.TXid)
					spentTX[prevTxId] = append(spentTX[prevTxId], vin.Vout) // inner index
				}
			}
		}
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	// this function should be called further to specify which output are exactly unspent
	//fmt.Println(len(unspentTX))
	return UTXO
}

// just pick the output
/*
a little bit of weird
*/
//func (bc *BlockChain) FindUTXO(PubKeyHash []byte) []TXOutput {
//	unspentTX := bc.FindUnspentTransactions(PubKeyHash)
//	var unspentTXO []TXOutput
//	for _, tx := range unspentTX {
//		for _, out := range tx.VOut {
//			if out.isLockedWithKey(PubKeyHash) { // only pick the unlocked output
//				unspentTXO = append(unspentTXO, out)
//			}
//		}
//	}
//	return unspentTXO
//}

// move to UTXOSet method
/*
func (bc *BlockChain) FindSpendableOutput(PubKeyHash []byte, amount int) (int, map[string][]int) {
	accumulated := 0
	unspentTXs := bc.FindUnspentTransactions(PubKeyHash) // spendable tx
	unspentOutputs := make(map[string][]int) // txid : outIdx

	Pick:
	for _, tx := range unspentTXs {
		txid := hex.EncodeToString(tx.ID)
		for outIdx, out := range tx.VOut { // outIdx is basically the idx, really? This way, there would be so many same output idx
			if out.isLockedWithKey(PubKeyHash) && accumulated < amount {
				accumulated += out.Value
				unspentOutputs[txid] = append(unspentOutputs[txid], outIdx)
				if accumulated >= amount {
					break Pick
				}
			}
		}
	}
	return accumulated, unspentOutputs
}*/

func (bc *BlockChain) FindPrevTransaction(ID []byte) (Transaction, error) {
	bci := bc.Iterator()
	for {
		block := bci.Next()
		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return Transaction{}, errors.New("transaction is not found")
}

func (bc *BlockChain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transaction)
	for _, vin := range tx.VIn {
		prevTX, err := bc.FindPrevTransaction(vin.TXid)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}
	tx.Sign(privKey, prevTXs)
}

func (bc *BlockChain) VerifyTransaction(tx *Transaction) bool {
	if tx.isCoinbaseTX() {
		return true
	}
	prevTXs := make(map[string]Transaction)
	for _, vin := range tx.VIn {
		prevTX, err := bc.FindPrevTransaction(vin.TXid)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}
	return tx.Verify(prevTXs)
}

func (bc *BlockChain) GetBlockHashes() [][]byte {
	var blockHashes [][]byte
	bci := bc.Iterator()
	for {
		b := bci.Next()
		blockHashes = append(blockHashes, b.Hash)
		if len(b.PrevBlockHash) == 0 {
			break
		}
	}
	return blockHashes
}

func (bc *BlockChain) GetBlock(blockhash []byte) (block, error) {
	var b block
	err := bc.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucket))
		blockData := bucket.Get(blockhash)
		if blockData == nil {
			return errors.New("Block is not found.")
		}
		b = *Deserialize(blockData)
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return b, nil
}

func (bc *BlockChain) GetBestHeight() int {
	var lastBlock block
	err := bc.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucket))
		lastHash := bucket.Get([]byte("l"))
		lastBlockData := bucket.Get(lastHash)
		lastBlock = *Deserialize(lastBlockData)
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return lastBlock.Height
}
