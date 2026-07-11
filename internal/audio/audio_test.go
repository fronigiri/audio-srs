package audio

import (
	"os"

	"testing"

	"github.com/fronigiri/audio-srs/internal/database"
	"github.com/gopxl/beep"
)

type fakeStreamer struct{}

func (f fakeStreamer) Stream(samples [][2]float64) (n int, ok bool) {
	return 0, false
}

func (f fakeStreamer) Err() error {
	return nil
}

func (f fakeStreamer) Len() int {
	return 1
}

func (f fakeStreamer) Position() int {
	return 0
}

func (f fakeStreamer) Seek(p int) error {
	return nil
}

func (f fakeStreamer) Close() error {
	return nil
}

func TestPlayCard(t *testing.T) {
	var openedPath string
	decodeCalled := false
	playCalled := false

	player := Player{
		Open: func(path string) (*os.File, error) {
			openedPath = path
			return os.CreateTemp("", "test")
		},
		Decode: func(f *os.File) (beep.StreamSeekCloser, beep.Format, error) {
			decodeCalled = true
			return fakeStreamer{}, beep.Format{}, nil
		},
		Play: func(streamer beep.StreamSeekCloser, format beep.Format) error {
			playCalled = true
			return nil
		},
	}

	card := database.Card{AudioPath: "media/test.mp3"}
	err := player.PlayCard(card)
	if err != nil {
		t.Fatal(err)
	}
	if openedPath != card.AudioPath {
		t.Fatal("wrong path used")
	}
	if !decodeCalled {
		t.Fatal("decode was not called")
	}
	if !playCalled {
		t.Fatal("play was not called")
	}
}
