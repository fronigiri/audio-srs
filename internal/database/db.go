package database

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	conn *sql.DB
}

func StartDB() (*DB, error) {
	conn, err := sql.Open("sqlite3", "./audio-srs.db")
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	if _, err := conn.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("error enabling foreign keys: %w", err)
	}

	return &DB{conn: conn}, nil
}

func (db *DB) Close() {
	db.conn.Close()
}
