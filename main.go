package main

import (
	"fmt"
	"log"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
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

type AppContext struct {
	Window fyne.Window
	DB     database.DB
	Config *Config
}

func main() {
	db, err := database.StartDB("./audio-srs.db", "./internal/sql/schema/001_schema.sql")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	a := app.NewWithID("com.fronigiri.audio-srs")
	w := a.NewWindow("Audio SRS")
	w.Resize(fyne.NewSize(900, 600))
	cfg := NewConfig()
	ShowPageTwo(w, cfg, *db)
	// TODO: Added page navigation and a main page
	w.ShowAndRun()
}

func SetupMainWindow(appCtx *AppContext) {
	w := appCtx.Window
	w.Resize(fyne.NewSize(900, 600))

	// 1. Dynamic content area (starts with Page Two / Decks)
	contentArea := container.NewStack()

	// Helper function to swap the active page
	setPage := func(pageContent fyne.CanvasObject) {
		contentArea.Objects = []fyne.CanvasObject{pageContent}
		contentArea.Refresh()
	}

	// 2. Navigation Sidebar Buttons
	btnDecks := widget.NewButton("Decks", func() {
		// Load Deck Browser view
		setPage(makePageTwo(appCtx, setPage))
	})

	btnBrowser := widget.NewButton("Music Browser", func() {
		// Load Song Browser view
		setPage(makePageThree(appCtx))
	})

	// Sidebar container (VBox stacks items vertically)
	sidebar := container.NewVBox(
		widget.NewLabelWithStyle("Audio SRS", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		btnDecks,
		btnBrowser,
	)

	// Wrap sidebar in a border to give it consistent padding or width if desired
	sidebarContainer := container.NewBorder(nil, nil, nil, nil, sidebar)

	// 3. Main Split View: Left = Sidebar, Right = Dynamic Content
	mainLayout := container.NewHSplit(sidebarContainer, contentArea)
	mainLayout.SetOffset(0.2) // Sidebar takes ~20% of width

	w.SetContent(mainLayout)

	// Set initial page (Page Two / Decks)
	setPage(makePageTwo(appCtx, setPage))
}

// Deck list page

func makePageTwo(ctx *AppContext, navigate func(fyne.CanvasObject)) fyne.CanvasObject {
	var deckNames []string
	selectedDeckName := "" // Tracks which deck the user clicked

	// 1. Study button starts disabled at the bottom
	startSRSButton := widget.NewButton("Select a Deck to Study", func() {
		if selectedDeckName == "" {
			return
		}

		// Look up the deck ID from the database
		deckID, err := ctx.DB.GetDeckID(selectedDeckName)
		if err != nil {
			dialog.ShowError(err, ctx.Window)
		}

		// Navigate to Page Four (the SRS review page)
		navigate(makePageFour(ctx, deckID, navigate))
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
		deckNames, err = ctx.DB.GetDeckList()
		if err != nil {
			dialog.ShowError(err, ctx.Window)
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
		ShowCreateDeckDialog(ctx.Window, ctx.DB, func() {
			refreshDecks()
		})
	})

	// Initial fetch
	refreshDecks()

	return container.NewBorder(newDeckBtn, startSRSButton, nil, nil, deckListWidget)
}

// Song Browser Page
func makePageThree(ctx *AppContext) fyne.CanvasObject {
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
	ctx.Window.Resize(fyne.NewSize(750, 450))

	var currentSongs []audio.Song
	var activeSong *audio.Song // Keep track of selected song for the button

	// --- Bottom Action Panel ---
	selectedLabel := widget.NewLabel("Select a song to play or add to a deck...")
	addBtn := widget.NewButton("Add to Deck", func() {
		if activeSong == nil {
			return
		}

		// Open the pop-up and provide what to do ONCE the deck is selected
		DeckPopUpPage(ctx.Window, ctx.DB, func(deckID int) {
			c := database.NewCard(activeSong.Path)
			err := ctx.DB.InsertCard(c, deckID)
			if err != nil {
				log.Printf("Failed to insert card: %v\n", err)
				dialog.ShowError(err, ctx.Window)
				return
			}

			dialog.ShowInformation("Success", fmt.Sprintf("Added '%s' to deck!", activeSong.Title), ctx.Window)
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
			err := player.PlaySong(selectedSong.Path)
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
	return mainLayout
}

func makePageFour(ctx *AppContext, deckID int, navigate func(fyne.CanvasObject)) fyne.CanvasObject {
	p := audio.NewPlayer()

	var currentCard database.Card

	// 1. UI Labels
	statusLabel := widget.NewLabel("")
	answerLabel := widget.NewLabel("") // Displays answer/metadata on reveal
	answerLabel.Hide()                 // Hidden initially

	replayBtn := widget.NewButton("Replay Audio", func() {
		p.Stop()
		go p.PlayCard(currentCard)
	})

	var loadNextCard func()

	// 2. Rating handler (Back Side)
	handleRating := func(rating int) {
		p.Stop()

		// Update card in SRS algorithm & DB
		srs.Review(&currentCard, rating)
		if err := ctx.DB.UpdateCard(currentCard); err != nil {
			log.Printf("Error updating card: %v", err)
		}

		// Advance to next card
		loadNextCard()
	}

	// 3. Create the 4 Rating Buttons (Back side controls)
	btnAgain := widget.NewButton("Again", func() { handleRating(1) })
	btnHard := widget.NewButton("Hard", func() { handleRating(2) })
	btnGood := widget.NewButton("Good", func() { handleRating(3) })
	btnEasy := widget.NewButton("Easy", func() { handleRating(4) })

	ratingBar := container.New(layout.NewGridLayoutWithColumns(4),
		btnAgain, btnHard, btnGood, btnEasy,
	)
	ratingBar.Hide() // Start hidden until user flips the card!

	// 4. "Show Answer" Button (Front side control)
	var showAnswerBtn *widget.Button
	showAnswerBtn = widget.NewButton("Show Answer", func() {

		answerLabel.Show()   // Reveal card details
		showAnswerBtn.Hide() // Hide the show answer button
		ratingBar.Show()     // Reveal the 4 rating buttons
	})

	// Bottom action bar holds either the "Show Answer" button OR the 4 rating buttons
	bottomControls := container.NewStack(showAnswerBtn, ratingBar)

	// 5. Load Next Card (Front Side Setup)
	loadNextCard = func() {
		card, err := ctx.DB.GetNextDueCard(deckID)
		if err != nil {
			log.Printf("No due cards: %v", err)
			ctx.Window.SetContent(container.NewCenter(widget.NewLabel("All caught up! No cards due for review.")))
			return
		}

		currentCard = card

		// Reset UI to FRONT side
		song, err := audio.ScanMetadata(card.AudioPath)
		if err != nil {
			log.Printf("Error: Unable to scan metadata: %v", err)
		}

		statusLabel.SetText(fmt.Sprintf("Card ID: %d", card.ID))
		answerLabel.SetText(fmt.Sprintf("%s - %s", song.Title, song.Album)) // or card notes/title
		answerLabel.Hide()

		showAnswerBtn.Show()
		ratingBar.Hide()

		// Play front-side audio in background
		p.Stop()
		go p.PlayCard(currentCard)
	}

	// 6. Layout
	centerContent := container.NewVBox(
		statusLabel,
		replayBtn,
		answerLabel,
	)

	// Start first card
	loadNextCard()

	return container.NewBorder(nil, bottomControls, nil, nil, container.NewCenter(centerContent))
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
