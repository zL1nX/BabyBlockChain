package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
)

const subsidy = 10

type Transaction struct {
	ID   []byte
	VIn  []TXInput
	VOut []TXOutput
} // a tx may have multiple input and output

func (tx *Transaction) isCoinbaseTX() bool {
	return len(tx.VIn) == 1 && tx.VIn[0].Vout == -1 && len(tx.VIn[0].TXid) == 0
	// genesis block 's vin has no TXid
}

// SetID func has been regrouped into Serialize and Hash

// serialize a tx, same code as serialize for block
func (tx Transaction) Serialize() []byte {
	var encoded bytes.Buffer
	e := gob.NewEncoder(&encoded)
	err := e.Encode(tx)
	if err != nil {
		log.Panic(err)
	}
	return encoded.Bytes()
}

func DeserializeTransaction(txData []byte) Transaction {
	var tx Transaction
	d := gob.NewDecoder(bytes.NewReader(txData))
	err := d.Decode(&tx) // remember the point
	if err != nil {
		log.Panic(err)
	}
	return tx
}

func (tx *Transaction) Hash() []byte {
	var hash [32]byte
	txCopy := *tx
	txCopy.ID = []byte{}
	hash = sha256.Sum256(txCopy.Serialize())
	return hash[:]
}

func (tx *Transaction) SetID() {
	var encoded bytes.Buffer
	var hash [32]byte
	buf := gob.NewEncoder(&encoded)
	err := buf.Encode(tx)
	if err != nil {
		log.Panic(err)
	}
	hash = sha256.Sum256(encoded.Bytes()) // Hash of the transaction, not the block
	tx.ID = hash[:]
}

// coinbase type of transaction doesn't need the last tx output
// when a miner mines a block, it will generate this kind of transaction
func NewCoinbaseTX(to, data string) *Transaction {
	if data == "" {
		randData := make([]byte, 20)
		_, err := rand.Read(randData)
		if err != nil {
			log.Panic(err)
		}
		data = fmt.Sprintf("%x", randData)
	}
	txin := TXInput{[]byte{}, -1, nil, []byte(data)} // remember this tx need no previous tx output
	txout := NewTXOutput(subsidy, to)
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{*txout}}
	tx.ID = tx.Hash() // New way
	return &tx
}

/*
// not sure what this part stands for
func (in *TXInput) CanUnlockOutputWith(unlockingData string) bool {
	return in.ScriptSig == unlockingData
}

func (out *TXOutput) CanBeUnlockedWith(unlockingData string) bool {
	return out.ScriptPubKey == unlockingData
}
*/

// coinbase-type tx can be used to generate the GENESIS BLOCK, aka the first block in blockchain

// a more general type of transaction
func NewUTXOTransaction(wallet *Wallet, to string, amount int, UTXO *UTXOSet) *Transaction {
	// find out all unspent tx to spend
	var inputs []TXInput
	var outputs []TXOutput
	//wallets, err := NewWallets()
	//if err != nil {
	//	log.Panic(err)
	//}
	//wallet := wallets.GetWallet(from) // who sent the coin
	FromPubKeyHash := HashPubKey(wallet.PublicKey)
	acc, validOutputs := UTXO.FindSpendableOutputs(FromPubKeyHash, amount)

	// validOutputs : map : string -> []int
	if acc < amount {
		log.Panic("Not Enough Coins")
	}
	for txid, outs := range validOutputs {
		hid, err := hex.DecodeString(txid)
		if err != nil {
			log.Panic(err)
		}
		for _, out := range outs {
			txinput := TXInput{hid, out, nil, wallet.PublicKey}
			inputs = append(inputs, txinput)
		}
	}
	// two outputs : one for specific tx (receiver address); one for coin change
	from := fmt.Sprintf("%s", wallet.GetAddress())
	outputs = append(outputs, *NewTXOutput(amount, to))
	if acc > amount { // why need this if statement
		outputs = append(outputs, *NewTXOutput(acc-amount, from))
	}
	tx := Transaction{nil, inputs, outputs}
	tx.ID = tx.Hash()
	UTXO.blockchain.SignTransaction(&tx, wallet.PrivateKey) // first sign then return
	return &tx
}

func (tx *Transaction) TrimmedCopy() Transaction {
	var txInput []TXInput
	var txOutput []TXOutput
	for _, in := range tx.VIn {
		txInput = append(txInput, TXInput{in.TXid, in.Vout, nil, nil})
	} // we leave the Signature and PubKey out of the input
	for _, out := range tx.VOut {
		txOutput = append(txOutput, TXOutput{out.Value, out.PubKeyHash})
	}
	txCopy := Transaction{tx.ID, txInput, txOutput}
	return txCopy
}

// sign the transaction
func (tx *Transaction) Sign(privKey ecdsa.PrivateKey, prevTXs map[string]Transaction) {
	if tx.isCoinbaseTX() {
		return // Coinbase-type transaction need no inputs
	}
	txCopy := tx.TrimmedCopy() // part of a tx need to be signed, not all of it
	for InId, vin := range txCopy.VIn {
		prevTx := prevTXs[hex.EncodeToString(vin.TXid)]            // get the tx in the input
		txCopy.VIn[InId].Signature = nil                           // make sure of that
		txCopy.VIn[InId].PubKey = prevTx.VOut[vin.Vout].PubKeyHash // set the pubkey of the input to be hash

		//txCopy.ID = txCopy.Hash() // hold it for now, hashes all the data above
		//txCopy.VIn[InId].PubKey = nil // after hash and serialize, we put it down again
		//r, s, err := ecdsa.Sign(rand.Reader, &privKey, txCopy.ID)

		dataSign := fmt.Sprintf("%x\n", txCopy)
		r, s, err := ecdsa.Sign(rand.Reader, &privKey, []byte(dataSign))
		if err != nil {
			log.Panic(err)
		}
		tx.VIn[InId].Signature = append(r.Bytes(), s.Bytes()...) // set it right
		// we sign the input seperately, so we need to set ID for Vin every time
		// tx is actually the transaction holding the signature
		txCopy.VIn[InId].PubKey = nil
	}
}

func (tx *Transaction) Verify(prevTxs map[string]Transaction) bool {
	txCopy := tx.TrimmedCopy()
	curve := elliptic.P256()

	for InId, vin := range tx.VIn {
		// regenerate the data to be signed, identical to the Sign
		prevTX := prevTxs[hex.EncodeToString(vin.TXid)]
		txCopy.VIn[InId].Signature = nil
		txCopy.VIn[InId].PubKey = prevTX.VOut[vin.Vout].PubKeyHash
		//txCopy.ID = txCopy.Hash()
		//txCopy.VIn[InId].PubKey = nil

		// set pubkey
		PubKeyBytes := vin.PubKey
		x, y := big.Int{}, big.Int{}
		x.SetBytes(PubKeyBytes[:len(PubKeyBytes)/2])
		y.SetBytes(PubKeyBytes[len(PubKeyBytes)/2:])
		PubKey := ecdsa.PublicKey{curve, &x, &y}

		// set signature
		signature := vin.Signature
		r, s := big.Int{}, big.Int{}
		r.SetBytes(signature[:len(signature)/2])
		s.SetBytes(signature[len(signature)/2:])
		dataVerify := fmt.Sprintf("%x\n", txCopy)
		if ecdsa.Verify(&PubKey, []byte(dataVerify), &r, &s) == false {
			return false
		}
		txCopy.VIn[InId].PubKey = nil
	}
	return true
}
