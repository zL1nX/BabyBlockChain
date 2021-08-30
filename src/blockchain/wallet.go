package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"golang.org/x/crypto/ripemd160"
	"log"
)

const version = byte(0x00)
const addressChecksumLen = 4

type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

func NewWallet() *Wallet {
	private, pubkey := NewKeyPair()
	wallet := Wallet{private, pubkey}
	return &wallet
}

func NewKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()
	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic(err)
	}
	pubkey := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...) // expand to append
	return *private, pubkey
}

func (w Wallet) GetAddress() []byte {
	pubKeyHash := HashPubKey(w.PublicKey)
	payload := append([]byte{version}, pubKeyHash...)
	checksem := CheckSum(payload)
	fullPayload := append(payload, checksem...)
	address := Base58Encode(fullPayload)
	return address
}

func VerifyAddress(address string) bool {
	fullPayload := Base58Decode([]byte(address))
	PayloadLen := len(fullPayload) - addressChecksumLen

	checksum := fullPayload[PayloadLen:]
	Payload := fullPayload[:PayloadLen]
	actualCheckSum := CheckSum(Payload)
	return bytes.Compare(checksum, actualCheckSum) == 0
}

func HashPubKey(pubkey []byte) []byte {
	pubkeySha := sha256.Sum256(pubkey)
	ripe := ripemd160.New()
	_, e := ripe.Write(pubkeySha[:])
	if e != nil {
		log.Panic(e)
	}
	ripeHash := ripe.Sum(nil)
	return ripeHash
}

func CheckSum(payload []byte) []byte {
	first := sha256.Sum256(payload)
	second := sha256.Sum256(first[:])
	return second[:addressChecksumLen] // only four bytes
}
