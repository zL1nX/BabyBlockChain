package main

import (
	"fmt"
	"log"
)

func (cli *CLI) startnode(nodeID, miningAddr string) {
	fmt.Printf("Starting node %s\n", nodeID)
	fmt.Printf("Starting node %s\n", nodeID)
	if len(miningAddr) > 0 {
		if VerifyAddress(miningAddr) {
			fmt.Println("Mining is on. Address to receive rewards: ", miningAddr)
		} else {
			log.Panic("Wrong miner address!")
		}
	}
	StartServer(nodeID, miningAddr)
}
