package main

import (
	"log"

	"encoding/base64"

	"../spear/audio"
	"../spear/crypto"
	"../spear/network"
	"github.com/gordonklaus/portaudio"

	"./config"
)

func main() {
	conf, err := config.ParseFile("/home/roger/spear/config.conf")
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

	stop := false
	done := make(chan bool, 1)
	go client.Start(&stop, done)
	go startAudioCallback(client)
	<-done
}

func startAudioCallback(client *network.Client) {
	if err := portaudio.Initialize(); err != nil {
		panic(err)
	}

	in := make([]float32, audio.FrameSize)
	out := make([]float32, audio.FrameSize)

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
		for _, peer := range client.PeerList {
			client.SendAudioData(peer, in)
			for i := 0; i < len(out); i++ {
				out[i] = 0
			}

			if packet := peer.GetAudioData(); packet != nil {
				client.SendAudioData(peer, in)
				for i := 0; i < len(out); i++ {
					out[i] += packet.AudioData[i]
				}
			}
			stream.Write()
		}
	}
}
