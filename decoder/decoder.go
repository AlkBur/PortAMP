package decoder

import (
	"PortAMP/blob"
	"errors"
	"path/filepath"
	"PortAMP/decoder/wav"
	"PortAMP/decoder/mp3"
	"fmt"
)

type DataFormat interface {
	NumChannels() int
	SamplesPerSecond() int32
	BitsPerSample() int
}

type Provider interface{
	//IsStreaming() bool
	Close()
	DataFormat
	GetDataSize() int
	//GetData() []byte
	Seek(offset int)
	StreamData(size int) []byte
	IsEndOfStream() bool
}


func New(data *blob.Data, isVerbose bool) (Provider, error) {
	if data == nil {
		return nil, errors.New("Error blod data")
	}
	Ext := GetFileExt(data.GetFileName())
	switch Ext {
	case ".mp3":
		return mp3.New(data, isVerbose)
	default:
		return wav.New(data, isVerbose)
	}
	return nil, fmt.Errorf("Unknown format: %s", Ext)
}

func GetFileExt(FileName string) string {
	return filepath.Ext(FileName)
}