package srs

import (
	"testing"
)

func testScheduler(t *testing.T) {
	cases := []struct {
		card   Card
		rating int
	}{
		// normal test
		{
			card:   Card{Repetition: 0, Interval: 1, EF: 2.5},
			rating: 4,
		},
		// forgotten card
		{
			card:   Card{Repetition: 5, Interval: 1, EF: 2.5},
			rating: 2,
		},
		// ease factor below 1.3
		{
			card:   Card{Repetition: 5, Interval: 1, EF: 0.9},
			rating: 3,
		},
	}
}
