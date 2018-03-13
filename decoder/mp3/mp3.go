package mp3

import (
	"PortAMP/blob"
	"errors"
	"github.com/hajimehoshi/go-mp3"
	"log"
)

var(
	errorFormat = errors.New("Error mp3 format")
)

type Data struct {
	buf []byte

	decoder *mp3.Decoder
	offset int
}

func New(data *blob.Data, isVerbose bool) (*Data, error) {
	d, err := mp3.NewDecoder(data)
	if err != nil {
		return nil, err
	}
	this := &Data{decoder: d}
	return this, nil
}

func (d *Data)IsStreaming() bool {
	return true
}

func (this *Data)Seek(offset int) {
	this.decoder.Seek(int64(offset), 0)
	this.offset = offset
}

func (this *Data)NumChannels() int {
	return 2
}

func (this *Data)SamplesPerSecond() int32 {
	return int32(this.decoder.SampleRate())
}

func (this *Data)BitsPerSample() int {
	return 16
}

func (this *Data)GetDataSize() int {
	return int(this.decoder.Length())
}

func (this *Data)Close() {
	this.buf = nil
	this.decoder.Close()
}

func (this *Data)StreamData(size int) []byte {
	if this.IsEndOfStream() {
		return nil
	}
	OldSize := len(this.buf)

	if ( size != OldSize ) {
		this.buf = make([]byte, size)
	}

	BytesRead := 0
	for BytesRead < size {
		i, err := this.decoder.Read(this.buf[BytesRead:])
		if err != nil {
			log.Printf("Read error: %v\n", err)
		}else if i <= 0 {
			break
		}
		BytesRead += i
	}
	if BytesRead < size {
		this.buf = this.buf[:BytesRead]
	}
	return this.buf
}

func (this *Data)IsEndOfStream() bool {
	return this.offset >= this.GetDataSize()
}