package main

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"strings"

	logging "github.com/op/go-logging"
	"github.com/urfave/cli"
)

// Configures the local logger
func setupLogging() {
	var format = logging.MustStringFormatter("%{color}%{time:15:04:05.000} " +
		"%{shortfunc} ▶ %{level} " +
		"%{color:reset} %{message}")
	backend := logging.NewLogBackend(os.Stderr, "", 0)
	formatter := logging.NewBackendFormatter(backend, format)
	logging.SetBackend(formatter)
	if enableDebugLogging {
		logging.SetLevel(logging.DEBUG, "uploader")
	} else {
		logging.SetLevel(logging.INFO, "uploader")
	}
}

// Asks the user interactively a confirmation question; if the user declines
// (answers anything but Y or defalut - empty string - stop execution.
func mustConfirm(format string, args ...interface{}) {
	if skipConfirmation {
		return
	}

	text := fmt.Sprintf(format, args...)
	fmt.Print(fmt.Sprintf("%v [Y/n] ", text))
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')

	if input != "Y\n" && input != "\n" {
		os.Exit(1)
	}
}

// Finds the longest file name
func findLongestName(infos []os.FileInfo) int {
	longest := 0

	for _, info := range infos {
		if len(info.Name()) > longest {
			longest = len(info.Name())
		}
	}

	return longest
}

// There seems to be a bug in the GolbalBoolT API, or I am using
// it incorrectly - however, this method checks for the presence of a BoolT
// flag and returns false if it is not specified.
func GlobalBoolT(c *cli.Context, name string) bool {
	if !c.IsSet(name) {
		return false
	} else {
		return c.GlobalBoolT(name)
	}
}

// Replaces substrings in the string with other strings, using strings.Replacer.
// The tokens parameter should be formatted as a valid CSV.
func replaceInString(s, tokens string) (string, error) {
	if tokens == "" {
		return s, nil
	}

	log.Debugf("Got tokens: %v", tokens)

	r := csv.NewReader(strings.NewReader(tokens))
	records, err := r.ReadAll()
	if err != nil {
		log.Errorf("Failed to read CSV: %v", err)
		return "", err
	}

	if len(records) != 1 {
		log.Errorf("Invalid number of CSV records: %v", len(records))
		return "", errors.New("Invalid number of CSV records. " +
			"Only single line CSV supported.")
	}

	tokenArray := records[0]

	log.Debugf("tokenArray: %v", tokenArray)

	replacer := strings.NewReplacer(tokenArray...)

	return replacer.Replace(s), nil
}
