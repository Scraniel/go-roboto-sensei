package mdb

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/Scraniel/go-roboto-sensei/command"
)

const (
	OneMillion = uint(1000000)
)

var (
	//TODO: Add proper embedding; go doesn't allow deserializtion direct from embed go:embed mdb.json
	questions map[string]string

	NoSuchQuestionIdError = errors.New("no question with that id exists in our question database")
)

// TODO: See https://github.com/Scraniel/go-roboto-sensei/issues/11. Extract stats functionality to a Stats struct and have both the bot and the handler take a dependency on that. This is the DB layer.
type Stats struct {
	Answered map[string]uint `json:"answered"`
}

func (s Stats) GetTotalMoney() uint {
	var totalMoney uint = 0
	for _, cost := range s.Answered {
		totalMoney += cost
	}

	return totalMoney
}

type MillionDollarBot struct {
	// May want to change this to be an interface or struct. It might be useful to have additional functionality in a "QuestionService" or something.
	questions           map[string]string
	askedQuestions      map[string]bool
	currentStats        map[string]Stats
	savePath            string
	willOverwriteSave   bool
	lock                sync.RWMutex
	Commands            []command.MessageCommand
	lastQuestionAskedId string
}

func NewMillionDollarBot(savePath string) *MillionDollarBot {
	bot := &MillionDollarBot{
		currentStats:      map[string]Stats{},
		savePath:          savePath,
		willOverwriteSave: true,
	}

	bot.Commands = []command.MessageCommand{
		{
			CommandInfo: mdbAnswerCommandInfo,
			Handler:     &AnswerHandler{bot},
			Key:         answerCommandId,
		},
	}

	return bot
}

// getStats returns the current stats for playerId
func (b *MillionDollarBot) getStats(playerId string) Stats {
	b.lock.RLock()
	defer b.lock.RUnlock()

	return b.currentStats[playerId]
}

// saveStats saves the stats currently in memory to disk
func (b *MillionDollarBot) saveStats() error {
	b.lock.Lock()
	defer b.lock.Unlock()

	return saveStats(b.currentStats, b.savePath, b.willOverwriteSave)
}

// loadStats loads the stats that are saved on disk, overwriting whatever is in memory
func (b *MillionDollarBot) loadStats() error {
	b.lock.Lock()
	defer b.lock.Unlock()

	var err error
	if b.currentStats, b.askedQuestions, err = loadStats(b.savePath); errors.Is(err, os.ErrNotExist) {
		b.currentStats = map[string]Stats{}
		b.askedQuestions = map[string]bool{}
	} else if err != nil {
		return fmt.Errorf("error loading stats: %v", err)
	}

	return nil
}

func (b *MillionDollarBot) getQuestion(id string) (string, error) {
	if question, ok := b.questions[id]; !ok {
		return "", NoSuchQuestionIdError
	} else {
		return question, nil
	}
}

func (b *MillionDollarBot) getUnaskedQuestion() string {
	// Simple right now - just generate a random number
	return ""
}

func (b *MillionDollarBot) hasQuestionBeenAsked(id string) bool {
	return b.askedQuestions[id]
}

func saveStats(stats map[string]Stats, filePath string, overwrite bool) error {
	if _, err := os.Stat(filePath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("error checking file on save: %v", err)
	} else if err == nil && !overwrite {
		return errors.New("file exists and overwrite was set to false")
	}

	var file *os.File
	var err error
	if file, err = os.Create(filePath); err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}

	defer file.Close()

	encoder := json.NewEncoder(file)
	if err = encoder.Encode(stats); err != nil {
		return fmt.Errorf("error writing to file: %v", err)
	}

	return nil
}

func loadStats(filePath string) (map[string]Stats, map[string]bool, error) {
	var file *os.File
	var err error
	if file, err = os.Open(filePath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil, err
		} else {
			return nil, nil, fmt.Errorf("error checking file on load: %v", err)
		}
	}

	defer file.Close()

	var stats map[string]Stats
	decoder := json.NewDecoder(file)
	if err = decoder.Decode(&stats); err != nil {
		return nil, nil, fmt.Errorf("error decoding json: %v", err)
	}

	askedQuestions := make(map[string]bool)
	for user := range stats {
		for key := range stats[user].Answered {
			askedQuestions[key] = true
		}
	}
	return stats, askedQuestions, nil
}
