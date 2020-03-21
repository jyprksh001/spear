package main

import (
	"fmt"
	"log"
	"os"

	"encoding/base64"

	"github.com/gordonklaus/portaudio"
	"github.com/hexdiract/spear/core/audio"
	"github.com/hexdiract/spear/core/crypto"
	"github.com/hexdiract/spear/core/network"

	"github.com/hexdiract/spear/frontend/config"
	"github.com/hexdiract/spear/frontend/ui"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: spear [config path]")
		return
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
	ui.NewLayout(client)
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
			//log.Println("Error while reading stream: " + err.Error())
		}

		for i := 0; i < len(out); i++ {
			out[i] = 0
		}

		data := audio.CompressAudio(encoder, in)
		for _, peer := range client.PeerList {
			peer.SendOpusData(data)
			if packet := peer.GetAudioData(); packet != nil && len(packet) == audio.FrameSize {
				for i := 0; i < len(packet); i++ {
					out[i] += packet[i] * peer.Volume
				}
			}
		}
		stream.Write()
	}
}
