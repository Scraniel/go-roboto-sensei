package storage

import (
	"crypto/rand"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"
	"strconv"
	"sync"
)

var (
	//go:embed mdb.json
	questionsSerialized []byte

	ErrNoSuchQuestionId         = errors.New("no question with that id exists in our question database")
	ErrNoQuestionsAsked         = errors.New("no questions have been asked yet")
	ErrNoMoreRemainingQuestions = errors.New("there are no remaining unasked questions")
)

type Storage interface {
	GetStats(playerId string) PlayerStats
	UpdateStats(questionId, playerId string, offer uint) PlayerStats

	GetQuestion(id string) (Question, error)
	GetMostRecentQuestionId() (string, error)
	GetUnaskedQuestion() (Question, error)
	HasQuestionBeenAsked(string) bool
}

type LocalStorage struct {
	statsLock            sync.RWMutex
	questionLock         sync.RWMutex
	currentStats         map[string]PlayerStats
	statsSavePath        string
	willOverwriteSave    bool
	askedQuestions       map[string]bool
	mostRecentQuestionId string
	questions            map[string]string
}

func NewLocalStorage(statsSavePath string) (*LocalStorage, error) {
	storage := &LocalStorage{
		currentStats:      map[string]PlayerStats{},
		statsSavePath:     statsSavePath,
		willOverwriteSave: true,
	}

	if err := storage.loadStats(); err != nil {
		return nil, fmt.Errorf("can't load stats from disk: %w", err)
	}

	if err := json.Unmarshal(questionsSerialized, &storage.questions); err != nil {
		return nil, fmt.Errorf("can't parse questions file: %w", err)
	}

	return storage, nil
}

type PlayerStats struct {
	Answered map[string]uint `json:"answered"`
}

func (s PlayerStats) GetTotalMoney() uint {
	var totalMoney uint = 0
	for _, cost := range s.Answered {
		totalMoney += cost
	}

	return totalMoney
}

type Question struct {
	Id   string
	Text string
}

// getStats returns the current stats for playerId
func (s *LocalStorage) GetStats(playerId string) PlayerStats {
	s.statsLock.RLock()
	defer s.statsLock.RUnlock()

	return s.currentStats[playerId]
}

// saveStats saves the stats currently in memory to disk
func (s *LocalStorage) saveStats() error {
	s.statsLock.Lock()
	defer s.statsLock.Unlock()

	return saveStats(s.currentStats, s.statsSavePath, s.willOverwriteSave)
}

func saveStats(stats map[string]PlayerStats, filePath string, overwrite bool) error {
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

// loadStats loads the stats that are saved on disk, overwriting whatever is in memory
func (s *LocalStorage) loadStats() error {
	s.statsLock.Lock()
	defer s.statsLock.Unlock()

	var err error
	if s.currentStats, s.askedQuestions, err = loadStats(s.statsSavePath); errors.Is(err, os.ErrNotExist) {
		s.currentStats = map[string]PlayerStats{}
		s.askedQuestions = map[string]bool{}
	} else if err != nil {
		return fmt.Errorf("error loading stats: %v", err)
	}

	return nil
}

func loadStats(filePath string) (map[string]PlayerStats, map[string]bool, error) {
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

	var stats map[string]PlayerStats
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

// RespondToAnswer stores the offer to questionId made by playerId and returns the total amount of money the player now has
func (s *LocalStorage) UpdateStats(questionId, playerId string, offer uint) PlayerStats {
	// TODO: revisit for perf. Probably not a concern unless you want other servers to use this bot.
	s.statsLock.Lock()
	defer s.statsLock.Unlock()

	var stats PlayerStats
	var ok bool
	if stats, ok = s.currentStats[playerId]; !ok {
		stats = PlayerStats{
			Answered: make(map[string]uint),
		}
	}

	stats.Answered[questionId] = offer
	s.currentStats[playerId] = stats
	return stats
}

func (s *LocalStorage) GetQuestion(id string) (Question, error) {
	if question, ok := s.questions[id]; !ok {
		return Question{}, ErrNoSuchQuestionId
	} else {
		return Question{Text: question, Id: id}, nil
	}
}

func (s *LocalStorage) GetUnaskedQuestion() (Question, error) {
	s.questionLock.Lock()
	defer s.questionLock.Unlock()

	// Simple right now - just generate a random number and iterate if we hit a collision.
	// Later, we should have a pool of IDs that get removed.
	numQuestions := len(s.questions)
	bigId, err := rand.Int(rand.Reader, big.NewInt(int64(numQuestions)))
	if err != nil {
		return Question{}, nil
	}

	// We know this is an int because we have far fewer than 2,147,483,647 hardcoded questions.
	intId := int(bigId.Int64())
	if s.HasQuestionBeenAsked(strconv.Itoa(intId)) {
		foundQuestion := false
		for i := intId + 1; i != intId; i = (i + 1) % numQuestions {
			if !s.HasQuestionBeenAsked(strconv.Itoa(i)) {
				intId = i
				foundQuestion = true
				break
			}
		}

		if !foundQuestion {
			return Question{}, ErrNoMoreRemainingQuestions
		}
	}

	stringId := strconv.Itoa(intId)
	questionText, ok := s.questions[stringId]
	if !ok {
		return Question{}, errors.New("an unknown question ID has been generated")
	}

	s.askedQuestions[stringId] = true
	s.mostRecentQuestionId = stringId

	log.Print(questionText)

	return Question{
		Text: questionText,
		Id:   stringId,
	}, nil
}

func (s *LocalStorage) HasQuestionBeenAsked(id string) bool {
	return s.askedQuestions[id]
}

func (s *LocalStorage) GetMostRecentQuestionId() (string, error) {
	if len(s.mostRecentQuestionId) == 0 {
		return "", ErrNoQuestionsAsked
	}

	return s.mostRecentQuestionId, nil
}
