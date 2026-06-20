package main

import (
	"log"

	"github.com/fronigiri/audio-srs/internal/database"
)

func main() {
	db, err := database.StartDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
}
