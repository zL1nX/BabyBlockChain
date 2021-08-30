package main

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
)

const protocol = "tcp"
const nodeVersion = 1
const commandLength = 12

var knownAddr = []string{"localhost:3000"}
var miningAddr string
var nodeAddr string
var blocksInTransit = [][]byte{}
var mempool = make(map[string]Transaction)

type addrMsg struct {
	Addrlist []string
}

type versionMsg struct {
	Version    int
	BestHeight int
	AddrFrom   string
}

/*show me what blocks you have, not give me your blocks*/
type getBlocksMsg struct {
	AddrFrom string // who's asking
}

/*
request for certain block or transaction, and it can contain only one block/transaction ID
*/
type getDataMsg struct {
	AddrFrom string
	Type     string
	ID       []byte
}

/*
Bitcoin uses inv to show other nodes what blocks or transactions current node has
it doesn’t contain whole blocks and transactions, just their hashes
*/
type invMsg struct {
	AddrFrom string // who will show you
	Type     string //blocks or transactions?
	Items    [][]byte
}

type blockMsg struct {
	AddrFrom string
	Block    []byte
}

type txMsg struct {
	AddrFrom string
	Tx       []byte
}

func sendAddr(addr string) {
	nodes := addrMsg{knownAddr}
	nodes.Addrlist = append(nodes.Addrlist, nodeAddr)
	payload := gobEncode(nodes)
	request := append(commandToBytes("addr"), payload...)

	sendData(addr, request)
}

func sendVersion(addr string, bc *BlockChain) {
	bestHeight := bc.GetBestHeight()
	payload := gobEncode(versionMsg{nodeVersion, bestHeight, nodeAddr})
	/*
		First 12 bytes specify command name (“version” in this case), and the latter bytes will contain gob-encoded message structure.
	*/
	request := append(commandToBytes("version"), payload...)

	sendData(addr, request)
}

func sendGetBlocks(addr string) {
	payload := gobEncode(getBlocksMsg{nodeAddr}) // my address
	request := append(commandToBytes("getblocks"), payload...)
	sendData(addr, request)
}

func sendInv(addr string, invType string, items [][]byte) {
	inventory := invMsg{nodeAddr, invType, items} // my current nodes
	payload := gobEncode(inventory)
	request := append(commandToBytes("inv"), payload...)
	sendData(addr, request)
}

func sendGetData(addr string, dataType string, targetData []byte) {
	fmt.Println("Send Get Data Request.")
	payload := gobEncode(getDataMsg{nodeAddr, dataType, targetData})
	request := append(commandToBytes("getdata"), payload...)
	sendData(addr, request)
}

func sendBlock(addr string, b *block) {
	fmt.Println("Send Block Data")
	payload := gobEncode(blockMsg{nodeAddr, b.Serialize()})
	request := append(commandToBytes("blocks"), payload...)
	sendData(addr, request)
}

func sendTx(addr string, tx *Transaction) {
	payload := gobEncode(txMsg{nodeAddr, tx.Serialize()})
	request := append(commandToBytes("txs"), payload...)
	sendData(addr, request)
}

func sendData(addr string, data []byte) {
	conn, err := net.Dial(protocol, addr)
	if err != nil {
		fmt.Printf("%s is not available. Updating Nodes\n", addr)
		var liveNodes []string
		for _, node := range knownAddr {
			if node != addr {
				liveNodes = append(liveNodes, node)
			}
		}
		knownAddr = liveNodes
		return
	}
	defer conn.Close()
	_, err = io.Copy(conn, bytes.NewReader(data))
	if err != nil {
		log.Panic(err)
	}
}

func requestBlocks() {
	for _, node := range knownAddr {
		sendGetBlocks(node)
	}
}

func gobEncode(data interface{}) []byte {
	var buff bytes.Buffer
	encoder := gob.NewEncoder(&buff)
	err := encoder.Encode(data)
	if err != nil {
		log.Panic(err)
	}
	return buff.Bytes()
}

/*
It creates a 12-byte buffer and fills it with the command name, leaving rest bytes empty.
*/
func commandToBytes(command string) []byte {
	var bcmd [commandLength]byte
	for i, c := range command {
		bcmd[i] = byte(c)
	}
	return bcmd[:]
}

func bytesToCommand(bcommand []byte) string {
	var command []byte
	for _, b := range bcommand {
		if b != 0x0 {
			command = append(command, b)
		}
	}
	return string(command)
}

func nodeIsKnown(addr string) bool {
	for _, node := range knownAddr {
		if addr == node {
			return true
		}
	}
	return false
}

func handleAddr(request []byte) {
	var buff bytes.Buffer
	var payload addrMsg
	buff.Write(request[commandLength:]) //  different than the previous method
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	knownAddr = append(knownAddr, payload.Addrlist...)
	fmt.Printf("There are %d known nodes now!\n", len(knownAddr))
	fmt.Println(knownAddr)
	requestBlocks()

}

func handleVersion(request []byte, bc *BlockChain) {
	var buff bytes.Buffer
	var payload versionMsg
	buff.Write(request[commandLength:]) //  different than the previous method
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	/*
		Then a node compares its BestHeight with the one from the message. If the node’s blockchain is longer,
		it’ll reply with version message; otherwise, it’ll send getblocks message.
	*/
	myBestHeight := bc.GetBestHeight()
	if myBestHeight > payload.BestHeight { // mine is longer
		sendVersion(payload.AddrFrom, bc)
	} else if myBestHeight < payload.BestHeight {
		sendGetBlocks(payload.AddrFrom)
	}

	if !nodeIsKnown(payload.AddrFrom) {
		knownAddr = append(knownAddr, payload.AddrFrom)
	}
}

func handleGetBlocks(request []byte, bc *BlockChain) {
	var buff bytes.Buffer
	var payload getBlocksMsg
	buff.Write(request[commandLength:]) //  different than the previous method
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	blocks := bc.GetBlockHashes()
	sendInv(payload.AddrFrom, "blocks", blocks) // send you my nodes
}

func handleInv(request []byte, bc *BlockChain) {
	var buff bytes.Buffer
	var payload invMsg
	buff.Write(request[commandLength:]) //  different than the previous method
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("Recevied inventory with %d %s from %s \n", len(payload.Items), payload.Type, payload.AddrFrom)
	if payload.Type == "blocks" {
		blocksInTransit = payload.Items     // these blocks need to be downloaded
		downloadedBlock := payload.Items[0] // one inv for one hash here
		// In our implementation, we’ll never send inv with multiple hashes
		sendGetData(payload.AddrFrom, "blocks", downloadedBlock) // download the actual block data
		var newTransit [][]byte
		fmt.Println("Preparing Downloading.")
		for _, b := range blocksInTransit {
			if bytes.Compare(b, downloadedBlock) != 0 {
				newTransit = append(newTransit, b) // store others
			}
		}
		/*
			download one block a time, and others store in newTransit to downloadedBlock
		*/
		blocksInTransit = newTransit
		fmt.Printf("Downloaded Blocks")
	}
	if payload.Type == "txs" {
		txId := payload.Items[0]
		if mempool[hex.EncodeToString(txId)].ID == nil { // this tx is not in our mempool
			sendGetData(payload.AddrFrom, "txs", txId)
		}
	}
}

func handleGetData(request []byte, bc *BlockChain) {
	var buff bytes.Buffer
	var payload getDataMsg
	buff.Write(request[commandLength:]) //  different than the previous method
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	fmt.Println("Handling Data Request.")
	if payload.Type == "blocks" {
		block, err := bc.GetBlock([]byte(payload.ID))
		if err != nil {
			log.Panic(err)
		}
		sendBlock(payload.AddrFrom, &block)
	}
	if payload.Type == "txs" {
		txId := hex.EncodeToString(payload.ID)
		tx := mempool[txId]
		sendTx(payload.AddrFrom, &tx)
	}
	/*
		Notice, that we don’t check if we actually have this block or transaction. This is a flaw
	*/
}

func handleBlocks(request []byte, bc *BlockChain) {
	var buff bytes.Buffer
	var payload blockMsg
	buff.Write(request[commandLength:]) //  different than the previous method
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	blockData := payload.Block
	b := Deserialize(blockData)
	fmt.Println("Recevied a new block!") // downloaded a block
	bc.AddBlock(b)                       // add to the chain
	fmt.Printf("Added block %x\n", b.Hash)

	if len(blocksInTransit) > 0 { // download one, still have these to go
		blockHash := blocksInTransit[0] // request to download next block
		sendGetData(payload.AddrFrom, "blocks", blockHash)
		blocksInTransit = blocksInTransit[1:] // update
	} else { // all downloaded and find unspent outputs
		utxo := UTXOSet{bc}
		utxo.Reindex()
	}
}

func handleTxs(request []byte, bc *BlockChain) {
	var buff bytes.Buffer
	var payload txMsg
	buff.Write(request[commandLength:]) //  different than the previous method
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	tx := DeserializeTransaction(payload.Tx)
	mempool[hex.EncodeToString(tx.ID)] = tx

	if nodeAddr == knownAddr[0] {
		fmt.Println("I'm central node")
		/*the central node won’t mine blocks.
		Instead, it’ll forward the new transactions to other nodes in the network.
		*/
		for _, node := range knownAddr {
			if node != knownAddr[0] && node != payload.AddrFrom {
				sendInv(node, "txs", [][]byte{tx.ID})
			}
		}
	} else if len(mempool) >= 1 && len(miningAddr) > 0 {
		fmt.Println("I'm miner node")

		/*
			miningAddress is only set on miner nodes
			. When there are 2 or more transactions in the mempool of the current (miner) node, mining begins.
		*/
	MiningTxs:
		var verifiedTxs []*Transaction
		for id := range mempool {
			t := mempool[id]
			if bc.VerifyTransaction(&t) {
				verifiedTxs = append(verifiedTxs, &t) // Invalid transactions are ignored,
			}
		}
		if len(verifiedTxs) == 0 {
			fmt.Println("All transactions are invalid! Waiting for new ones...")
			return
		}

		cbTx := NewCoinbaseTX(miningAddr, "") // coinbase transaction with the reward
		verifiedTxs = append(verifiedTxs, cbTx)
		newBlock := bc.MineBlock(verifiedTxs) // mined newblock
		utxo := UTXOSet{bc}
		utxo.Reindex()

		fmt.Println("New block is mined!")

		for _, txs := range verifiedTxs {
			txId := hex.EncodeToString(txs.ID)
			delete(mempool, txId) //delete old txs
		}

		/* Every other nodes the current node is aware of*/
		for _, node := range knownAddr {
			if node != nodeAddr {
				sendInv(node, "blocks", [][]byte{newBlock.Hash})
			}
		}

		if len(mempool) > 0 { // still txs need to be mined
			goto MiningTxs
		}
	}

}

func handleConncetion(conn net.Conn, bc *BlockChain) {
	request, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Panic(err)
	}
	command := bytesToCommand(request[:commandLength])
	fmt.Printf("Received %s command\n", command)

	switch command {
	case "addr":
		handleAddr(request)
	case "version":
		handleVersion(request, bc)
	case "getblocks":
		handleGetBlocks(request, bc)
	case "inv":
		handleInv(request, bc)
	case "getdata":
		handleGetData(request, bc)
	case "blocks":
		handleBlocks(request, bc)
	case "txs":
		handleTxs(request, bc)
	default:
		fmt.Println("Command Unknown")
	}
	conn.Close()
}

func StartServer(nodeID, minerAddr string) {
	miningAddr = minerAddr
	nodeAddr = fmt.Sprintf("localhost:%s", nodeID)
	listener, err := net.Listen(protocol, nodeAddr)
	if err != nil {
		log.Panic(err)
	}
	defer listener.Close()

	bc := NewBlockChain(nodeID)

	/*
	 if current node is not the central one, it must send version message to the central node
	*/
	if nodeAddr != knownAddr[0] {
		sendVersion(knownAddr[0], bc)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Panic(err)
		}
		go handleConncetion(conn, bc)
	}
}
