package audio

import (
	"os"
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

func (p Player) PlayCard(c database.Card) error {
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
