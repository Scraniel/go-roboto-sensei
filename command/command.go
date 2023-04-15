package command

import "github.com/bwmarrin/discordgo"

type Command struct {
	CommandInfo *discordgo.ApplicationCommand
	Handler     func(s *discordgo.Session, i *discordgo.InteractionCreate)
	Key         string
}

func ToMap(options []*discordgo.ApplicationCommandInteractionDataOption) map[string]*discordgo.ApplicationCommandInteractionDataOption {
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
	}

	return optionMap
}
