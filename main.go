package main

import (
	"fmt"
	"log"

	"github.com/fronigiri/audio-srs/internal/database"
)

func main() {
	fmt.Print("welcome")
	db, err := database.StartDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
}
