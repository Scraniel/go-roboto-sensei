package mdb

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/Scraniel/go-roboto-sensei/command"
)

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

const (
	// TODO: Probably move these to a different file that contains more discord related stuff.
	userErrorResponse      = "Dude c'mon, to answer a MDB question please use the following format: `/answer <question ID> <response>`. For example: `/answer 40 yes`"
	noSuchQuestionResponse = "Uhh, there's no question with ID %s."
	answerYesOrNoResponse  = "What are you trying to do? You gotta answer with `yes`, `no`, or `maybe...` along with your `counter-offer`."
	ValidAnswerResponsefmt = "Cool, answer recorded. <@%s>, you've currently got $%d million! To see your full stats, try `/stats`"

	OneMillion = uint(1000000)
)

type MillionDollarBot struct {
	askedQuestions      map[string]bool
	currentStats        map[string]Stats
	savePath            string
	willOverwriteSave   bool
	lock                sync.RWMutex
	Commands            []command.Command
	lastQuestionAskedId string
}

func NewMillionDollarBot(savePath string) *MillionDollarBot {
	bot := &MillionDollarBot{
		currentStats:      map[string]Stats{},
		savePath:          savePath,
		willOverwriteSave: true,
	}

	bot.Commands = []command.Command{
		{
			CommandInfo: mdbCommandInfo,
			Handler:     bot.handleMdb,
			Key:         mdbKey,
		},
	}

	return bot
}

// GetStats returns the current stats for playerId
func (b *MillionDollarBot) GetStats(playerId string) Stats {
	b.lock.RLock()
	defer b.lock.RUnlock()

	return b.currentStats[playerId]
}

// SaveStats saves the stats currently in memory to disk
func (b *MillionDollarBot) SaveStats() error {
	b.lock.Lock()
	defer b.lock.Unlock()

	return saveStats(b.currentStats, b.savePath, b.willOverwriteSave)
}

// LoadStats loads the stats that are saved on disk, overwriting whatever is in memory
func (b *MillionDollarBot) LoadStats() error {
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

func (b *MillionDollarBot) HasQuestionBeenAsked(id string) bool {
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
