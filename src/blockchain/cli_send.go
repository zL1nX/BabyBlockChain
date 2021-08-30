package main

import (
	"fmt"
	"log"
)

func (cli *CLI) send(from, to string, amount int, nodeID string, mineNow bool) {
	//bc := NewBlockChain(from)
	if !VerifyAddress(from) {
		log.Panic("ERROR: Sender address is not valid")
	}
	if !VerifyAddress(to) {
		log.Panic("ERROR: Recipient address is not valid")
	}
	bc := NewBlockChain(nodeID)
	UTXO := UTXOSet{bc}
	defer bc.db.Close()

	/*
		tx := NewUTXOTransaction(from, to, amount, &UTXO)
		cbtx := NewCoinbaseTX(from, "") // the reward
	*/
	wallets, err := NewWallets(nodeID)
	if err != nil {
		log.Panic(err)
	}
	wallet := wallets.GetWallet(from)
	tx := NewUTXOTransaction(&wallet, to, amount, &UTXO)
	if mineNow {
		cbtx := NewCoinbaseTX(from, "")                    // the reward
		newBlock := bc.MineBlock([]*Transaction{cbtx, tx}) // add it to the chain
		UTXO.Update(newBlock)
	} else {
		sendTx(knownAddr[0], tx)
	}

	fmt.Println("Send Coin Success")
}
