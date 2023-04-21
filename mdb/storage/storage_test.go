package storage

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var (
	//go:embed test_stats.json
	testStatsSerialized []byte

	expectedStats = map[string]PlayerStats{
		"first": {
			Answered: map[string]uint{"0": 1000000, "1": 1000000, "2": 2000000, "3": 0, "4": 0, "5": 0},
		},
		"second": {
			Answered: map[string]uint{"0": 1000000, "15": 0},
		},
	}

	expectedAsked = map[string]bool{
		"0":  true,
		"1":  true,
		"2":  true,
		"3":  true,
		"4":  true,
		"5":  true,
		"15": true,
	}
)

const (
	testFileName = "/teststats.json"
)

func TestMain(m *testing.M) {
	testStatsSerialized = stripWhitespace(testStatsSerialized)
	m.Run()
}

func TestSaveStats(t *testing.T) {
	t.Run("serializes to file", func(t *testing.T) {
		savePath := t.TempDir() + testFileName
		err := saveStats(expectedStats, savePath, true)
		assert.NoError(t, err)

		file, err := os.Open(savePath)
		assert.NoError(t, err)
		defer file.Close()

		savedStats, err := ioutil.ReadAll(file)
		assert.NoError(t, err)
		savedStats = stripWhitespace(savedStats)
		assert.Equal(t, testStatsSerialized, savedStats)
	})

	t.Run("file exists", func(t *testing.T) {
		savePath := t.TempDir() + testFileName
		destination, err := os.Create(savePath)
		assert.NoError(t, err)

		written, err := io.Copy(destination, bytes.NewReader(testStatsSerialized))
		assert.NoError(t, err)
		assert.Equal(t, written, int64(len(testStatsSerialized)))
		err = destination.Close()
		assert.NoError(t, err)

		testStats := map[string]PlayerStats{
			"first": {
				Answered: map[string]uint{"0": 2000000},
			},
		}

		t.Run("errors if no overwrite", func(t *testing.T) {
			err := saveStats(testStats, savePath, false)
			assert.EqualError(t, err, "file exists and overwrite was set to false")
		})

		t.Run("overwrites if specified", func(t *testing.T) {
			err := saveStats(testStats, savePath, true)
			assert.NoError(t, err)

			file, err := os.Open(savePath)
			assert.NoError(t, err)
			defer file.Close()

			decoder := json.NewDecoder(file)

			var savedStats map[string]PlayerStats

			err = decoder.Decode(&savedStats)
			assert.NoError(t, err)
			assert.Equal(t, testStats, savedStats)
		})
	})
}

func TestLoadStats(t *testing.T) {
	t.Run("decodes stats successfully", func(t *testing.T) {
		actualStats, actualAsked, err := loadStats("./test_stats.json")
		assert.NoError(t, err)

		assert.Equal(t, expectedStats, actualStats)
		assert.Equal(t, expectedAsked, actualAsked)
	})

	t.Run("surfaces os error", func(t *testing.T) {
		_, _, err := loadStats("./not_a_real_json.json")
		assert.Error(t, err)
		assert.ErrorIs(t, err, os.ErrNotExist)
	})
}

func stripWhitespace(input []byte) []byte {
	buffer := new(bytes.Buffer)
	if err := json.Compact(buffer, input); err != nil {
		fmt.Println(err)
	}

	return buffer.Bytes()
}

func TestRespondToAnswer(t *testing.T) {
	questionId := uuid.NewString()
	path := "fake_path"
	player := uuid.NewString()
	t.Run("initializes stats to 0", func(t *testing.T) {
		storage, err := NewLocalStorage(path)
		assert.NoError(t, err)

		offer := uint(123456)
		response := storage.UpdateStats(questionId, player, offer)
		assert.Equal(t, offer, response.GetTotalMoney())
	})

	t.Run("subsequent answers add to total", func(t *testing.T) {
		storage, err := NewLocalStorage(path)
		assert.NoError(t, err)

		offer := uint(123456)
		storage.UpdateStats(questionId, player, offer)
		response := storage.UpdateStats(questionId+"2", player, offer)
		assert.Equal(t, offer*2, response.GetTotalMoney())
	})

	t.Run("reanswering same question updates", func(t *testing.T) {
		storage, err := NewLocalStorage(path)
		assert.NoError(t, err)

		offer := uint(123456)
		storage.UpdateStats(questionId, player, offer)

		offer = 1
		response := storage.UpdateStats(questionId, player, offer)
		assert.Equal(t, offer, response.GetTotalMoney())
	})
}
