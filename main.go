package main

import (
	"fmt"
	"log"
	"strings"
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
	db, err := database.StartDB("./audio-srs.db", "./internal/sql/schema/001_schema.sql")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	a := app.NewWithID("com.fronigiri.audio-srs")
	w := a.NewWindow("Audio SRS")
	w.Resize(fyne.NewSize(600, 600))
	cfg := NewConfig()
	ShowPageTwo(w, cfg, *db)
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
	var deckNames []string
	selectedDeckName := "" // Tracks which deck the user clicked

	// 1. Study button starts disabled at the bottom
	startSRSButton := widget.NewButton("Select a Deck to Study", func() {
		if selectedDeckName == "" {
			return
		}

		// Look up the deck ID from the database
		deckID, err := db.GetDeckID(selectedDeckName)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}

		// Navigate to Page Four (the SRS review page)
		ShowPageFour(w, cfg, db, deckID)
	})
	startSRSButton.Disable() // Disabled until a deck is clicked

	// 2. Deck List Widget
	deckListWidget := widget.NewList(
		func() int {
			return len(deckNames)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Deck Name Template")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(deckNames[id])
		},
	)

	// When a deck is clicked: activate and update the study button!
	deckListWidget.OnSelected = func(id widget.ListItemID) {
		selectedDeckName = deckNames[id]
		startSRSButton.SetText(fmt.Sprintf("Study: %s", selectedDeckName))
		startSRSButton.Enable()
	}

	// 3. Helper function to query DB and update UI
	refreshDecks := func() {
		var err error
		deckNames, err = db.GetDeckList()
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		// Reset selection on refresh
		selectedDeckName = ""
		startSRSButton.SetText("Select a Deck to Study")
		startSRSButton.Disable()

		deckListWidget.UnselectAll()
		deckListWidget.Refresh()
	}

	// 4. Button to trigger the New Deck form dialog
	newDeckBtn := widget.NewButton("Create New Deck", func() {
		ShowCreateDeckDialog(w, db, func() {
			refreshDecks()
		})
	})

	// Initial fetch
	refreshDecks()

	// 5. Layout: Top has Create button, Bottom has Study button, Center has Deck list
	layout := container.NewBorder(newDeckBtn, startSRSButton, nil, nil, deckListWidget)
	w.SetContent(layout)
}

func ShowPageThree(w fyne.Window, cfg *Config, db database.DB) {
	player := audio.NewPlayer()

	// 1. Scan files
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
	var activeSong *audio.Song // Keep track of selected song for the button

	// --- Bottom Action Panel ---
	selectedLabel := widget.NewLabel("Select a song to play or add to a deck...")
	addBtn := widget.NewButton("Add to Deck", func() {
		if activeSong == nil {
			return
		}

		// Open the pop-up and provide what to do ONCE the deck is selected
		DeckPopUpPage(w, db, func(deckID int) {
			c := database.NewCard(activeSong.Path)
			err := db.InsertCard(c, deckID)
			if err != nil {
				log.Printf("Failed to insert card: %v\n", err)
				dialog.ShowError(err, w)
				return
			}

			dialog.ShowInformation("Success", fmt.Sprintf("Added '%s' to deck!", activeSong.Title), w)
		})
	})

	addBtn.Disable() // Disabled until a song is selected

	bottomBar := container.NewBorder(nil, nil, nil, addBtn, selectedLabel)

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

	// Action when clicking a song: Play + update bottom panel
	songList.OnSelected = func(id widget.ListItemID) {
		selectedSong := currentSongs[id]
		activeSong = &selectedSong

		// Update bottom panel state
		selectedLabel.SetText(fmt.Sprintf("Selected: %s - %s", selectedSong.Title, selectedSong.Artist))
		addBtn.Enable()

		// Play song
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

	// Pin bottomBar to the bottom, splitView fills remaining space
	mainLayout := container.NewBorder(nil, bottomBar, nil, nil, splitView)

	w.SetContent(mainLayout)
	w.ShowAndRun()
}

func ShowPageFour(w fyne.Window, cfg *Config, db database.DB, deckID int) {
	p := audio.NewPlayer()

	var currentCard database.Card

	// 1. Labels for card information
	statusLabel := widget.NewLabel("")
	replayBtn := widget.NewButton("Replay Audio", func() {
		p.Stop()
		go p.PlayCard(currentCard)
	})

	// 2. Define loadNextCard (replaces the 'for' loop)
	var loadNextCard func()

	// 3. Helper to handle grading when any button is tapped
	handleRating := func(rating int) {
		// Stop any audio currently playing
		p.Stop()

		// Update card in memory with SRS algorithm
		srs.Review(&currentCard, rating)

		// Save the updated card to the database
		err := db.UpdateCard(currentCard)
		if err != nil {
			log.Printf("Error updating card: %v", err)
		}

		// Advance to next card!
		loadNextCard()
	}

	// 4. Create the 4 bottom rating buttons
	btnAgain := widget.NewButton("Again", func() { handleRating(1) })
	btnHard := widget.NewButton("Hard", func() { handleRating(2) })
	btnGood := widget.NewButton("Good", func() { handleRating(3) })
	btnEasy := widget.NewButton("Easy", func() { handleRating(4) })

	bottomBar := container.New(layout.NewGridLayoutWithColumns(4),
		btnAgain,
		btnHard,
		btnGood,
		btnEasy,
	)

	// 5. Implementation of loadNextCard
	loadNextCard = func() {
		card, err := db.GetNextDueCard(deckID)
		if err != nil {
			log.Printf("Error getting due card: %v", err)
			w.SetContent(widget.NewLabel("All caught up! No cards due for review."))
			return
		}

		// Check if the card is actually due today or earlier
		if card.DueDate.After(time.Now()) {
			w.SetContent(container.NewCenter(widget.NewLabel("All caught up! No cards due for review.")))
			return
		}

		currentCard = card
		statusLabel.SetText(fmt.Sprintf("Studying Card ID: %d", card.ID))

		// Play audio in a goroutine so UI does not freeze
		p.Stop()
		go p.PlayCard(currentCard)
	}

	// 6. Layout Composition
	centerContent := container.NewVBox(
		statusLabel,
		replayBtn,
	)

	// Pin the 4 buttons to the bottom, card info in center
	mainLayout := container.NewBorder(nil, bottomBar, nil, nil, container.NewCenter(centerContent))
	w.SetContent(mainLayout)

	// Kick off the first card
	loadNextCard()
}

func DeckPopUpPage(w fyne.Window, db database.DB, onConfirm func(int)) {
	//Shows the list of decks and send deck id back
	deckList, err := db.GetDeckList()
	if err != nil {
		log.Println("Error: unable to get deck list")
	}

	if len(deckList) == 0 {
		dialog.ShowInformation("No Decks", "No decks found. Please create a deck first.", w)
		return
	}

	// 1. Prepare string slice for the dropdown options
	deckNames := make([]string, len(deckList))
	for i, name := range deckList {
		deckNames[i] = name
	}

	// 2. Create the dropdown selector
	selectedName := deckNames[0]
	selectWidget := widget.NewSelect(deckNames, func(chosen string) {
		selectedName = chosen
	})
	selectWidget.SetSelectedIndex(0)

	// 3. Modal content
	content := container.NewVBox(
		widget.NewLabel("Select a deck to add this song to:"),
		selectWidget,
	)

	// dialog.ShowCustomConfirm handles the pop-up modal
	dialog.ShowCustomConfirm("Add to Deck", "Add", "Cancel", content, func(ok bool) {
		if !ok {
			return // User canceled
		}

		// Look up the ID for the chosen name
		deckID, err := db.GetDeckID(selectedName)
		if err != nil {
			log.Println("Error fetching deck ID:", err)
			return
		}

		// Trigger the callback with the chosen ID!
		onConfirm(deckID)
	}, w)
}

func ShowCreateDeckDialog(w fyne.Window, db database.DB, onDeckCreated func()) {
	// 1. Create the entry input field
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("e.g. Jazz Standards, Ear Training...")

	// 2. Define the form item(s)
	items := []*widget.FormItem{
		widget.NewFormItem("Deck Name", nameEntry),
	}

	// 3. Show the form dialog
	dialog.ShowForm(
		"Create New Deck",
		"Create",
		"Cancel",
		items,
		func(confirmed bool) {
			if !confirmed {
				return
			}

			deckName := strings.TrimSpace(nameEntry.Text)
			if deckName == "" {
				return
			}

			// Insert into the database
			err := db.CreateDeck(deckName)
			if err != nil {
				dialog.ShowError(err, w)
				return
			}

			// Notify caller to reload/refresh the list of decks
			if onDeckCreated != nil {
				onDeckCreated()
			}
		},
		w,
	)
}
