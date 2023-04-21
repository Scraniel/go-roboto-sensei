package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/Scraniel/go-roboto-sensei/command"
	"github.com/Scraniel/go-roboto-sensei/mdb"
	"github.com/bwmarrin/discordgo"
)

// Bot parameters
var (
	GuildID  = flag.String("guild", "", "Test guild ID. If not passed - bot registers commands globally")
	BotToken = flag.String("token", "", "Bot access token")
	SavePath = flag.String("savePath", "./stats.json", "The file to save / load from.")

	mdbBot *mdb.MillionDollarBot
)

func init() { flag.Parse() }

var (
	session         *discordgo.Session
	commands        []*discordgo.ApplicationCommand
	commandHandlers map[string]command.MessageHandler
)

// Initializes discord library
func init() {
	var err error
	session, err = discordgo.New("Bot " + *BotToken)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}
}

// initializes MDB bot
func init() {
	log.Println("Starting mdb...")
	mdbBot, err := mdb.NewMillionDollarBot(*SavePath)
	if err != nil {
		log.Fatalf("something broke while starting the bot: %v", err)
	}

	commands = make([]*discordgo.ApplicationCommand, 0, len(mdbBot.Commands))
	commandHandlers = make(map[string]command.MessageHandler, len(mdbBot.Commands))

	for _, command := range mdbBot.Commands {
		commands = append(commands, command.CommandInfo)
		commandHandlers[command.Key] = command.Handler
	}

	// This adds the handlers themselves. When a person interacts with the bot via a command, this hook is called and
	// the relevant handler is fired if present.
	session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		optionMap := command.ToMap(i.ApplicationCommandData().Options)
		log.Printf("%s command recieved from %s", i.ApplicationCommandData().Name, i.Member.Nick)
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

		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			messageContent = h.Handle(i.Member, optionMap)
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

	// This part registers the commands in Discord so they pop up when you type '/'.
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
