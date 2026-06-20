package database

import (
	"time"
)

type Deck struct {
	ID        int
	Name      string
	CreatedAt time.Time
}
