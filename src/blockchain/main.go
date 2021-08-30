package main

func main() {
	//bc := NewBlockChain() // need some work to initialize
	//// clearly slower
	//
	//bc.AddBlock("Send 1 BTC to Ivan")
	//bc.AddBlock("Send 2 more BTC to Ivan")
	//for _, block := range bc.blocks {
	//	fmt.Printf("Prev. Hash is %x\n", block.PrevBlockHash)
	//	fmt.Printf("Data is %s\n", block.Data)
	//	fmt.Printf("Current Hash is %x\n", block.Hash)
	//	pow := NewProofOfWork(block) // the satisfied nonce and hash
	//	fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
	//	fmt.Println()
	//}

	//bc := NewBlockChain()
	//defer bc.db.Close()
	//
	//cli := CLI{bc}
	//cli.Run()

	cli := CLI{}
	cli.Run()
}
