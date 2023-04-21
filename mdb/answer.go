package mdb

import (
	"fmt"
	"log"

	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/Scraniel/go-roboto-sensei/mdb/storage"
	"github.com/bwmarrin/discordgo"
)

const (
	answerCommandVersion = "0.1"
	answerCommandId      = "answer"

	counterOfferOptionId = "counter-offer"

	choiceOptionId = "choice"
	yesChoiceKey   = "yes"
	noChoiceKey    = "no"
	maybeChoiceKey = "maybe..."

	questionIdOptionId = "id"
)

var (
	// Unfortunately must be variables instead of constants so that they're addressable.
	minCounterOfferDollars = float64(1)
	maxCounterOfferDollars = float64(5000000)

	answerCommandInfo = &discordgo.ApplicationCommand{
		Version:     answerCommandVersion,
		Type:        discordgo.ChatApplicationCommand,
		Name:        answerCommandId,
		Description: "Would you take the million dollars? Answer here!",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        choiceOptionId,
				Description: "Would you take the million dollars? Answer `yes`, `no`, or `maybe...` (with a `counter-offer`).",
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
				Description: "Used with `maybe...`, include a `counter-offer`. Must be between `0` and `5000000`.",
				MinValue:    &minCounterOfferDollars,
				MaxValue:    maxCounterOfferDollars,
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        questionIdOptionId,
				Description: "Optional: ID of a previously asked question. Defaults to most recently asked question.",
				Required:    false,
			},
		},
	}
)

type AnswerHandler struct {
	storage storage.Storage
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
		return "Something fucky's going on if you're getting this response. Please tell Danny."
	}

	if questionId == "" {
		mostRecentQuestion, err := h.storage.GetMostRecentQuestionId()
		if err == storage.ErrNoQuestionsAsked {
			return fmt.Sprintf("No one has asked for any questions yet (or my memory has been reset)! Try `/%s`", questionCommandId)
		} else if err != nil {
			log.Printf("GetMostRecentQuestionId returned an error: %v.", err)
			return "You shouldn't be able to get this message. Good job. Plase tell Danny."
		}

		questionId = mostRecentQuestion
	} else if !h.storage.HasQuestionBeenAsked(questionId) {
		return fmt.Sprintf("No question with that ID has been asked! Try `/%s` for a new qustion.", questionCommandId)
	}

	stats := h.storage.UpdateStats(questionId, caller.User.ID, offer)
	return getResponse(questionId, caller.User.ID, offer, stats)
}

func getResponse(questionId string, username string, offer uint, stats storage.PlayerStats) string {
	var answer string
	if offer == 0 {
		answer = "no"
	} else if offer == OneMillion {
		answer = "yes"
	} else {
		printer := message.NewPrinter(language.English)
		answer = printer.Sprintf("yes... but only if you give me $%d!", offer)
	}

	millions := float64(stats.GetTotalMoney()) / float64(OneMillion)
	return fmt.Sprintf("Thanks @%s, you answered `%s`! You've currently got $%d million! To see your full stats, try `/stats`", username, answer, millions)
}
