package main

import (
	"fmt"
	"os"
	"time"

	"encoding/base64"

	"../spear/crypto"
	"../spear/network"

	"./config"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "genkey" {
		privateKey := crypto.RandomBytes(32)
		publicKey := crypto.CreatePublicKey(privateKey)
		fmt.Printf("Public Key: %s\n", base64.StdEncoding.EncodeToString(publicKey))
		fmt.Printf("Secret Key: %s\n", base64.StdEncoding.EncodeToString(privateKey))
		return
	}

	conf, err := config.ParseFile("/home/roger/spear/config.conf")
	if err != nil {
		panic(err)
	}

	client, err := config.CreateClient(conf)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%d peers found.\n", len(client.PeerList))

	if err := client.Initialize(); err != nil {
		panic(err)
	}

	fmt.Println("Starting client.")

	stop := false
	done := make(chan bool, 1)
	go client.Start(&stop, done)
	go sendTrash(client)

	for {
		for _, peer := range client.PeerList {
			packet := peer.GetNewPacket()
			if packet != nil {
				fmt.Println(packet.RawData)
			}
		}
		time.Sleep(5000000)
	}
}

func sendTrash(client *network.Client) {
	for {
		client.SendBytesToAll([]byte("Password123456"))
		time.Sleep(5000000)
	}
}
