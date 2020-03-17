package network

import (
	"../audio"
	"../crypto"
)

//Peer refers to another spear user
type Peer struct {
	PublicKey []byte
	Addr      DeterminableAddr

	receiveAudioPacket func(*Packet)
	GetAudioData       func() []float32
	SendOpusData       func([]byte)
}

func (peer *Peer) init(client *Client) {
	audioBuffer := PacketBuffer{}
	opusDecoder := audio.NewDecoder()
	audioPacketID := uint32(0)

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
