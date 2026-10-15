package audio

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/flac"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/speaker"
	"github.com/gopxl/beep/wav"

	"github.com/fronigiri/audio-srs/internal/database"
)

type Player struct {
	Open   func(string) (*os.File, error)
	Decode func(*os.File) (beep.StreamSeekCloser, beep.Format, error)
	Play   func(beep.StreamSeekCloser, beep.Format) error
	Cancel context.CancelFunc
	Mu     sync.Mutex
}

func NewPlayer() Player {
	return Player{
		Open: os.Open,
		Decode: func(f *os.File) (beep.StreamSeekCloser, beep.Format, error) {
			return mp3.Decode(f)
		},
		Play: func(streamer beep.StreamSeekCloser, format beep.Format) error {
			speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
			done := make(chan bool)
			speaker.Play(beep.Seq(streamer, beep.Callback(func() {
				done <- true
			})))
			<-done
			return nil
		},
	}
}

func (p *Player) DecodeFile(filePath string) (beep.StreamSeekCloser, beep.Format, *os.File, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, beep.Format{}, nil, err
	}

	ext := strings.ToLower(filepath.Ext(filePath))

	var streamer beep.StreamSeekCloser
	var format beep.Format

	switch ext {
	case ".mp3":
		streamer, format, err = mp3.Decode(f)
	case ".flac":
		streamer, format, err = flac.Decode(f)
	case ".wav":
		streamer, format, err = wav.Decode(f)
	default:
		f.Close()
		return nil, beep.Format{}, nil, fmt.Errorf("unsupported audio format: %s", ext)
	}

	if err != nil {
		f.Close()
		return nil, beep.Format{}, nil, fmt.Errorf("failed to decode %s: %w", ext, err)
	}

	return streamer, format, f, nil
}

func (p *Player) PlayCard(c database.Card) error {
	streamer, format, file, err := p.DecodeFile(c.AudioPath)
	if err != nil {
		return err
	}
	defer file.Close()
	defer streamer.Close()

	return p.Play(streamer, format)
}

func (p *Player) PlaySong(filePath string) error {
	streamer, format, file, err := p.DecodeFile(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	defer streamer.Close()

	return p.Play(streamer, format)
}

func (p *Player) Stop() {
	p.Mu.Lock()
	defer p.Mu.Unlock()

	// speaker.Clear removes all audio streams from the speaker
	speaker.Clear()
}
