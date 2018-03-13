package blob

import (
	"os"
	"io/ioutil"
	"encoding/binary"
	"bytes"
	"errors"
	"io"
)

var (
	ErrorSeek = errors.New("Error seek")
)

type Data struct {
	fileName string
	m_Data []byte
	off int64
}

func ReadFile(fileName string) (*Data, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	b := &Data{fileName: fileName}
	if b.m_Data, err = ioutil.ReadAll(file); err != nil {
		return nil, err
	}
	return b, nil
}

func (b *Data)GetDataSize() int {
	return len(b.m_Data)
}

func (b *Data)GetFileName() string {
	return b.fileName
}

func (this *Data) ReadData(v interface{}, offset int) (int, error) {
	buf := bytes.NewReader(this.m_Data[offset:])
	size := binary.Size(v)
	err := binary.Read(buf, binary.LittleEndian, v)
	if err != nil {
		return -1, err
	}
	return size, nil
}

func (this *Data) Read(p []byte) (n int, err error) {
	if this.empty() {
		if len(p) == 0 {
			return 0, nil
		}
		return 0, io.EOF
	}
	n = copy(p, this.m_Data[this.off:])
	this.off += int64(n)
	return n, nil
}

func (this *Data) Seek(offset int64, whence int) (n int64, err error) {
	off := this.off
	switch whence {
	case 0:
		off = offset
		if off >= this.Size() || off < 0 {
			err = ErrorSeek
			return
		}
	case 1:
		off = offset + this.off
		if off >= this.Size() || off < 0 {
			err = ErrorSeek
			return
		}
	case 2:
		off = this.Size() - offset
		if off >= this.Size() || off < 0 {
			err = ErrorSeek
			return
		}
	}
	this.off = off
	return this.off, nil
}

func (this *Data) Size() int64 {
	return int64(len(this.m_Data))
}

func (this *Data) empty() bool {
	return int64(len(this.m_Data)) <= this.off
}

func (this *Data) Close() error {
	this.m_Data = this.m_Data[:0]
	this.fileName = ""
	return nil
}

func (this *Data) ReadAt(b []byte, off int) (n int, err error) {
	// cannot modify state - see io.ReaderAt
	if off < 0 {
		return 0, errors.New("blob.ReadAt: negative offset")
	}
	if off >= len(this.m_Data) {
		return 0, io.EOF
	}
	n = copy(b, this.m_Data[off:])
	if n < len(b) {
		err = io.EOF
	}
	return
}

//func (b *Data) Reinterpret_Cast(v interface{}, ptr uintptr) error {
//	buf := bytes.NewReader(unsafeToByte(ptr, binary.Size(v)))
//	return binary.Read(buf, binary.LittleEndian, v)
//}
//
//func (b *Data) GetDataPtr() uintptr {
//	return uintptr(unsafe.Pointer(&b.m_Data[0]))
//}
//
//func unsafeToByte(ptr uintptr, len int) []byte {
//	var b []byte
//	byteHeader := (*reflect.SliceHeader)(unsafe.Pointer(&b))
//	byteHeader.Data = ptr
//	byteHeader.Len = len
//	byteHeader.Cap = len
//	return b
//}

