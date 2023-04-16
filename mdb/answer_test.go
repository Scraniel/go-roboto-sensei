package mdb

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestRespondToAnswer(t *testing.T) {
	questionId := uuid.NewString()
	path := "fake_path"
	player := uuid.NewString()
	t.Run("initializes stats to 0", func(t *testing.T) {
		bot := NewMillionDollarBot(path)

		offer := uint(123456)
		response := bot.updateAnswerStats(questionId, player, offer)
		assert.Equal(t, offer, response)
	})

	t.Run("subsequent answers add to total", func(t *testing.T) {
		bot := NewMillionDollarBot(path)

		offer := uint(123456)
		bot.updateAnswerStats(questionId, player, offer)
		response := bot.updateAnswerStats(questionId+"2", player, offer)
		assert.Equal(t, offer*2, response)
	})

	t.Run("reanswering same question updates", func(t *testing.T) {
		bot := NewMillionDollarBot(path)

		offer := uint(123456)
		bot.updateAnswerStats(questionId, player, offer)

		offer = 1
		response := bot.updateAnswerStats(questionId, player, offer)
		assert.Equal(t, offer, response)
	})
}
