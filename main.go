package main

import (
	"fmt"
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

func main() {
	db, err := database.StartDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	a := app.New()
	w := a.NewWindow("Audio SRS")
	w.Resize(fyne.NewSize(600, 600))

	button := widget.NewButtonWithIcon("Choose Library Path...", theme.FileIcon(), func() {
		dialog.ShowFolderOpen(func(rc fyne.ListableURI, err error) {
			if err != nil || rc == nil {
				return
			}
			fmt.Println(rc.Path())
		}, w)
	})

	button2 := widget.NewButton(
		"Click me",
		func() { println("Five Seven!") },
	)
	sidebar := container.New(layout.NewGridLayoutWithColumns(2), button, button2)
	content := container.New(layout.NewVBoxLayout(), sidebar)

	w.SetContent(content)

	w.ShowAndRun()

}
