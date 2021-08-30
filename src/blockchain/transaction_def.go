package main

import (
	"bytes"
	"encoding/gob"
	"log"
)

type TXInput struct {
	TXid []byte // trans id
	Vout int    // index of an output in all outputs
	//ScriptSig string // signature of the transcation input
	Signature []byte // signature of the last tx
	PubKey    []byte // signature of the last user
}

type TXOutput struct {
	Value int // the bitcoin
	//ScriptPubKey string // the public key to verify a transaction
	PubKeyHash []byte // hash my pubkey for you to check
}
type TXOutputs struct {
	Outputs []TXOutput
}

func (in *TXInput) UseKey(pubKeyHash []byte) bool {
	actualHashKey := HashPubKey(in.PubKey)
	return bytes.Compare(actualHashKey, pubKeyHash) == 0
}

func (out *TXOutput) Lock(address []byte) {
	// address to verify
	decoded := Base58Decode(address)
	pubKeyHash := decoded[1 : len(decoded)-4] // the middle part
	out.PubKeyHash = pubKeyHash
}

func (out *TXOutput) isLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(out.PubKeyHash, pubKeyHash) == 0
}

func NewTXOutput(coins int, address string) *TXOutput {
	tx := &TXOutput{coins, nil}
	tx.Lock([]byte(address))
	return tx
}

func (outs TXOutputs) Serialize() []byte {
	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff)
	err := enc.Encode(outs)
	if err != nil {
		log.Panic(err)
	}
	return buff.Bytes()
}

func DeserializeOutputs(value []byte) TXOutputs {
	var outs TXOutputs
	dec := gob.NewDecoder(bytes.NewReader(value))
	err := dec.Decode(&outs)
	if err != nil {
		log.Panic(err)
	}
	return outs
}
