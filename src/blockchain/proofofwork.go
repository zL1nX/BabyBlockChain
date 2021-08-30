package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"
)

const targetBits = 16
const hashLength = 256

type ProofOfWork struct {
	b      *block
	target *big.Int
}

// initialize a new pow
func NewProofOfWork(b *block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(hashLength-targetBits)) // left shift those bits
	// like 0x10000000000000000000000000000000000000000000000000000000000 as a target
	pow := &ProofOfWork{b, target}
	return pow
}

func (pow *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.b.PrevBlockHash,
			//pow.b.Data,
			pow.b.HashTransactions(), // we use hash of trans now to refer the actual data
			IntToHex(pow.b.Timestamp),
			IntToHex(int64(targetBits)),
			IntToHex(int64(nonce)),
		},
		[]byte{}, // use an empty []byte to expand the two-row bytes array
	)
	return data // data to sha256
}

func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int
	data := pow.prepareData(pow.b.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])
	isValid := hashInt.Cmp(pow.target) == -1
	return isValid
}

// actually is the Mining
// the core idea is the nonce as requested
func (pow *ProofOfWork) Run() (int, []byte) {
	var hashInt big.Int
	var hash [32]byte
	var nonce, maxNonce int = 0, math.MaxInt64
	//fmt.Printf("Mining the block containing \"%s\"\n", pow.b.Data)
	fmt.Println("Mining a block...")
	for nonce < maxNonce {
		data := pow.prepareData(nonce)
		hash = sha256.Sum256(data)
		hashInt.SetBytes(hash[:]) // we compare the Int, not the bytes
		if hashInt.Cmp(pow.target) == -1 {
			// hashInt < target
			fmt.Printf("\rfound: %x", hash)
			break
		} else {
			nonce++
		}
	}
	fmt.Printf("\n\n")
	return nonce, hash[:]
}
