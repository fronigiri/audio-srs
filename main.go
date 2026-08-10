package main

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/fronigiri/audio-srs/internal/database"
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

	a := app.New()
	w := a.NewWindow("Audio SRS")
	w.Resize(fyne.NewSize(600, 600))
	cfg := NewConfig()
	ShowHomePage(w, cfg)
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
	sidebar := container.New(layout.NewGridLayoutWithRows(2), button, button2)
	content := container.New(layout.NewVBoxLayout(), sidebar)

	w.SetContent(content)
}

func ShowPageTwo(w fyne.Window, cfg *Config, db database.DB) {
	//Display available decks and create new decks
	decks, err := db.GetDeckList()
	if err != nil {
		log.Println("Error: unable to list available decks deck list")
	}

	button := widget.NewButton("Create New Deck", func() {
		d := database.Deck{}
		db.CreateDeck(d)
	},
	)
	content := container.New(layout.NewCenterLayout(), button)
	w.SetContent(content)
}
