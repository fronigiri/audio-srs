package srs

import (
	"testing"
)

func testScheduler(t *testing.T) {
	cases := []struct {
		card        Card
		rating      int
		expectedInterval int
		expectedEF  int
	}{
		// normal test
		{
			card:   Card{Repetition: 5, Interval: 4, EF: 2.5},
			rating: 4,
			expectedInterval: 1,
			expectedEF: 
		},
		// forgotten card
		{
			card:   Card{Repetition: 1, Interval: 6, EF: 2.5},
			rating: 2,
			expectedInterval: 1,
			expectedEF:
		},
		// ease factor below 1.3
		{
			card:   Card{Repetition: 5, Interval: 1, EF: 0.9},
			rating: 3,
			
		},
		//first successful review should be 1 day
		{
			card:   Card{Repetition: 0, Interval: 1, EF: 0.9},
			rating: 3,
			expectedInterval: 1,
			expectedEF:
		},
		//second successful review should be 6 days
		{
			card:   Card{Repetition: 1, Interval: 1, EF: 0.9},
			rating: 3,
			expectedInterval: 6,
			expectedEF:
		},
	}
	for _, c := range cases{
		review(c)


	}
}
