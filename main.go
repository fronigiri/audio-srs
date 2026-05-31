package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", "./myapp.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Verify the connection
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

}
