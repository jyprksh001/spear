package network

//Packet refers to a decrypted incoming packet sent by a peer
type Packet struct {
	ID           uint32
	RawData      []byte
	ReceivedTime int64
}

//AudioData contains uncompressed audio data
type AudioData struct {
	AudioData []byte
}

//VideoData contains video data
type VideoData struct {
	Metadata  byte
	VideoData []byte
}

//IsAudioPacket returns true if p is an audio packet, false if it is a video packet
func (p *Packet) IsAudioPacket() bool {
	return p.RawData[0] == 0
}

//ToAudioData turns Packet into AudioData
func (p *Packet) ToAudioData() *AudioData {
	if !p.IsAudioPacket() {
		panic("Packet isn't audio packet")
	}

	data := p.RawData[1:]
	//TODO: Handle opus decompression!
	return &AudioData{
		AudioData: data,
	}
}

//ToVideoData turns Packet into VideoData
func (p *Packet) ToVideoData() *VideoData {
	if !p.IsAudioPacket() {
		panic("Packet isn't audio packet")
	}

	data := p.RawData[1:]
	//TODO: Handle H.264 decompression!
	return &VideoData{
		Metadata:  p.RawData[0],
		VideoData: data,
	}
}

//ToBytes turns p into raw bytes
func (p *AudioData) ToBytes() []byte {
	//TODO: Opus compression!
	return p.AudioData
}

//ToBytes turns p into raw bytes
func (p *VideoData) ToBytes() []byte {
	//TODO: H.264 compression!
	return p.VideoData
}
