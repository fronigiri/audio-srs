package audio

import (
	"log"
	"os"
	"time"

	"github.com/fronigiri/audio-srs/internal/database"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/speaker"
)

func PlayCard(c database.Card) {
	f, err := os.Open(c.AudioPath)
	if err != nil {
		log.Fatal(err)
	}
	streamer, format, err := mp3.Decode(f)
	if err != nil {
		log.Fatal(err)
	}
	defer streamer.Close()

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	speaker.Play(streamer)

}
