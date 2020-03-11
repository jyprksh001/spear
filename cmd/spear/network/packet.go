package network

//Packet refers to a decrypted incoming packet sent by a peer
type Packet struct {
	ID           []byte
	RawData      []byte
	ReceivedTime int64
}

//PacketAudio contains uncompressed audio data
type PacketAudio struct {
	AudioData []byte
}

//PacketVideo contains video data
type PacketVideo struct {
	Metadata  byte
	VideoData []byte
}

//IsAudioPacket returns true if p is an audio packet, false if it is a video packet
func (p *Packet) IsAudioPacket() bool {
	return p.RawData[0] == 0
}

//ToPacketAudio turns Packet into PacketAudio
func (p *Packet) ToPacketAudio() *PacketAudio {
	if !p.IsAudioPacket() {
		panic("Packet isn't audio packet")
	}

	data := p.RawData[1:]
	//TODO: Handle opus decompression!
	return &PacketAudio{
		AudioData: data,
	}
}

//ToPacketVideo turns Packet into PacketAudio
func (p *Packet) ToPacketVideo() *PacketVideo {
	if !p.IsAudioPacket() {
		panic("Packet isn't audio packet")
	}

	data := p.RawData[1:]
	//TODO: Handle H.264 decompression!
	return &PacketVideo{
		Metadata:  p.RawData[0],
		VideoData: data,
	}
}

//ToBytes turns p into raw bytes
func (p *PacketAudio) ToBytes() []byte {
	//TODO: Opus compression!
	return p.AudioData
}

//ToBytes turns p into raw bytes
func (p *PacketVideo) ToBytes() []byte {
	//TODO: H.264 compression!
	return p.VideoData
}
