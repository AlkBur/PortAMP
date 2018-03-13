package main

import (
	"strings"
)

var playlist = NewPlaylist()

type Files []string

type Playlist struct {
	FileNames Files
}

func NewPlaylist() *Playlist {
	return &Playlist{FileNames: make(Files, 0)}
}

func (f *Files) Set(value string) error {
	*f = append(*f, value)
	return nil
}

func (f *Files) String() string {
	buf := strings.Builder{}
	for _, file := range *f {
		if buf.Len() > 0 {
			buf.WriteByte('\n')
		}
		buf.WriteString(file)
	}
	return buf.String()
}

func (p *Playlist)IsEmpty() bool {
	return len(p.FileNames)==0
}

func (p *Playlist)EnqueueTrack(file string) {
	p.FileNames.Set(file)
}

func (p *Playlist)GetNumTracks() int {
	return len(p.FileNames)
}

func (pl *Playlist)GetAndPopNextTrack(loop bool) string {
	result := ""
	if pl.GetNumTracks() > 0 {
		result = pl.FileNames[0]
		pl.FileNames = pl.FileNames[1:]
		if loop {
			pl.EnqueueTrack(result)
		}
	}
	return result
}
