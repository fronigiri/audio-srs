package main

import (
	"fmt"
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/fronigiri/audio-srs/internal/audio"
	"github.com/fronigiri/audio-srs/internal/database"
	"github.com/fronigiri/audio-srs/internal/srs"
)

func chooseFolder(cfg *Config, w fyne.Window) {
	dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
		if err != nil || uri == nil {
			return
		}
		cfg.LibraryPath = uri.Path()
	}, w)
}

func main() {
	db, err := database.StartDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	a := app.NewWithID("com.fronigiri.audio-srs")
	w := a.NewWindow("Audio SRS")
	w.Resize(fyne.NewSize(600, 600))
	cfg := NewConfig()
	ShowPageThree(w, cfg, *db)
	w.ShowAndRun()

}

func ShowHomePage(w fyne.Window, cfg *Config) {
	button := widget.NewButtonWithIcon("Choose Library Path...", theme.FileIcon(), func() {
		chooseFolder(cfg, w)
	})

	button2 := widget.NewButton(
		"Decks",
		func() { println("Deck Button") },
	)
	button3 := widget.NewButton(
		"Browse Library",
		func() { println("Library Button") },
	)
	sidebar := container.New(layout.NewGridLayoutWithRows(3), button, button2, button3)
	content := container.New(layout.NewVBoxLayout(), sidebar)

	w.SetContent(content)
}

func ShowPageTwo(w fyne.Window, cfg *Config, db database.DB) {

	decks, err := db.GetDeckList()
	if err != nil {
		log.Println("Error: unable to list available decks deck list")
	}
	for _, deck := range decks {
		fmt.Println(deck)

	}

	button := widget.NewButton("Create New Deck", func() {
		d := database.Deck{}
		db.CreateDeck(d)
	},
	)
	content := container.New(layout.NewCenterLayout(), button)
	w.SetContent(content)
}

func ShowPageThree(w fyne.Window, cfg *Config, db database.DB) {
	player := audio.NewPlayer()
	// 1. Scan your files
	albums, err := audio.ScanLibrary("./library")
	if err != nil {
		log.Fatalf("Error loading music: %v", err)
	}

	// 2. Prepare the album keys list
	albumNames := make([]string, 0, len(albums))
	for name := range albums {
		albumNames = append(albumNames, name)
	}

	// 3. Fyne Window setup
	w.Resize(fyne.NewSize(750, 450))

	var currentSongs []audio.Song

	// Right list: Songs in selected album
	songList := widget.NewList(
		func() int {
			return len(currentSongs)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Song Title")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			s := currentSongs[id]
			obj.(*widget.Label).SetText(fmt.Sprintf("%s - %s", s.Title, s.Artist))
		},
	)

	// Action when clicking a song: Play
	songList.OnSelected = func(id widget.ListItemID) {
		selectedSong := currentSongs[id]
		player.Stop()
		go func() {
			err := player.PlaySong(selectedSong)
			if err != nil {
				log.Printf("Error playing %s: %v", selectedSong.Title, err)
			}

		}()
	}

	// Left list: Albums
	albumList := widget.NewList(
		func() int {
			return len(albumNames)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Album Name")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(albumNames[id])
		},
	)

	// Action when clicking an album: Update the song list
	albumList.OnSelected = func(id widget.ListItemID) {
		selectedAlbum := albumNames[id]
		currentSongs = albums[selectedAlbum]
		songList.UnselectAll()
		songList.Refresh()
	}

	// Put album list on the left, song list on the right
	splitView := container.NewHSplit(albumList, songList)
	splitView.SetOffset(0.35)

	w.SetContent(splitView)
	w.ShowAndRun()
}

func ShowPageFour(w fyne.Window, cfg *Config, db database.DB, deckID int) {
	//Run the audio SRS

	//Make new player (only need to do this once)
	p := audio.NewPlayer()

	//get cards from deck
	for {
		card, err := db.GetNextDueCard(deckID)
		if err != nil {
			log.Fatal(err)
		}
		if card.DueDate.Day() != time.Now().Day() {
			break
		}

		//stop card audio before playing the next one
		p.Stop()

		//play card and get rating
		p.PlayCard(card)
		rating := 3
		button := widget.NewButton(
			"Again",
			func() { rating = 1 },
		)
		button2 := widget.NewButton(
			"Hard",
			func() { rating = 2 },
		)
		button3 := widget.NewButton(
			"Good",
			func() { rating = 3 },
		)
		button4 := widget.NewButton(
			"Easy",
			func() { rating = 4 },
		)
		container.New(layout.NewGridLayoutWithRows(4), button, button2, button3, button4)

		//schedule said card
		srs.Review(&card, rating)

		//continue until no cards are due
	}
}
