package audio

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/speaker"

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

func (p *Player) PlayCard(c database.Card) error {
	f, err := p.Open(c.AudioPath)
	if err != nil {
		return err
	}
	defer f.Close()

	streamer, format, err := p.Decode(f)
	if err != nil {
		return err
	}
	defer streamer.Close()

	return p.Play(streamer, format)
}

func (p *Player) PlaySong(s Song) error {
	// 1. Stop any currently playing track
	p.Stop()

	// 2. Create a new cancellable context for this song
	p.Mu.Lock()
	ctx, cancel := context.WithCancel(context.Background())
	p.Cancel = cancel
	p.Mu.Unlock()

	// 3. Open and decode audio
	f, err := p.Open(s.Path)
	if err != nil {
		return err
	}
	defer f.Close()

	streamer, format, err := p.Decode(f)
	if err != nil {
		return err
	}
	defer streamer.Close()

	// 4. Play until finished OR until ctx is cancelled
	done := make(chan error, 1)
	go func() {
		done <- p.Play(streamer, format)
	}()

	select {
	case <-ctx.Done():
		// User picked a new song or hit Stop
		return nil
	case err := <-done:
		// Song finished naturally or errored
		return err
	}
}

func (p *Player) Stop() {
	if p.Cancel != nil {
		p.Cancel()
		p.Cancel = nil
	}
}
