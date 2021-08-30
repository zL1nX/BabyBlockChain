package main

import (
	"encoding/hex"
	"github.com/boltdb/bolt"
	"log"
)

const utxoBucket = "chainstate"

type UTXOSet struct {
	blockchain *BlockChain
}

/*
uses FindUTXO to find unspent outputs, and stores them in a database. This is where caching happens.
*/
func (utxo UTXOSet) Reindex() {
	db := utxo.blockchain.db // use the same db but different bucket
	bucketName := []byte(utxoBucket)
	err := db.Update(func(tx *bolt.Tx) error {
		err := tx.DeleteBucket(bucketName)
		if err != nil && err != bolt.ErrBucketNotFound {
			log.Panic(err)
		}
		_, err = tx.CreateBucket(bucketName)

		if err != nil {
			log.Panic(err)
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	UTXO := utxo.blockchain.FindUTXO()
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		for txId, outs := range UTXO {
			key, e := hex.DecodeString(txId)
			if e != nil {
				log.Panic(e)
			}
			e = b.Put(key, outs.Serialize())
			if e != nil {
				log.Panic(e)
			}
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

// core part is the same as the blockchain's method
//  used to send coins:
func (utxo UTXOSet) FindSpendableOutputs(PubKeyHash []byte, amount int) (int, map[string][]int) {
	accumulated := 0
	unspentOutputs := make(map[string][]int) // txid : outIdx
	db := utxo.blockchain.db

	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(utxoBucket))
		c := bucket.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			txid := hex.EncodeToString(k)
			outs := DeserializeOutputs(v)           // remember the Reindex
			for outIdx, out := range outs.Outputs { // outIdx is basically the idx, really? This way, there would be so many same output idx
				if out.isLockedWithKey(PubKeyHash) && accumulated < amount {
					accumulated += out.Value
					unspentOutputs[txid] = append(unspentOutputs[txid], outIdx)
				}
			}
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return accumulated, unspentOutputs
}

//check balance:
func (utxo UTXOSet) FindUTXO(PubKeyHash []byte) []TXOutput {
	var unspentTXO []TXOutput
	db := utxo.blockchain.db
	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(utxoBucket))
		c := bucket.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			outs := DeserializeOutputs(v)
			for _, out := range outs.Outputs {
				if out.isLockedWithKey(PubKeyHash) { // only pick the unlocked output
					unspentTXO = append(unspentTXO, out)
				}
			}
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return unspentTXO
}

// Such separation requires solid synchronization mechanism
// But we don’t want to reindex every time a new block is mined
// Thus, we need a mechanism of updating the UTXO set:

func (utxo UTXOSet) Update(block *block) {
	db := utxo.blockchain.db
	err := db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(utxoBucket))
		for _, tx := range block.Transactions {
			if tx.isCoinbaseTX() == false {
				for _, vin := range tx.VIn {
					updateOuts := TXOutputs{}
					oldOuts := DeserializeOutputs(bucket.Get(vin.TXid))
					for outIdx, out := range oldOuts.Outputs {
						if outIdx != vin.Vout {
							updateOuts.Outputs = append(updateOuts.Outputs, out)
						}
					}
					if len(updateOuts.Outputs) == 0 {
						e := bucket.Delete(vin.TXid) // no need to cache
						if e != nil {
							log.Panic(e)
						}
					} else {
						e := bucket.Put(vin.TXid, updateOuts.Serialize()) // only cache the new
						// store outputs of most recent transactions.
						if e != nil {
							log.Panic(e)
						}
					}
				}
			}
			// tx is the candidate (new, because it is in the block), bucket get is the target (old)
			newOuts := TXOutputs{}
			for _, out := range tx.VOut {
				newOuts.Outputs = append(newOuts.Outputs, out)
			}
			e := bucket.Put(tx.ID, newOuts.Serialize()) // store outputs of most recent transactions.
			if e != nil {
				log.Panic(e)
			}
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	/*
		Updating means removing spent outputs and
		adding unspent outputs from newly mined transactions.
	*/
}

func (utx UTXOSet) CountTransactions() int {
	count := 0
	db := utx.blockchain.db
	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(utxoBucket))
		c := bucket.Cursor()
		for i, _ := c.First(); i != nil; i, _ = c.Next() {
			count++
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return count
}
