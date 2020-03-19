package network

import (
	"encoding/base64"
	"time"

	"github.com/hexdiract/spear/core/audio"
	"github.com/hexdiract/spear/core/crypto"
)

//Peer refers to another spear user
type Peer struct {
	PublicKey []byte
	Addr      DeterminableAddr
	Volume    float32
	Name      string

	lastPacketReceived int64
	receiveAudioPacket func(*Packet)
	GetAudioData       func() []float32
	SendOpusData       func([]byte)
}

func (peer *Peer) init(client *Client) {
	audioBuffer := PacketBuffer{}
	opusDecoder := audio.NewDecoder()
	audioPacketID := uint32(0)

	peer.Volume = 1

	peer.receiveAudioPacket = audioBuffer.Push
	peer.GetAudioData = func() []float32 {
		var content []byte = nil
		if packet := audioBuffer.Pop(); packet != nil {
			content = packet.RawData[1:]
		}
		data, err := audio.DecompressAudio(opusDecoder, content)
		if err != nil {
			return nil
		}
		return data
	}
	peer.SendOpusData = func(data []byte) {
		ciphertext := crypto.EncryptBytes(peer.PublicKey, client.SecretKey, append([]byte{0}, data...), audioPacketID)
		client.writeTo(peer, ciphertext)
		audioPacketID++
	}
}

//Status returns connection status from a peer
func (peer *Peer) Status() string {
	if time.Now().Unix()-peer.lastPacketReceived > 5 {
		return "Timeout"
	}
	return "Connected"
}

//DisplayName returns the displayed name on CUI
func (peer *Peer) DisplayName() string {
	if len(peer.Name) == 0 {
		peer.Name = base64.StdEncoding.EncodeToString(peer.PublicKey)
	}
	return peer.Name
}
