package main

import (
	"PortAMP/al"
	"PortAMP/decoder"
	"sync"
	"errors"
	"fmt"
	"time"
	"log"
)

const BUFFER_DURATION = 250 // milliseconds
const BUFFER_SIZE = 44100 * 2 * 2 * BUFFER_DURATION / 1000

var errorProvider = errors.New("Error BindDataProvider")

type AudioSystem struct {
	m_IsPendingExit chan struct{}

	m_ActiveSources []*AudioSource
	sync.RWMutex
}


type AudioSource struct {
	m_Looping bool

	m_AudioSystem *AudioSystem
	m_DataProvider decoder.Provider

	// OpenAL stuff
	m_SourceID al.Source
	m_BufferID []al.Buffer
}

func NewAudioSource(audio *AudioSystem) *AudioSource {
	this := &AudioSource{
		m_AudioSystem: audio,
		m_Looping: false,
	}

	this.m_SourceID = al.GenSources( 1)[0]

	return this
}

func NewAudioSystem() (*AudioSystem, error) {
	err := al.OpenDevice()
	if err != nil {
		return nil, fmt.Errorf("al: cannot open the default audio device: %v", err)
	}

	s := &AudioSystem{
		m_IsPendingExit: make(chan struct{}, 1),
		m_ActiveSources: make([]*AudioSource, 0, 2),
	}
	if IsVerbose() {
		s.DebugPrintVersion()
	}
	return s, nil
}

func (this *AudioSource) genBuffers(n int) error {
	if len(this.m_BufferID) > 0 {
		al.DeleteBuffers(this.m_BufferID...)
	}
	this.m_BufferID = al.GenBuffers(n)
	return al.GetError()
}

func (this *AudioSource) BindDataProvider(Provider decoder.Provider) error {
	this.m_DataProvider = Provider

	if this.m_DataProvider == nil {
		return errorProvider
	}

	return this.genBuffers(2)
}

func (s *AudioSource) PrepareBuffers() error {
	if s.m_DataProvider == nil {
		return errorProvider
	}

	State := s.m_SourceID.State()

	if State != al.Paused {
		s.UnqueueAllBuffers()

		BuffersToQueue := 2

		s.StreamBuffer( s.m_BufferID[0], BUFFER_SIZE )
		if s.StreamBuffer( s.m_BufferID[1], BUFFER_SIZE ) == 0 {
			if s.IsLooping() {
				s.m_DataProvider.Seek(0)
				s.StreamBuffer(s.m_BufferID[1], BUFFER_SIZE)
			} else {
				BuffersToQueue = 1
			}
		}

		s.m_SourceID.QueueBuffers(s.m_BufferID[:BuffersToQueue]...)

		s.m_AudioSystem.RegisterSource(s)
	}
	return al.GetError()
}

func (s *AudioSource) UnqueueAllBuffers() {
	NumQueued := s.m_SourceID.BuffersQueued()

	if NumQueued > 0 {
		s.m_SourceID.UnqueueBuffers(s.m_BufferID[:NumQueued]...)
	}
}

func (s *AudioSource) StreamBuffer(BufferID al.Buffer, Size int) int {
	if s.m_DataProvider == nil {
		return 0
	}

	Data := s.m_DataProvider.StreamData(Size)
	ActualSize := len(Data)

	if ActualSize == 0 {
		return 0
	}

	BufferID.BufferData(
		s.FormatOpenAL(),
		Data,
		s.m_DataProvider.SamplesPerSecond(),
	)

	return ActualSize
}

func (this *AudioSource) EnqueueOneBuffer() {
	BufID := [1]al.Buffer{}
	this.m_SourceID.UnqueueBuffers(BufID[:]...)

	Size := this.StreamBuffer(BufID[0], BUFFER_SIZE )

	if this.m_DataProvider.IsEndOfStream() {
		fmt.Println("IsEndOfStream")
		if this.IsLooping() {
			this.m_DataProvider.Seek(0)
			if Size==0 {
				Size = this.StreamBuffer(BufID[0], BUFFER_SIZE)
			}
		}
	}

	if Size > 0 {
		this.m_SourceID.QueueBuffers(BufID[:]...)
	} else {
		if !this.IsPlaying() {
			this.Stop()
		}
	}
}

func (this *AudioSource) UpdateBuffers() {
	if this.m_DataProvider == nil {
		return
	}

	if err := al.GetError(); err != nil {
		log.Printf("Update Buffers: %v\n", err)
		return
	}

	for bp := this.m_SourceID.BuffersProcessed(); bp > 0; bp--{
		this.EnqueueOneBuffer()
	}
}

func (s *AudioSource) Play() error {
	if s.IsPlaying() {
		return nil
	}

	if s.m_DataProvider == nil {
		return errorProvider
	}

	err := s.PrepareBuffers()
	if err != nil {
		return  err
	}

	al.PlaySources(s.m_SourceID)
	return al.GetError()
}

func (s *AudioSource) Stop() {
	al.StopSources(s.m_SourceID)

	s.UnqueueAllBuffers()

	s.m_AudioSystem.UnregisterSource(s)

	if s.m_DataProvider != nil {
		s.m_DataProvider.Seek(0)
	}
}

func (s *AudioSource) IsPlaying() bool {
	State := s.m_SourceID.State()

	return State == al.Playing
}

func (s *AudioSource) SetLooping( Looping bool ) {
	s.m_Looping = Looping
}

func (this *AudioSystem)Close() {
	this.Stop()
	defer close(this.m_IsPendingExit)

	this.Lock()
	defer this.Unlock()

	for i, s := range this.m_ActiveSources {
		s.Close()
		this.m_ActiveSources[i] = nil
	}
	this.m_ActiveSources = this.m_ActiveSources[:0]

	al.CloseDevice()
}

func (s *AudioSystem) DebugPrintVersion() {
	fmt.Printf( "OpenAL version : %s\n", al.Vendor() )
	fmt.Printf( "OpenAL vendor  : %s\n", al.Vendor() )
	fmt.Printf( "OpenAL renderer: %s\n", al.Renderer() )
	fmt.Printf( "OpenAL extensions:\n%s\n\n", al.Extensions() )
}

func (s *AudioSystem) Start() {
	go func() {
		bExit := false
		for !bExit {
			select {
			case <-s.m_IsPendingExit:
				bExit = true
			case <-time.After(time.Millisecond * 10):
				sources := s.GetLockedSources()

				for _, i := range sources {
					i.UpdateBuffers()
				}
			}
		}
	}()
}

func (s *AudioSystem) RegisterSource(Source *AudioSource) {
	s.Lock()
	defer s.Unlock()

	for _, i := range s.m_ActiveSources {
		if i == Source {
			return
		}
	}
	s.m_ActiveSources = append(s.m_ActiveSources, Source)
}

func (source *AudioSource) Close() {
	source.Stop()
	al.DeleteSources(source.m_SourceID)
	if len(source.m_BufferID) > 0 {
		al.DeleteBuffers(source.m_BufferID...)
	}
}

func (s *AudioSource)IsLooping() bool {
	return s.m_Looping
}

func (s *AudioSystem) GetLockedSources() []*AudioSource {
	s.RLock()
	defer s.RUnlock()
	return s.m_ActiveSources
}

func (s *AudioSystem) UnregisterSource(Source *AudioSource) {
	s.Lock()
	defer s.Unlock()

	for index, i := range s.m_ActiveSources {
		if i == Source {
			s.m_ActiveSources[index] = nil
			s.m_ActiveSources = append(s.m_ActiveSources[:index], s.m_ActiveSources[index+1:]...)
			return
		}
	}
}

func (s *AudioSystem) Stop() {
	s.m_IsPendingExit <- struct{}{}
}

func (s *AudioSystem) CreateAudioSource() *AudioSource {
	return NewAudioSource(s)
}

func (this *AudioSource)FormatOpenAL() uint32 {
	if this.m_DataProvider != nil {
		switch this.m_DataProvider.BitsPerSample() {
		case 8:
			if this.m_DataProvider.NumChannels() == 2 {
				return al.FormatStereo8
			} else {
				return al.FormatMono8
			}
		case 16:
			if this.m_DataProvider.NumChannels() == 2 {
				return al.FormatStereo16
			} else {
				return al.FormatMono16
			}
		}
	}
	return al.FormatMono8
}

