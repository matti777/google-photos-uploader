package util

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/pkg/errors"

	"github.com/matti777/google-photos-uploader/internal/config"
	"github.com/matti777/google-photos-uploader/internal/logging"

	"github.com/urfave/cli"
)

var (
	log      = logging.MustGetLogger()
	settings = config.MustGetSettings()

	albumYearRegex *regexp.Regexp
)

// Asks the user interactively a confirmation question; if the user declines
// (answers anything but Y or defalut - empty string - stop execution.
func MustConfirm(format string, args ...interface{}) {
	text := fmt.Sprintf(format, args...)
	fmt.Print(fmt.Sprintf("%v\nContinue? [Y/n] ", text))
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')

	if input != "Y\n" && input != "\n" {
		os.Exit(1)
	}
}

// Finds the longest file name
func FindLongestName(infos []os.FileInfo) int {
	longest := 0

	for _, info := range infos {
		if len(info.Name()) > longest {
			longest = len(info.Name())
		}
	}

	return longest
}

// GlobalBoolT checks for the presence of a BoolT
// flag and returns false if it is not specified.
func GlobalBoolT(c *cli.Context, name string) bool {
	if !c.IsSet(name) {
		return false
	}

	return c.GlobalBoolT(name)
}

// Replaces substrings in the string with other strings, using strings.Replacer.
// The tokens parameter should be formatted as a valid CSV.
func ReplaceInString(s, tokens string) (string, error) {
	if tokens == "" {
		return s, nil
	}

	r := csv.NewReader(strings.NewReader(tokens))
	records, err := r.ReadAll()
	if err != nil {
		return "", errors.Wrap(err, "failed to read CSV")
	}

	if len(records) != 1 {
		return "", errors.Errorf("Invalid number of CSV records. " +
			"Only single line CSV supported.")
	}

	tokenArray := records[0]
	replacer := strings.NewReplacer(tokenArray...)

	return replacer.Replace(s), nil
}

// Chunked returns an array of arrays so that the original array is divided
// into chunks of equal size (except for the remainder chunk).
func Chunked(arr []string, chunkSize int) [][]string {
	arrayLen := len(arr)
	numChunks := arrayLen / chunkSize
	if arrayLen%chunkSize > 0 {
		numChunks++
	}

	chunks := make([][]string, 0, numChunks)

	for i := 0; i < arrayLen; i += chunkSize {
		chunkEnd := i + chunkSize
		if chunkEnd > arrayLen {
			chunkEnd = arrayLen
		}
		chunks = append(chunks, arr[i:chunkEnd])
	}

	return chunks
}

// parseAlbumYear tries to parse the year of the album from a string
// (directory name) using a certain regex pattern (eg. 'Trip to X - 2009')
// Returns empty string if not found.
func parseAlbumYear(name string) string {
	res := albumYearRegex.FindStringSubmatch(name)

	if len(res) == 2 {
		return res[1]
	}

	return ""
}

func init() {
	albumYearRegex = regexp.MustCompile("^.+[-_ ]([12]\\d{3})$")
}
