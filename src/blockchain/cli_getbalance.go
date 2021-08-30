package main

import (
	"fmt"
	"log"
)

func (cli *CLI) getBalance(address string, nodeID string) {
	if !VerifyAddress(address) {
		log.Panic("ERROR: Address is not valid")

	}
	bc := NewBlockChain(nodeID)
	utxo := UTXOSet{bc}

	defer bc.db.Close()

	balance := 0
	decoded := Base58Decode([]byte(address))
	pubKeyHash := decoded[1 : len(decoded)-4]
	UTXO := utxo.FindUTXO(pubKeyHash)
	for _, out := range UTXO {
		balance += out.Value // the coin change output
	}
	fmt.Printf("Balance of '%s': %d\n", address, balance)
}
