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
	crypto.Init()

	channel := make(chan int)

	if len(os.Args) > 1 && os.Args[1] == "genkey" {
		privateKey := crypto.RandomBytes(32)
		publicKey := crypto.CreatePublicKey(privateKey)
		fmt.Printf("Public Key: %s\n", base64.StdEncoding.EncodeToString(publicKey))
		fmt.Printf("Secret Key: %s\n", base64.StdEncoding.EncodeToString(privateKey))
		return
	}

	var conf config.Config
	conf.ReadFile("config.json")

	var client network.Client
	conf.LoadToClient(&client)

	if err := client.Initialize(); err != nil {
		panic(err)
	}

	fmt.Println("Starting client.")

	stop := false
	done := make(chan bool, 1)
	go client.Start(&stop, done)

	for _, peer := range client.PeerList {
		for i := 0; i < 10; i++ {
			client.SendBytes(peer, []byte("I love you donald!!!"))
		}
		time.Sleep(3000000000)
		fmt.Println(peer.GetNewPacket())
	}

	fmt.Println("Stopping server")
	stop = true

	<-done
	fmt.Println("Server stopped")
}
