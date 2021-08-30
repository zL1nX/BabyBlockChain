package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

// we want to manipulate the cmd

//type CLI struct {
//	bc *BlockChain
//}
//func (cli *CLI) printUsage() {
//	fmt.Println("Usage:")
//	fmt.Println("addblock -data BLOCK_DATA - add a block to the blockchain")
//	fmt.Println("printchain - print all the blocks of the blockchain")
//}
//
//func (cli *CLI) validateArgs() {
//	if len(os.Args) < 2{
//		cli.printUsage()
//		os.Exit(1)
//	}
//}
//
//func (cli *CLI) addBlock(data string) {
//	cli.bc.AddBlock(data)
//	fmt.Println("Add Block Data Success!...")
//}
//
//func (cli *CLI) printChain() {
//	it := cli.bc.Iterator()
//	for {
//		b := it.Next()
//		fmt.Printf("Prev. hash: %x\n\n", b.PrevBlockHash)
//		fmt.Printf("Data: %s\n", b.Data)
//		fmt.Printf("Hash: %x\n", b.Hash)
//		pow := NewProofOfWork(b)
//		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
//		fmt.Println()
//		if len(b.PrevBlockHash) == 0 {
//			break
//		}
//	}
//}
//
//func (cli *CLI) Run() {
//	cli.validateArgs()
//	addBlockCmd := flag.NewFlagSet("addblock", flag.ExitOnError)
//	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
//	addBlockData := addBlockCmd.String("data", "", "Block Data") // a string pointer, binding the block data
//
//	// parse the args
//	switch os.Args[1] {
//	case "addblock":
//		err := addBlockCmd.Parse(os.Args[2:])
//		if err != nil {
//			log.Panic(err)
//		}
//	case "printchain" :
//		err := printChainCmd.Parse(os.Args[2:])
//		if err != nil {
//			log.Panic(err)
//		}
//	default:
//		cli.printUsage()
//		os.Exit(1)
//	}
//
//	// now let's look at which cmd is parsed by the user
//	if addBlockCmd.Parsed() {
//		if *addBlockData == "" { // the block data is empty
//			addBlockCmd.Usage()
//			os.Exit(1)
//		}
//		cli.addBlock(*addBlockData)
//	}
//
//	if printChainCmd.Parsed() {
//		cli.printChain()
//	}
//
//
//}

type CLI struct{}

func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  getbalance -address ADDRESS - Get balance of ADDRESS")
	fmt.Println("  createblockchain -address ADDRESS - Create a blockchain and send genesis block reward to ADDRESS")
	fmt.Println("  printchain - Print all the blocks of the blockchain")
	//fmt.Println("  send -from FROM -to TO -amount AMOUNT - Send AMOUNT of coins from FROM address to TO")
	fmt.Println("  send -from FROM -to TO -amount AMOUNT -mine - Send AMOUNT of coins from FROM address to TO. Mine on the same node, when -mine is set.")
	fmt.Println("  startnode -miner ADDRESS - Start a node with ID specified in NODE_ID env. var. -miner enables mining")
}

func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

func (cli *CLI) Run() {
	cli.validateArgs()

	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" {
		fmt.Printf("NODE_ID env. var is not set!")
		os.Exit(1)
	}
	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	printchainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	listAddressCmd := flag.NewFlagSet("listaddress", flag.ExitOnError)
	reindexCmd := flag.NewFlagSet("reindex", flag.ExitOnError)
	startNodeCmd := flag.NewFlagSet("startnode", flag.ExitOnError)

	createBlockchainAddr := createBlockchainCmd.String("address", "", "The address to send genesis block reward to")
	getBalanceValue := getBalanceCmd.String("address", "", "The address to get balance for")
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")
	sendMine := sendCmd.Bool("mine", false, "Mine immediately on the same node")
	startNodeMinder := startNodeCmd.String("miner", "", "Enable mining mode and send reward to ADDRESS")

	switch os.Args[1] {
	case "createblockchain":
		err := createBlockchainCmd.Parse(os.Args[2:]) // [2:] not [2]
		if err != nil {
			log.Panic(err)
		}
	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:]) // [2:] not [2]
		if err != nil {
			log.Panic(err)
		}
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "printchain":
		err := printchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createwallet":
		err := createWalletCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "listaddress":
		err := listAddressCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "reindex":
		err := reindexCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "startnode":
		err := startNodeCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		cli.printUsage()
		os.Exit(1)
	}

	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddr == "" {
			createBlockchainCmd.Usage()
			os.Exit(1)
		}
		cli.createBlockchain(*createBlockchainAddr, nodeID)
	}
	if getBalanceCmd.Parsed() {
		if *getBalanceValue == "" {
			getBalanceCmd.Usage()
			os.Exit(1)
		}
		cli.getBalance(*getBalanceValue, nodeID)
	}
	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount == 0 {
			sendCmd.Usage()
			os.Exit(1)
		}
		cli.send(*sendFrom, *sendTo, *sendAmount, nodeID, *sendMine)
	}
	if printchainCmd.Parsed() {
		cli.printChain(nodeID)
	}
	if createWalletCmd.Parsed() {
		cli.createWallet(nodeID)
	}
	if listAddressCmd.Parsed() {
		cli.listAddress(nodeID)
	}
	if reindexCmd.Parsed() {
		cli.reindex(nodeID)
	}
	if startNodeCmd.Parsed() {
		nodeID := os.Getenv("NODE_ID")
		if nodeID == "" {
			startNodeCmd.Usage()
			os.Exit(1)
		}
		cli.startnode(nodeID, *startNodeMinder)
	}
}
