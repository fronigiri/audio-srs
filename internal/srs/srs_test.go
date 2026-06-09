package srs

import (
	"testing"
)

func testScheduler(t *testing.T) {
	cases := []struct {
		card             Card
		rating           int
		expectedInterval int
		expectedEF       float64
	}{
		// normal test
		{
			card:             Card{Repetition: 5, Interval: 4, EF: 2.5},
			rating:           4,
			expectedInterval: 10,
			expectedEF:       2.5,
		},
		// forgotten card
		{
			card:             Card{Repetition: 1, Interval: 6, EF: 2.5},
			rating:           2,
			expectedInterval: 1,
			expectedEF:       2.18,
		},
		// ease factor below 1.3. Should be 1.3
		{
			card:             Card{Repetition: 5, Interval: 1, EF: 1.3},
			rating:           3,
			expectedInterval: 1,
			expectedEF:       1.3,
		},
		//first successful review should be 1 day
		{
			card:             Card{Repetition: 0, Interval: 1, EF: 1.3},
			rating:           3,
			expectedInterval: 1,
			expectedEF:       1.3,
		},
		//second successful review should be 6 days
		{
			card:             Card{Repetition: 1, Interval: 1, EF: 1.3},
			rating:           3,
			expectedInterval: 6,
			expectedEF:       1.3,
		},
	}
	for _, c := range cases {
		review(&c.card, c.rating)
		if c.expectedEF != c.card.EF {
			t.Errorf("Error: The expected EF is not the same. Expected: %v, Given: %v", c.expectedEF, c.card.EF)
		}
		if c.expectedInterval != c.card.Interval {
			t.Errorf("Error: The expected interval is not the same. Expected: %d, Given: %d", c.expectedInterval, c.card.Interval)
		}

	}
}
