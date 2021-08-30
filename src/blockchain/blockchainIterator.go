package main

import (
	"github.com/boltdb/bolt"
	"log"
)

// we dont want all blocks in memory, thus we need an dedicated iterator for block chain
type BlockChainIterator struct {
	currentHash []byte
	db          *bolt.DB
}

func (chain *BlockChain) Iterator() *BlockChainIterator {
	it := &BlockChainIterator{chain.tip, chain.db} // from the newest block
	return it
}

func (it *BlockChainIterator) Next() *block {
	var b *block
	err := it.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucket))
		Encoded := bucket.Get(it.currentHash)
		b = Deserialize(Encoded) // get the current (you can say "next") block
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	it.currentHash = b.PrevBlockHash // move iterator backward ( from the newest to the oldest)
	return b
}
