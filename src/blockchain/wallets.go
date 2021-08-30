package main

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

const walletFile = "wallet_%s.dat"

type Wallets struct {
	Wallets map[string]*Wallet
}

func NewWallets(nodeID string) (*Wallets, error) {
	wallets := Wallets{}
	wallets.Wallets = make(map[string]*Wallet)
	err := wallets.LoadFromFile(nodeID) // load many times
	return &wallets, err
}

func (wallets *Wallets) CreateWallet() string {
	wallet := NewWallet()
	address := fmt.Sprintf("%s", wallet.GetAddress())
	wallets.Wallets[address] = wallet
	return address

}

func (w Wallets) GetWallet(address string) Wallet {
	return *w.Wallets[address]
}

func (wallets *Wallets) LoadFromFile(nodeID string) error {
	thiswalletFile := fmt.Sprintf(walletFile, nodeID)
	fmt.Println("here")
	if _, err := os.Stat(thiswalletFile); os.IsNotExist(err) {
		return err
	}
	fmt.Println("here")
	fileContent, err := ioutil.ReadFile(thiswalletFile)
	if err != nil {
		log.Panic(err)
	}
	var ws Wallets
	gob.Register(elliptic.P256()) // gob need to know what kind of type we need to map
	// wallet stores the ecdsa priv key and pubkey
	decoded := gob.NewDecoder(bytes.NewReader(fileContent))
	err = decoded.Decode(&ws) // output to ws
	if err != nil {
		log.Panic(err)
	}
	wallets.Wallets = ws.Wallets
	return nil
}

func (wallets *Wallets) SaveToFile(nodeID string) {
	var content bytes.Buffer
	thiswalletFile := fmt.Sprintf(walletFile, nodeID)

	gob.Register(elliptic.P256())

	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(wallets)
	if err != nil {
		log.Panic(err)
	}

	err = ioutil.WriteFile(thiswalletFile, content.Bytes(), 0644)
	if err != nil {
		log.Panic(err)
	}
}

func (wallets *Wallets) GetAddresses() []string {
	var addresses []string
	for address, _ := range wallets.Wallets {
		addresses = append(addresses, address)
	}
	return addresses
}
