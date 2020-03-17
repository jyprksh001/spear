package network

import (
	"math"
	"time"
)

//List of packet ids
const (
	AudioID = 0
	VideoID = 1
)

//Packet refers to a decrypted incoming packet sent by a peer
type Packet struct {
	ID           uint32
	RawData      []byte
	ReceivedTime int64
}

const minimumBufferSize = 5
const maximumFailedPacket = 20
const maximumTimeDifference = 500

//PacketBuffer is a queue of packet, handles packet ordering internally
type PacketBuffer struct {
	idToPacket            map[uint32]*Packet
	currentID             uint32
	rejectedPacketCounter uint
}

func (buffer *PacketBuffer) findSmallestPacket() uint32 {
	index := uint32(math.MaxUint32)
	for k := range buffer.idToPacket {
		if k < index {
			index = k
		}
	}
	return index
}

func (buffer *PacketBuffer) reset() {
	buffer.idToPacket = map[uint32]*Packet{}
	buffer.rejectedPacketCounter = 0
	buffer.currentID = 0
}

//Push a packet to the buffer
func (buffer *PacketBuffer) Push(packet *Packet) {
	if buffer.idToPacket == nil || buffer.rejectedPacketCounter > maximumFailedPacket {
		buffer.reset()
	}

	//Check for outdated packets
	now := time.Now().UnixNano() / 1000000
	outdatedPackets := []uint32{}
	for k, v := range buffer.idToPacket {
		if now-v.ReceivedTime > maximumTimeDifference || v.ID < buffer.currentID {
			outdatedPackets = append(outdatedPackets, k)
		}
	}
	for _, k := range outdatedPackets {
		delete(buffer.idToPacket, k)
	}

	//Append data
	if buffer.currentID <= packet.ID {
		buffer.idToPacket[packet.ID] = packet
	} else {
		buffer.rejectedPacketCounter++
	}
}

//Pop a packet from the buffer
func (buffer *PacketBuffer) Pop() *Packet {
	if len(buffer.idToPacket) < minimumBufferSize {
		return nil
	}
	if buffer.currentID == 0 {
		buffer.currentID = buffer.findSmallestPacket()
	}
	if packet, ok := buffer.idToPacket[buffer.currentID]; ok {
		buffer.currentID++
		buffer.rejectedPacketCounter = 0
		return packet
	}
	buffer.currentID++
	return nil
}
