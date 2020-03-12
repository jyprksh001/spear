package network

import (
	"math"
	"time"
)

//Peer refers to another spear user
type Peer struct {
	PublicKey []byte
	Addr      Addr

	receivedAudioData []*Packet
	receivedVideoData []*Packet
	packetID          uint32
}

//GetAudioData returns a new audio data in the peer buffer
func (peer *Peer) GetAudioData() *AudioData {
	if len(peer.receivedAudioData) < minimumPacketBufferSize {
		return nil
	}

	packet, err := peer.popReceivedPackets(true).ToAudioData()
	if err != nil {
		return nil
	}
	return packet
}

//GetVideoData returns a new audio data in the peer buffer
func (peer *Peer) GetVideoData() *VideoData {
	if peer.receivedVideoData == nil {
		return nil
	}
	if len(peer.receivedVideoData) < minimumPacketBufferSize {
		return nil
	}

	return peer.popReceivedPackets(false).ToVideoData()
}

func (peer *Peer) popReceivedPackets(audio bool) *Packet {
	var receivedPackets []*Packet
	if audio {
		receivedPackets = peer.receivedAudioData
	} else {
		receivedPackets = peer.receivedVideoData
	}

	index := -1
	var smallestID uint32 = math.MaxUint32
	for i, p := range receivedPackets {
		if p.ID < smallestID {
			index = i
			smallestID = p.ID
		}
	}

	if index < 0 {
		return nil
	}

	packet := receivedPackets[index]
	receivedPackets[index] = receivedPackets[len(receivedPackets)-1]
	receivedPackets = receivedPackets[:len(receivedPackets)-1]
	if audio {
		peer.receivedAudioData = receivedPackets
	} else {
		peer.receivedVideoData = receivedPackets
	}

	return packet
}

func (peer *Peer) receivePacket(packet *Packet) {
	var receivedPackets []*Packet
	if packet.IsAudioData() {
		receivedPackets = peer.receivedAudioData
	} else {
		receivedPackets = peer.receivedVideoData
	}

	receivedPackets = append(receivedPackets, packet)

	//Check size max limit
	if len(receivedPackets) > maximumPacketBufferSize {
		peer.popReceivedPackets(packet.IsAudioData())
	}

	keptPackets := []*Packet{}

	for _, p := range receivedPackets {
		currentTimeMillis := time.Now().UnixNano() / 1000000
		if currentTimeMillis-p.ReceivedTime < maximumTimeDifference {
			keptPackets = append(keptPackets, p)
		}
	}

	if packet.IsAudioData() {
		peer.receivedAudioData = keptPackets
	} else {
		peer.receivedVideoData = keptPackets
	}
}
