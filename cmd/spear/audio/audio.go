package audio

import (
	"time"

	"gopkg.in/hraban/opus.v2"
)

const (
	//FrameSize is the preset frame size for each audio data
	FrameSize = 2880
	//SampleRate is the preset sample rate for each audio data
	SampleRate = 48000

	//FrameDuration is the duration of audio data in each frame
	FrameDuration = time.Second * FrameSize / SampleRate

	channels = 1
)

//CompressAudio uses opus codec to compress raw MONO audio data
func CompressAudio(encoder *opus.Encoder, raw []float32) []byte {
	data := make([]byte, 1024)
	n, err := encoder.EncodeFloat32(raw, data)
	if err != nil {
		panic(err)
	}
	return data[:n]
}

//DecompressAudio uses opus codec to decompress opus data to raw MONO audio data
func DecompressAudio(decoder *opus.Decoder, data []byte) ([]float32, error) {
	var frameSizeMs float32 = 60
	frameSize := channels * frameSizeMs * SampleRate / 1000
	pcm := make([]float32, int(frameSize))
	n, err := decoder.DecodeFloat32(data, pcm)
	if err != nil {
		return nil, err
	}

	return pcm[:n], nil
}

//NewEncoder creates a new Opus encoder
func NewEncoder() *opus.Encoder {
	enc, err := opus.NewEncoder(SampleRate, channels, opus.AppVoIP)
	if err != nil {
		panic(err)
	}
	return enc
}

//NewDecoder creates a new Opus decoder
func NewDecoder() *opus.Decoder {
	dec, err := opus.NewDecoder(SampleRate, channels)
	if err != nil {
		panic(err)
	}
	return dec
}
