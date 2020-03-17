package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"time"

	"encoding/base64"

	"../spear/audio"
	"../spear/crypto"
	"../spear/network"
	"github.com/gordonklaus/portaudio"

	"./config"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s [config path]\n", os.Args[0])
	}

	conf, err := config.ParseFile(os.Args[1])
	if err != nil {
		panic(err)
	}

	client, err := config.CreateClient(conf)
	if err != nil {
		panic(err)
	}

	log.Println("Current public key: " + base64.StdEncoding.EncodeToString(crypto.CreatePublicKey(client.SecretKey)))
	log.Printf("%d peers found\n", len(client.PeerList))

	if err := client.Initialize(); err != nil {
		panic(err)
	}

	log.Println("Starting client")

	go startAudioCallback(client)
	time.Sleep(time.Second * math.MaxUint32)
}

func startAudioCallback(client *network.Client) {
	if err := portaudio.Initialize(); err != nil {
		panic(err)
	}

	in := make([]float32, audio.FrameSize)
	out := make([]float32, audio.FrameSize)

	encoder := audio.NewEncoder()

	stream, err := portaudio.OpenDefaultStream(1, 1, audio.SampleRate, audio.FrameSize, in, out)
	if err != nil {
		panic(err)
	}

	if err := stream.Start(); err != nil {
		panic(err)
	}

	for {
		if err := stream.Read(); err != nil {
			log.Println("Error while reading stream: " + err.Error())
		}

		for i := 0; i < len(out); i++ {
			out[i] = 0
		}

		data := audio.CompressAudio(encoder, in)
		for _, peer := range client.PeerList {
			peer.SendOpusData(data)
			if packet := peer.GetAudioData(); packet != nil {
				for i := 0; i < len(packet); i++ {
					out[i] += packet[i]
				}
			}
		}
		stream.Write()
	}
}
