package database

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	conn *sql.DB
}

func StartDB(dbPath, schemaPath string) (*DB, error) {
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	if _, err := conn.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("error enabling foreign keys: %w", err)
	}

	schemaBytes, err := os.ReadFile(schemaPath)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("could not read schema file: %w", err)
	}

	if _, err := conn.Exec(string(schemaBytes)); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to run schema: %w", err)
	}

	return &DB{conn: conn}, nil
}

func (db *DB) Close() {
	db.conn.Close()
}
