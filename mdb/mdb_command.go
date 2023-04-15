package mdb

import (
	"fmt"
	"log"

	"github.com/Scraniel/go-roboto-sensei/command"
	"github.com/bwmarrin/discordgo"
)

const (
	commandVersion = "0.1"

	mdbAnswerKey                     = "answer"
	yesAnswerKey                     = "yes"
	noAnswerKey                      = "no"
	maybeAnswerKey                   = "maybe..."
	mdbKey                           = "mdb"
	mdbCounterOfferKey               = "counter-offer"
	mdbIdKey                         = "id"
	millionDollarsButQuestionCommand = "mdb?"
)

var (
	// Unfortunately must be variables instead of constants so that they're addressable.
	minCounterOfferDollars = float64(1)
	maxCounterOfferDollars = float64(5000000)

	mdbCommandInfo = &discordgo.ApplicationCommand{
		Version:     commandVersion,
		Type:        discordgo.ChatApplicationCommand,
		Name:        "million-dollars-but",
		Description: "You get a million dollars, but...",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        mdbAnswerKey,
				Description: "Would you take the million dollars? Answer `yes`, `no`, or `maybe...` along with your `counter-offer`.",
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{
						Name:  "Yes, I would take the million dollars.",
						Value: yesAnswerKey,
					},
					{
						Name:  "No, I would not take the million dollars.",
						Value: noAnswerKey,
					},
					{
						Name:  "Maybe... would you give me this many dollars instead?",
						Value: maybeAnswerKey,
					},
				},
				Required: true,
			},

			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        mdbCounterOfferKey,
				Description: "If you answered `maybe`, include your `counter-offer` in whole dollars. Must be between `0` and `5000000`",
				MinValue:    &minCounterOfferDollars,
				MaxValue:    maxCounterOfferDollars,
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        mdbIdKey,
				Description: "Optionally include question ID to answer a previously asked question. If not provided, the most recently asked question is answered.",
				Required:    false,
			},
		},
	}
)

func (m *MillionDollarBot) handleMdb(s *discordgo.Session, i *discordgo.InteractionCreate) {
	optionMap := command.ToMap(i.ApplicationCommandData().Options)

	var messageContent string

	defer func() {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseType(discordgo.InteractionResponseChannelMessageWithSource),
			Data: &discordgo.InteractionResponseData{
				Content: messageContent,
			},
		})
	}()

	if i.Member == nil {
		messageContent = "Sorry, you can't use this bot in DMs."
		return
	}

	// Get question ID
	var questionId string
	if val, ok := optionMap[mdbIdKey]; !ok {
		if m.lastQuestionAskedId == "" {
			messageContent = fmt.Sprintf("No one has asked for any questions yet (or my memory has been reset)! Try `/%s`", millionDollarsButQuestionCommand)
			return
		}

		questionId = m.lastQuestionAskedId
	} else {
		questionId = val.StringValue()
		if !m.HasQuestionBeenAsked(questionId) {
			messageContent = fmt.Sprintf("No question with that ID has been asked! Try `/%s` for a new qustion.", millionDollarsButQuestionCommand)
		}
	}

	// Find playerId, parse answer
	playerId := i.Member.User.ID

	var offer uint
	answer := optionMap[mdbAnswerKey].StringValue()
	switch answer {
	case yesAnswerKey:
		offer = OneMillion
	case noAnswerKey:
		offer = 0
	case maybeAnswerKey:
		if val, ok := optionMap[mdbCounterOfferKey]; !ok || val == nil {
			messageContent = "Make sure to include your `counter-offer` if you're answering `maybe...`!"
			return
		} else {
			offer = uint(val.IntValue())
		}
	default:
		messageContent = "You really fucked something up by getting this response. Please tell Danny."
		log.Printf("we don't know how to handle the answer: %v.", answer)
		return
	}

	totalMoney := m.respondToAnswer(questionId, playerId, offer)
	messageContent = fmt.Sprintf(ValidAnswerResponsefmt, i.Member.User.Username, totalMoney)
}

// RespondToAnswer stores the offer to questionId made by playerId and returns the total amount of money the player now has
func (b *MillionDollarBot) respondToAnswer(questionId, playerId string, offer uint) uint {
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
