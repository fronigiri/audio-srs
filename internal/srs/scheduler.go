package srs

import "github.com/fronigiri/audio-srs/internal/database"

func generateEase(card *database.Card, rating int) {
	ef := card.EF + (0.1 - float64(5-rating)*(0.08+float64(5-rating)*0.02))

	if ef < 1.3 {
		ef = 1.3
	}
	card.EF = ef
}

func generateInterval(card *database.Card, rating int) {
	if rating < 3 {
		card.Repetition = 0
		card.Interval = 1
		return
	}
	switch card.Repetition {
	case 0:
		card.Interval = 1
	case 1:
		card.Interval = 6
	default:
		card.Interval = int(float64(card.Interval) * card.EF)
	}
}

func review(card *database.Card, rating int) {
	generateInterval(card, rating)
	card.Repetition++
	generateEase(card, rating)
}
