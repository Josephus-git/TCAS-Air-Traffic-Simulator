package util

import (
	"os"
	"strings"
)

// CleanInput processes a string, returning a slice of lowercase words with leading/trailing spaces and empty strings removed.
func CleanInput(text string) []string {
	words := []string{}
	sText := strings.Split(strings.TrimSpace(text), " ")
	for _, word := range sText {
		if len(word) != 0 {
			words = append(words, strings.ToLower(word))
		}
	}
	return words
}

// ResetLog removes all logs in logs/
func ResetLog() {
	filesToDelete := []string{
		"logs/airportDetails.txt",
		"logs/airplaneDetails.txt",
		"logs/flightDetails.txt",
		"logs/console_log.txt",
		"logs/tcasLog.txt",
	}

	for _, filePath := range filesToDelete {
		_, err := os.Stat(filePath)
		if os.IsNotExist(err) {
			continue
		}
		os.Remove(filePath)
	}
}
