package mdb

import (
	_ "embed"
	"fmt"

	"github.com/Scraniel/go-roboto-sensei/command"
	"github.com/Scraniel/go-roboto-sensei/mdb/storage"
)

const (
	OneMillion = uint(1000000)
)

type MillionDollarBot struct {
	storage  storage.Storage
	Commands []command.MessageCommand
}

func NewMillionDollarBot(savePath string) (*MillionDollarBot, error) {
	storage, err := storage.NewLocalStorage(savePath)
	if err != nil {
		return nil, fmt.Errorf("can't create local storage: %w", err)
	}

	bot := &MillionDollarBot{
		storage: storage,
	}

	bot.Commands = []command.MessageCommand{
		{
			CommandInfo: answerCommandInfo,
			Handler:     &AnswerHandler{storage},
			Key:         answerCommandId,
		},
		{
			CommandInfo: questionCommandInfo,
			Handler:     &QuestionHandler{storage},
			Key:         questionCommandId,
		},
	}

	return bot, nil
}
