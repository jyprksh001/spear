package network

import "../audio"

//Packet refers to a decrypted incoming packet sent by a peer
type Packet struct {
	ID           uint32
	RawData      []byte
	ReceivedTime int64
}

//AudioData contains uncompressed audio data
type AudioData struct {
	AudioData []float32
}

//VideoData contains video data
type VideoData struct {
	Metadata  byte
	VideoData []byte
}

//IsAudioData returns true if p is an audio packet, false if it is a video packet
func (p *Packet) IsAudioData() bool {
	return p.RawData[0] == 0
}

//ToAudioData turns Packet into AudioData
func (p *Packet) ToAudioData() (*AudioData, error) {
	if !p.IsAudioData() {
		panic("Packet isn't audio packet")
	}

	data, err := audio.DecompressAudio(p.RawData[1:])
	return &AudioData{
		AudioData: data,
	}, err
}

//ToVideoData turns Packet into VideoData
func (p *Packet) ToVideoData() *VideoData {
	if p.IsAudioData() {
		panic("Packet is audio packet")
	}

	data := p.RawData[1:]
	//TODO: Handle H.264 decompression!
	return &VideoData{
		Metadata:  p.RawData[0],
		VideoData: data,
	}
}
