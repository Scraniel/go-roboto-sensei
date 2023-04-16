package mdb

import (
	"github.com/bwmarrin/discordgo"
)

const (
	questionCommandId = "question"
)

var (
	questionCommandInfo = &discordgo.ApplicationCommand{
		Version:     commandVersion,
		Type:        discordgo.ChatApplicationCommand,
		Name:        "million-dollars-but-question",
		Description: "You get a million dollars, but...",
	}
)

type QuestionHandler struct {
	bot *MillionDollarBot
}

func (h *QuestionHandler) Handle(caller *discordgo.Member, options map[string]interface{}) string {

	return ""
}
