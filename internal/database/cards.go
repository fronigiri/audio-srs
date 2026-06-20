package database

import (
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

func (db *DB) InsertCard(c Card) error {
	_, err := db.conn.Exec(`
        INSERT INTO cards (audio_path, interval, ease_factor, due_date, created_at)
        VALUES (?, ?, ?, ?, ?)`,
		c.AudioPath,
		c.Interval,
		c.EF,
		c.DueDate.Format(time.RFC3339),
		c.CreatedAt.Format(time.RFC3339),
	)
	return err
}

func (db *DB) RemoveCard(id int) error {
	_, err := db.conn.Exec("DELETE FROM cards WHERE id = ?", id)
	return err
}
