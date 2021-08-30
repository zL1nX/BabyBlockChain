package main

import "fmt"

func (cli *CLI) createBlockchain(address string, nodeID string) {
	bc := CreateBlockChain(address, nodeID)
	defer bc.db.Close()

	utxo := UTXOSet{bc}
	utxo.Reindex()

	fmt.Println("Create New BlockChain Done.")
}
