package main

import "fmt"

func (cli *CLI) reindex(nodeID string) {
	bc := NewBlockChain(nodeID)
	utx := UTXOSet{bc}
	utx.Reindex()
	cnt := utx.CountTransactions()
	fmt.Printf("Done! There are %d transactions in the UTXO set.\n", cnt)
}
