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

func (db *DB) RemoveDeck(id int) error {
	_, err := db.conn.Exec("DELETE FROM decks WHERE id = ?", id)
	return err
}

func (db *DB) GetDeckList() ([]string, error) {
	rows, err := db.conn.Query(`
		SELECT name FROM decks
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var decks []string

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		decks = append(decks, name)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return decks, nil
}
