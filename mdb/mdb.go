package mdb

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

type Stats struct {
	totalMoney     uint
	AnsweredYesIds map[uint]uint `json:"answered_yes"`
	AnsweredNoIds  []uint        `json:"answered_no"`
}

func (s Stats) GetTotalMoney() uint {
	if s.totalMoney == 0 {
		for _, cost := range s.AnsweredYesIds {
			s.totalMoney += cost
		}
	}

	return s.totalMoney
}

const (
	userErrorResponse       = "Dude c'mon, to answer a MDB question please use the following format: `/answer <question ID> <response>`. For example: `/answer 40 yes`"
	noSuchQuestionResponse  = "Uhh, there's no question with ID %s."
	answerYesOrNoResponse   = "What are you trying to do? You gotta answer with either 'yes' or 'no'."
	alreadyAnsweredResponse = "You already answered that one, cheater. Try /stats to see what you've got."
	validAnswerResponsefmt  = "Cool, answer recorded. <@%s>, you've currently got $%d million! To see your full stats, try /stats"
)

var (
	currentStats map[string]Stats
)

func RespondToAnswer(wouldTakeMoney bool, counterOffer *uint) (int, error) {
	if !wouldTakeMoney {

	} else if counterOffer != nil {

	} else {

	}

	return 0, nil
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

func loadStats(filePath string) (map[string]Stats, error) {
	var file *os.File
	var err error
	if file, err = os.Open(filePath); err != nil {
		return nil, fmt.Errorf("error checking file on load: %v", err)
	}

	defer file.Close()

	var stats map[string]Stats
	decoder := json.NewDecoder(file)
	if err = decoder.Decode(&stats); err != nil {
		return nil, fmt.Errorf("error decoding json: %v", err)
	}

	return stats, nil
}
