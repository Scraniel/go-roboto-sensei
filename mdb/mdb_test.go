package mdb

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	//go:embed test_stats.json
	testStatsSerialized []byte

	expectedStats = map[string]Stats{
		"first": Stats{
			AnsweredYesIds: map[uint]uint{0: 1000000, 1: 1000000, 2: 2000000},
			AnsweredNoIds:  []uint{3, 4, 5},
		},
		"second": Stats{
			AnsweredYesIds: map[uint]uint{0: 1000000},
			AnsweredNoIds:  []uint{},
		},
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

		testStats := map[string]Stats{
			"first": Stats{
				AnsweredYesIds: map[uint]uint{0: 2000000},
				AnsweredNoIds:  []uint{},
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

			var savedStats map[string]Stats

			err = decoder.Decode(&savedStats)
			assert.NoError(t, err)
			assert.Equal(t, testStats, savedStats)
		})
	})
}

func TestLoadStats(t *testing.T) {
	t.Run("decodes stats successfully", func(t *testing.T) {
		actualStats, err := loadStats("./test_stats.json")
		assert.NoError(t, err)

		assert.Equal(t, expectedStats, actualStats)
	})

	t.Run("surfaces os error", func(t *testing.T) {
		_, err := loadStats("./not_a_real_json.json")
		assert.Error(t, err)
		assert.True(t, strings.HasPrefix(err.Error(), "error checking file on load:"))
	})
}

func stripWhitespace(input []byte) []byte {
	buffer := new(bytes.Buffer)
	if err := json.Compact(buffer, input); err != nil {
		fmt.Println(err)
	}

	return buffer.Bytes()
}
