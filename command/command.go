package command

import "github.com/bwmarrin/discordgo"

type MessageHandler interface {
	Handle(caller *discordgo.Member, options map[string]interface{}) string
}

type MessageCommand struct {
	CommandInfo *discordgo.ApplicationCommand
	Handler     MessageHandler
	Key         string
}

func ToMap(options []*discordgo.ApplicationCommandInteractionDataOption) map[string]interface{} {
	optionMap := make(map[string]interface{}, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt.Value
	}

	return optionMap
}
