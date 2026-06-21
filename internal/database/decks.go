package database

import (
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Deck struct {
	ID        int
	Name      string
	CreatedAt time.Time
}

func (db *DB) CreateDeck(d Deck) error {
	_, err := db.conn.Exec(`
	INSERT INTO decks (id, name, created_at) VALUES (?,?,?)
	`,
		d.ID,
		d.Name,
		d.CreatedAt.Format(time.RFC3339),
	)
	return err
}
