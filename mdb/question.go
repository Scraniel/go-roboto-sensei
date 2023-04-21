package mdb

import (
	"fmt"
	"log"

	"github.com/Scraniel/go-roboto-sensei/mdb/storage"
	"github.com/bwmarrin/discordgo"
)

const (
	questionCommandVersion = "0.1"

	questionCommandId = "question"

	questionFormat = "You get a million dollars, but... %s (ID: `%s`)"
)

var (
	questionCommandInfo = &discordgo.ApplicationCommand{
		Version:     questionCommandVersion,
		Type:        discordgo.ChatApplicationCommand,
		Name:        questionCommandId,
		Description: "You get a million dollars, but...",
	}
)

type QuestionHandler struct {
	storage storage.Storage
}

func (h *QuestionHandler) Handle(caller *discordgo.Member, options map[string]interface{}) string {

	question, err := h.storage.GetUnaskedQuestion()
	if err == storage.ErrNoMoreRemainingQuestions {
		return "Whoops, all the prewritten questions have been asked! Tell Danny to add more!"
	} else if err != nil {
		log.Printf("unknown error from storage: %v", err)
		return "You shouldn't be able to get here!! Tell Danny please!"
	}

	return fmt.Sprintf(questionFormat, question.Text, question.Id)
}
