package main

import (
	"bufio"
	"fmt"
	"os"

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

	log.Debugf("input: '%v'", input)

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
