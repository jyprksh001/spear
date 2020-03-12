package network

import (
	"math"
	"time"
)

//Peer refers to another spear user
type Peer struct {
	PublicKey []byte
	Addr      Addr

	receivedPackets []*Packet
	packetID        uint32
}

//GetNewPacket returns a new packet in the peer buffer
func (peer *Peer) GetNewPacket() *Packet {
	if peer.receivedPackets == nil {
		return nil
	}
	if len(peer.receivedPackets) < minimumPacketBufferSize {
		return nil
	}

	return peer.popReceivedPackets()
}

func (peer *Peer) popReceivedPackets() *Packet {
	index := -1
	var smallestID uint32 = math.MaxUint32
	for i, p := range peer.receivedPackets {
		if p.ID < smallestID {
			index = i
			smallestID = p.ID
		}
	}

	if index < 0 {
		return nil
	}

	packet := peer.receivedPackets[index]
	peer.receivedPackets[index] = peer.receivedPackets[len(peer.receivedPackets)-1]
	peer.receivedPackets = peer.receivedPackets[:len(peer.receivedPackets)-1]
	return packet
}

func (peer *Peer) receivePacket(packet *Packet) {
	peer.receivedPackets = append(peer.receivedPackets, packet)

	//Check size max limit
	if len(peer.receivedPackets) > maximumPacketBufferSize {
		peer.popReceivedPackets()
	}

	keptPackets := []*Packet{}

	for _, p := range peer.receivedPackets {
		currentTimeMillis := time.Now().UnixNano() / 1000000
		if currentTimeMillis-p.ReceivedTime < maximumTimeDifference {
			keptPackets = append(keptPackets, p)
		}
	}

	peer.receivedPackets = keptPackets
}
