package srs

import (
	"testing"
)

func testScheduler(t *testing.T) {
	cases := []struct {
		card        Card
		rating      int
		expectedRep int
		expectedEF  int
	}{
		// normal test
		{
			card:   Card{Repetition: 0, Interval: 1, EF: 2.5},
			rating: 4,
		},
		// forgotten card
		{
			card:   Card{Repetition: 5, Interval: 4, EF: 2.5},
			rating: 2,
		},
		// ease factor below 1.3
		{
			card:   Card{Repetition: 5, Interval: 1, EF: 0.9},
			rating: 3,
		},
		//first successful review should be 1 day
		{
			card:   Card{Repetition: 5, Interval: 1, EF: 0.9},
			rating: 3,
		},
		//second successful review should be 6 days
		{
			card:   Card{Repetition: 5, Interval: 1, EF: 0.9},
			rating: 3,
		},
	}
}
