package mdb

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
)

const (
	commandVersion = "0.1"

	answerCommandId = "answer"

	questionIdOptionId = "id"

	counterOfferOptionId = "counter-offer"

	choiceOptionId = "choice"
	yesChoiceKey   = "yes"
	noChoiceKey    = "no"
	maybeChoiceKey = "maybe..."

	ValidAnswerResponsefmt = "Cool, answer recorded. <@%s>, you've currently got $%d million! To see your full stats, try `/stats`"
)

var (
	// Unfortunately must be variables instead of constants so that they're addressable.
	minCounterOfferDollars = float64(1)
	maxCounterOfferDollars = float64(5000000)

	mdbAnswerCommandInfo = &discordgo.ApplicationCommand{
		Version:     commandVersion,
		Type:        discordgo.ChatApplicationCommand,
		Name:        "million-dollars-but-answer",
		Description: "Would you take the million dollars? Answer here!",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        choiceOptionId,
				Description: "Would you take the million dollars? Answer `yes`, `no`, or `maybe...` along with your `counter-offer`.",
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{
						Name:  "Yes, I would take the million dollars.",
						Value: yesChoiceKey,
					},
					{
						Name:  "No, I would not take the million dollars.",
						Value: noChoiceKey,
					},
					{
						Name:  "Maybe... I'd do it for this much:",
						Value: maybeChoiceKey,
					},
				},
				Required: true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        counterOfferOptionId,
				Description: "If you answered `maybe...`, include your `counter-offer` in whole dollars. Must be between `0` and `5000000`. Will be ignored if you answered `yes` or `no`.",
				MinValue:    &minCounterOfferDollars,
				MaxValue:    maxCounterOfferDollars,
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        questionIdOptionId,
				Description: "Optionally include question ID to answer a previously asked question. If not provided, the most recently asked question is answered.",
				Required:    false,
			},
		},
	}
)

type AnswerHandler struct {
	bot *MillionDollarBot
}

func (h *AnswerHandler) Handle(caller *discordgo.Member, options map[string]interface{}) string {
	var questionId string
	if val, ok := options[questionIdOptionId]; !ok {
		// questionId is already empty string
	} else {
		questionId = val.(string)
	}

	var offer uint
	choice := options[choiceOptionId].(string)
	switch choice {
	case yesChoiceKey:
		offer = OneMillion
	case noChoiceKey:
		offer = 0
	case maybeChoiceKey:
		if val, ok := options[counterOfferOptionId]; !ok || val == nil {
			return "Make sure to include your `counter-offer` if you're answering `maybe...`!"
		} else {
			offer = val.(uint)
		}
	default:
		log.Printf("we don't know how to handle the answer: %v.", choice)
		return "You really fucked something up by getting this response. Please tell Danny."
	}

	return h.answer(caller, questionId, offer)
}

func (h *AnswerHandler) answer(caller *discordgo.Member, questionId string, offer uint) string {
	if questionId == "" {
		if h.bot.lastQuestionAskedId == "" {
			return fmt.Sprintf("No one has asked for any questions yet (or my memory has been reset)! Try `/%s`", questionCommandId)
		}

		questionId = h.bot.lastQuestionAskedId
	} else if !h.bot.hasQuestionBeenAsked(questionId) {
		return fmt.Sprintf("No question with that ID has been asked! Try `/%s` for a new qustion.", questionCommandId)
	}

	totalMoney := h.bot.updateAnswerStats(questionId, caller.User.ID, offer)
	return fmt.Sprintf(ValidAnswerResponsefmt, caller.User.Username, totalMoney)
}

// RespondToAnswer stores the offer to questionId made by playerId and returns the total amount of money the player now has
// TODO: Should this be in MillionDollarBot or AnswerHandler?
func (b *MillionDollarBot) updateAnswerStats(questionId, playerId string, offer uint) uint {
	// TODO: revisit for perf. Probably not a concern unless you want other servers to use this bot.
	b.lock.Lock()
	defer b.lock.Unlock()

	var playerStats Stats
	var ok bool
	if playerStats, ok = b.currentStats[playerId]; !ok {
		playerStats = Stats{
			Answered: make(map[string]uint),
		}
	}

	var money uint
	playerStats.Answered[questionId] = offer

	money = playerStats.GetTotalMoney()

	b.currentStats[playerId] = playerStats
	return money
}
