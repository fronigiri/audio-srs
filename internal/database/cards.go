package database

import (
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Card struct {
	ID         int
	AudioPath  string
	Repetition int
	Interval   int
	EF         float64
	DueDate    time.Time
	CreatedAt  time.Time
	DeckID     int
}

func NewCard(audioPath string) Card {
	now := time.Now()
	return Card{
		AudioPath: audioPath,
		Interval:  1,
		EF:        2.5,
		DueDate:   now,
		CreatedAt: now,
	}
}

func (db *DB) InsertCard(c Card, deckID int) error {
	_, err := db.conn.Exec(`
        INSERT INTO cards (audio_path, interval, ease_factor, due_date, created_at, deck_id)
        VALUES (?, ?, ?, ?, ?, ?)`,
		c.AudioPath,
		c.Interval,
		c.EF,
		c.DueDate.UTC().Format(time.RFC3339),
		c.CreatedAt.UTC().Format(time.RFC3339),
		deckID,
	)
	return err
}

func (db *DB) RemoveCard(id int) error {
	_, err := db.conn.Exec("DELETE FROM cards WHERE id = ?", id)
	return err
}

func (db *DB) GetNextDueCard(deckID int) (Card, error) {
	var c Card
	var dueDateStr string
	var createdAtStr string

	err := db.conn.QueryRow(`
        SELECT id, audio_path, interval, ease_factor, due_date, created_at 
        FROM cards 
        WHERE deck_id = ? AND due_date <= ? 
        ORDER BY due_date ASC 
        LIMIT 1`,
		deckID,
		time.Now().UTC().Format(time.RFC3339),
	).Scan(
		&c.ID,
		&c.AudioPath,
		&c.Interval,
		&c.EF,
		&dueDateStr,
		&createdAtStr,
	)
	if err != nil {
		return Card{}, err
	}
	c.DueDate, err = time.Parse(time.RFC3339, dueDateStr)
	if err != nil {
		return Card{}, fmt.Errorf("failed to parse due_date: %w", err)
	}

	c.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return Card{}, fmt.Errorf("failed to parse created_at: %w", err)
	}

	return c, nil
}

func (db *DB) UpdateCard(c Card) error {
	return nil
}
