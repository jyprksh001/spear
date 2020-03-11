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

	var conf config.Config
	conf.ReadFile("/home/roger/spear/config.json")
	fmt.Println(conf.Peers)

	var client network.Client
	conf.LoadToClient(&client)

	if err := client.Initialize(); err != nil {
		panic(err)
	}

	fmt.Println("Starting client.")

	stop := false
	done := make(chan bool, 1)
	go client.Start(&stop, done)
	go sendTrash(&client)

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
		client.SendBytesToAll([]byte("I love you donald!!!"))
	}
}
