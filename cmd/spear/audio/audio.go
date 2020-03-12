package audio

import (
	"gopkg.in/hraban/opus.v2"
)

const (
	//FrameSize is the preset frame size for each audio data
	FrameSize = 2880
	//SampleRate is the preset sample rate for each audio data
	SampleRate = 48000

	channels = 1
)

var encoder *opus.Encoder = nil
var decoder *opus.Decoder = nil

//CompressAudio uses opus codec to compress raw MONO audio data
func CompressAudio(raw []float32) []byte {
	data := make([]byte, 1024)
	n, err := getEncoder().EncodeFloat32(raw, data)
	if err != nil {
		panic(err)
	}
	return data[:n]
}

//DecompressAudio uses opus codec to decompress opus data to raw MONO audio data
func DecompressAudio(data []byte) ([]float32, error) {
	var frameSizeMs float32 = 60
	frameSize := channels * frameSizeMs * SampleRate / 1000
	pcm := make([]float32, int(frameSize))
	n, err := getDecoder().DecodeFloat32(data, pcm)
	if err != nil {
		return nil, err
	}

	return pcm[:n], nil
}

func getEncoder() *opus.Encoder {
	if encoder != nil {
		return encoder
	}

	enc, err := opus.NewEncoder(SampleRate, channels, opus.AppVoIP)
	if err != nil {
		panic(err)
	}
	encoder = enc
	return encoder
}

func getDecoder() *opus.Decoder {
	if decoder != nil {
		return decoder
	}

	dec, err := opus.NewDecoder(SampleRate, channels)
	if err != nil {
		panic(err)
	}
	decoder = dec
	return decoder
}
