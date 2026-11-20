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
	Open        func(string) (*os.File, error)
	Decode      func(*os.File) (beep.StreamSeekCloser, beep.Format, error)
	Play        func(beep.StreamSeekCloser, beep.Format) error
	Cancel      context.CancelFunc
	Mu          sync.Mutex
	Ctrl        *beep.Ctrl
	Streamer    beep.StreamSeekCloser
	Path        string
	Initialized bool
}

func NewPlayer() *Player {
	p := &Player{}

	p.Play = func(streamer beep.StreamSeekCloser, format beep.Format) error {
		p.Mu.Lock()
		// 1. Only initialize the speaker once
		if !p.Initialized {
			err := speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
			if err != nil {
				p.Mu.Unlock()
				return err
			}
			p.Initialized = true
		}

		// 2. Clear any old streams from the speaker
		speaker.Clear()

		// 3. Save references on p and wrap in beep.Ctrl
		p.Streamer = streamer
		p.Ctrl = &beep.Ctrl{Streamer: streamer, Paused: false}
		p.Mu.Unlock()

		done := make(chan bool)

		// 4. Play p.ctrl through the speaker, not the raw streamer!
		speaker.Play(beep.Seq(p.Ctrl, beep.Callback(func() {
			done <- true
		})))

		<-done

		p.Mu.Lock()
		p.Ctrl = nil
		p.Streamer = nil
		p.Mu.Unlock()

		return nil
	}

	return p
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

func (p *Player) Restart() {
	p.Mu.Lock()
	defer p.Mu.Unlock()

	if p.Streamer != nil {
		speaker.Lock()
		// Seek(0) rewinds to sample 0
		p.Streamer.Seek(0)
		p.Ctrl.Paused = false // Ensure it's unpaused
		speaker.Unlock()
	}
}

// TogglePause flips between paused and playing
func (p *Player) TogglePause() {
	p.Mu.Lock()
	defer p.Mu.Unlock()

	if p.Ctrl != nil {
		speaker.Lock()
		p.Ctrl.Paused = !p.Ctrl.Paused
		speaker.Unlock()
	}
}
