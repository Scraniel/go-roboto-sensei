package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/Scraniel/go-roboto-sensei/mdb"
	"github.com/bwmarrin/discordgo"
)

// Bot parameters
var (
	GuildID  = flag.String("guild", "", "Test guild ID. If not passed - bot registers commands globally")
	BotToken = flag.String("token", "", "Bot access token")
	SavePath = flag.String("savePath", "./stats.json", "The file to save / load from.")

	// Unfortunately must be variables instead of constants so that they're addressable.
	minCounterOfferDollars = float64(0)
	maxCounterOfferDollars = float64(5000000)

	mdbBot              *mdb.MillionDollarBot
	lastQuestionAskedId string
)

const (
	// TODO: These should probably be in the mdb package, along with the commands, maybe.
	yesAnswerKey   = "yes"
	noAnswerKey    = "no"
	maybeAnswerKey = "maybe..."

	millionDollarsButCommand         = "mdb"
	millionDollarsButQuestionCommand = "mdb?"
	mdbAnswerKey                     = "answer"
	mdbCounterOfferKey               = "counter-offer"
	mdbIdKey                         = "id"
)

func init() { flag.Parse() }

var (
	session  *discordgo.Session
	commands = []*discordgo.ApplicationCommand{
		{
			Name: "basic-command",
			// All commands and options must have a description
			// Commands/options without description will fail the registration
			// of the command.
			Description: "Basic command",
		},
		{
			Version:     "0.1",
			Type:        0,
			Name:        millionDollarsButCommand,
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
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"basic-command": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Hey there! Congratulations, you just executed your first slash command",
				},
			})
		},
		millionDollarsButCommand: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			// convert the slice into a map
			options := i.ApplicationCommandData().Options
			optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
			for _, opt := range options {
				optionMap[opt.Name] = opt
			}

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
				if lastQuestionAskedId == "" {
					messageContent = fmt.Sprintf("No one has asked for any questions yet! Try /%s", millionDollarsButQuestionCommand)
					return
				}

				questionId = lastQuestionAskedId
			} else {
				// TODO: Validate question has been asked
				questionId = val.StringValue()
			}

			// Find playerId, parse answer
			playerId := i.Member.User.ID

			var counterOffer uint
			answer := optionMap[mdbAnswerKey].StringValue()
			switch answer {
			case yesAnswerKey:
				counterOffer = mdb.OneMillion
			case noAnswerKey:
				counterOffer = 0
			case maybeAnswerKey:
				if val, ok := optionMap[mdbCounterOfferKey]; !ok || val == nil {
					messageContent = "Make sure to include your `counter-offer` if you're answering `maybe...`!"
					return
				} else {
					counterOffer = uint(val.IntValue())
				}
			default:
				messageContent = "You really fucked something up by getting this response. Please tell Danny."
				log.Printf("we don't know how to handle the answer: %v.", answer)
				return
			}

			totalMoney := mdbBot.RespondToAnswer(questionId, playerId, counterOffer)
			messageContent = fmt.Sprintf(mdb.ValidAnswerResponsefmt, i.Member.User.Username, totalMoney)
		},
	}
)

func init() {
	var err error
	session, err = discordgo.New("Bot " + *BotToken)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}
}

func init() {
	session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})
}

func main() {
	session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})

	err := session.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	log.Println("Adding commands...")
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := session.ApplicationCommandCreate(session.State.User.ID, *GuildID, v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
		log.Printf("Command %s added.", v.Name)
	}

	defer session.Close()

	log.Println("Starting mdb...")
	mdbBot = mdb.NewMillionDollarBot(*SavePath)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop

	log.Println("Removing commands...")

	for _, v := range registeredCommands {
		err := session.ApplicationCommandDelete(session.State.User.ID, *GuildID, v.ID)
		if err != nil {
			log.Panicf("Cannot delete '%v' command: %v", v.Name, err)
		}

		log.Printf("Command %s removed.", v.Name)
	}

	log.Println("Gracefully shutting down.")
}
