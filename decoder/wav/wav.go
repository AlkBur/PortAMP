package wav

import (
	"PortAMP/blob"
	"errors"
	"log"
	"encoding/binary"
	"fmt"
)

const (
	headerSize  = 36
	chunkHeaderSize = 8
)

const (
	FORMAT_PCM uint16 = 0x0001
	FORMAT_FLOAT = 0x0003
	FORMAT_EXT   = 0xFFFE
	FORMAT_ADPCM_MS  = 0x0002
	FORMAT_ADPCM_IMA = 0x0011
	FORMAT_ALAW  = 0x0006
	FORMAT_MULAW = 0x0007

	///
	debug = false
)

var(
	errorFormat = errors.New("Error: Unsupported WAV file")
)

type waveHeader struct {
	// RIFF header
	RIFF [4]byte 		//0...3		Содержит символы “RIFF” в ASCII кодировке
	FileSize uint32 	//4...7 	Оставшийся размер цепочки
	WAVE [4]byte 		//8...11 	Содержит символы “WAVE”
	FMT [4]byte 		//12...15 	Содержит символы “fmt “
	SizeFmt uint32 		//16...19 	Оставшийся размер цепочки
	FormatTag uint16 	//20...21 	Аудио формат
	Channels uint16 	//22...23	Количество каналов
	SampleRate uint32 	//24...27	Частота дискретизации
	AvgBytesPerSec uint32 //28...31 Количество байт, переданных за секунду воспроизведения
	NBlockAlign uint16 	//32...33 	Количество байт для одного сэмпла, включая все каналы.
	NBitsperSample uint16 //34...35	Количество бит в сэмпле
}

type chunkHeader struct{
	ID [4]byte	//4
	Size uint32 //8
};

type Data struct {
	data []byte
	header *waveHeader
	offset int
}

func memcmp(src []byte, dst string, size int) bool {
	if string(src[:size])== dst {
		return true
	}
	return false
}

func New(data *blob.Data, isVerbose bool) (*Data, error) {
	if data != nil && data.GetDataSize() > headerSize {
		Offset := 0
		this := &Data{data: make([]byte, 0), header : &waveHeader{}}
		size, err := data.ReadData(this.header, Offset)
		if err != nil {
			return nil, err
		}else if size != headerSize {
			return nil, errors.New("Error size header")
		}
		Offset += headerSize

		IsPCM := this.header.FormatTag == FORMAT_PCM
		IsExtFormat := this.header.FormatTag == FORMAT_EXT
		IsFloat := this.header.FormatTag == FORMAT_FLOAT
		IsRIFF := memcmp( this.header.RIFF[:], "RIFF", 4 )
		IsWAVE := memcmp( this.header.WAVE[:], "WAVE", 4 )
		IsADPCM_MS := this.header.FormatTag == FORMAT_ADPCM_MS
		IsADPCM_IMA := this.header.FormatTag == FORMAT_ADPCM_IMA
		IsALaw := this.header.FormatTag == FORMAT_ALAW
		IsMuLaw := this.header.FormatTag == FORMAT_MULAW

		if IsRIFF && IsWAVE && ( !IsPCM || IsADPCM_MS || IsADPCM_IMA || IsALaw || IsMuLaw ) && debug {
			log.Printf( "Channels       : %i\n", this.header.Channels )
			log.Printf( "Sample rate    : %i\n", this.header.SampleRate )
			log.Printf( "Bits per sample: %i\n", this.header.NBitsperSample )
			log.Printf( "Format tag     : %x\n", this.header.FormatTag )
		}

		IsSupportedCodec := IsPCM || IsFloat || IsExtFormat || IsADPCM_MS || IsADPCM_IMA || IsALaw || IsMuLaw

		if IsRIFF && IsWAVE && IsSupportedCodec {
			var ExtraParamSize uint16
			if !IsPCM {
				var CBSize uint16
				if _, err = data.ReadData(&CBSize, Offset ); err != nil {
					return nil, err
				}
				Offset += 2

				ExtraParamSize = CBSize
			}

			if IsExtFormat {
				var SubFormatTag uint16
				data.ReadData(&SubFormatTag, Offset + 6)

				if SubFormatTag == FORMAT_PCM {
					IsFloat = false
				}
				if SubFormatTag == FORMAT_FLOAT {
					IsFloat = true
				}
			}

			Offset = Offset + int(ExtraParamSize)


			var ChunkHeader *chunkHeader

			for {
				LocalChunkHeader := new(chunkHeader)
				size, err := data.ReadData(LocalChunkHeader, Offset)
				if err != nil {
					return nil, err
				}else if size != chunkHeaderSize {
					return nil, errors.New("Error size chunk header")
				}

				if memcmp( LocalChunkHeader.ID[:], "data", 4 ) {
					ChunkHeader = LocalChunkHeader
					break
				} else if memcmp( LocalChunkHeader.ID[:], "fact", 4 ) ||
					memcmp( LocalChunkHeader.ID[:], "LIST", 4 ) ||
					memcmp( LocalChunkHeader.ID[:], "PAD ", 4 ) ||
					memcmp( LocalChunkHeader.ID[:], "JUNK", 4 ) ||
					memcmp( LocalChunkHeader.ID[:], "INFO", 4 ) ||
					memcmp( LocalChunkHeader.ID[:], "CSET", 4 ) ||
					memcmp( LocalChunkHeader.ID[:], "bext", 4 ){

					if debug {
						log.Println("ID=", string(LocalChunkHeader.ID[:]))
					}

					Offset += chunkHeaderSize
					Offset += int(LocalChunkHeader.Size)
				} else {
					return nil, fmt.Errorf("Unknown chunk ID: %s; size = %d; offset = %d\n", string(LocalChunkHeader.ID[:]), LocalChunkHeader.Size, Offset)
				}
			}

			m_DataSize := 0
			if ChunkHeader != nil {
				m_DataSize = int(ChunkHeader.Size)
			}else{
				return nil, errors.New("WAVE: Not found data")
			}

			if IsALaw  {
				this.data = make([]byte, m_DataSize * 2)
				Src := make([]byte, m_DataSize)
				data.ReadAt(Src, Offset + chunkHeaderSize )
				ConvertClamp_ALawToInt16( Src, this.data)
				this.header.NBitsperSample = 16
			} else if IsMuLaw {
				this.data = make([]byte, m_DataSize * 2)
				Src := make([]byte, m_DataSize)
				data.ReadAt(Src, Offset + chunkHeaderSize )
				ConvertClamp_MuLawToInt16( Src, this.data );
				this.header.NBitsperSample = 16
			}else if IsADPCM_MS {
				this.data = make([]byte, m_DataSize * 4)
				Src := make([]byte, m_DataSize)
				data.ReadAt(Src, Offset + chunkHeaderSize )
				ConvertClamp_MSADPCMToInt16( Src, this.data, this.header.NBlockAlign, this.header.Channels == 2 )
				this.header.NBitsperSample = 16
			} else if IsADPCM_IMA {
				this.data = make([]byte, m_DataSize * 4)
				Src := make([]byte, m_DataSize)
				data.ReadAt(Src, Offset + chunkHeaderSize )
				ConvertClamp_IMAADPCMToInt16( Src, this.data, this.header.NBlockAlign, this.header.Channels == 2 )
				this.header.NBitsperSample = 16
			} else if IsFloat {
				if this.header.NBitsperSample == 32 {
					this.data = make([]byte, m_DataSize / 2)
					Src := make([]byte, m_DataSize)
					data.ReadAt(Src, Offset + chunkHeaderSize )
					ConvertClamp_IEEEToInt16(Src, this.data)
				} else if this.header.NBitsperSample == 64 {
					this.data = make([]byte, m_DataSize / 4)
					Src := make([]byte, m_DataSize)
					data.ReadAt(Src, Offset + chunkHeaderSize )
					ConvertClamp_IEEEToInt16(Src, this.data)
				} else {
					return nil, fmt.Errorf("Unknown float format in WAV: %s\n", string(ChunkHeader.ID[:]))
				}
				this.header.NBitsperSample = 16
			} else if ( this.header.NBitsperSample == 24 ) {
				// replace the blob and convert data to 16-bit
				this.data = make([]byte, m_DataSize / 3 * 2)

				Src := make([]byte, m_DataSize)
				data.ReadAt(Src, Offset + chunkHeaderSize )
				ConvertClamp_Int24ToInt16(Src, this.data)

				this.header.NBitsperSample = 16
			} else if ( this.header.NBitsperSample == 32 ) {
				// replace the blob and convert data to 16-bit
				this.data = make([]byte, m_DataSize / 2)

				Src := make([]byte, m_DataSize)
				data.ReadAt(Src, Offset + chunkHeaderSize )
				ConvertClamp_Int32ToInt16(Src, this.data)

				this.header.NBitsperSample = 16
			} else {
				this.data = make([]byte, m_DataSize)
				data.ReadAt(this.data, Offset + chunkHeaderSize )
			}
			if debug || isVerbose {
				fmt.Printf( "PCM WAVE\n" )

				fmt.Printf( "Channels    = %v\n", this.header.Channels )
				fmt.Printf( "Samples/S   = %v\n", this.header.SampleRate )
				fmt.Printf( "Bits/Sample = %v\n", this.header.NBitsperSample )
				fmt.Printf( "Format tag  = %x\n", this.header.FormatTag )
				fmt.Printf( "m_DataSize = %v\n\n", this.GetDataSize() )
			}
			return this, nil
		}else{
			return nil, errorFormat
		}
	}
	return nil, errorFormat
}

func (this *Data)IsStreaming() bool {
	return true
}

func (this *Data)Seek(offset int) {
	this.offset = offset
}

func (this *Data)NumChannels() int {
	return int(this.header.Channels)
}

func (this *Data)SamplesPerSecond() int32 {
	return int32(this.header.SampleRate)
}

func (this *Data)BitsPerSample() int {
	return int(this.header.NBitsperSample)
}

func (this *Data)GetDataSize() int {
	return len(this.data)
}

func (this *Data)Close() {
	this.data = this.data[:0]
}

func (this *Data)GetData() []byte {
	return this.data
}

func (this *Data)StreamData(size int) []byte {
	if this.offset >= len(this.data){
		return nil
	}
	end := this.offset + size
	if end > len(this.data) {
		end = len(this.data)
	}
	defer func(this *Data) {
		this.offset = end
	}(this)
	return this.data[this.offset:end]
}

////////////////////////////////////////////////////
func ConvertClamp_ALawToInt16( Src, Dst []byte) {
	b := make([]byte, 2)
	for i, s := range Src {
		binary.BigEndian.PutUint16(b, uint16(ALawDecodeSample(s)))
		Dst[i*2]=b[0]
		Dst[i*2+1]=b[1]
	}
}

func ConvertClamp_MuLawToInt16( Src, Dst []byte ) {
	b := make([]byte, 2)
	for i, s := range Src {
		binary.BigEndian.PutUint16(b, uint16(MLawDecodeSample(s)))
		Dst[i*2]=b[0]
		Dst[i*2+1]=b[1]
	}
}

func ConvertClamp_MSADPCMToInt16(Src, Dst []byte, BlockAlign uint16, IsStereo bool){
	log.Fatalln("ConvertClamp_MSADPCMToInt16")
	//TODO: ConvertClamp_MSADPCMToInt16
}

func ConvertClamp_IMAADPCMToInt16(Src, Dst []byte, BlockAlign uint16, IsStereo bool){
	log.Fatalln("ConvertClamp_IMAADPCMToInt16")
	//TODO: ConvertClamp_IMAADPCMToInt16
}

func ConvertClamp_IEEEToInt16(Src, Dst []byte)  {
	log.Fatalln("ConvertClamp_IEEEToInt16")
	//TODO: ConvertClamp_IEEEToInt16
}

func ConvertClamp_Int24ToInt16(Src, Dst []byte)  {
	log.Fatalln("ConvertClamp_Int24ToInt16")
	//TODO: ConvertClamp_Int24ToInt16
}

func ConvertClamp_Int32ToInt16(Src, Dst []byte)  {
	log.Fatalln("ConvertClamp_Int32ToInt16")
	//TODO: ConvertClamp_Int32ToInt16
}

func (this *Data)IsEndOfStream() bool {
	return this.offset >= len(this.data)
}